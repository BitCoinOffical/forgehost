package services

import (
	"errors"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	store_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/mock/gomock"
)

func messageMatcher(chatID, userID, text string) gomock.Matcher {
	return gomock.Cond(func(m *models.Message) bool {
		if chatID != "" && m.ChatID != chatID {
			return false
		}
		return m.UserID == userID && m.Text == text && time.Since(m.CreatedAt) < time.Minute
	})
}

func TestSaveDirectMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockMessagesRepo(ctrl)
	mrepo := store_mocks.NewMockMessagerRepo(ctrl)

	message := NewMessageService(repo, mrepo)

	oid := bson.NewObjectID()

	testCases := []struct {
		caseName string
		data     struct {
			userID   string
			targetID string
			text     string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "save direct message",
			data: struct {
				userID   string
				targetID string
				text     string
			}{userID: "user_id", targetID: "target_id", text: "text"},
			prepare: func() {
				repo.EXPECT().CheckBlock(gomock.Any(), "user_id", "target_id").
					Return(false, nil).Times(1)
				mrepo.EXPECT().SaveMessage(gomock.Any(), messageMatcher("", "user_id", "text")).
					Return(&oid, nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "save direct message to blocked user",
			data: struct {
				userID   string
				targetID string
				text     string
			}{userID: "user_id", targetID: "target_id", text: "text"},
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
				text     string
			}{userID: "user_id", targetID: "target_id", text: "text"},
			prepare: func() {
				repo.EXPECT().CheckBlock(gomock.Any(), "user_id", "target_id").
					Return(false, errors.New("check block error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "save message error",
			data: struct {
				userID   string
				targetID string
				text     string
			}{userID: "user_id", targetID: "target_id", text: "text"},
			prepare: func() {
				repo.EXPECT().CheckBlock(gomock.Any(), "user_id", "target_id").
					Return(false, nil).Times(1)
				mrepo.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("save message error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.MessageDTO{
				Text: tc.data.text,
			}
			got, err := message.SaveDirectMessage(t.Context(), tc.data.userID, tc.data.targetID, &req)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, &oid, got)
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

func TestSaveGroupMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockMessagesRepo(ctrl)
	mrepo := store_mocks.NewMockMessagerRepo(ctrl)

	message := NewMessageService(repo, mrepo)

	oid := bson.NewObjectID()

	testCases := []struct {
		caseName string
		data     struct {
			userID string
			chatID string
			text   string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "save group message",
			data: struct {
				userID string
				chatID string
				text   string
			}{userID: "user_id", chatID: "chat_id", text: "text"},
			prepare: func() {
				repo.EXPECT().CheckBanUser(gomock.Any(), "chat_id", "user_id").
					Return(false, nil).Times(1)
				mrepo.EXPECT().SaveMessage(gomock.Any(), messageMatcher("chat_id", "user_id", "text")).
					Return(&oid, nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "save group message by banned user",
			data: struct {
				userID string
				chatID string
				text   string
			}{userID: "user_id", chatID: "chat_id", text: "text"},
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
				userID string
				chatID string
				text   string
			}{userID: "user_id", chatID: "chat_id", text: "text"},
			prepare: func() {
				repo.EXPECT().CheckBanUser(gomock.Any(), "chat_id", "user_id").
					Return(false, errors.New("check ban user error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "save message error",
			data: struct {
				userID string
				chatID string
				text   string
			}{userID: "user_id", chatID: "chat_id", text: "text"},
			prepare: func() {
				repo.EXPECT().CheckBanUser(gomock.Any(), "chat_id", "user_id").
					Return(false, nil).Times(1)
				mrepo.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("save message error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.MessageDTO{
				Text: tc.data.text,
			}
			got, err := message.SaveGroupMessage(t.Context(), tc.data.userID, tc.data.chatID, &req)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, &oid, got)
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

func TestGetMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockMessagesRepo(ctrl)
	mrepo := store_mocks.NewMockMessagerRepo(ctrl)

	message := NewMessageService(repo, mrepo)

	testCases := []struct {
		caseName string
		data     struct {
			chatID string
		}
		prepare func()
		want    []models.Message
		success bool
	}{
		{
			caseName: "get messages",
			data: struct {
				chatID string
			}{chatID: "chat_id"},
			prepare: func() {
				mrepo.EXPECT().GetMessages(gomock.Any(), "chat_id").
					Return([]models.Message{{Text: "first"}, {Text: "second"}}, nil).Times(1)
			},
			want:    []models.Message{{Text: "first"}, {Text: "second"}},
			success: true,
		},
		{
			caseName: "get messages from empty chat",
			data: struct {
				chatID string
			}{chatID: "chat_id"},
			prepare: func() {
				mrepo.EXPECT().GetMessages(gomock.Any(), "chat_id").
					Return(nil, nil).Times(1)
			},
			want:    nil,
			success: true,
		},
		{
			caseName: "get messages error",
			data: struct {
				chatID string
			}{chatID: "chat_id"},
			prepare: func() {
				mrepo.EXPECT().GetMessages(gomock.Any(), "chat_id").
					Return(nil, errors.New("get messages error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := message.GetMessages(t.Context(), tc.data.chatID)
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

func TestMessageGetUsersFromChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockMessagesRepo(ctrl)
	mrepo := store_mocks.NewMockMessagerRepo(ctrl)

	message := NewMessageService(repo, mrepo)

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

			got, err := message.GetUsersFromChat(t.Context(), tc.data.chatID)
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
