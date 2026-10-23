package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckBlockMessage(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewMessagersRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (userID, targetID string)
		canceledCtx bool
		want        bool
		success     bool
	}{
		{
			caseName: "no block",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), newUser(t, pool)
			},
			want:    false,
			success: true,
		},
		{
			caseName: "user blocked target",
			f: func(t *testing.T) (string, string) {
				userID, targetID := newUser(t, pool), newUser(t, pool)
				seedBlock(t, pool, userID, targetID)
				return userID, targetID
			},
			want:    true,
			success: true,
		},
		{
			caseName: "target blocked user",
			f: func(t *testing.T) (string, string) {
				userID, targetID := newUser(t, pool), newUser(t, pool)
				seedBlock(t, pool, targetID, userID)
				return userID, targetID
			},
			want:    true,
			success: true,
		},
		{
			caseName: "block of other users does not affect",
			f: func(t *testing.T) (string, string) {
				userID, targetID, other := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				seedBlock(t, pool, userID, other)
				seedBlock(t, pool, other, targetID)
				return userID, targetID
			},
			want:    false,
			success: true,
		},
		{
			caseName: "check block with invalid id",
			f: func(t *testing.T) (string, string) {
				return "invalid_uuid", newUser(t, pool)
			},
			success: false,
		},
		{
			caseName: "check block with canceled context",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), newUser(t, pool)
			},
			canceledCtx: true,
			success:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, targetID := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			blocked, err := repo.CheckBlock(ctx, userID, targetID)
			if !tc.success {
				require.Error(t, err)
				assert.False(t, blocked)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, blocked)
		})
	}
}

func TestGetUsersFromChatMessage(t *testing.T) {
	pool := setupTestDB(t)
	chats := NewChatsRepo(pool)
	repo := NewMessagersRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (chatID string, wantIDs []string)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "chat with owner only",
			f: func(t *testing.T) (string, []string) {
				chatID, ownerID := seedChat(t, pool, chats)
				return chatID, []string{ownerID}
			},
			success: true,
		},
		{
			caseName: "chat with members",
			f: func(t *testing.T) (string, []string) {
				chatID, ownerID := seedChat(t, pool, chats)
				first := seedMember(t, pool, chats, chatID, ownerID, "member")
				second := seedMember(t, pool, chats, chatID, ownerID, "member")
				return chatID, []string{ownerID, first, second}
			},
			success: true,
		},
		{
			caseName: "chat excludes members of other chats",
			f: func(t *testing.T) (string, []string) {
				chatID, ownerID := seedChat(t, pool, chats)
				otherChatID, otherOwnerID := seedChat(t, pool, chats)
				seedMember(t, pool, chats, otherChatID, otherOwnerID, "member")
				return chatID, []string{ownerID}
			},
			success: true,
		},
		{
			caseName: "not found chat",
			f: func(t *testing.T) (string, []string) {
				return uuid.NewString(), []string{}
			},
			success: true,
		},
		{
			caseName: "get users with invalid chat id",
			f: func(t *testing.T) (string, []string) {
				return "invalid_uuid", nil
			},
			success: false,
		},
		{
			caseName: "get users with canceled context",
			f: func(t *testing.T) (string, []string) {
				chatID, _ := seedChat(t, pool, chats)
				return chatID, nil
			},
			canceledCtx: true,
			success:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, wantIDs := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			members, err := repo.GetUsersFromChat(ctx, chatID)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, members)
				return
			}
			require.NoError(t, err)

			gotIDs := make([]string, 0, len(members))
			for _, m := range members {
				gotIDs = append(gotIDs, m.UserID)
				assert.Equal(t, chatID, m.ChatID)
				assert.Equal(t, "user_"+m.UserID[:8], *m.UserName)
				assert.False(t, m.JoinedAt.IsZero())
			}
			assert.ElementsMatch(t, wantIDs, gotIDs)
		})
	}
}

func TestCheckBanUserMessage(t *testing.T) {
	pool := setupTestDB(t)
	chats := NewChatsRepo(pool)
	repo := NewMessagersRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (chatID, targetID string)
		canceledCtx bool
		want        bool
		success     bool
	}{
		{
			caseName: "user is banned",
			f: func(t *testing.T) (string, string) {
				chatID, _ := seedChat(t, pool, chats)
				targetID := newUser(t, pool)
				seedChatBan(t, pool, chatID, targetID, newUser(t, pool))
				return chatID, targetID
			},
			want:    true,
			success: true,
		},
		{
			caseName: "user is not banned",
			f: func(t *testing.T) (string, string) {
				chatID, _ := seedChat(t, pool, chats)
				return chatID, newUser(t, pool)
			},
			want:    false,
			success: true,
		},
		{
			caseName: "user is banned in other chat",
			f: func(t *testing.T) (string, string) {
				chatID, _ := seedChat(t, pool, chats)
				otherChatID, _ := seedChat(t, pool, chats)
				targetID := newUser(t, pool)
				seedChatBan(t, pool, otherChatID, targetID, newUser(t, pool))
				return chatID, targetID
			},
			want:    false,
			success: true,
		},
		{
			caseName: "other user is banned in chat",
			f: func(t *testing.T) (string, string) {
				chatID, _ := seedChat(t, pool, chats)
				seedChatBan(t, pool, chatID, newUser(t, pool), newUser(t, pool))
				return chatID, newUser(t, pool)
			},
			want:    false,
			success: true,
		},
		{
			caseName: "check ban with invalid id",
			f: func(t *testing.T) (string, string) {
				return "invalid_uuid", newUser(t, pool)
			},
			success: false,
		},
		{
			caseName: "check ban with canceled context",
			f: func(t *testing.T) (string, string) {
				chatID, _ := seedChat(t, pool, chats)
				return chatID, newUser(t, pool)
			},
			canceledCtx: true,
			success:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			chatID, targetID := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			banned, err := repo.CheckBanUser(ctx, chatID, targetID)
			if !tc.success {
				require.Error(t, err)
				assert.False(t, banned)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, banned)
		})
	}
}
