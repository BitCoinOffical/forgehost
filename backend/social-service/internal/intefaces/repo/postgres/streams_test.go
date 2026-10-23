package postgres

import (
	"context"
	"testing"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveChannel(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewStreamRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (channelID, userID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "save channel",
			f: func(t *testing.T) (string, string) {
				return newChannelID(), newUser(t, pool)
			},
			success: true,
		},
		{
			caseName: "overwrite channel",
			f: func(t *testing.T) (string, string) {
				userID := newUser(t, pool)
				setChannel(t, pool, userID, newChannelID())
				return newChannelID(), userID
			},
			success: true,
		},
		{
			caseName: "channel already linked to other user",
			f: func(t *testing.T) (string, string) {
				channelID := newChannelID()
				setChannel(t, pool, newUser(t, pool), channelID)
				return channelID, newUser(t, pool)
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			caseName: "save channel for not found user",
			f: func(t *testing.T) (string, string) {
				return newChannelID(), uuid.NewString()
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "save channel with invalid user id",
			f: func(t *testing.T) (string, string) {
				return newChannelID(), "invalid_uuid"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			channelID, userID := tc.f(t)

			err := repo.SaveChannel(context.Background(), channelID, userID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, channelOf(t, pool, userID))
			assert.Equal(t, channelID, *channelOf(t, pool, userID))
		})
	}
}

func TestGetChannel(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewStreamRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (userID, want string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "get channel",
			f: func(t *testing.T) (string, string) {
				userID, channelID := newUser(t, pool), newChannelID()
				setChannel(t, pool, userID, channelID)
				return userID, channelID
			},
			success: true,
		},
		{
			caseName: "channel not linked",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), ""
			},
			wantErr: domain.ErrChannelNotLinked,
		},
		{
			caseName: "get channel of not found user",
			f: func(t *testing.T) (string, string) {
				return uuid.NewString(), ""
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "get channel with invalid user id",
			f: func(t *testing.T) (string, string) {
				return "invalid_uuid", ""
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, want := tc.f(t)

			channelID, err := repo.GetChannel(context.Background(), userID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				assert.Empty(t, channelID)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, want, channelID)
		})
	}
}

func TestDeleteStream(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewStreamRepo(pool)

	testCases := []struct {
		caseName   string
		f          func(t *testing.T) (videoID, userID string)
		wantExists bool
		wantErr    error
		success    bool
	}{
		{
			caseName: "delete stream",
			f: func(t *testing.T) (string, string) {
				userID, videoID := newUser(t, pool), newVideoID()
				seedStream(t, pool, userID, videoID, "live")
				return videoID, userID
			},
			wantExists: false,
			success:    true,
		},
		{
			caseName: "delete stream of other user",
			f: func(t *testing.T) (string, string) {
				videoID := newVideoID()
				seedStream(t, pool, newUser(t, pool), videoID, "live")
				return videoID, newUser(t, pool)
			},
			wantExists: true,
			wantErr:    domain.ErrStreamNotFound,
		},
		{
			caseName: "delete keeps other streams of user",
			f: func(t *testing.T) (string, string) {
				userID, videoID := newUser(t, pool), newVideoID()
				seedStream(t, pool, userID, videoID, "live")
				seedStream(t, pool, userID, newVideoID(), "live")
				return videoID, userID
			},
			wantExists: false,
			success:    true,
		},
		{
			caseName: "delete not found stream",
			f: func(t *testing.T) (string, string) {
				return newVideoID(), newUser(t, pool)
			},
			wantErr: domain.ErrStreamNotFound,
		},
		{
			caseName: "delete stream with invalid user id",
			f: func(t *testing.T) (string, string) {
				return newVideoID(), "invalid_uuid"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			videoID, userID := tc.f(t)

			err := repo.DeleteStream(context.Background(), videoID, userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				requireErr(t, err, tc.wantErr)
			}
			if userID != "invalid_uuid" {
				n := countRows(t, pool, `SELECT COUNT(*) FROM streams WHERE video_id = $1`, videoID)
				assert.Equal(t, tc.wantExists, n > 0)
			}
		})
	}
}

func TestDeleteComplatedStreams(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewStreamRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (wantGone, wantKept []string)
		success  bool
	}{
		{
			caseName: "delete only complete streams",
			f: func(t *testing.T) ([]string, []string) {
				userID := newUser(t, pool)
				done1 := seedStream(t, pool, userID, newVideoID(), "complete")
				done2 := seedStream(t, pool, userID, newVideoID(), "complete")
				live := seedStream(t, pool, userID, newVideoID(), "live")
				created := seedStream(t, pool, userID, newVideoID(), "created")
				return []string{done1, done2}, []string{live, created}
			},
			success: true,
		},
		{
			caseName: "nothing to delete is not an error",
			f: func(t *testing.T) ([]string, []string) {
				_ = repo.DeleteComplatedStreams(context.Background())
				live := seedStream(t, pool, newUser(t, pool), newVideoID(), "live")
				return nil, []string{live}
			},
			success: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			wantGone, wantKept := tc.f(t)

			err := repo.DeleteComplatedStreams(context.Background())
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			for _, id := range wantGone {
				assert.Equal(t, 0, countRows(t, pool, `SELECT COUNT(*) FROM streams WHERE id = $1`, id))
			}
			for _, id := range wantKept {
				assert.Equal(t, 1, countRows(t, pool, `SELECT COUNT(*) FROM streams WHERE id = $1`, id))
			}
		})
	}
}
