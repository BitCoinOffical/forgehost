package repo

import (
	"context"
	"strings"
	"testing"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/migrations"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:18",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)

	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	err = migrations.RunMigrations(pool)
	require.NoError(t, err)

	t.Cleanup(func() {
		migrations.RollbackLast(ctx, pool)
		require.NoError(t, err)
	})

	return pool
}

func cleanDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public' AND tablename <> 'goose_db_version'
	`)
	require.NoError(t, err)

	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)
	if len(tables) == 0 {
		return
	}

	quoted := make([]string, len(tables))

	for i, name := range tables {
		quoted[i] = pgx.Identifier{name}.Sanitize()
	}

	_, err = pool.Exec(ctx, "TRUNCATE TABLE "+strings.Join(quoted, ", ")+" RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

func TestSaveUser(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanDB(t, pool)
	repo := NewAuthRepo(pool)

	testCases := []struct {
		caseName string
		data     *models.User
		succes   bool
	}{
		{
			caseName: "save user",
			data: &models.User{
				Email:        "user@gmail.com",
				PasswordHash: []byte("password_hash"),
			},
			succes: true,
		},
		{
			caseName: "duplicate email",
			data: &models.User{
				Email:        "user@gmail.com",
				PasswordHash: []byte("password_hash"),
			},
			succes: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, err := repo.SaveUser(context.Background(), tc.data)
			if tc.succes {
				require.NoError(t, err)
				assert.NotEmpty(t, userID)
				return
			}
			require.Error(t, err)
		})
	}

}

func TestSaveGoogleUser(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanDB(t, pool)
	repo := NewAuthRepo(pool)

	testCases := []struct {
		caseName string
		data     *models.User
		oathData *models.OAuthAccount
		f        func(ctx context.Context, req *models.User) (uuid.UUID, error)
		succes   bool
	}{
		{
			caseName: "save google user",
			data: &models.User{
				Email:         "user@gmail.com",
				EmailVerified: true,
			},
			oathData: &models.OAuthAccount{
				Provider:       "google",
				ProviderUserID: "g-123",
				GivenName:      "Ivan",
				FamilyName:     "Ivanov",
			},
			succes: true,
		},
		{
			caseName: "save dublicate google user",
			data: &models.User{
				Email:         "user@gmail.com",
				EmailVerified: true,
			},
			oathData: &models.OAuthAccount{
				Provider:       "google",
				ProviderUserID: "g-123",
				GivenName:      "Ivan",
				FamilyName:     "Ivanov",
			},
			succes: true,
		},
		{
			caseName: "links to existing unverified user",
			data: &models.User{
				Email: "user2@gmail.com",
			},
			oathData: &models.OAuthAccount{
				Provider:       "google",
				ProviderUserID: "g-123",
				GivenName:      "Ivan",
				FamilyName:     "Ivanov",
			},
			f:      repo.SaveUser,
			succes: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.f != nil {
				tc.f(context.Background(), tc.data)
			}
			userID, err := repo.SaveGoogleUser(context.Background(), tc.data, tc.oathData)
			if tc.succes {
				require.NoError(t, err)
				assert.NotEmpty(t, userID)
				return
			}
			require.Error(t, err)
		})
	}
}

func TestGetUserByID(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanDB(t, pool)
	repo := NewAuthRepo(pool)

	testCases := []struct {
		caseName string
		data     *models.User
		f        func(ctx context.Context, req *models.User) (uuid.UUID, error)
		succes   bool
	}{
		{
			caseName: "get user by id",
			data: &models.User{
				Email:        "user@gmail.com",
				PasswordHash: []byte("password_hash"),
			},
			f:      repo.SaveUser,
			succes: true,
		},
		{
			caseName: "not found user by id",
			data: &models.User{
				Email:        "notfound@gmail.com",
				PasswordHash: []byte("password_hash"),
			},
			succes: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			var UserID uuid.UUID
			if tc.f != nil {
				userID, err := tc.f(context.Background(), tc.data)
				require.NoError(t, err)
				UserID = userID
			}

			user, err := repo.GetUserByID(context.Background(), UserID)
			if tc.succes {
				require.NoError(t, err)
				assert.NotEmpty(t, user)
				return
			}
			require.Error(t, err)
			assert.Nil(t, user)
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanDB(t, pool)
	repo := NewAuthRepo(pool)

	testCases := []struct {
		caseName string
		data     *models.User
		f        func(ctx context.Context, req *models.User) (uuid.UUID, error)
		succes   bool
	}{
		{
			caseName: "get user by email",
			data: &models.User{
				Email:        "user@gmail.com",
				PasswordHash: []byte("password_hash"),
			},
			f:      repo.SaveUser,
			succes: true,
		},
		{
			caseName: "not found user by email",
			data: &models.User{
				Email:        "notfound@gmail.com",
				PasswordHash: []byte("password_hash"),
			},
			succes: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.f != nil {
				_, err := tc.f(context.Background(), tc.data)
				require.NoError(t, err)
			}

			user, err := repo.GetUserByEmail(context.Background(), tc.data.Email)
			if tc.succes {
				require.NoError(t, err)
				assert.NotEmpty(t, user)
				return
			}
			require.Error(t, err)
			assert.Nil(t, user)
		})
	}
}

func TestUpdateUserPassword(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanDB(t, pool)
	repo := NewAuthRepo(pool)

	testCases := []struct {
		caseName string
		data     *models.User
		f        func(ctx context.Context, req *models.User) (uuid.UUID, error)
		succes   bool
	}{
		{
			caseName: "update user password",
			data: &models.User{
				Email:        "user@gmail.com",
				PasswordHash: []byte("new_password_hash"),
			},
			f:      repo.SaveUser,
			succes: true,
		},
		{
			caseName: "failed update user password",
			data: &models.User{
				ID:           uuid.New(),
				Email:        "notfound@gmail.com",
				PasswordHash: []byte("new_password_hash"),
			},
			succes: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			user := models.User{
				ID:           tc.data.ID,
				Email:        tc.data.Email,
				PasswordHash: tc.data.PasswordHash,
			}
			if tc.f != nil {
				userID, err := tc.f(context.Background(), tc.data)
				require.NoError(t, err)
				user.ID = userID
			}

			err := repo.UpdateUserPassword(context.Background(), &user)
			if tc.succes {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
		})
	}
}
