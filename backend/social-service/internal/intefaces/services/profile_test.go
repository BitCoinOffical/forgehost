package services

import (
	"errors"
	"testing"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	store_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetProfileByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockProfileRepo(ctrl)

	profile := NewProfileService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
		}
		prepare func()
		want    *models.ProfileResponse
		wantErr error
		success bool
	}{
		{
			caseName: "get profile by id",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetProfileByID(gomock.Any(), "user_id").
					Return(&models.Profile{UserID: "user_id", UserName: strPtr("name")},
						[]models.FeedPost{{PostID: "1"}, {PostID: "2"}}, nil).Times(1)
			},
			want: &models.ProfileResponse{
				Profile: models.Profile{UserID: "user_id", UserName: strPtr("name")},
				Posts:   []models.FeedPost{{PostID: "1"}, {PostID: "2"}},
			},
			success: true,
		},
		{
			caseName: "get profile without posts",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetProfileByID(gomock.Any(), "user_id").
					Return(&models.Profile{UserID: "user_id"}, nil, nil).Times(1)
			},
			want: &models.ProfileResponse{
				Profile: models.Profile{UserID: "user_id"},
			},
			success: true,
		},
		{
			caseName: "get not found profile",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetProfileByID(gomock.Any(), "user_id").
					Return(nil, nil, domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "get profile by id error",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetProfileByID(gomock.Any(), "user_id").
					Return(nil, nil, errors.New("get profile by id error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := profile.GetProfileByID(t.Context(), tc.data.userID)
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

func TestSaveProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockProfileRepo(ctrl)

	profile := NewProfileService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "save profile",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().SaveProfile(gomock.Any(), "user_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "save existing profile",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().SaveProfile(gomock.Any(), "user_id").
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "save profile error",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().SaveProfile(gomock.Any(), "user_id").
					Return(errors.New("save profile error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			event := dto.UserProfileDTO{
				UserID: tc.data.userID,
			}
			err := profile.SaveProfile(t.Context(), &event)
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

func TestUpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockProfileRepo(ctrl)

	profile := NewProfileService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID    string
			userName  string
			bio       string
			avatarURL string
		}
		prepare func()
		want    *models.Profile
		wantErr error
		success bool
	}{
		{
			caseName: "update profile",
			data: struct {
				userID    string
				userName  string
				bio       string
				avatarURL string
			}{userID: "user_id", userName: "name", bio: "bio", avatarURL: "https://avatar.test/1.png"},
			prepare: func() {
				repo.EXPECT().BuildUpdateProfile(gomock.Any(), &models.Profile{
					UserID:    "user_id",
					UserName:  strPtr("name"),
					Bio:       strPtr("bio"),
					AvatarUrl: strPtr("https://avatar.test/1.png"),
				}).Return(&models.Profile{UserID: "user_id", UserName: strPtr("name")}, nil).Times(1)
			},
			want:    &models.Profile{UserID: "user_id", UserName: strPtr("name")},
			success: true,
		},
		{
			caseName: "update profile with empty fields",
			data: struct {
				userID    string
				userName  string
				bio       string
				avatarURL string
			}{userID: "user_id", userName: "name"},
			prepare: func() {
				repo.EXPECT().BuildUpdateProfile(gomock.Any(), &models.Profile{
					UserID:    "user_id",
					UserName:  strPtr("name"),
					Bio:       strPtr(""),
					AvatarUrl: strPtr(""),
				}).Return(&models.Profile{UserID: "user_id", UserName: strPtr("name")}, nil).Times(1)
			},
			want:    &models.Profile{UserID: "user_id", UserName: strPtr("name")},
			success: true,
		},
		{
			caseName: "update profile with taken username",
			data: struct {
				userID    string
				userName  string
				bio       string
				avatarURL string
			}{userID: "user_id", userName: "name", bio: "bio", avatarURL: "url"},
			prepare: func() {
				repo.EXPECT().BuildUpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "update not found profile",
			data: struct {
				userID    string
				userName  string
				bio       string
				avatarURL string
			}{userID: "user_id", userName: "name", bio: "bio", avatarURL: "url"},
			prepare: func() {
				repo.EXPECT().BuildUpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "update profile error",
			data: struct {
				userID    string
				userName  string
				bio       string
				avatarURL string
			}{userID: "user_id", userName: "name", bio: "bio", avatarURL: "url"},
			prepare: func() {
				repo.EXPECT().BuildUpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("update profile error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.UpdateProfileDTO{
				UserName:  &tc.data.userName,
				Bio:       &tc.data.bio,
				AvatarUrl: &tc.data.avatarURL,
			}
			got, err := profile.UpdateProfile(t.Context(), &req, tc.data.userID)
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

func TestGetSubscriptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockProfileRepo(ctrl)

	profile := NewProfileService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
		}
		prepare func()
		want    []models.Subscriptions
		success bool
	}{
		{
			caseName: "get subscriptions",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetSubscriptions(gomock.Any(), "user_id").
					Return([]models.Subscriptions{{UserID: "user_1"}, {UserID: "user_2"}}, nil).Times(1)
			},
			want:    []models.Subscriptions{{UserID: "user_1"}, {UserID: "user_2"}},
			success: true,
		},
		{
			caseName: "get empty subscriptions",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetSubscriptions(gomock.Any(), "user_id").
					Return(nil, nil).Times(1)
			},
			want:    nil,
			success: true,
		},
		{
			caseName: "get subscriptions error",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetSubscriptions(gomock.Any(), "user_id").
					Return(nil, errors.New("get subscriptions error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := profile.GetSubscriptions(t.Context(), tc.data.userID)
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

func TestGetSubscribers(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockProfileRepo(ctrl)

	profile := NewProfileService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
		}
		prepare func()
		want    []models.Subscribes
		success bool
	}{
		{
			caseName: "get subscribers",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetSubscribers(gomock.Any(), "user_id").
					Return([]models.Subscribes{{TargetID: "target_1"}, {TargetID: "target_2"}}, nil).Times(1)
			},
			want:    []models.Subscribes{{TargetID: "target_1"}, {TargetID: "target_2"}},
			success: true,
		},
		{
			caseName: "get empty subscribers",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetSubscribers(gomock.Any(), "user_id").
					Return(nil, nil).Times(1)
			},
			want:    nil,
			success: true,
		},
		{
			caseName: "get subscribers error",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetSubscribers(gomock.Any(), "user_id").
					Return(nil, errors.New("get subscribers error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := profile.GetSubscribers(t.Context(), tc.data.userID)
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

func TestSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockProfileRepo(ctrl)

	profile := NewProfileService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID   string
			targetID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "subscribe",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().Subscribe(gomock.Any(), "user_id", "target_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "subscribe to self",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "user_id"},
			prepare: func() {},
			wantErr: domain.ErrSubscribeToSelf,
			success: false,
		},
		{
			caseName: "subscribe twice",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().Subscribe(gomock.Any(), "user_id", "target_id").
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "subscribe error",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().Subscribe(gomock.Any(), "user_id", "target_id").
					Return(errors.New("subscribe error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := profile.Subscribe(t.Context(), tc.data.userID, tc.data.targetID)
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

func TestUnSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockProfileRepo(ctrl)

	profile := NewProfileService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID   string
			targetID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "unsubscribe",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().UnSubscribe(gomock.Any(), "user_id", "target_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "unsubscribe without subscription",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().UnSubscribe(gomock.Any(), "user_id", "target_id").
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "unsubscribe error",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().UnSubscribe(gomock.Any(), "user_id", "target_id").
					Return(errors.New("unsubscribe error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := profile.UnSubscribe(t.Context(), tc.data.userID, tc.data.targetID)
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

func TestReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockProfileRepo(ctrl)

	profile := NewProfileService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID   string
			targetID string
			cause    string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "report profile",
			data: struct {
				userID   string
				targetID string
				cause    string
			}{userID: "user_id", targetID: "target_id", cause: "spam"},
			prepare: func() {
				repo.EXPECT().CreateProfileReport(gomock.Any(), "user_id", "target_id", "spam").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "report self",
			data: struct {
				userID   string
				targetID string
				cause    string
			}{userID: "user_id", targetID: "user_id", cause: "spam"},
			prepare: func() {},
			wantErr: domain.ErrSubscribeToSelf,
			success: false,
		},
		{
			caseName: "report profile twice",
			data: struct {
				userID   string
				targetID string
				cause    string
			}{userID: "user_id", targetID: "target_id", cause: "spam"},
			prepare: func() {
				repo.EXPECT().CreateProfileReport(gomock.Any(), "user_id", "target_id", "spam").
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "report profile error",
			data: struct {
				userID   string
				targetID string
				cause    string
			}{userID: "user_id", targetID: "target_id", cause: "spam"},
			prepare: func() {
				repo.EXPECT().CreateProfileReport(gomock.Any(), "user_id", "target_id", "spam").
					Return(errors.New("report profile error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.ReportDTO{
				Cause: tc.data.cause,
			}
			err := profile.Report(t.Context(), tc.data.userID, tc.data.targetID, &req)
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

func TestBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockProfileRepo(ctrl)

	profile := NewProfileService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID   string
			targetID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "block user",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().Block(gomock.Any(), "user_id", "target_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "block user twice",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().Block(gomock.Any(), "user_id", "target_id").
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "block user error",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().Block(gomock.Any(), "user_id", "target_id").
					Return(errors.New("block user error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := profile.Block(t.Context(), tc.data.userID, tc.data.targetID)
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
