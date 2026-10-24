package services

import (
	"encoding/base64"
	v2 "encoding/json/v2"
	"errors"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	store_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func makeStreamCursor(t *testing.T, streamID string, createdAt time.Time) string {
	t.Helper()
	raw, err := v2.Marshal(dto.CursorStreamDTO{StreamID: streamID, CreatedAt: createdAt})
	require.NoError(t, err)
	return base64.URLEncoding.EncodeToString(raw)
}

func TestGetStreamHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		data     struct {
			videoID string
		}
		prepare func()
		want    []models.Message
		success bool
	}{
		{
			caseName: "get stream history",
			data: struct {
				videoID string
			}{videoID: "video_id"},
			prepare: func() {
				store.EXPECT().GetStreamMessages(gomock.Any(), "video_id").
					Return([]models.Message{{Text: "first"}, {Text: "second"}}, nil).Times(1)
			},
			want:    []models.Message{{Text: "first"}, {Text: "second"}},
			success: true,
		},
		{
			caseName: "get empty stream history",
			data: struct {
				videoID string
			}{videoID: "video_id"},
			prepare: func() {
				store.EXPECT().GetStreamMessages(gomock.Any(), "video_id").
					Return(nil, nil).Times(1)
			},
			want:    nil,
			success: true,
		},
		{
			caseName: "get stream history error",
			data: struct {
				videoID string
			}{videoID: "video_id"},
			prepare: func() {
				store.EXPECT().GetStreamMessages(gomock.Any(), "video_id").
					Return(nil, errors.New("get stream history error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := stream.GetStreamHistory(t.Context(), tc.data.videoID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				assert.Nil(t, got)
			}
		})
	}
}

func TestSaveStreamMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID  string
			videoID string
			text    string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "save stream message",
			data: struct {
				userID  string
				videoID string
				text    string
			}{userID: "user_id", videoID: "video_id", text: "text"},
			prepare: func() {
				store.EXPECT().SaveStreamMessage(gomock.Any(), "video_id", messageMatcher("video_id", "user_id", "text")).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "save stream message error",
			data: struct {
				userID  string
				videoID string
				text    string
			}{userID: "user_id", videoID: "video_id", text: "text"},
			prepare: func() {
				store.EXPECT().SaveStreamMessage(gomock.Any(), "video_id", gomock.Any()).
					Return(errors.New("save stream message error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.MessageDTO{
				Text: tc.data.text,
			}
			err := stream.SaveStreamMessage(t.Context(), tc.data.userID, tc.data.videoID, &req)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestSaveStreamUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID  string
			videoID string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "save stream user",
			data: struct {
				userID  string
				videoID string
			}{userID: "user_id", videoID: "video_id"},
			prepare: func() {
				store.EXPECT().SaveStreamUser(gomock.Any(), "user_id", "video_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "save stream user error",
			data: struct {
				userID  string
				videoID string
			}{userID: "user_id", videoID: "video_id"},
			prepare: func() {
				store.EXPECT().SaveStreamUser(gomock.Any(), "user_id", "video_id").
					Return(errors.New("save stream user error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := stream.SaveStreamUser(t.Context(), tc.data.userID, tc.data.videoID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetStreamUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		data     struct {
			videoID string
		}
		prepare func()
		want    []string
		success bool
	}{
		{
			caseName: "get stream users",
			data: struct {
				videoID string
			}{videoID: "video_id"},
			prepare: func() {
				store.EXPECT().GetStreamUsers(gomock.Any(), "video_id").
					Return([]string{"user_1", "user_2"}, nil).Times(1)
			},
			want:    []string{"user_1", "user_2"},
			success: true,
		},
		{
			caseName: "get empty stream users",
			data: struct {
				videoID string
			}{videoID: "video_id"},
			prepare: func() {
				store.EXPECT().GetStreamUsers(gomock.Any(), "video_id").
					Return(nil, nil).Times(1)
			},
			want:    nil,
			success: true,
		},
		{
			caseName: "get stream users error",
			data: struct {
				videoID string
			}{videoID: "video_id"},
			prepare: func() {
				store.EXPECT().GetStreamUsers(gomock.Any(), "video_id").
					Return(nil, errors.New("get stream users error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := stream.GetStreamUsers(t.Context(), tc.data.videoID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				assert.Nil(t, got)
			}
		})
	}
}

func TestSaveChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		data     struct {
			channelID string
			userID    string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "save channel",
			data: struct {
				channelID string
				userID    string
			}{channelID: "channel_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().SaveChannel(gomock.Any(), "channel_id", "user_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "save channel linked to other user",
			data: struct {
				channelID string
				userID    string
			}{channelID: "channel_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().SaveChannel(gomock.Any(), "channel_id", "user_id").
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "save channel for not found profile",
			data: struct {
				channelID string
				userID    string
			}{channelID: "channel_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().SaveChannel(gomock.Any(), "channel_id", "user_id").
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "save channel error",
			data: struct {
				channelID string
				userID    string
			}{channelID: "channel_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().SaveChannel(gomock.Any(), "channel_id", "user_id").
					Return(errors.New("save channel error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := stream.SaveChannel(t.Context(), tc.data.channelID, tc.data.userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

func TestGetChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
		}
		prepare func()
		want    string
		wantErr error
		success bool
	}{
		{
			caseName: "get channel",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetChannel(gomock.Any(), "user_id").
					Return("channel_id", nil).Times(1)
			},
			want:    "channel_id",
			success: true,
		},
		{
			caseName: "get not linked channel",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetChannel(gomock.Any(), "user_id").
					Return("", domain.ErrChannelNotLinked).Times(1)
			},
			wantErr: domain.ErrChannelNotLinked,
			success: false,
		},
		{
			caseName: "get channel of not found profile",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetChannel(gomock.Any(), "user_id").
					Return("", domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "get channel error",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetChannel(gomock.Any(), "user_id").
					Return("", errors.New("get channel error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := stream.GetChannel(t.Context(), tc.data.userID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
				assert.Empty(t, got)
			}
		})
	}
}

func TestSaveStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID   string
			videoID  string
			title    string
			status   *string
			embedURL string
			viewers  *uint64
			likes    *uint64
		}
		prepare func()
		want    string
		wantErr error
		success bool
	}{
		{
			caseName: "save stream",
			data: struct {
				userID   string
				videoID  string
				title    string
				status   *string
				embedURL string
				viewers  *uint64
				likes    *uint64
			}{userID: "user_id", videoID: "video_id", title: "title", status: strPtr("live"), embedURL: "https://embed.test/1", viewers: uintPtr(10), likes: uintPtr(5)},
			prepare: func() {
				repo.EXPECT().SaveStream(gomock.Any(), &models.Stream{
					UserID:   "user_id",
					VideoID:  "video_id",
					Title:    strPtr("title"),
					Status:   strPtr("live"),
					EmbedURL: strPtr("https://embed.test/1"),
					Viewers:  uintPtr(10),
					Likes:    uintPtr(5),
				}).Return("stream_id", nil).Times(1)
			},
			want:    "stream_id",
			success: true,
		},
		{
			caseName: "save existing stream",
			data: struct {
				userID   string
				videoID  string
				title    string
				status   *string
				embedURL string
				viewers  *uint64
				likes    *uint64
			}{userID: "user_id", videoID: "video_id", title: "title"},
			prepare: func() {
				repo.EXPECT().SaveStream(gomock.Any(), gomock.Any()).
					Return("", domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "save stream error",
			data: struct {
				userID   string
				videoID  string
				title    string
				status   *string
				embedURL string
				viewers  *uint64
				likes    *uint64
			}{userID: "user_id", videoID: "video_id", title: "title"},
			prepare: func() {
				repo.EXPECT().SaveStream(gomock.Any(), gomock.Any()).
					Return("", errors.New("save stream error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.StreamDTO{
				VideoID:           tc.data.videoID,
				Title:             tc.data.title,
				Status:            tc.data.status,
				EmbedURL:          tc.data.embedURL,
				ConcurrentViewers: tc.data.viewers,
				LikeCount:         tc.data.likes,
			}
			got, err := stream.SaveStream(t.Context(), tc.data.userID, &req)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
				assert.Empty(t, got)
			}
		})
	}
}

func TestBuildUpdateStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID   string
			videoID  string
			title    *string
			status   *string
			embedURL *string
			viewers  *uint64
			likes    *uint64
		}
		prepare func()
		want    *models.Stream
		wantErr error
		success bool
	}{
		{
			caseName: "update stream",
			data: struct {
				userID   string
				videoID  string
				title    *string
				status   *string
				embedURL *string
				viewers  *uint64
				likes    *uint64
			}{userID: "user_id", videoID: "video_id", title: strPtr("new_title"), status: strPtr("complete"), embedURL: strPtr("new_embed"), viewers: uintPtr(20), likes: uintPtr(8)},
			prepare: func() {
				repo.EXPECT().BuildUpdateStream(gomock.Any(), &models.Stream{
					UserID:   "user_id",
					VideoID:  "video_id",
					Title:    strPtr("new_title"),
					Status:   strPtr("complete"),
					EmbedURL: strPtr("new_embed"),
					Viewers:  uintPtr(20),
					Likes:    uintPtr(8),
				}).Return(&models.Stream{ID: "stream_id", VideoID: "video_id"}, nil).Times(1)
			},
			want:    &models.Stream{ID: "stream_id", VideoID: "video_id"},
			success: true,
		},
		{
			caseName: "update not found stream",
			data: struct {
				userID   string
				videoID  string
				title    *string
				status   *string
				embedURL *string
				viewers  *uint64
				likes    *uint64
			}{userID: "user_id", videoID: "video_id", title: strPtr("new_title")},
			prepare: func() {
				repo.EXPECT().BuildUpdateStream(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrStreamNotFound).Times(1)
			},
			wantErr: domain.ErrStreamNotFound,
			success: false,
		},
		{
			caseName: "update stream error",
			data: struct {
				userID   string
				videoID  string
				title    *string
				status   *string
				embedURL *string
				viewers  *uint64
				likes    *uint64
			}{userID: "user_id", videoID: "video_id", title: strPtr("new_title")},
			prepare: func() {
				repo.EXPECT().BuildUpdateStream(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("update stream error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.UpdateStreamDTO{
				Title:             tc.data.title,
				Status:            tc.data.status,
				EmbedURL:          tc.data.embedURL,
				ConcurrentViewers: tc.data.viewers,
				LikeCount:         tc.data.likes,
			}
			got, err := stream.BuildUpdateStream(t.Context(), tc.data.userID, tc.data.videoID, &req)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
				assert.Nil(t, got)
			}
		})
	}
}

func TestDeleteStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		data     struct {
			videoID string
			userID  string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "delete stream",
			data: struct {
				videoID string
				userID  string
			}{videoID: "video_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().DeleteStream(gomock.Any(), "video_id", "user_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "delete not found stream",
			data: struct {
				videoID string
				userID  string
			}{videoID: "video_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().DeleteStream(gomock.Any(), "video_id", "user_id").
					Return(domain.ErrStreamNotFound).Times(1)
			},
			wantErr: domain.ErrStreamNotFound,
			success: false,
		},
		{
			caseName: "delete stream error",
			data: struct {
				videoID string
				userID  string
			}{videoID: "video_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().DeleteStream(gomock.Any(), "video_id", "user_id").
					Return(errors.New("delete stream error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := stream.DeleteStream(t.Context(), tc.data.videoID, tc.data.userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

func TestDeleteComplatedStreams(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		prepare  func()
		success  bool
	}{
		{
			caseName: "delete completed streams",
			prepare: func() {
				repo.EXPECT().DeleteComplatedStreams(gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "delete completed streams error",
			prepare: func() {
				repo.EXPECT().DeleteComplatedStreams(gomock.Any()).
					Return(errors.New("delete completed streams error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := stream.DeleteComplatedStreams(t.Context())
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	testCases := []struct {
		caseName string
		data     struct {
			videoID string
			userID  string
		}
		prepare func()
		want    *models.Stream
		wantErr error
		success bool
	}{
		{
			caseName: "get stream",
			data: struct {
				videoID string
				userID  string
			}{videoID: "video_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetStream(gomock.Any(), "video_id", "user_id").
					Return(&models.Stream{ID: "stream_id", VideoID: "video_id"}, nil).Times(1)
			},
			want:    &models.Stream{ID: "stream_id", VideoID: "video_id"},
			success: true,
		},
		{
			caseName: "get not found stream",
			data: struct {
				videoID string
				userID  string
			}{videoID: "video_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetStream(gomock.Any(), "video_id", "user_id").
					Return(nil, domain.ErrStreamNotFound).Times(1)
			},
			wantErr: domain.ErrStreamNotFound,
			success: false,
		},
		{
			caseName: "get stream error",
			data: struct {
				videoID string
				userID  string
			}{videoID: "video_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetStream(gomock.Any(), "video_id", "user_id").
					Return(nil, errors.New("get stream error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := stream.GetStream(t.Context(), tc.data.videoID, tc.data.userID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
				assert.Nil(t, got)
			}
		})
	}
}

func TestGetCursorStreams(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := store_mocks.NewMockStreamStore(ctrl)
	repo := store_mocks.NewMockStreamRepo(ctrl)

	stream := NewStreamService(store, repo)

	createdAt := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	validCursor := makeStreamCursor(t, "stream_id", createdAt)
	badJSONCursor := base64.URLEncoding.EncodeToString([]byte("{invalid"))

	createdAtMatcher := gomock.Cond(func(v time.Time) bool { return v.Equal(createdAt) })

	feed := []models.FeedStream{{ID: "1"}, {ID: "2"}}

	testCases := []struct {
		caseName string
		data     struct {
			cursor string
		}
		prepare func()
		want    []models.FeedStream
		success bool
	}{
		{
			caseName: "get cursor streams first page",
			data: struct {
				cursor string
			}{cursor: ""},
			prepare: func() {
				repo.EXPECT().GetCursorStreams(gomock.Any(), "", time.Time{}).
					Return(feed, nil).Times(1)
			},
			want:    feed,
			success: true,
		},
		{
			caseName: "get cursor streams with cursor",
			data: struct {
				cursor string
			}{cursor: validCursor},
			prepare: func() {
				repo.EXPECT().GetCursorStreams(gomock.Any(), "stream_id", createdAtMatcher).
					Return(feed, nil).Times(1)
			},
			want:    feed,
			success: true,
		},
		{
			caseName: "invalid base64 cursor",
			data: struct {
				cursor string
			}{cursor: "!!!"},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "invalid json in cursor",
			data: struct {
				cursor string
			}{cursor: badJSONCursor},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "get cursor streams error",
			data: struct {
				cursor string
			}{cursor: ""},
			prepare: func() {
				repo.EXPECT().GetCursorStreams(gomock.Any(), "", time.Time{}).
					Return(nil, errors.New("get cursor streams error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := stream.GetCursorStreams(t.Context(), tc.data.cursor)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				assert.Nil(t, got)
			}
		})
	}
}
