package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ws_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/api/ws/mock"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestSendStreamMessageHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := ws_mocks.NewMockStreamService(ctrl)

	streams := NewStreamHandler(NewHub(zaptest.NewLogger(t)), srvc, zaptest.NewLogger(t))

	body := `{"text":"hello"}`

	testCases := []msgCase{
		{
			caseName: "send stream message",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetStreamUsers(gomock.Any(), "video-id").
					Return([]string{"user-id", "other-id"}, nil).Times(1)
				srvc.EXPECT().SaveStreamMessage(gomock.Any(), "user-id", "video-id", gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, req *dto.MessageDTO) error {
						assert.Equal(t, "hello", req.Text)
						return nil
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
			caseName: "get stream users error",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetStreamUsers(gomock.Any(), "video-id").
					Return(nil, errors.New("get stream users error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
		{
			caseName: "sender is not in stream",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetStreamUsers(gomock.Any(), "video-id").
					Return([]string{"other-id"}, nil).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "stream without users",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetStreamUsers(gomock.Any(), "video-id").Return(nil, nil).Times(1)
			},
			status: http.StatusForbidden,
		},
		{
			caseName: "save stream message error",
			data:     msgInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetStreamUsers(gomock.Any(), "video-id").
					Return([]string{"user-id"}, nil).Times(1)
				srvc.EXPECT().SaveStreamMessage(gomock.Any(), "user-id", "video-id", gomock.Any()).
					Return(errors.New("save stream message error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveMsg(t, http.MethodGet, "/ws/message/stream/:video_id", "/ws/message/stream/video-id", tc.data, streams.SendStreamMessage)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestGetStreamHistoryHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := ws_mocks.NewMockStreamService(ctrl)

	streams := NewStreamHandler(NewHub(zaptest.NewLogger(t)), srvc, zaptest.NewLogger(t))

	testCases := []struct {
		caseName string
		prepare  func()
		status   int
		wantBody string
	}{
		{
			caseName: "get stream history",
			prepare: func() {
				srvc.EXPECT().GetStreamHistory(gomock.Any(), "video-id").
					Return([]models.Message{{ChatID: "video-id", UserID: "user-id", Text: "hi"}}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "empty history",
			prepare: func() {
				srvc.EXPECT().GetStreamHistory(gomock.Any(), "video-id").Return([]models.Message{}, nil).Times(1)
			},
			status:   http.StatusOK,
			wantBody: "[]",
		},
		{
			caseName: "get stream history error",
			prepare: func() {
				srvc.EXPECT().GetStreamHistory(gomock.Any(), "video-id").
					Return(nil, errors.New("get stream history error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveMsg(t, http.MethodGet, "/message/stream/:video_id", "/message/stream/video-id", msgInput{}, streams.GetStreamHistory)

			assert.Equal(t, tc.status, w.Code)
			if tc.wantBody != "" {
				assert.Equal(t, tc.wantBody, w.Body.String())
			}
			if tc.caseName == "get stream history" {
				assert.Contains(t, w.Body.String(), "hi")
			}
		})
	}
}

func TestSaveStreamUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := ws_mocks.NewMockStreamService(ctrl)

	streams := NewStreamHandler(NewHub(zaptest.NewLogger(t)), srvc, zaptest.NewLogger(t))

	testCases := []msgCase{
		{
			caseName: "save stream user",
			data:     msgInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().SaveStreamUser(gomock.Any(), "user-id", "video-id").Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "save stream user error",
			data:     msgInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().SaveStreamUser(gomock.Any(), "user-id", "video-id").
					Return(errors.New("save stream user error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveMsg(t, http.MethodPost, "/streams/:video_id/join", "/streams/video-id/join", tc.data, streams.SaveStreamUser)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestSendStreamMessageBroadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := ws_mocks.NewMockStreamService(ctrl)

	hub := NewHub(zaptest.NewLogger(t))
	messages := NewMessageHandler(hub, ws_mocks.NewMockMessageService(ctrl), zaptest.NewLogger(t))
	streams := NewStreamHandler(hub, srvc, zaptest.NewLogger(t))

	r := gin.New()
	r.Use(func(c *gin.Context) {
		withUser(userClaims(c.GetHeader("X-Test-User")))(c)
	})
	r.GET("/ws/connect", messages.Connect)
	r.GET("/ws/message/stream/:video_id", streams.SendStreamMessage)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws/connect", &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Test-User": {"viewer-id"}},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.CloseNow())
	})

	time.Sleep(100 * time.Millisecond)

	srvc.EXPECT().GetStreamUsers(gomock.Any(), "video-id").
		Return([]string{"sender-id", "viewer-id"}, nil).Times(1)
	srvc.EXPECT().SaveStreamMessage(gomock.Any(), "sender-id", "video-id", gomock.Any()).
		Return(nil).Times(1)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/ws/message/stream/video-id",
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
	assert.Equal(t, "video-id", got["video_id"])
	assert.Equal(t, "sender-id", got["from"])
	assert.Equal(t, "hello", got["text"])
}
