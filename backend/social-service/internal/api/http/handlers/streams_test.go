package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers/mock"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
	"golang.org/x/oauth2"
)

const testVideoID = "video-id"

func newStreamOauthConfig(tokenURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "http://localhost:8080/api/v1/social/youtube/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/youtube.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.test/auth",
			TokenURL: tokenURL,
		},
	}
}

func TestStreamGoogleLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockStreamService(ctrl)

	stream := NewStreamHandler(zaptest.NewLogger(t), newStreamOauthConfig("https://oauth.test/token"), srvc)

	testCases := []chatCase{
		{
			caseName: "google login",
			data:     chatInput{},
			prepare:  func() {},
			status:   http.StatusTemporaryRedirect,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/youtube/connect", "/youtube/connect", tc.data, stream.GoogleLogin)

			assert.Equal(t, tc.status, w.Code)

			var state string
			for _, cookie := range w.Result().Cookies() {
				if cookie.Name == "oauth_state" {
					state = cookie.Value
					assert.True(t, cookie.HttpOnly)
					assert.True(t, cookie.Secure)
				}
			}
			require.NotEmpty(t, state)
			assert.Contains(t, w.Header().Get("Location"), "https://accounts.test/auth")
			assert.Contains(t, w.Header().Get("Location"), "state="+state)
		})
	}
}

func TestStreamGoogleCallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockStreamService(ctrl)

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer tokenServer.Close()

	stream := NewStreamHandler(zaptest.NewLogger(t), newStreamOauthConfig(tokenServer.URL), srvc)

	testCases := []struct {
		caseName string
		cookie   string
		query    string
		cleared  bool
		status   int
	}{
		{
			caseName: "missing state cookie",
			cookie:   "",
			query:    "state=state&code=code",
			status:   http.StatusBadRequest,
		},
		{
			caseName: "state mismatch",
			cookie:   "state",
			query:    "state=other&code=code",
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty code",
			cookie:   "state",
			query:    "state=state",
			cleared:  true,
			status:   http.StatusBadRequest,
		},
		{
			caseName: "exchange code error",
			cookie:   "state",
			query:    "state=state&code=code",
			cleared:  true,
			status:   http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			r := gin.New()
			r.Use(gin.Recovery())
			r.GET("/youtube/callback", stream.GoogleCallback)

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/youtube/callback?"+tc.query, nil)
			require.NoError(t, err)
			if tc.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "oauth_state", Value: tc.cookie})
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)

			cleared := false
			for _, cookie := range w.Result().Cookies() {
				if cookie.Name == "oauth_state" && cookie.MaxAge < 0 {
					cleared = true
				}
			}
			assert.Equal(t, tc.cleared, cleared)
		})
	}
}

func TestGetChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockStreamService(ctrl)

	stream := NewStreamHandler(zaptest.NewLogger(t), nil, srvc)

	testCases := []chatCase{
		{
			caseName: "get channel",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetChannel(gomock.Any(), testUserID).
					Return("channel-id", nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "channel not linked",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetChannel(gomock.Any(), testUserID).
					Return("", fmt.Errorf("s.repo.GetChannel: %w", domain.ErrChannelNotLinked)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "profile not found",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetChannel(gomock.Any(), testUserID).
					Return("", fmt.Errorf("s.repo.GetChannel: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "get channel error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetChannel(gomock.Any(), testUserID).
					Return("", errors.New("get channel error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/youtube/channel", "/youtube/channel", tc.data, stream.GetChannel)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestSaveStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockStreamService(ctrl)

	stream := NewStreamHandler(zaptest.NewLogger(t), nil, srvc)

	body := `
	{
	"videoId":"video-id",
	"title":"title",
	"lifeCycleStatus":"live",
	"embedUrl":"https://www.youtube.com/embed/1" 
	}`
	reqMatcher := gomock.Cond(func(r *dto.StreamDTO) bool {
		return r.VideoID == testVideoID && r.Title == "title"
	})

	testCases := []chatCase{
		{
			caseName: "save stream",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().SaveStream(gomock.Any(), testUserID, reqMatcher).
					Return("stream-id", nil).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{userID: userClaims(testUserID), body: `{invalid`},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "save existing stream",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().SaveStream(gomock.Any(), testUserID, gomock.Any()).
					Return("", fmt.Errorf("s.repo.SaveStream: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "save stream error",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().SaveStream(gomock.Any(), testUserID, gomock.Any()).
					Return("", errors.New("save stream error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, "/streams", "/streams", tc.data, stream.SaveStream)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestBuildUpdateStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockStreamService(ctrl)

	stream := NewStreamHandler(zaptest.NewLogger(t), nil, srvc)

	body := `{"title":"new_title"}`
	reqMatcher := gomock.Cond(func(r *dto.UpdateStreamDTO) bool {
		return r.Title != nil && *r.Title == "new_title"
	})

	testCases := []chatCase{
		{
			caseName: "update stream",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().BuildUpdateStream(gomock.Any(), testUserID, testVideoID, reqMatcher).
					Return(&models.Stream{ID: "stream-id", VideoID: testVideoID}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{userID: userClaims(testUserID), body: `{invalid`},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "update not found stream",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().BuildUpdateStream(gomock.Any(), testUserID, testVideoID, gomock.Any()).
					Return(nil, fmt.Errorf("s.repo.BuildUpdateStream: %w", domain.ErrStreamNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "update stream error",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().BuildUpdateStream(gomock.Any(), testUserID, testVideoID, gomock.Any()).
					Return(nil, errors.New("update stream error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPatch, "/streams/:video_id", "/streams/"+testVideoID, tc.data, stream.BuildUpdateStream)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestDeleteStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockStreamService(ctrl)

	stream := NewStreamHandler(zaptest.NewLogger(t), nil, srvc)

	testCases := []chatCase{
		{
			caseName: "delete stream",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().DeleteStream(gomock.Any(), testVideoID, testUserID).
					Return(nil).Times(1)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "delete not found stream",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().DeleteStream(gomock.Any(), testVideoID, testUserID).
					Return(fmt.Errorf("s.repo.DeleteStream: %w", domain.ErrStreamNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "delete stream error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().DeleteStream(gomock.Any(), testVideoID, testUserID).
					Return(errors.New("delete stream error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodDelete, "/streams/:video_id", "/streams/"+testVideoID, tc.data, stream.DeleteStream)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestGetStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockStreamService(ctrl)

	stream := NewStreamHandler(zaptest.NewLogger(t), nil, srvc)

	testCases := []chatCase{
		{
			caseName: "get stream",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetStream(gomock.Any(), testVideoID, testUserID).
					Return(&models.Stream{ID: "stream-id", VideoID: testVideoID}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "get not found stream",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetStream(gomock.Any(), testVideoID, testUserID).
					Return(nil, fmt.Errorf("s.repo.GetStream: %w", domain.ErrStreamNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "get stream error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetStream(gomock.Any(), testVideoID, testUserID).
					Return(nil, errors.New("get stream error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/streams/:video_id", "/streams/"+testVideoID, tc.data, stream.GetStream)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestGetCursorStreams(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockStreamService(ctrl)

	stream := NewStreamHandler(zaptest.NewLogger(t), nil, srvc)

	testCases := []chatCase{
		{
			caseName: "get cursor streams",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetCursorStreams(gomock.Any(), "cursor-token").
					Return([]models.FeedStream{{ID: "1"}, {ID: "2"}}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid cursor",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetCursorStreams(gomock.Any(), "cursor-token").
					Return(nil, fmt.Errorf("base64.URLEncoding.DecodeString: %w", domain.ErrInvalidCursor)).Times(1)
			},
			status: http.StatusBadRequest,
		},
		{
			caseName: "get cursor streams error",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetCursorStreams(gomock.Any(), "cursor-token").
					Return(nil, errors.New("get cursor streams error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/streams/list/:cursor", "/streams/list/cursor-token", tc.data, stream.GetCursorStreams)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestGetCursorStreamsFirstPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockStreamService(ctrl)

	stream := NewStreamHandler(zaptest.NewLogger(t), nil, srvc)

	testCases := []chatCase{
		{
			caseName: "get first page",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetCursorStreams(gomock.Any(), "").
					Return([]models.FeedStream{{ID: "1"}}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "get first page error",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetCursorStreams(gomock.Any(), "").
					Return(nil, errors.New("get first page error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/streams/list", "/streams/list", tc.data, stream.GetCursorStreams)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}
