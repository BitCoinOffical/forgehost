package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	store_mocks "github.com/BitCoinOffical/forgehost/hosting-service/internal/interfaces/services/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func worldsTestModel(id string) *models.MinecraftServer {
	return &models.MinecraftServer{
		ID:      id,
		UserID:  "user-id",
		Name:    "test-world",
		Address: "mc.test.com",
		Port:    25565,
		Edition: "java",
	}
}

func TestUpsert(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockWorldsRepo(ctrl)
	cache := store_mocks.NewMockWorldsCache(ctrl)

	srvc := NewWorldsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
			req    dto.UpsertMinecraftServerDTO
		}
		prepare func()
		want    *models.MinecraftServer
		wantErr error
		success bool
	}{
		{
			caseName: "upsert world",
			data: struct {
				userID string
				req    dto.UpsertMinecraftServerDTO
			}{
				userID: "user-id",
				req: dto.UpsertMinecraftServerDTO{
					Name:    " test-world ",
					Address: " MC.TEST.COM ",
					Port:    25565,
					Edition: "java",
				},
			},
			prepare: func() {
				repo.EXPECT().Upsert(
					gomock.Any(),
					&models.MinecraftServer{
						UserID:  "user-id",
						Name:    "test-world",
						Address: "mc.test.com",
						Port:    25565,
						Edition: "java",
					},
				).Return(worldsTestModel("world-id"), nil).Times(1)

				cache.EXPECT().Remove(gomock.Any(), "world-id").
					Return(nil).Times(1)
			},
			want:    worldsTestModel("world-id"),
			success: true,
		},
		{
			caseName: "upsert world with already exists endpoint",
			data: struct {
				userID string
				req    dto.UpsertMinecraftServerDTO
			}{
				userID: "user-id",
				req: dto.UpsertMinecraftServerDTO{
					Name:    "test-world",
					Address: "mc.test.com",
					Port:    25565,
					Edition: "java",
				},
			},
			prepare: func() {
				repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("endpoint exists: %w", domain.ErrAlreadyExists)).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "upsert world repo error",
			data: struct {
				userID string
				req    dto.UpsertMinecraftServerDTO
			}{
				userID: "user-id",
				req:    dto.UpsertMinecraftServerDTO{Name: "test-world"},
			},
			prepare: func() {
				repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("upsert world error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "upsert world cache error",
			data: struct {
				userID string
				req    dto.UpsertMinecraftServerDTO
			}{
				userID: "user-id",
				req:    dto.UpsertMinecraftServerDTO{Name: "test-world"},
			},
			prepare: func() {
				repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).
					Return(worldsTestModel("world-id"), nil).Times(1)

				cache.EXPECT().Remove(gomock.Any(), "world-id").
					Return(errors.New("remove cache error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := srvc.Upsert(
				t.Context(),
				&tc.data.req,
				tc.data.userID,
			)

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

func TestGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockWorldsRepo(ctrl)
	cache := store_mocks.NewMockWorldsCache(ctrl)

	srvc := NewWorldsService(repo, cache)

	testCases := []struct {
		caseName string
		prepare  func()
		want     []models.MinecraftServer
		success  bool
	}{
		{
			caseName: "get all worlds",
			prepare: func() {
				cache.EXPECT().GetAll(gomock.Any()).
					Return([]models.MinecraftServer{
						*worldsTestModel("world-1"),
						*worldsTestModel("world-2"),
					}, nil).Times(1)
			},
			want: []models.MinecraftServer{
				*worldsTestModel("world-1"),
				*worldsTestModel("world-2"),
			},
			success: true,
		},
		{
			caseName: "get empty worlds",
			prepare: func() {
				cache.EXPECT().GetAll(gomock.Any()).
					Return([]models.MinecraftServer{}, nil).Times(1)
			},
			want:    []models.MinecraftServer{},
			success: true,
		},
		{
			caseName: "get all worlds error",
			prepare: func() {
				cache.EXPECT().GetAll(gomock.Any()).
					Return(nil, errors.New("get worlds error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := srvc.GetAll(t.Context())

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

func TestGetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockWorldsRepo(ctrl)
	cache := store_mocks.NewMockWorldsCache(ctrl)

	srvc := NewWorldsService(repo, cache)

	testCases := []struct {
		caseName string
		worldID  string
		prepare  func()
		want     *models.MinecraftServer
		wantErr  error
		success  bool
	}{
		{
			caseName: "get world by id",
			worldID:  "world-id",
			prepare: func() {
				cache.EXPECT().GetByID(gomock.Any(), "world-id").
					Return(worldsTestModel("world-id"), nil).Times(1)
			},
			want:    worldsTestModel("world-id"),
			success: true,
		},
		{
			caseName: "get not found world",
			worldID:  "world-id",
			prepare: func() {
				cache.EXPECT().GetByID(gomock.Any(), "world-id").
					Return(nil, domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "get world by id error",
			worldID:  "world-id",
			prepare: func() {
				cache.EXPECT().GetByID(gomock.Any(), "world-id").
					Return(nil, errors.New("get world error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := srvc.GetByID(t.Context(), tc.worldID)

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

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockWorldsRepo(ctrl)
	cache := store_mocks.NewMockWorldsCache(ctrl)

	srvc := NewWorldsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			worldID string
			userID  string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "delete world",
			data: struct {
				worldID string
				userID  string
			}{
				worldID: "world-id",
				userID:  "user-id",
			},
			prepare: func() {
				repo.EXPECT().Delete(gomock.Any(), "world-id", "user-id").
					Return(nil).Times(1)

				cache.EXPECT().Remove(gomock.Any(), "world-id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "delete not found world",
			data: struct {
				worldID string
				userID  string
			}{
				worldID: "world-id",
				userID:  "user-id",
			},
			prepare: func() {
				repo.EXPECT().Delete(gomock.Any(), "world-id", "user-id").
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "delete world repo error",
			data: struct {
				worldID string
				userID  string
			}{
				worldID: "world-id",
				userID:  "user-id",
			},
			prepare: func() {
				repo.EXPECT().Delete(gomock.Any(), "world-id", "user-id").
					Return(errors.New("delete world error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "delete world cache error",
			data: struct {
				worldID string
				userID  string
			}{
				worldID: "world-id",
				userID:  "user-id",
			},
			prepare: func() {
				repo.EXPECT().Delete(gomock.Any(), "world-id", "user-id").
					Return(nil).Times(1)

				cache.EXPECT().Remove(gomock.Any(), "world-id").
					Return(errors.New("remove cache error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := srvc.Delete(
				t.Context(),
				tc.data.worldID,
				tc.data.userID,
			)

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
