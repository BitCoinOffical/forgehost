package s3store

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/adapters/seaweedfs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testBucket = "forgehost-test"

var testS3 *S3

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "chrislusf/seaweedfs:latest",
			Cmd:          []string{"server", "-s3", "-dir=/data"},
			ExposedPorts: []string{"8333/tcp"},
			WaitingFor:   wait.ForListeningPort("8333/tcp").WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "start seaweedfs:", err)
		os.Exit(1)
	}

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "container host:", err)
		os.Exit(1)
	}
	port, err := container.MappedPort(ctx, "8333/tcp")
	if err != nil {
		fmt.Fprintln(os.Stderr, "mapped port:", err)
		os.Exit(1)
	}

	storage, err := seaweedfs.NewSeaweedfs(ctx, seaweedfs.S3Config{
		Endpoint:  fmt.Sprintf("http://%s:%s", host, port.Port()),
		AccessKey: "test",
		SecretKey: "test",
		Bucket:    testBucket,
		Region:    "us-east-1",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "new storage:", err)
		os.Exit(1)
	}

	var createErr error
	for i := 0; i < 30; i++ {
		_, createErr = storage.Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(testBucket),
		})
		if createErr == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if createErr != nil {
		fmt.Fprintln(os.Stderr, "create bucket:", createErr)
		os.Exit(1)
	}

	testS3 = NewS3Loader(storage)

	code := m.Run()
	_ = container.Terminate(ctx)
	os.Exit(code)
}

func newKey() string {
	return "posts/" + uuid.NewString() + ".bin"
}

func TestUpload(t *testing.T) {
	testCases := []struct {
		caseName    string
		contentType string
		body        []byte
	}{
		{caseName: "png", contentType: "image/png", body: []byte("fake png content")},
		{caseName: "jpeg", contentType: "image/jpeg", body: []byte("fake jpeg content")},
		{caseName: "large body", contentType: "image/webp", body: bytes.Repeat([]byte("a"), 1<<20)},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			key := newKey()
			t.Cleanup(func() { _ = testS3.Delete(context.Background(), key) })

			err := testS3.Upload(t.Context(), key, tc.contentType, bytes.NewReader(tc.body))
			require.NoError(t, err)

			out, err := testS3.storage.Client.GetObject(t.Context(), &s3.GetObjectInput{
				Bucket: aws.String(testBucket),
				Key:    aws.String(key),
			})
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, out.Body.Close())
			})

			got, err := io.ReadAll(out.Body)
			require.NoError(t, err)
			assert.Equal(t, tc.body, got)
			assert.Equal(t, tc.contentType, aws.ToString(out.ContentType))
		})
	}
}

func TestDelete(t *testing.T) {
	t.Run("deletes existing object", func(t *testing.T) {
		key := newKey()
		require.NoError(t, testS3.Upload(t.Context(), key, "image/png", strings.NewReader("data")))

		require.NoError(t, testS3.Delete(t.Context(), key))

		_, err := testS3.storage.Client.GetObject(t.Context(), &s3.GetObjectInput{
			Bucket: aws.String(testBucket),
			Key:    aws.String(key),
		})
		require.Error(t, err)
	})

	t.Run("deleting missing object is not an error", func(t *testing.T) {
		require.NoError(t, testS3.Delete(t.Context(), newKey()))
	})
}

func TestGetURL(t *testing.T) {
	body := []byte("presigned content")
	key := newKey()
	require.NoError(t, testS3.Upload(t.Context(), key, "image/png", bytes.NewReader(body)))
	t.Cleanup(func() { _ = testS3.Delete(context.Background(), key) })

	url, err := testS3.GetURL(t.Context(), key, time.Minute)
	require.NoError(t, err)
	assert.Contains(t, url, testBucket)
	assert.Contains(t, url, key)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, resp.Body.Close())
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	got, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, body, got)
}
