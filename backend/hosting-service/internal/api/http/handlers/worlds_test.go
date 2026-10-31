package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	srvc_mocks "github.com/BitCoinOffical/forgehost/hosting-service/internal/api/http/handlers/mock"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	jwtpkg "github.com/BitCoinOffical/forgehost/hosting-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

const (
	testUserID  = "user-id"
	testWorldID = "world-id"
)

func newWorldsHandlerRouter(t *testing.T, srvc *srvc_mocks.MockWorldsService, userID string) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()

	if userID != "" {
		r.Use(func(c *gin.Context) {
			c.Set("user_claims", &jwtpkg.Claims{
				UserID: userID,
			})
			c.Next()
		})
	}

	handler := NewWorldsHandler(srvc, zaptest.NewLogger(t))

	r.POST("/minecraft", handler.Create)
	r.GET("/minecraft", handler.GetAll)
	r.GET("/minecraft/:world_id", handler.GetByID)
	r.DELETE("/minecraft/:world_id", handler.Delete)

	return r
}

func performWorldsRequest(r http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func worldsHandlerTestModel() *models.MinecraftServer {
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)

	return &models.MinecraftServer{
		ID:        testWorldID,
		UserID:    testUserID,
		Name:      "Test world",
		Address:   "mc.test.com",
		Port:      25565,
		Edition:   "java",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestCreate(t *testing.T) {
	reqBody := `{
		"name": "Test world",
		"address": "mc.test.com",
		"port": 25565,
		"edition": "java"
	}`

	wantReq := dto.UpsertMinecraftServerDTO{
		Name:    "Test world",
		Address: "mc.test.com",
		Port:    25565,
		Edition: "java",
	}

	wantWorld := worldsHandlerTestModel()

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		userID  string
		prepare func(*testing.T, *srvc_mocks.MockWorldsService)
		status  int
		want    *models.MinecraftServer
	}{
		{
			caseName: "create world",
			data: struct {
				body string
			}{body: reqBody},
			userID: testUserID,
			prepare: func(t *testing.T, srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), testUserID).
					DoAndReturn(func(
						_ context.Context,
						req *dto.UpsertMinecraftServerDTO,
						userID string,
					) (*models.MinecraftServer, error) {
						assert.Equal(t, wantReq, *req)
						assert.Equal(t, testUserID, userID)

						return wantWorld, nil
					})
			},
			status: http.StatusOK,
			want:   wantWorld,
		},
		{
			caseName: "missing user id",
			data: struct {
				body string
			}{body: reqBody},
			status: http.StatusUnauthorized,
		},
		{
			caseName: "invalid request body",
			data: struct {
				body string
			}{body: `{"name":`},
			userID: testUserID,
			status: http.StatusBadRequest,
		},
		{
			caseName: "world already exists",
			data: struct {
				body string
			}{body: reqBody},
			userID: testUserID,
			prepare: func(t *testing.T, srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), testUserID).
					Return(nil, domain.ErrAlreadyExists)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "service error",
			data: struct {
				body string
			}{body: reqBody},
			userID: testUserID,
			prepare: func(t *testing.T, srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), testUserID).
					Return(nil, errors.New("database error"))
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockWorldsService(ctrl)

			if tt.prepare != nil {
				tt.prepare(t, srvc)
			}

			r := newWorldsHandlerRouter(t, srvc, tt.userID)
			rec := performWorldsRequest(r, http.MethodPost, "/minecraft", tt.data.body)

			assert.Equal(t, tt.status, rec.Code)

			if tt.want != nil {
				expected, err := json.Marshal(tt.want)
				require.NoError(t, err)

				assert.JSONEq(t, string(expected), rec.Body.String())
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	want := []models.MinecraftServer{
		*worldsHandlerTestModel(),
	}

	testCases := []struct {
		caseName string
		prepare  func(*srvc_mocks.MockWorldsService)
		status   int
		want     []models.MinecraftServer
	}{
		{
			caseName: "get all worlds",
			prepare: func(srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					GetAll(gomock.Any()).
					Return(want, nil)
			},
			status: http.StatusOK,
			want:   want,
		},
		{
			caseName: "empty worlds",
			prepare: func(srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					GetAll(gomock.Any()).
					Return([]models.MinecraftServer{}, nil)
			},
			status: http.StatusOK,
			want:   []models.MinecraftServer{},
		},
		{
			caseName: "service error",
			prepare: func(srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					GetAll(gomock.Any()).
					Return(nil, errors.New("redis error"))
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockWorldsService(ctrl)

			if tt.prepare != nil {
				tt.prepare(srvc)
			}

			r := newWorldsHandlerRouter(t, srvc, "")
			rec := performWorldsRequest(r, http.MethodGet, "/minecraft", "")

			assert.Equal(t, tt.status, rec.Code)

			if tt.want != nil {
				expected, err := json.Marshal(tt.want)
				require.NoError(t, err)

				assert.JSONEq(t, string(expected), rec.Body.String())
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	want := worldsHandlerTestModel()

	testCases := []struct {
		caseName string
		data     struct {
			worldID string
		}
		prepare func(*srvc_mocks.MockWorldsService)
		status  int
		want    *models.MinecraftServer
	}{
		{
			caseName: "get world",
			data: struct {
				worldID string
			}{worldID: testWorldID},
			prepare: func(srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					GetByID(gomock.Any(), testWorldID).
					Return(want, nil)
			},
			status: http.StatusOK,
			want:   want,
		},
		{
			caseName: "world not found",
			data: struct {
				worldID string
			}{worldID: testWorldID},
			prepare: func(srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					GetByID(gomock.Any(), testWorldID).
					Return(nil, domain.ErrNotFound)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "service error",
			data: struct {
				worldID string
			}{worldID: testWorldID},
			prepare: func(srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					GetByID(gomock.Any(), testWorldID).
					Return(nil, errors.New("redis error"))
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockWorldsService(ctrl)

			if tt.prepare != nil {
				tt.prepare(srvc)
			}

			r := newWorldsHandlerRouter(t, srvc, "")
			path := "/minecraft/" + tt.data.worldID

			rec := performWorldsRequest(r, http.MethodGet, path, "")

			assert.Equal(t, tt.status, rec.Code)

			if tt.want != nil {
				expected, err := json.Marshal(tt.want)
				require.NoError(t, err)

				assert.JSONEq(t, string(expected), rec.Body.String())
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		caseName string
		userID   string
		prepare  func(*srvc_mocks.MockWorldsService)
		status   int
	}{
		{
			caseName: "delete world",
			userID:   testUserID,
			prepare: func(srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					Delete(gomock.Any(), testWorldID, testUserID).
					Return(nil)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "missing user id",
			status:   http.StatusUnauthorized,
		},
		{
			caseName: "world not found",
			userID:   testUserID,
			prepare: func(srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					Delete(gomock.Any(), testWorldID, testUserID).
					Return(domain.ErrNotFound)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "service error",
			userID:   testUserID,
			prepare: func(srvc *srvc_mocks.MockWorldsService) {
				srvc.EXPECT().
					Delete(gomock.Any(), testWorldID, testUserID).
					Return(errors.New("database error"))
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockWorldsService(ctrl)

			if tt.prepare != nil {
				tt.prepare(srvc)
			}

			r := newWorldsHandlerRouter(t, srvc, tt.userID)
			rec := performWorldsRequest(
				r,
				http.MethodDelete,
				"/minecraft/"+testWorldID,
				"",
			)

			assert.Equal(t, tt.status, rec.Code)
			if tt.status == http.StatusNoContent {
				assert.Empty(t, rec.Body.String())
			}
		})
	}
}
