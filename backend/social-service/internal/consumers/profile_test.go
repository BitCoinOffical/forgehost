package consumers

import (
	"context"
	v2 "encoding/json/v2"
	"errors"
	"fmt"
	"testing"
	"time"

	prof_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/consumers/mock"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	kafkatc "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const testClusterID = "social-test-cluster"

func setupTestKafka(t *testing.T) []string {
	t.Helper()

	container, err := kafkatc.Run(t.Context(), "confluentinc/confluent-local:7.5.0",
		kafkatc.WithClusterID(testClusterID),
	)
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	brokers, err := container.Brokers(t.Context())
	require.NoError(t, err)

	return brokers
}

func createTopic(t *testing.T, brokers []string, name string) string {
	t.Helper()

	admin, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	require.NoError(t, err)
	t.Cleanup(admin.Close)

	_, err = kadm.NewClient(admin).CreateTopic(t.Context(), 1, 1, nil, name)
	require.NoError(t, err)

	return name
}

func produce(t *testing.T, brokers []string, topic string, values ...[]byte) {
	t.Helper()

	producer, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.DefaultProduceTopic(topic))
	require.NoError(t, err)
	defer producer.Close()

	for _, v := range values {
		require.NoError(t, producer.ProduceSync(t.Context(), &kgo.Record{Value: v}).FirstErr())
	}
}

func newConsumerClient(t *testing.T, brokers []string, topic string) *kgo.Client {
	t.Helper()

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	return client
}

func waitEvents(t *testing.T, saved <-chan string, errs <-chan error, wantSaved, wantErrs int) ([]string, []error) {
	t.Helper()

	var gotSaved []string
	var gotErrs []error
	timeout := time.After(15 * time.Second)

	for len(gotSaved) < wantSaved || len(gotErrs) < wantErrs {
		select {
		case id := <-saved:
			gotSaved = append(gotSaved, id)
		case err, ok := <-errs:
			require.True(t, ok, "errs channel closed unexpectedly")
			gotErrs = append(gotErrs, err)
		case <-timeout:
			t.Fatalf("timeout: saved %d/%d, errors %d/%d", len(gotSaved), wantSaved, len(gotErrs), wantErrs)
		}
	}

	select {
	case id := <-saved:
		t.Fatalf("unexpected profile saved: %s", id)
	case err, ok := <-errs:
		if ok {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
	}

	return gotSaved, gotErrs
}

func TestConsumerRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := prof_mocks.NewMockProfileService(ctrl)
	brokers := setupTestKafka(t)

	saved := make(chan string, 16)
	saveProfile := func(_ context.Context, e *dto.UserProfileDTO) error {
		saved <- e.UserID
		return nil
	}
	event := func(userID string) []byte {
		b, err := v2.Marshal(dto.UserProfileDTO{UserID: userID})
		require.NoError(t, err)
		return b
	}

	testCases := []struct {
		caseName string
		data     struct {
			messages [][]byte
		}
		prepare   func()
		wantSaved []string
		wantErrs  int
		wantErr   string
	}{
		{
			caseName: "save profile",
			data: struct {
				messages [][]byte
			}{messages: [][]byte{event("user_1")}},
			prepare: func() {
				service.EXPECT().SaveProfile(gomock.Any(), gomock.Any()).
					DoAndReturn(saveProfile).Times(1)
			},
			wantSaved: []string{"user_1"},
		},
		{
			caseName: "save several profiles",
			data: struct {
				messages [][]byte
			}{messages: [][]byte{event("user_1"), event("user_2"), event("user_3")}},
			prepare: func() {
				service.EXPECT().SaveProfile(gomock.Any(), gomock.Any()).
					DoAndReturn(saveProfile).Times(3)
			},
			wantSaved: []string{"user_1", "user_2", "user_3"},
		},
		{
			caseName: "invalid json does not stop consumer",
			data: struct {
				messages [][]byte
			}{messages: [][]byte{[]byte("{invalid"), event("user_4")}},
			prepare: func() {
				service.EXPECT().SaveProfile(gomock.Any(), gomock.Any()).
					DoAndReturn(saveProfile).Times(1)
			},
			wantSaved: []string{"user_4"},
			wantErrs:  1,
			wantErr:   "v2.Unmarshal",
		},
		{
			caseName: "save profile error",
			data: struct {
				messages [][]byte
			}{messages: [][]byte{event("user_5")}},
			prepare: func() {
				service.EXPECT().SaveProfile(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, e *dto.UserProfileDTO) error {
						saved <- e.UserID
						return errors.New("save profile error")
					}).Times(1)
			},
			wantSaved: []string{"user_5"},
			wantErrs:  1,
			wantErr:   "save profile error",
		},
	}

	for i, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			topic := createTopic(t, brokers, fmt.Sprintf("profiles-%d", i))
			produce(t, brokers, topic, tc.data.messages...)

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			consumer := NewConsumer(newConsumerClient(t, brokers, topic), zap.NewNop(), service)
			errs := consumer.Run(ctx)

			gotSaved, gotErrs := waitEvents(t, saved, errs, len(tc.wantSaved), tc.wantErrs)
			assert.ElementsMatch(t, tc.wantSaved, gotSaved)
			for _, err := range gotErrs {
				assert.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func TestConsumerRunStops(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := prof_mocks.NewMockProfileService(ctrl)
	brokers := setupTestKafka(t)

	testCases := []struct {
		caseName string
		stop     func(cancel context.CancelFunc, client *kgo.Client)
	}{
		{
			caseName: "stop on context cancel",
			stop:     func(cancel context.CancelFunc, _ *kgo.Client) { cancel() },
		},
		{
			caseName: "stop on client close",
			stop:     func(_ context.CancelFunc, client *kgo.Client) { client.Close() },
		},
	}

	for i, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			topic := createTopic(t, brokers, fmt.Sprintf("stop-%d", i))
			client := newConsumerClient(t, brokers, topic)

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			errs := NewConsumer(client, zaptest.NewLogger(t), service).Run(ctx)

			time.Sleep(500 * time.Millisecond)
			tc.stop(cancel, client)

			timeout := time.After(5 * time.Second)
			for {
				select {
				case _, ok := <-errs:
					if !ok {
						return
					}
				case <-timeout:
					t.Fatal("consumer did not stop: errs channel is not closed")
				}
			}
		})
	}
}
