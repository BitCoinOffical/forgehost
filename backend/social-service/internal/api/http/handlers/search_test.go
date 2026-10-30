package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers/mock"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockSearchService(ctrl)

	h := NewSearchHandler(srvc, zap.NewNop())
	r := gin.New()
	r.GET("/search", h.Search)

	serviceErr := errors.New("service error")

	postsResp := &models.SearchResponse{
		Posts: []models.SearchPost{{ID: [16]byte{1}}},
	}
	streamsResp := &models.SearchResponse{
		Streams: []models.SearchStream{{ID: "stream_id", Title: "minecraft"}},
	}
	usersResp := &models.SearchResponse{
		Users: []models.SearchUser{{Username: "mike"}},
	}

	testCases := []struct {
		caseName string
		data     struct {
			query string
		}
		prepare func()
		status  int
		body    any
	}{
		{
			caseName: "search posts",
			data:     struct{ query string }{query: "?q=mine&type=posts"},
			prepare: func() {
				srvc.EXPECT().Search(gomock.Any(), "mine", "posts", limit).
					Return(postsResp, nil).Times(1)
			},
			status: http.StatusOK,
			body:   postsResp,
		},
		{
			caseName: "search streams",
			data:     struct{ query string }{query: "?q=mine&type=streams"},
			prepare: func() {
				srvc.EXPECT().Search(gomock.Any(), "mine", "streams", limit).
					Return(streamsResp, nil).Times(1)
			},
			status: http.StatusOK,
			body:   streamsResp,
		},
		{
			caseName: "search users",
			data:     struct{ query string }{query: "?q=mike&type=users"},
			prepare: func() {
				srvc.EXPECT().Search(gomock.Any(), "mike", "users", limit).
					Return(usersResp, nil).Times(1)
			},
			status: http.StatusOK,
			body:   usersResp,
		},
		{
			caseName: "unknown type",
			data:     struct{ query string }{query: "?q=mine&type=comments"},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "type is missing",
			data:     struct{ query string }{query: "?q=mine"},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "service error",
			data:     struct{ query string }{query: "?q=mine&type=posts"},
			prepare: func() {
				srvc.EXPECT().Search(gomock.Any(), "mine", "posts", limit).
					Return(nil, serviceErr).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := httptest.NewRequest(http.MethodGet, "/search"+tc.data.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, tc.status, w.Code)
			if tc.body != nil {
				assert.JSONEq(t, mustJSON(t, tc.body), w.Body.String())
			}
		})
	}
}

func TestSuggest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockSearchService(ctrl)

	h := NewSearchHandler(srvc, zaptest.NewLogger(t))

	serviceErr := errors.New("service error")

	suggestions := []models.Suggestion{
		{Type: "user", ID: "user_id", Text: "mike"},
		{Type: "stream", ID: "stream_id", Text: "minecraft"},
	}
	testSearchUserID := "11111111-1111-1111-1111-111111111111"
	testCases := []struct {
		caseName string
		data     struct {
			userID string
			query  string
		}
		prepare func()
		status  int
		body    any
	}{
		{
			caseName: "suggest success",
			data: struct {
				userID string
				query  string
			}{userID: testSearchUserID, query: "?q=mi"},
			prepare: func() {
				srvc.EXPECT().Suggest(gomock.Any(), testSearchUserID, "mi", limit).
					Return(suggestions, nil).Times(1)
			},
			status: http.StatusOK,
			body:   suggestions,
		},
		{
			caseName: "no suggestions",
			data: struct {
				userID string
				query  string
			}{userID: testSearchUserID, query: "?q=zzz"},
			prepare: func() {
				srvc.EXPECT().Suggest(gomock.Any(), testSearchUserID, "zzz", limit).
					Return([]models.Suggestion{}, nil).Times(1)
			},
			status: http.StatusOK,
			body:   []models.Suggestion{},
		},
		{
			caseName: "user is not authorized",
			data: struct {
				userID string
				query  string
			}{userID: "", query: "?q=mi"},
			prepare: func() {},
			status:  http.StatusUnauthorized,
		},
		{
			caseName: "service error",
			data: struct {
				userID string
				query  string
			}{userID: testSearchUserID, query: "?q=mi"},
			prepare: func() {
				srvc.EXPECT().Suggest(gomock.Any(), testSearchUserID, "mi", limit).
					Return(nil, serviceErr).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			r := gin.New()
			if tc.data.userID != "" {
				r.Use(withUserID(tc.data.userID))
			}
			r.GET("/search/suggest", h.Suggest)

			req := httptest.NewRequest(http.MethodGet, "/search/suggest"+tc.data.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, tc.status, w.Code)
			if tc.body != nil {
				assert.JSONEq(t, mustJSON(t, tc.body), w.Body.String())
			}
		})
	}
}
