package store

import (
	"context"
	"testing"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveUser(t *testing.T) {
	client := setupTestDB(t)
	store := NewUserStore(client)
	testCases := []struct {
		caseName string
		randStr  string
		user     *models.UserStored
		success  bool
	}{
		{
			caseName: "save user",
			randStr:  "random",
			user: &models.UserStored{
				Email:        "user@gmail.com",
				PasswordHash: []byte("password_hash"),
			},
			success: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			err := store.SaveUser(context.Background(), tc.randStr, tc.user)
			if tc.success {
				require.NoError(t, err)

				user, err := store.GetUser(context.Background(), tc.randStr)
				require.NoError(t, err)
				assert.Equal(t, tc.user, user)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	client := setupTestDB(t)
	store := NewUserStore(client)
	testCases := []struct {
		caseName string
		randStr  string
		user     *models.UserStored
		f        func(ctx context.Context, randStr string, user *models.UserStored) error
		success  bool
	}{
		{
			caseName: "get user",
			randStr:  "random",
			user: &models.UserStored{
				Email:        "user@gmail.com",
				PasswordHash: []byte("password_hash"),
			},
			f:       store.SaveUser,
			success: true,
		},
		{
			caseName: "get not found user",
			randStr:  "not found",
			success:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.f != nil {
				err := tc.f(context.Background(), tc.randStr, tc.user)
				require.NoError(t, err)
			}
			user, err := store.GetUser(context.Background(), tc.randStr)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.user, user)
			} else {
				require.Error(t, err)
				assert.Empty(t, user)
			}
		})
	}
}
