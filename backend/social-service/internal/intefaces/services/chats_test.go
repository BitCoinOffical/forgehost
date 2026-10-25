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

func TestCreateGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			ownerID     string
			title       string
			description *string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "create group",
			data: struct {
				ownerID     string
				title       string
				description *string
			}{ownerID: "owner_id", title: "title", description: strPtr("description")},
			prepare: func() {
				repo.EXPECT().CreateGroup(gomock.Any(), "owner_id",
					&models.Chat{Title: strPtr("title"), Description: strPtr("description")}).
					Return("chat_id", nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "create group without description",
			data: struct {
				ownerID     string
				title       string
				description *string
			}{ownerID: "owner_id", title: "title", description: nil},
			prepare: func() {
				repo.EXPECT().CreateGroup(gomock.Any(), "owner_id",
					&models.Chat{Title: strPtr("title")}).
					Return("chat_id", nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "create group error",
			data: struct {
				ownerID     string
				title       string
				description *string
			}{ownerID: "owner_id", title: "title", description: strPtr("description")},
			prepare: func() {
				repo.EXPECT().CreateGroup(gomock.Any(), "owner_id", gomock.Any()).
					Return("", errors.New("create group error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.CreateChatDTO{
				Title:       tc.data.title,
				Description: tc.data.description,
			}
			chatID, err := chat.CreateGroup(t.Context(), tc.data.ownerID, &req)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, "chat_id", chatID)
			} else {
				require.Error(t, err)
				assert.Empty(t, chatID)
			}
		})
	}
}

func TestCreateDirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

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
			caseName: "create direct",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().CheckBlock(gomock.Any(), "user_id", "target_id").
					Return(false, nil).Times(1)
				repo.EXPECT().CreateDirect(gomock.Any(), "user_id", "target_id").
					Return("chat_id", nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "create direct with blocked user",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().CheckBlock(gomock.Any(), "user_id", "target_id").
					Return(true, nil).Times(1)
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
		{
			caseName: "check block error",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().CheckBlock(gomock.Any(), "user_id", "target_id").
					Return(false, errors.New("check block error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "create direct error",
			data: struct {
				userID   string
				targetID string
			}{userID: "user_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().CheckBlock(gomock.Any(), "user_id", "target_id").
					Return(false, nil).Times(1)
				repo.EXPECT().CreateDirect(gomock.Any(), "user_id", "target_id").
					Return("", errors.New("create direct error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			chatID, err := chat.CreateDirect(t.Context(), tc.data.userID, tc.data.targetID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, "chat_id", chatID)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
				assert.Empty(t, chatID)
			}
		})
	}
}

func TestGetChatList(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
		}
		prepare func()
		want    []models.FeedDirects
		success bool
	}{
		{
			caseName: "get chat list",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetChatList(gomock.Any(), "user_id").
					Return([]models.FeedDirects{{ID: "chat_1"}, {ID: "chat_2"}}, nil).Times(1)
			},
			want:    []models.FeedDirects{{ID: "chat_1"}, {ID: "chat_2"}},
			success: true,
		},
		{
			caseName: "get empty chat list",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetChatList(gomock.Any(), "user_id").
					Return(nil, nil).Times(1)
			},
			want:    nil,
			success: true,
		},
		{
			caseName: "get chat list error",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetChatList(gomock.Any(), "user_id").
					Return(nil, errors.New("get chat list error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := chat.GetChatList(t.Context(), tc.data.userID)
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

func TestMarkAsRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID string
			userID string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "mark as read",
			data: struct {
				chatID string
				userID string
			}{chatID: "chat_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().MarkAsRead(gomock.Any(), "chat_id", "user_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "mark as read error",
			data: struct {
				chatID string
				userID string
			}{chatID: "chat_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().MarkAsRead(gomock.Any(), "chat_id", "user_id").
					Return(errors.New("mark as read error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := chat.MarkAsRead(t.Context(), tc.data.chatID, tc.data.userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetChatByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID string
		}
		prepare func()
		want    *models.Chat
		wantErr error
		success bool
	}{
		{
			caseName: "get chat by id",
			data: struct {
				chatID string
			}{chatID: "chat_id"},
			prepare: func() {
				repo.EXPECT().GetChatByID(gomock.Any(), "chat_id").
					Return(&models.Chat{Title: strPtr("title")}, nil).Times(1)
			},
			want:    &models.Chat{Title: strPtr("title")},
			success: true,
		},
		{
			caseName: "get not found chat",
			data: struct {
				chatID string
			}{chatID: "chat_id"},
			prepare: func() {
				repo.EXPECT().GetChatByID(gomock.Any(), "chat_id").
					Return(nil, domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "get chat by id error",
			data: struct {
				chatID string
			}{chatID: "chat_id"},
			prepare: func() {
				repo.EXPECT().GetChatByID(gomock.Any(), "chat_id").
					Return(nil, errors.New("get chat by id error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := chat.GetChatByID(t.Context(), tc.data.chatID)
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

func TestDeleteChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID  string
			ownerID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "delete chat",
			data: struct {
				chatID  string
				ownerID string
			}{chatID: "chat_id", ownerID: "owner_id"},
			prepare: func() {
				repo.EXPECT().DeleteChat(gomock.Any(), "chat_id", "owner_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "delete not found chat",
			data: struct {
				chatID  string
				ownerID string
			}{chatID: "chat_id", ownerID: "owner_id"},
			prepare: func() {
				repo.EXPECT().DeleteChat(gomock.Any(), "chat_id", "owner_id").
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "delete chat error",
			data: struct {
				chatID  string
				ownerID string
			}{chatID: "chat_id", ownerID: "owner_id"},
			prepare: func() {
				repo.EXPECT().DeleteChat(gomock.Any(), "chat_id", "owner_id").
					Return(errors.New("delete chat error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := chat.DeleteChat(t.Context(), tc.data.chatID, tc.data.ownerID)
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

func TestLeaveChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID string
			userID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "leave chat",
			data: struct {
				chatID string
				userID string
			}{chatID: "chat_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().LeaveChat(gomock.Any(), "chat_id", "user_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "leave not found chat",
			data: struct {
				chatID string
				userID string
			}{chatID: "chat_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().LeaveChat(gomock.Any(), "chat_id", "user_id").
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "leave chat error",
			data: struct {
				chatID string
				userID string
			}{chatID: "chat_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().LeaveChat(gomock.Any(), "chat_id", "user_id").
					Return(errors.New("leave chat error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := chat.LeaveChat(t.Context(), tc.data.chatID, tc.data.userID)
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

func TestBuildUpdateChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID      string
			userID      string
			title       *string
			description *string
		}
		prepare func()
		want    *models.Chat
		wantErr error
		success bool
	}{
		{
			caseName: "update title and description",
			data: struct {
				chatID      string
				userID      string
				title       *string
				description *string
			}{chatID: "chat_id", userID: "user_id", title: strPtr("new_title"), description: strPtr("new_description")},
			prepare: func() {
				repo.EXPECT().BuildUpdateChat(gomock.Any(), "chat_id", "user_id",
					&models.Chat{Title: strPtr("new_title"), Description: strPtr("new_description")}).
					Return(&models.Chat{Title: strPtr("new_title"), Description: strPtr("new_description")}, nil).Times(1)
			},
			want:    &models.Chat{Title: strPtr("new_title"), Description: strPtr("new_description")},
			success: true,
		},
		{
			caseName: "update only title",
			data: struct {
				chatID      string
				userID      string
				title       *string
				description *string
			}{chatID: "chat_id", userID: "user_id", title: strPtr("new_title")},
			prepare: func() {
				repo.EXPECT().BuildUpdateChat(gomock.Any(), "chat_id", "user_id",
					&models.Chat{Title: strPtr("new_title")}).
					Return(&models.Chat{Title: strPtr("new_title")}, nil).Times(1)
			},
			want:    &models.Chat{Title: strPtr("new_title")},
			success: true,
		},
		{
			caseName: "update chat without permission",
			data: struct {
				chatID      string
				userID      string
				title       *string
				description *string
			}{chatID: "chat_id", userID: "user_id", title: strPtr("new_title")},
			prepare: func() {
				repo.EXPECT().BuildUpdateChat(gomock.Any(), "chat_id", "user_id", gomock.Any()).
					Return(nil, domain.ErrForbidden).Times(1)
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
		{
			caseName: "update chat error",
			data: struct {
				chatID      string
				userID      string
				title       *string
				description *string
			}{chatID: "chat_id", userID: "user_id", title: strPtr("new_title")},
			prepare: func() {
				repo.EXPECT().BuildUpdateChat(gomock.Any(), "chat_id", "user_id", gomock.Any()).
					Return(nil, errors.New("update chat error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.UpdateChatDTO{
				Title:       tc.data.title,
				Description: tc.data.description,
			}
			got, err := chat.BuildUpdateChat(t.Context(), tc.data.chatID, tc.data.userID, &req)
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

func TestSetUserRoleChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID   string
			ownerID  string
			targetID string
			role     string
		}
		prepare func()
		want    *models.ChatMember
		wantErr error
		success bool
	}{
		{
			caseName: "set user role",
			data: struct {
				chatID   string
				ownerID  string
				targetID string
				role     string
			}{chatID: "chat_id", ownerID: "owner_id", targetID: "target_id", role: "moderator"},
			prepare: func() {
				repo.EXPECT().SetUserRoleChat(gomock.Any(), "owner_id",
					&models.ChatMember{ChatID: "chat_id", UserID: "target_id", Role: models.MemberRole("moderator")}).
					Return(&models.ChatMember{ChatID: "chat_id", UserID: "target_id", Role: models.MemberRole("moderator")}, nil).Times(1)
			},
			want:    &models.ChatMember{ChatID: "chat_id", UserID: "target_id", Role: models.MemberRole("moderator")},
			success: true,
		},
		{
			caseName: "set user role without permission",
			data: struct {
				chatID   string
				ownerID  string
				targetID string
				role     string
			}{chatID: "chat_id", ownerID: "owner_id", targetID: "target_id", role: "moderator"},
			prepare: func() {
				repo.EXPECT().SetUserRoleChat(gomock.Any(), "owner_id", gomock.Any()).
					Return(nil, domain.ErrForbidden).Times(1)
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
		{
			caseName: "set user role error",
			data: struct {
				chatID   string
				ownerID  string
				targetID string
				role     string
			}{chatID: "chat_id", ownerID: "owner_id", targetID: "target_id", role: "moderator"},
			prepare: func() {
				repo.EXPECT().SetUserRoleChat(gomock.Any(), "owner_id", gomock.Any()).
					Return(nil, errors.New("set user role error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.SetMemberRoleDTO{
				UserID: tc.data.targetID,
				Role:   dto.MemberRole(tc.data.role),
			}
			got, err := chat.SetUserRoleChat(t.Context(), tc.data.chatID, tc.data.ownerID, &req)
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

func TestAddUserInChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID string
			userID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "add user in chat",
			data: struct {
				chatID string
				userID string
			}{chatID: "chat_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().CheckBanUser(gomock.Any(), "chat_id", "user_id").
					Return(false, nil).Times(1)
				repo.EXPECT().AddUserInChat(gomock.Any(), "chat_id", "user_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "add banned user in chat",
			data: struct {
				chatID string
				userID string
			}{chatID: "chat_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().CheckBanUser(gomock.Any(), "chat_id", "user_id").
					Return(true, nil).Times(1)
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
		{
			caseName: "check ban user error",
			data: struct {
				chatID string
				userID string
			}{chatID: "chat_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().CheckBanUser(gomock.Any(), "chat_id", "user_id").
					Return(false, errors.New("check ban user error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "add user in chat error",
			data: struct {
				chatID string
				userID string
			}{chatID: "chat_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().CheckBanUser(gomock.Any(), "chat_id", "user_id").
					Return(false, nil).Times(1)
				repo.EXPECT().AddUserInChat(gomock.Any(), "chat_id", "user_id").
					Return(errors.New("add user in chat error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := chat.AddUserInChat(t.Context(), tc.data.chatID, tc.data.userID)
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

func TestKickUserFromChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID   string
			targetID string
			userID   string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "kick user from chat",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().KickUserFromChat(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "kick user without permission",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().KickUserFromChat(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(domain.ErrForbidden).Times(1)
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
		{
			caseName: "kick owner",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().KickUserFromChat(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(domain.ErrCannotKickOwner).Times(1)
			},
			wantErr: domain.ErrCannotKickOwner,
			success: false,
		},
		{
			caseName: "kick user error",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().KickUserFromChat(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(errors.New("kick user error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := chat.KickUserFromChat(t.Context(), tc.data.chatID, tc.data.targetID, tc.data.userID)
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

func TestGetUsersFromChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID string
		}
		prepare func()
		want    []models.FeedChatMember
		success bool
	}{
		{
			caseName: "get users from chat",
			data: struct {
				chatID string
			}{chatID: "chat_id"},
			prepare: func() {
				repo.EXPECT().GetUsersFromChat(gomock.Any(), "chat_id").
					Return([]models.FeedChatMember{{UserID: "user_1"}, {UserID: "user_2"}}, nil).Times(1)
			},
			want:    []models.FeedChatMember{{UserID: "user_1"}, {UserID: "user_2"}},
			success: true,
		},
		{
			caseName: "get users from chat error",
			data: struct {
				chatID string
			}{chatID: "chat_id"},
			prepare: func() {
				repo.EXPECT().GetUsersFromChat(gomock.Any(), "chat_id").
					Return(nil, errors.New("get users from chat error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := chat.GetUsersFromChat(t.Context(), tc.data.chatID)
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

func TestBanUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID   string
			targetID string
			userID   string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "ban user",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().BanUser(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "ban user without permission",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().BanUser(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(domain.ErrForbidden).Times(1)
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
		{
			caseName: "ban owner",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().BanUser(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(domain.ErrCannotKickOwner).Times(1)
			},
			wantErr: domain.ErrCannotKickOwner,
			success: false,
		},
		{
			caseName: "ban user error",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().BanUser(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(errors.New("ban user error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := chat.BanUser(t.Context(), tc.data.chatID, tc.data.targetID, tc.data.userID)
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

func TestUnbanUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID   string
			targetID string
			userID   string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "unban user",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().UnbanUser(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "unban user without permission",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().UnbanUser(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(domain.ErrForbidden).Times(1)
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
		{
			caseName: "unban user error",
			data: struct {
				chatID   string
				targetID string
				userID   string
			}{chatID: "chat_id", targetID: "target_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().UnbanUser(gomock.Any(), "chat_id", "target_id", "user_id").
					Return(errors.New("unban user error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := chat.UnbanUser(t.Context(), tc.data.chatID, tc.data.targetID, tc.data.userID)
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

func TestCheckBanUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockChatsRepo(ctrl)

	chat := NewChatService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID   string
			targetID string
		}
		prepare    func()
		wantBanned bool
		success    bool
	}{
		{
			caseName: "user is banned",
			data: struct {
				chatID   string
				targetID string
			}{chatID: "chat_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().CheckBanUser(gomock.Any(), "chat_id", "target_id").
					Return(true, nil).Times(1)
			},
			wantBanned: true,
			success:    true,
		},
		{
			caseName: "user is not banned",
			data: struct {
				chatID   string
				targetID string
			}{chatID: "chat_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().CheckBanUser(gomock.Any(), "chat_id", "target_id").
					Return(false, nil).Times(1)
			},
			wantBanned: false,
			success:    true,
		},
		{
			caseName: "check ban user error",
			data: struct {
				chatID   string
				targetID string
			}{chatID: "chat_id", targetID: "target_id"},
			prepare: func() {
				repo.EXPECT().CheckBanUser(gomock.Any(), "chat_id", "target_id").
					Return(false, errors.New("check ban user error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			banned, err := chat.CheckBanUser(t.Context(), tc.data.chatID, tc.data.targetID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.wantBanned, banned)
			} else {
				require.Error(t, err)
				assert.False(t, banned)
			}
		})
	}
}
