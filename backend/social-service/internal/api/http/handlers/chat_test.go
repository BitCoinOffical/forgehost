package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers/mock"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestCreateGroupHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "create group",
			data:     chatInput{body: `{"title":"group","description":"desc"}`, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().CreateGroup(gomock.Any(), testUserID, gomock.Any()).
					Return("new-chat-id", nil).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty body fails validation",
			data:     chatInput{body: `{}`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "create group error",
			data:     chatInput{body: `{"title":"group"}`, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().CreateGroup(gomock.Any(), testUserID, gomock.Any()).
					Return("", errors.New("create group error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(`{"title":"group"}`)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, "/chats", "/chats", tc.data, chat.CreateGroup)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusCreated {
				assert.Contains(t, w.Body.String(), "new-chat-id")
			}
		})
	}
}

func TestCreateDirectHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/directs/:user_id", "/chats/directs/"+testTargetID

	testCases := []chatCase{
		{
			caseName: "create direct",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().CreateDirect(gomock.Any(), testUserID, testTargetID).
					Return("direct-id", nil).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "blocked user",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().CreateDirect(gomock.Any(), testUserID, testTargetID).
					Return("", fmt.Errorf("s.repo.CheckBlock: %w", domain.ErrForbidden)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "create direct error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().CreateDirect(gomock.Any(), testUserID, testTargetID).
					Return("", errors.New("create direct error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, route, url, tc.data, chat.CreateDirect)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusCreated {
				assert.Contains(t, w.Body.String(), "direct-id")
			}
		})
	}
}

func TestGetChatListHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "get chat list",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetChatList(gomock.Any(), testUserID).
					Return([]models.FeedDirects{}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "get chat list error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetChatList(gomock.Any(), testUserID).
					Return(nil, errors.New("get chat list error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/chats/list", "/chats/list", tc.data, chat.GetChatList)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestMarkAsReadHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id/read", "/chats/"+testChatID+"/read"

	testCases := []chatCase{
		{
			caseName: "mark as read",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().MarkAsRead(gomock.Any(), testChatID, testUserID).Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "chat not found",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().MarkAsRead(gomock.Any(), testChatID, testUserID).
					Return(fmt.Errorf("s.repo.MarkAsRead: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "mark as read error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().MarkAsRead(gomock.Any(), testChatID, testUserID).
					Return(errors.New("mark as read error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, route, url, tc.data, chat.MarkAsRead)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestGetChatByIDHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id", "/chats/"+testChatID

	testCases := []chatCase{
		{
			caseName: "get chat by id",
			prepare: func() {
				srvc.EXPECT().GetChatByID(gomock.Any(), testChatID).Return(&models.Chat{}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "chat not found",
			prepare: func() {
				srvc.EXPECT().GetChatByID(gomock.Any(), testChatID).
					Return(nil, fmt.Errorf("s.repo.GetChatByID: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "get chat by id error",
			prepare: func() {
				srvc.EXPECT().GetChatByID(gomock.Any(), testChatID).
					Return(nil, errors.New("get chat error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, route, url, tc.data, chat.GetChatByID)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestDeleteChatHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id", "/chats/"+testChatID

	testCases := []chatCase{
		{
			caseName: "delete chat",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().DeleteChat(gomock.Any(), testChatID, testUserID).Return(nil).Times(1)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "chat not found",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().DeleteChat(gomock.Any(), testChatID, testUserID).
					Return(fmt.Errorf("s.repo.DeleteChat: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "not owner",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().DeleteChat(gomock.Any(), testChatID, testUserID).
					Return(fmt.Errorf("s.repo.DeleteChat: %w", domain.ErrForbidden)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "delete chat error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().DeleteChat(gomock.Any(), testChatID, testUserID).
					Return(errors.New("delete chat error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodDelete, route, url, tc.data, chat.DeleteChat)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestLeaveChatHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id/leave", "/chats/"+testChatID+"/leave"

	testCases := []chatCase{
		{
			caseName: "leave chat",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().LeaveChat(gomock.Any(), testChatID, testUserID).Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "chat not found",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().LeaveChat(gomock.Any(), testChatID, testUserID).
					Return(fmt.Errorf("s.repo.LeaveChat: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "leave chat error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().LeaveChat(gomock.Any(), testChatID, testUserID).
					Return(errors.New("leave chat error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, route, url, tc.data, chat.LeaveChat)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUpdateChatHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id", "/chats/"+testChatID
	body := `{"title":"new title"}`

	testCases := []chatCase{
		{
			caseName: "update chat",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().BuildUpdateChat(gomock.Any(), testChatID, testUserID, gomock.Any()).
					Return(&models.Chat{}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "not enough rights",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().BuildUpdateChat(gomock.Any(), testChatID, testUserID, gomock.Any()).
					Return(nil, fmt.Errorf("s.repo.BuildUpdateChat: %w", domain.ErrForbidden)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "chat not found",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().BuildUpdateChat(gomock.Any(), testChatID, testUserID, gomock.Any()).
					Return(nil, fmt.Errorf("s.repo.BuildUpdateChat: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "update chat error",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().BuildUpdateChat(gomock.Any(), testChatID, testUserID, gomock.Any()).
					Return(nil, errors.New("update chat error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPatch, route, url, tc.data, chat.UpdateChat)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestSetUserRoleChatHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id/members/:user_id/role", "/chats/"+testChatID+"/members/"+testTargetID+"/role"
	body := `{"user_id":"target-id","role":"moderator"}`

	testCases := []chatCase{
		{
			caseName: "set user role",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().SetUserRoleChat(gomock.Any(), testChatID, testUserID, gomock.Any()).
					Return(&models.ChatMember{}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty body fails validation",
			data:     chatInput{body: `{}`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "not enough rights",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().SetUserRoleChat(gomock.Any(), testChatID, testUserID, gomock.Any()).
					Return(nil, fmt.Errorf("s.repo.SetUserRoleChat: %w", domain.ErrForbidden)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "set user role error",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().SetUserRoleChat(gomock.Any(), testChatID, testUserID, gomock.Any()).
					Return(nil, errors.New("set role error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPatch, route, url, tc.data, chat.SetUserRoleChat)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestJoinInChatHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id/members", "/chats/"+testChatID+"/members"

	testCases := []chatCase{
		{
			caseName: "join in chat",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().AddUserInChat(gomock.Any(), testChatID, testUserID).Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "user is banned",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().AddUserInChat(gomock.Any(), testChatID, testUserID).
					Return(fmt.Errorf("s.repo.CheckBanUser: %w", domain.ErrForbidden)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "already a member",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().AddUserInChat(gomock.Any(), testChatID, testUserID).
					Return(fmt.Errorf("s.repo.AddUserInChat: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "join in chat error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().AddUserInChat(gomock.Any(), testChatID, testUserID).
					Return(errors.New("join error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, route, url, tc.data, chat.JoinInChat)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestKickUserFromChatHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id/members/:user_id", "/chats/"+testChatID+"/members/"+testTargetID

	testCases := []chatCase{
		{
			caseName: "kick user",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().KickUserFromChat(gomock.Any(), testChatID, testTargetID, testUserID).
					Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "not enough rights",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().KickUserFromChat(gomock.Any(), testChatID, testTargetID, testUserID).
					Return(fmt.Errorf("s.repo.KickUserFromChat: %w", domain.ErrForbidden)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "cannot kick owner",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().KickUserFromChat(gomock.Any(), testChatID, testTargetID, testUserID).
					Return(fmt.Errorf("s.repo.KickUserFromChat: %w", domain.ErrCannotKickOwner)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "kick user error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().KickUserFromChat(gomock.Any(), testChatID, testTargetID, testUserID).
					Return(errors.New("kick error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodDelete, route, url, tc.data, chat.KickUserFromChat)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestGetUsersFromChatHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id/members", "/chats/"+testChatID+"/members"

	testCases := []chatCase{
		{
			caseName: "get users from chat",
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), testChatID).
					Return([]models.FeedChatMember{}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "get users from chat error",
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), testChatID).
					Return(nil, errors.New("get users error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, route, url, tc.data, chat.GetUsersFromChat)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestBanUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id/bans/:user_id", "/chats/"+testChatID+"/bans/"+testTargetID

	testCases := []chatCase{
		{
			caseName: "ban user",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().BanUser(gomock.Any(), testChatID, testTargetID, testUserID).Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "not enough rights",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().BanUser(gomock.Any(), testChatID, testTargetID, testUserID).
					Return(fmt.Errorf("s.repo.BanUser: %w", domain.ErrForbidden)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "cannot ban owner",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().BanUser(gomock.Any(), testChatID, testTargetID, testUserID).
					Return(fmt.Errorf("s.repo.BanUser: %w", domain.ErrCannotKickOwner)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "ban user error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().BanUser(gomock.Any(), testChatID, testTargetID, testUserID).
					Return(errors.New("ban error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, route, url, tc.data, chat.BanUser)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUnbanUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id/bans/:user_id", "/chats/"+testChatID+"/bans/"+testTargetID

	testCases := []chatCase{
		{
			caseName: "unban user",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UnbanUser(gomock.Any(), testChatID, testTargetID, testUserID).Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "not enough rights",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UnbanUser(gomock.Any(), testChatID, testTargetID, testUserID).
					Return(fmt.Errorf("s.repo.UnbanUser: %w", domain.ErrForbidden)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "unban user error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UnbanUser(gomock.Any(), testChatID, testTargetID, testUserID).
					Return(errors.New("unban error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodDelete, route, url, tc.data, chat.UnbanUser)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestCheckBanUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockChatService(ctrl)

	chat := NewChatHandler(srvc, zaptest.NewLogger(t))
	route, url := "/chats/:chat_id/bans/:user_id", "/chats/"+testChatID+"/bans/"+testTargetID

	testCases := []struct {
		caseName string
		prepare  func()
		status   int
		wantBody string
	}{
		{
			caseName: "user is banned",
			prepare: func() {
				srvc.EXPECT().CheckBanUser(gomock.Any(), testChatID, testTargetID).Return(true, nil).Times(1)
			},
			status:   http.StatusOK,
			wantBody: "true",
		},
		{
			caseName: "user is not banned",
			prepare: func() {
				srvc.EXPECT().CheckBanUser(gomock.Any(), testChatID, testTargetID).Return(false, nil).Times(1)
			},
			status:   http.StatusOK,
			wantBody: "false",
		},
		{
			caseName: "check ban error",
			prepare: func() {
				srvc.EXPECT().CheckBanUser(gomock.Any(), testChatID, testTargetID).
					Return(false, errors.New("check ban error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, route, url, chatInput{}, chat.CheckBanUser)

			assert.Equal(t, tc.status, w.Code)
			if tc.wantBody != "" {
				assert.Equal(t, tc.wantBody, w.Body.String())
			}
		})
	}
}
