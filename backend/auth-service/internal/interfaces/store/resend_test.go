package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResendLimitAdd(t *testing.T) {
	client := setupTestDB(t)
	store := NewResendStore(client)
	testCases := []struct {
		caseName string
		email    string
		success  bool
	}{
		{
			caseName: "resend limit add",
			email:    "user@gmail.com",
			success:  true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			err := store.ResendLimitAdd(context.Background(), tc.email)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestResendLimitCheck(t *testing.T) {
	client := setupTestDB(t)
	store := NewResendStore(client)
	testCases := []struct {
		caseName string
		email    string
		f        func(ctx context.Context, email string) error
		success  bool
	}{
		{
			caseName: "resend limit check",
			email:    "user@gmail.com",
			f:        store.ResendLimitAdd,
			success:  true,
		},
		{
			caseName: "empty limit check",
			email:    "empty@gmail.com",
			success:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.f != nil {
				err := store.ResendLimitAdd(context.Background(), tc.email)
				require.NoError(t, err)
			}
			email, err := store.ResendLimitCheck(context.Background(), tc.email)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.email, email)
			} else {
				require.Error(t, err)
				assert.Equal(t, "", email)
			}
		})
	}
}
