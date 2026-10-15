package store

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	redistc "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()
	container, err := redistc.Run(ctx,
		"redis:7", testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections"),
		))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	opt, err := redis.ParseURL(connStr)
	require.NoError(t, err)

	client := redis.NewClient(opt)

	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	return client
}

func TestSaveVerificationCode(t *testing.T) {
	client := setupTestDB(t)
	store := NewCodeStore(client)
	testCases := []struct {
		caseName string
		data     struct {
			randStr    string
			code       int
			expiration time.Duration
		}
		success bool
	}{
		{
			caseName: "save verification code",
			data: struct {
				randStr    string
				code       int
				expiration time.Duration
			}{randStr: "user@gmail.com", code: 123456, expiration: 5 * time.Minute},
			success: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			err := store.SaveVerificationCode(context.Background(), tc.data.randStr, tc.data.code, tc.data.expiration)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestSaveResetPasswordCode(t *testing.T) {
	client := setupTestDB(t)
	store := NewCodeStore(client)
	testCases := []struct {
		caseName string
		data     struct {
			randStr    string
			code       int
			expiration time.Duration
		}
		success bool
	}{
		{
			caseName: "save verification code",
			data: struct {
				randStr    string
				code       int
				expiration time.Duration
			}{randStr: "user@gmail.com", code: 123456, expiration: 5 * time.Minute},
			success: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			err := store.SaveResetPasswordCode(context.Background(), tc.data.randStr, tc.data.code, tc.data.expiration)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
func TestSaveOauthCode(t *testing.T) {
	client := setupTestDB(t)
	store := NewCodeStore(client)
	testCases := []struct {
		caseName string
		data     struct {
			oauthCode string
			userID    string
		}
		success bool
	}{
		{
			caseName: "save oauth code",
			data: struct {
				oauthCode string
				userID    string
			}{oauthCode: "user@gmail.com",
				userID: "user_id"},
			success: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			err := store.SaveOauthCode(context.Background(), tc.data.oauthCode, tc.data.userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
func TestGetOauthCode(t *testing.T) {
	client := setupTestDB(t)
	store := NewCodeStore(client)
	testCases := []struct {
		caseName  string
		userID    string
		oauthCode string
		f         func(ctx context.Context, oauthCode, userID string) error
		success   bool
	}{
		{
			caseName:  "get oauth code",
			userID:    "user_id",
			oauthCode: "code",
			f:         store.SaveOauthCode,
			success:   true,
		},
		{
			caseName:  "get not found oauth code",
			oauthCode: "not found",
			success:   false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.f != nil {
				err := tc.f(context.Background(), tc.oauthCode, tc.userID)
				require.NoError(t, err)
			}
			userID, err := store.GetOauthCode(context.Background(), tc.oauthCode)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.userID, userID)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDeleteOauthCode(t *testing.T) {
	client := setupTestDB(t)
	store := NewCodeStore(client)
	testCases := []struct {
		caseName  string
		userID    string
		oauthCode string
		f         func(ctx context.Context, oauthCode, userID string) error
	}{
		{
			caseName:  "delete oauth code",
			userID:    "user_id",
			oauthCode: "code",
			f:         store.SaveOauthCode,
		},
		{
			caseName:  "delete not found oauth code",
			oauthCode: "not found",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.f != nil {
				err := tc.f(context.Background(), tc.oauthCode, tc.userID)
				require.NoError(t, err)
			}
			err := store.DeleteOauthCode(context.Background(), tc.oauthCode)
			require.NoError(t, err)
			userID, err := store.GetOauthCode(context.Background(), tc.oauthCode)
			require.Error(t, err)
			assert.Equal(t, "", userID)
		})
	}
}

func TestGetResetPasswordCode(t *testing.T) {
	client := setupTestDB(t)
	store := NewCodeStore(client)
	testCases := []struct {
		caseName string
		randStr  string
		code     int
		f        func(ctx context.Context, randStr string, code int, expiration time.Duration) error
		success  bool
	}{
		{
			caseName: "get reset password code",
			randStr:  "code",
			code:     1234,
			f:        store.SaveResetPasswordCode,
			success:  true,
		},
		{
			caseName: "get not found reset password code",
			randStr:  "not found",
			code:     1234,
			success:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.f != nil {
				err := tc.f(context.Background(), tc.randStr, tc.code, 5*time.Minute)
				require.NoError(t, err)
			}
			codeStr, err := store.GetResetPasswordCode(context.Background(), tc.randStr)
			code, err := strconv.Atoi(codeStr)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.code, code)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetVerificationCode(t *testing.T) {
	client := setupTestDB(t)
	store := NewCodeStore(client)
	testCases := []struct {
		caseName string
		randStr  string
		code     int
		f        func(ctx context.Context, randStr string, code int, expiration time.Duration) error
		success  bool
	}{
		{
			caseName: "get verification code",
			randStr:  "user_id",
			code:     1234,
			f:        store.SaveVerificationCode,
			success:  true,
		},
		{
			caseName: "get not found verification code",
			randStr:  "not found",
			code:     1234,
			success:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.f != nil {
				err := tc.f(context.Background(), tc.randStr, tc.code, 5*time.Second)
				require.NoError(t, err)
			}
			codeStr, err := store.GetVerificationCode(context.Background(), tc.randStr)
			code, err := strconv.Atoi(codeStr)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.code, code)
			} else {
				require.Error(t, err)
			}
		})
	}
}
