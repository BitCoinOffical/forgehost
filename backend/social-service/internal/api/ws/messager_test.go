package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ws_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/api/ws/mock"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestSendMessageGroupHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := ws_mocks.NewMockMessageService(ctrl)

	msgs := NewMessageHandler(NewHub(zaptest.NewLogger(t)), srvc, zaptest.NewLogger(t))

	objID := bson.NewObjectID()
	body := `{"text":"hello"}`
	members := []models.FeedChatMember{{UserID: "user-id"}, {UserID: "other-id"}}

	testCases := []msgCase{
		{
			caseName: "send group message",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").Return(members, nil).Times(1)
				srvc.EXPECT().SaveGroupMessage(gomock.Any(), "user-id", "chat-id", gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, req *dto.MessageDTO) (*bson.ObjectID, error) {
						assert.Equal(t, "hello", req.Text)
						return &objID, nil
					}).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data:     msgInput{body: `{invalid`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty json fails validation",
			data:     msgInput{body: `{}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty text",
			data:     msgInput{body: `{"text":""}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "text too long",
			data:     msgInput{body: `{"text":"` + strings.Repeat("a", 9001) + `"}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "get members error",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").
					Return(nil, errors.New("get members error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
		{
			caseName: "sender is not a member",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").
					Return([]models.FeedChatMember{{UserID: "other-id"}}, nil).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "chat without members",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").Return(nil, nil).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "sender is banned",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").Return(members, nil).Times(1)
				srvc.EXPECT().SaveGroupMessage(gomock.Any(), "user-id", "chat-id", gomock.Any()).
					Return(nil, fmt.Errorf("s.repo.CheckBanUser: %w", domain.ErrForbidden)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "save message error",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").Return(members, nil).Times(1)
				srvc.EXPECT().SaveGroupMessage(gomock.Any(), "user-id", "chat-id", gomock.Any()).
					Return(nil, errors.New("save message error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveMsg(t, http.MethodGet, "/ws/message/group/:chat_id", "/ws/message/group/chat-id", tc.data, msgs.SendMessageGroup)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.Contains(t, w.Body.String(), objID.Hex())
			}
		})
	}
}

func TestSendMessageDirectHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := ws_mocks.NewMockMessageService(ctrl)

	msgs := NewMessageHandler(NewHub(zaptest.NewLogger(t)), srvc, zaptest.NewLogger(t))

	objID := bson.NewObjectID()
	body := `{"text":"hello"}`
	members := []models.FeedChatMember{{UserID: "user-id"}, {UserID: "other-id"}}

	testCases := []msgCase{
		{
			caseName: "send direct message",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").Return(members, nil).Times(1)
				srvc.EXPECT().SaveDirectMessage(gomock.Any(), "user-id", "chat-id", gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, req *dto.MessageDTO) (*bson.ObjectID, error) {
						assert.Equal(t, "hello", req.Text)
						return &objID, nil
					}).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data:     msgInput{body: `{invalid`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty json fails validation",
			data:     msgInput{body: `{}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "text too long",
			data:     msgInput{body: `{"text":"` + strings.Repeat("a", 9001) + `"}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "get members error",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").
					Return(nil, errors.New("get members error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
		{
			caseName: "sender is not a participant",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").
					Return([]models.FeedChatMember{{UserID: "other-id"}, {UserID: "third-id"}}, nil).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "participant is blocked",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").Return(members, nil).Times(1)
				srvc.EXPECT().SaveDirectMessage(gomock.Any(), "user-id", "chat-id", gomock.Any()).
					Return(nil, fmt.Errorf("s.repo.CheckBlock: %w", domain.ErrForbidden)).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "save message error",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").Return(members, nil).Times(1)
				srvc.EXPECT().SaveDirectMessage(gomock.Any(), "user-id", "chat-id", gomock.Any()).
					Return(nil, errors.New("save message error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveMsg(t, http.MethodGet, "/ws/message/direct/:chat_id", "/ws/message/direct/chat-id", tc.data, msgs.SendMessageDirect)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.Contains(t, w.Body.String(), objID.Hex())
			}
		})
	}
}

func TestGetChatHistoryHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := ws_mocks.NewMockMessageService(ctrl)

	msgs := NewMessageHandler(NewHub(zaptest.NewLogger(t)), srvc, zaptest.NewLogger(t))

	testCases := []struct {
		caseName string
		prepare  func()
		status   int
		wantBody string
	}{
		{
			caseName: "get history",
			prepare: func() {
				srvc.EXPECT().GetMessages(gomock.Any(), "chat-id").
					Return([]models.Message{{ChatID: "chat-id", UserID: "user-id", Text: "hi"}}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "empty history",
			prepare: func() {
				srvc.EXPECT().GetMessages(gomock.Any(), "chat-id").Return([]models.Message{}, nil).Times(1)
			},
			status:   http.StatusOK,
			wantBody: "[]",
		},
		{
			caseName: "get history error",
			prepare: func() {
				srvc.EXPECT().GetMessages(gomock.Any(), "chat-id").
					Return(nil, errors.New("get history error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveMsg(t, http.MethodGet, "/message/chat/:chat_id", "/message/chat/chat-id", msgInput{}, msgs.GetChatHistory)

			assert.Equal(t, tc.status, w.Code)
			if tc.wantBody != "" {
				assert.Equal(t, tc.wantBody, w.Body.String())
			}
			if tc.caseName == "get history" {
				assert.Contains(t, w.Body.String(), "hi")
			}
		})
	}
}

func TestConnectHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := ws_mocks.NewMockMessageService(ctrl)

	msgs := NewMessageHandler(NewHub(zaptest.NewLogger(t)), srvc, zaptest.NewLogger(t))

	testCases := []msgCase{
		{
			caseName: "user id not found in context",
			data:     msgInput{},
			prepare:  func() {},
			status:   http.StatusUnauthorized,
		},
		{
			caseName: "incorrect user id type",
			data:     msgInput{userID: "not_claims"},
			prepare:  func() {},
			status:   http.StatusUnauthorized,
		},
		{
			caseName: "not a websocket request",
			data:     msgInput{userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusUpgradeRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveMsg(t, http.MethodGet, "/ws/connect", "/ws/connect", tc.data, msgs.Connect)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestSendMessageGroupBroadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := ws_mocks.NewMockMessageService(ctrl)

	msgs := NewMessageHandler(NewHub(zaptest.NewLogger(t)), srvc, zaptest.NewLogger(t))

	r := gin.New()
	r.Use(func(c *gin.Context) {
		withUser(userClaims(c.GetHeader("X-Test-User")))(c)
	})
	r.GET("/ws/connect", msgs.Connect)
	r.GET("/ws/message/group/:chat_id", msgs.SendMessageGroup)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws/connect", &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Test-User": {"receiver-id"}},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.CloseNow())
	})

	time.Sleep(100 * time.Millisecond)

	objID := bson.NewObjectID()
	srvc.EXPECT().GetUsersFromChat(gomock.Any(), "chat-id").
		Return([]models.FeedChatMember{{UserID: "sender-id"}, {UserID: "receiver-id"}}, nil).Times(1)
	srvc.EXPECT().SaveGroupMessage(gomock.Any(), "sender-id", "chat-id", gomock.Any()).
		Return(&objID, nil).Times(1)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/ws/message/group/chat-id",
		strings.NewReader(`{"text":"hello"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", "sender-id")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, resp.Body.Close())
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	_, data, err := conn.Read(ctx)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, objID.Hex(), got["id"])
	assert.Equal(t, "chat-id", got["chat_id"])
	assert.Equal(t, "sender-id", got["from"])
	assert.Equal(t, "hello", got["text"])
}
