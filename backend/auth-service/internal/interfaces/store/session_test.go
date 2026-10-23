package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveToken(t *testing.T) {
	client := setupTestDB(t)
	store := NewSessionStore(client)
	testCases := []struct {
		caseName string
		userID   uuid.UUID
		token    string
		success  bool
	}{
		{
			caseName: "save token",
			userID:   uuid.New(),
			token:    "token",
			success:  true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			err := store.SaveToken(context.Background(), tc.userID, tc.token, 5*time.Minute)
			if tc.success {
				require.NoError(t, err)

				token, err := store.GetToken(context.Background(), tc.userID)
				require.NoError(t, err)
				assert.Equal(t, tc.token, token)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetToken(t *testing.T) {
	client := setupTestDB(t)
	store := NewSessionStore(client)
	testCases := []struct {
		caseName string
		data     struct {
			userID uuid.UUID
			token  string
		}
		f       func(ctx context.Context, userID uuid.UUID, token string, RefreshTTL time.Duration) error
		success bool
	}{
		{
			caseName: "get token",
			data: struct {
				userID uuid.UUID
				token  string
			}{
				userID: uuid.New(),
				token:  "token",
			},
			f:       store.SaveToken,
			success: true,
		},
		{
			caseName: "get not found token",
			data: struct {
				userID uuid.UUID
				token  string
			}{
				userID: uuid.New(),
				token:  "not found",
			},
			success: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.f != nil {
				err := tc.f(context.Background(), tc.data.userID, tc.data.token, 5*time.Minute)
				require.NoError(t, err)
			}
			token, err := store.GetToken(context.Background(), tc.data.userID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.data.token, token)
			} else {
				require.Error(t, err)
				assert.Equal(t, "", token)
			}
		})
	}
}

func TestDeleteToken(t *testing.T) {
	client := setupTestDB(t)
	store := NewSessionStore(client)
	testCases := []struct {
		caseName string
		data     struct {
			userID uuid.UUID
			token  string
		}
		f func(ctx context.Context, userID uuid.UUID, token string, RefreshTTL time.Duration) error
	}{
		{
			caseName: "delete token",
			data: struct {
				userID uuid.UUID
				token  string
			}{
				userID: uuid.New(),
				token:  "token",
			},
			f: store.SaveToken,
		},
		{
			caseName: "delete not found token",
			data: struct {
				userID uuid.UUID
				token  string
			}{
				userID: uuid.New(),
				token:  "not found",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.f != nil {
				err := tc.f(context.Background(), tc.data.userID, tc.data.token, 5*time.Minute)
				require.NoError(t, err)
			}
			err := store.DeleteToken(context.Background(), tc.data.userID)
			require.NoError(t, err)

			token, err := store.GetToken(context.Background(), tc.data.userID)
			require.Error(t, err)
			assert.Equal(t, "", token)
		})
	}
}
