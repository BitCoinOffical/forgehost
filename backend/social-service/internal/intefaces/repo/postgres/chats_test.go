package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGroup(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName    string
		ownerID     string
		title       string
		description string
		canceledCtx bool
		success     bool
	}{
		{
			caseName:    "create group",
			title:       "title",
			description: "description",
			success:     true,
		},
		{
			caseName: "create group with invalid owner id",
			ownerID:  "invalid_uuid",
			title:    "title",
			success:  false,
		},
		{
			caseName:    "create group with canceled context",
			title:       "title",
			canceledCtx: true,
			success:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ownerID := tc.ownerID
			if ownerID == "" {
				ownerID = newUser(t, pool)
			}
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			chatID, err := repo.CreateGroup(ctx, ownerID,
				&models.Chat{Title: strPtr(tc.title), Description: strPtr(tc.description)})
			if tc.success {
				require.NoError(t, err)
				require.NotEmpty(t, chatID)

				chat, err := repo.GetChatByID(context.Background(), chatID)
				require.NoError(t, err)
				require.NotNil(t, chat.Title)
				assert.Equal(t, tc.title, *chat.Title)
				require.NotNil(t, chat.Description)
				assert.Equal(t, tc.description, *chat.Description)
				assert.Equal(t, ownerID, chat.OwnerID)

				members, err := repo.GetUsersFromChat(context.Background(), chatID)
				require.NoError(t, err)
				require.Len(t, members, 1)
				assert.Equal(t, ownerID, members[0].UserID)
				assert.EqualValues(t, "owner", members[0].Role)
				return
			}
			require.Error(t, err)
			assert.Empty(t, chatID)
		})
	}
}

func TestCreateDirect(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (string, string, string)
		sameChat bool
		success  bool
	}{
		{
			caseName: "create direct",
			f: func(t *testing.T) (string, string, string) {
				return newUser(t, pool), newUser(t, pool), ""
			},
			success: true,
		},
		{
			caseName: "create direct in reversed order returns same chat",
			f: func(t *testing.T) (string, string, string) {
				a, b := newUser(t, pool), newUser(t, pool)
				existing, err := repo.CreateDirect(context.Background(), a, b)
				require.NoError(t, err)
				return b, a, existing
			},
			sameChat: true,
			success:  true,
		},
		{
			caseName: "create direct repeated returns same chat",
			f: func(t *testing.T) (string, string, string) {
				a, b := newUser(t, pool), newUser(t, pool)
				existing, err := repo.CreateDirect(context.Background(), a, b)
				require.NoError(t, err)
				return a, b, existing
			},
			sameChat: true,
			success:  true,
		},
		{
			caseName: "create direct with other user returns new chat",
			f: func(t *testing.T) (string, string, string) {
				a, b := newUser(t, pool), newUser(t, pool)
				existing, err := repo.CreateDirect(context.Background(), a, b)
				require.NoError(t, err)
				return a, newUser(t, pool), existing
			},
			sameChat: false,
			success:  true,
		},
		{
			caseName: "create direct with invalid user id",
			f: func(t *testing.T) (string, string, string) {
				return "invalid_uuid", newUser(t, pool), ""
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, targetID, existing := tc.f(t)

			chatID, err := repo.CreateDirect(context.Background(), userID, targetID)
			if !tc.success {
				require.Error(t, err)
				assert.Empty(t, chatID)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, chatID)
			if existing != "" {
				if tc.sameChat {
					assert.Equal(t, existing, chatID)
				} else {
					assert.NotEqual(t, existing, chatID)
				}
			}
		})
	}
}

func TestGetChatByID(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) string
		wantErr  error
		success  bool
	}{
		{
			caseName: "get chat by id",
			f: func(t *testing.T) string {
				chatID, _ := seedChat(t, pool, repo)
				return chatID
			},
			success: true,
		},
		{
			caseName: "get not found chat",
			f:        func(t *testing.T) string { return uuid.NewString() },
			wantErr:  domain.ErrNotFound,
			success:  false,
		},
		{
			caseName: "get chat with invalid id",
			f:        func(t *testing.T) string { return "invalid_uuid" },
			success:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID := tc.f(t)

			chat, err := repo.GetChatByID(context.Background(), chatID)
			if tc.success {
				require.NoError(t, err)
				require.NotNil(t, chat)
				require.NotNil(t, chat.Title)
				assert.Equal(t, "title", *chat.Title)
				require.NotNil(t, chat.Description)
				assert.Equal(t, "description", *chat.Description)
				assert.NotEmpty(t, chat.OwnerID)
				assert.NotEmpty(t, chat.CreatedAt)
				return
			}
			requireErr(t, err, tc.wantErr)
			assert.Nil(t, chat)
		})
	}
}

func TestDeleteChat(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (chatID, userID string)
		stillExists bool
		wantErr     error
		success     bool
	}{
		{
			caseName: "delete chat",
			f: func(t *testing.T) (string, string) {
				return seedChat(t, pool, repo)
			},
			success: true,
		},
		{
			caseName: "delete chat by not owner",
			f: func(t *testing.T) (string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "moderator")
			},
			stillExists: true,
			wantErr:     domain.ErrNotFound,
			success:     false,
		},
		{
			caseName: "delete not found chat",
			f: func(t *testing.T) (string, string) {
				return uuid.NewString(), newUser(t, pool)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, userID := tc.f(t)

			err := repo.DeleteChat(context.Background(), chatID, userID)
			if tc.success {
				require.NoError(t, err)
				_, err := repo.GetChatByID(context.Background(), chatID)
				require.ErrorIs(t, err, domain.ErrNotFound)
				return
			}
			requireErr(t, err, tc.wantErr)
			if tc.stillExists {
				_, err := repo.GetChatByID(context.Background(), chatID)
				require.NoError(t, err)
			}
		})
	}
}

func TestLeaveChat(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (chatID, userID string)
		chatDeleted bool
		success     bool
	}{
		{
			caseName: "leave chat",
			f: func(t *testing.T) (string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "member")
			},
			success: true,
		},
		{
			caseName: "owner leaves chat and chat is deleted",
			f: func(t *testing.T) (string, string) {
				return seedChat(t, pool, repo)
			},
			chatDeleted: true,
			success:     true,
		},
		{
			caseName: "leave chat by not a member",
			f: func(t *testing.T) (string, string) {
				chatID, _ := seedChat(t, pool, repo)
				return chatID, newUser(t, pool)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, userID := tc.f(t)

			err := repo.LeaveChat(context.Background(), chatID, userID)
			if !tc.success {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.chatDeleted {
				_, err := repo.GetChatByID(context.Background(), chatID)
				require.ErrorIs(t, err, domain.ErrNotFound)
				return
			}
			assert.NotContains(t, memberIDs(t, repo, chatID), userID)
		})
	}
}

func TestBuildUpdateChat(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName        string
		f               func(t *testing.T) (chatID, userID string)
		newTitle        *string
		newDescription  *string
		wantTitle       string
		wantDescription string
		wantErr         error
		success         bool
	}{
		{
			caseName: "update title",
			f: func(t *testing.T) (string, string) {
				return seedChat(t, pool, repo)
			},
			newTitle:        strPtr("new_title"),
			wantTitle:       "new_title",
			wantDescription: "description",
			success:         true,
		},
		{
			caseName: "update description",
			f: func(t *testing.T) (string, string) {
				return seedChat(t, pool, repo)
			},
			newDescription:  strPtr("new_description"),
			wantTitle:       "title",
			wantDescription: "new_description",
			success:         true,
		},
		{
			caseName: "update title and description",
			f: func(t *testing.T) (string, string) {
				return seedChat(t, pool, repo)
			},
			newTitle:        strPtr("new_title"),
			newDescription:  strPtr("new_description"),
			wantTitle:       "new_title",
			wantDescription: "new_description",
			success:         true,
		},
		{
			caseName: "update without fields keeps data",
			f: func(t *testing.T) (string, string) {
				return seedChat(t, pool, repo)
			},
			wantTitle:       "title",
			wantDescription: "description",
			success:         true,
		},
		{
			caseName: "update by moderator",
			f: func(t *testing.T) (string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "moderator")
			},
			newTitle: strPtr("new_title"),
			wantErr:  domain.ErrForbidden,
			success:  false,
		},
		{
			caseName: "update by not a member",
			f: func(t *testing.T) (string, string) {
				chatID, _ := seedChat(t, pool, repo)
				return chatID, newUser(t, pool)
			},
			newTitle: strPtr("new_title"),
			success:  false,
		},
		{
			caseName: "update not found chat",
			f: func(t *testing.T) (string, string) {
				return uuid.NewString(), newUser(t, pool)
			},
			newTitle: strPtr("new_title"),
			success:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, userID := tc.f(t)

			chat, err := repo.BuildUpdateChat(context.Background(), chatID, userID,
				&models.Chat{Title: tc.newTitle, Description: tc.newDescription})
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				assert.Nil(t, chat)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, chat)
			assert.Equal(t, chatID, chat.ID)
			require.NotNil(t, chat.Title)
			assert.Equal(t, tc.wantTitle, *chat.Title)
			require.NotNil(t, chat.Description)
			assert.Equal(t, tc.wantDescription, *chat.Description)
			assert.NotEmpty(t, chat.UpdatedAt)

			saved, err := repo.GetChatByID(context.Background(), chatID)
			require.NoError(t, err)
			require.NotNil(t, saved.Title)
			assert.Equal(t, tc.wantTitle, *saved.Title)
		})
	}
}

func TestSetUserRoleChat(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (chatID, actorID, targetID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "set user role",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, ownerID, seedMember(t, pool, repo, chatID, ownerID, "member")
			},
			success: true,
		},
		{
			caseName: "set user role by moderator",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				moderatorID := seedMember(t, pool, repo, chatID, ownerID, "moderator")
				return chatID, moderatorID, seedMember(t, pool, repo, chatID, ownerID, "member")
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
		{
			caseName: "set role for not a member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, ownerID, newUser(t, pool)
			},
			success: false,
		},
		{
			caseName: "set role by not a member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, newUser(t, pool), ownerID
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, actorID, targetID := tc.f(t)

			member, err := repo.SetUserRoleChat(context.Background(), actorID,
				&models.ChatMember{ChatID: chatID, UserID: targetID, Role: "moderator"})
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				assert.Nil(t, member)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, member)
			assert.Equal(t, chatID, member.ChatID)
			assert.Equal(t, targetID, member.UserID)
			assert.EqualValues(t, "moderator", member.Role)
			assert.NotEmpty(t, member.JoinedAt)
		})
	}
}

func TestAddUserInChat(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (chatID, userID string)
		success  bool
	}{
		{
			caseName: "add user in chat",
			f: func(t *testing.T) (string, string) {
				chatID, _ := seedChat(t, pool, repo)
				return chatID, newUser(t, pool)
			},
			success: true,
		},
		{
			caseName: "add user already in chat",
			f: func(t *testing.T) (string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, ownerID
			},
			success: false,
		},
		{
			caseName: "add user in chat with invalid id",
			f: func(t *testing.T) (string, string) {
				return "invalid_uuid", newUser(t, pool)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, userID := tc.f(t)

			err := repo.AddUserInChat(context.Background(), chatID, userID)
			if !tc.success {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Contains(t, memberIDs(t, repo, chatID), userID)
		})
	}
}

func TestKickUserFromChat(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (chatID, targetID, actorID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "owner kicks member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "member"), ownerID
			},
			success: true,
		},
		{
			caseName: "moderator kicks member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				moderatorID := seedMember(t, pool, repo, chatID, ownerID, "moderator")
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "member"), moderatorID
			},
			success: true,
		},
		{
			caseName: "member kicks member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				actorID := seedMember(t, pool, repo, chatID, ownerID, "member")
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "member"), actorID
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
		{
			caseName: "moderator kicks owner",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, ownerID, seedMember(t, pool, repo, chatID, ownerID, "moderator")
			},
			wantErr: domain.ErrCannotKickOwner,
			success: false,
		},
		{
			caseName: "kick not a member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, newUser(t, pool), ownerID
			},
			success: false,
		},
		{
			caseName: "kick by not a member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "member"), newUser(t, pool)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, targetID, actorID := tc.f(t)

			err := repo.KickUserFromChat(context.Background(), chatID, targetID, actorID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotContains(t, memberIDs(t, repo, chatID), targetID)
		})
	}
}

func TestGetUsersFromChat(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (chatID string, wantIDs []string)
	}{
		{
			caseName: "get users from chat",
			f: func(t *testing.T) (string, []string) {
				chatID, ownerID := seedChat(t, pool, repo)
				m1 := seedMember(t, pool, repo, chatID, ownerID, "member")
				m2 := seedMember(t, pool, repo, chatID, ownerID, "moderator")
				return chatID, []string{ownerID, m1, m2}
			},
		},
		{
			caseName: "get users from chat without members",
			f: func(t *testing.T) (string, []string) {
				return uuid.NewString(), nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, wantIDs := tc.f(t)

			members, err := repo.GetUsersFromChat(context.Background(), chatID)
			require.NoError(t, err)
			require.Len(t, members, len(wantIDs))

			ids := make([]string, 0, len(members))
			for _, m := range members {
				ids = append(ids, m.UserID)
				assert.Equal(t, chatID, m.ChatID)
				assert.NotEmpty(t, m.UserName)
				assert.NotEmpty(t, m.Role)
			}
			assert.ElementsMatch(t, wantIDs, ids)
		})
	}
}

func TestBanUser(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (chatID, targetID, actorID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "owner bans member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "member"), ownerID
			},
			success: true,
		},
		{
			caseName: "moderator bans member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				moderatorID := seedMember(t, pool, repo, chatID, ownerID, "moderator")
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "member"), moderatorID
			},
			success: true,
		},
		{
			caseName: "member bans member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				actorID := seedMember(t, pool, repo, chatID, ownerID, "member")
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "member"), actorID
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
		{
			caseName: "ban owner",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, ownerID, seedMember(t, pool, repo, chatID, ownerID, "moderator")
			},
			wantErr: domain.ErrCannotKickOwner,
			success: false,
		},
		{
			caseName: "ban not a member",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, newUser(t, pool), ownerID
			},
			success: false,
		},
		{
			caseName: "ban already banned user",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				targetID := seedMember(t, pool, repo, chatID, ownerID, "member")
				require.NoError(t, repo.BanUser(context.Background(), chatID, targetID, ownerID))
				return chatID, targetID, ownerID
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, targetID, actorID := tc.f(t)

			err := repo.BanUser(context.Background(), chatID, targetID, actorID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			banned, err := repo.CheckBanUser(context.Background(), chatID, targetID)
			require.NoError(t, err)
			assert.True(t, banned)
		})
	}
}

func TestUnbanUser(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (chatID, targetID, actorID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "owner unbans user",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				targetID := seedMember(t, pool, repo, chatID, ownerID, "member")
				require.NoError(t, repo.BanUser(context.Background(), chatID, targetID, ownerID))
				return chatID, targetID, ownerID
			},
			success: true,
		},
		{
			caseName: "moderator unbans user",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				moderatorID := seedMember(t, pool, repo, chatID, ownerID, "moderator")
				targetID := seedMember(t, pool, repo, chatID, ownerID, "member")
				require.NoError(t, repo.BanUser(context.Background(), chatID, targetID, ownerID))
				return chatID, targetID, moderatorID
			},
			success: true,
		},
		{
			caseName: "unban not banned user",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "member"), ownerID
			},
			success: true,
		},
		{
			caseName: "member unbans user",
			f: func(t *testing.T) (string, string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				actorID := seedMember(t, pool, repo, chatID, ownerID, "member")
				targetID := seedMember(t, pool, repo, chatID, ownerID, "member")
				require.NoError(t, repo.BanUser(context.Background(), chatID, targetID, ownerID))
				return chatID, targetID, actorID
			},
			wantErr: domain.ErrForbidden,
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, targetID, actorID := tc.f(t)

			err := repo.UnbanUser(context.Background(), chatID, targetID, actorID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)

				banned, err := repo.CheckBanUser(context.Background(), chatID, targetID)
				require.NoError(t, err)
				assert.True(t, banned)
				return
			}
			require.NoError(t, err)

			banned, err := repo.CheckBanUser(context.Background(), chatID, targetID)
			require.NoError(t, err)
			assert.False(t, banned)
		})
	}
}

func TestCheckBanUser(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName   string
		f          func(t *testing.T) (chatID, targetID string)
		wantBanned bool
	}{
		{
			caseName: "user is banned",
			f: func(t *testing.T) (string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				targetID := seedMember(t, pool, repo, chatID, ownerID, "member")
				require.NoError(t, repo.BanUser(context.Background(), chatID, targetID, ownerID))
				return chatID, targetID
			},
			wantBanned: true,
		},
		{
			caseName: "user is not banned",
			f: func(t *testing.T) (string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, seedMember(t, pool, repo, chatID, ownerID, "member")
			},
			wantBanned: false,
		},
		{
			caseName: "user is banned in other chat",
			f: func(t *testing.T) (string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				targetID := seedMember(t, pool, repo, chatID, ownerID, "member")
				require.NoError(t, repo.BanUser(context.Background(), chatID, targetID, ownerID))
				otherChatID, _ := seedChat(t, pool, repo)
				return otherChatID, targetID
			},
			wantBanned: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, targetID := tc.f(t)

			banned, err := repo.CheckBanUser(context.Background(), chatID, targetID)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBanned, banned)
		})
	}
}

func TestCheckBlock(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (userID, targetID string)
		wantBlocked bool
	}{
		{
			caseName: "user blocked target",
			f: func(t *testing.T) (string, string) {
				a, b := newUser(t, pool), newUser(t, pool)
				seedBlock(t, pool, a, b)
				return a, b
			},
			wantBlocked: true,
		},
		{
			caseName: "block works in both directions",
			f: func(t *testing.T) (string, string) {
				a, b := newUser(t, pool), newUser(t, pool)
				seedBlock(t, pool, a, b)
				return b, a
			},
			wantBlocked: true,
		},
		{
			caseName: "no block",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), newUser(t, pool)
			},
			wantBlocked: false,
		},
		{
			caseName: "block with other user does not count",
			f: func(t *testing.T) (string, string) {
				a, b, c := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				seedBlock(t, pool, a, c)
				return a, b
			},
			wantBlocked: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, targetID := tc.f(t)

			blocked, err := repo.CheckBlock(context.Background(), userID, targetID)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBlocked, blocked)
		})
	}
}

func TestGetChatList(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (userID string, wantChatIDs []string)
	}{
		{
			caseName: "get chat list ordered by last activity",
			f: func(t *testing.T) (string, []string) {
				ctx := context.Background()
				a, b, c := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				ab, err := repo.CreateDirect(ctx, a, b)
				require.NoError(t, err)
				ac, err := repo.CreateDirect(ctx, a, c)
				require.NoError(t, err)
				require.NoError(t, repo.TouchChat(ctx, ab))
				return a, []string{ab, ac}
			},
		},
		{
			caseName: "get chat list without chats",
			f: func(t *testing.T) (string, []string) {
				return newUser(t, pool), nil
			},
		},
		{
			caseName: "get chat list excludes chats of other users",
			f: func(t *testing.T) (string, []string) {
				ctx := context.Background()
				a, b, c := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				ab, err := repo.CreateDirect(ctx, a, b)
				require.NoError(t, err)
				_, err = repo.CreateDirect(ctx, b, c)
				require.NoError(t, err)
				return a, []string{ab}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, wantChatIDs := tc.f(t)

			chats, err := repo.GetChatList(context.Background(), userID)
			require.NoError(t, err)
			require.Len(t, chats, len(wantChatIDs))

			ids := make([]string, 0, len(chats))
			for _, chat := range chats {
				ids = append(ids, chat.ID)
				assert.NotEmpty(t, chat.Username)
				assert.NotEqual(t, userID, chat.UserID)
			}
			if len(wantChatIDs) > 0 {
				assert.Equal(t, wantChatIDs, ids)
			}
		})
	}
}

func TestTouchChat(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName  string
		f         func(t *testing.T) string
		checkTime bool
		success   bool
	}{
		{
			caseName: "touch chat",
			f: func(t *testing.T) string {
				chatID, err := repo.CreateDirect(context.Background(), newUser(t, pool), newUser(t, pool))
				require.NoError(t, err)
				return chatID
			},
			checkTime: true,
			success:   true,
		},
		{
			caseName: "touch not found chat",
			f:        func(t *testing.T) string { return uuid.NewString() },
			success:  true,
		},
		{
			caseName: "touch chat with invalid id",
			f:        func(t *testing.T) string { return "invalid_uuid" },
			success:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID := tc.f(t)

			var before time.Time
			if tc.checkTime {
				before = directUpdatedAt(t, pool, chatID)
			}

			err := repo.TouchChat(context.Background(), chatID)
			if !tc.success {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.checkTime {
				assert.True(t, directUpdatedAt(t, pool, chatID).After(before))
			}
		})
	}
}

func TestMarkAsRead(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewChatsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (chatID, userID string)
		success  bool
	}{
		{
			caseName: "mark as read",
			f: func(t *testing.T) (string, string) {
				chatID, ownerID := seedChat(t, pool, repo)
				return chatID, ownerID
			},
			success: true,
		},
		{
			caseName: "mark as read with invalid chat id",
			f: func(t *testing.T) (string, string) {
				return "invalid_uuid", newUser(t, pool)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, userID := tc.f(t)

			err := repo.MarkAsRead(context.Background(), chatID, userID)
			if !tc.success {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			first := lastReadAt(t, pool, chatID, userID)
			assert.False(t, first.IsZero())

			require.NoError(t, repo.MarkAsRead(context.Background(), chatID, userID))
			assert.True(t, lastReadAt(t, pool, chatID, userID).After(first))
		})
	}
}
