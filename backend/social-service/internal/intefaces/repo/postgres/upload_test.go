package postgres

import (
	"context"
	"testing"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUpload(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewUploadRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (imageID, userID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "create upload",
			f: func(t *testing.T) (string, string) {
				return uuid.NewString(), newUser(t, pool)
			},
			success: true,
		},
		{
			caseName: "create upload with existing image id",
			f: func(t *testing.T) (string, string) {
				imageID := uuid.NewString()
				seedUpload(t, pool, imageID, newUser(t, pool))
				return imageID, newUser(t, pool)
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			caseName: "create upload for not found user",
			f: func(t *testing.T) (string, string) {
				return uuid.NewString(), uuid.NewString()
			},
		},
		{
			caseName: "create upload with invalid user id",
			f: func(t *testing.T) (string, string) {
				return uuid.NewString(), "invalid_uuid"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			imageID, userID := tc.f(t)

			err := repo.CreateUpload(context.Background(), imageID, userID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, countRows(t, pool,
				`SELECT COUNT(*) FROM uploads WHERE image_id = $1 AND user_id = $2 AND attached = false`,
				imageID, userID))
		})
	}
}
