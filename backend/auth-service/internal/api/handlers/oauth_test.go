package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/forgehost/auth-service/internal/api/handlers/mock"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
	"golang.org/x/oauth2"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fakeGoogle(tokenStatus int, userinfoBody string, userinfoErr error) roundTripFunc {
	return func(r *http.Request) (*http.Response, error) {
		resp := &http.Response{Header: make(http.Header), Request: r}
		switch r.URL.Host {
		case "oauth.test":
			resp.StatusCode = tokenStatus
			resp.Header.Set("Content-Type", "application/json")
			body := `{"access_token":"token","token_type":"Bearer"}`
			if tokenStatus != http.StatusOK {
				body = `{"error":"invalid_grant"}`
			}
			resp.Body = io.NopCloser(strings.NewReader(body))
		case "openidconnect.googleapis.com":
			if userinfoErr != nil {
				return nil, userinfoErr
			}
			resp.StatusCode = http.StatusOK
			resp.Body = io.NopCloser(strings.NewReader(userinfoBody))
		default:
			return nil, fmt.Errorf("unexpected host %s", r.URL.Host)
		}
		return resp, nil
	}
}

func newTestOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "http://localhost:8080/auth/login/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.test/auth",
			TokenURL: "https://oauth.test/token",
		},
	}
}

func TestGoogleLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockGoogleAuthService(ctrl)

	google := NewGoogleAuthHandler(srvc, newTestOauthConfig(), zaptest.NewLogger(t))

	r := gin.New()
	r.GET("/auth/login/google", google.GoogleLogin)

	testCases := []struct {
		caseName string
		prepare  func()
		status   int
	}{
		{
			caseName: "google login",
			prepare:  func() {},
			status:   http.StatusTemporaryRedirect,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/auth/login/google", nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)

			var state string
			for _, cookie := range w.Result().Cookies() {
				if cookie.Name == "oauth_state" {
					state = cookie.Value
				}
			}
			require.NotEmpty(t, state)
			assert.Contains(t, w.Header().Get("Location"), "https://accounts.test/auth")
			assert.Contains(t, w.Header().Get("Location"), "state="+state)
		})
	}
}

func TestGoogleCallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockGoogleAuthService(ctrl)

	google := NewGoogleAuthHandler(srvc, newTestOauthConfig(), zaptest.NewLogger(t))

	r := gin.New()
	r.GET("/auth/login/google/callback", google.GoogleCallback)

	verifiedUser := `{"sub":"google_sub","email":"user@gmail.com","email_verified":true}`

	testCases := []struct {
		caseName string
		data     struct {
			cookie    string
			state     string
			code      string
			transport http.RoundTripper
		}
		prepare func()
		status  int
	}{
		{
			caseName: "google callback",
			data: struct {
				cookie    string
				state     string
				code      string
				transport http.RoundTripper
			}{cookie: "state", state: "state", code: "code", transport: fakeGoogle(http.StatusOK, verifiedUser, nil)},
			prepare: func() {
				srvc.EXPECT().GoogleCallback(gomock.Any(), gomock.Any()).
					Return("oauth_code", nil).Times(1)
			},
			status: http.StatusTemporaryRedirect,
		},
		{
			caseName: "missing state cookie",
			data: struct {
				cookie    string
				state     string
				code      string
				transport http.RoundTripper
			}{cookie: "", state: "state", code: "code", transport: fakeGoogle(http.StatusOK, verifiedUser, nil)},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "state mismatch",
			data: struct {
				cookie    string
				state     string
				code      string
				transport http.RoundTripper
			}{cookie: "state", state: "other_state", code: "code", transport: fakeGoogle(http.StatusOK, verifiedUser, nil)},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "empty code",
			data: struct {
				cookie    string
				state     string
				code      string
				transport http.RoundTripper
			}{cookie: "state", state: "state", code: "", transport: fakeGoogle(http.StatusOK, verifiedUser, nil)},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "exchange code error",
			data: struct {
				cookie    string
				state     string
				code      string
				transport http.RoundTripper
			}{cookie: "state", state: "state", code: "code", transport: fakeGoogle(http.StatusBadRequest, "", nil)},
			prepare: func() {},
			status:  http.StatusUnauthorized,
		},
		{
			caseName: "userinfo request error",
			data: struct {
				cookie    string
				state     string
				code      string
				transport http.RoundTripper
			}{cookie: "state", state: "state", code: "code", transport: fakeGoogle(http.StatusOK, "", errors.New("userinfo error"))},
			prepare: func() {},
			status:  http.StatusBadGateway,
		},
		{
			caseName: "invalid userinfo json",
			data: struct {
				cookie    string
				state     string
				code      string
				transport http.RoundTripper
			}{cookie: "state", state: "state", code: "code", transport: fakeGoogle(http.StatusOK, `{invalid`, nil)},
			prepare: func() {},
			status:  http.StatusInternalServerError,
		},
		{
			caseName: "email not verified",
			data: struct {
				cookie    string
				state     string
				code      string
				transport http.RoundTripper
			}{cookie: "state", state: "state", code: "code", transport: fakeGoogle(http.StatusOK, `{"sub":"google_sub","email":"user@gmail.com","email_verified":false}`, nil)},
			prepare: func() {},
			status:  http.StatusForbidden,
		},
		{
			caseName: "google callback service error",
			data: struct {
				cookie    string
				state     string
				code      string
				transport http.RoundTripper
			}{cookie: "state", state: "state", code: "code", transport: fakeGoogle(http.StatusOK, verifiedUser, nil)},
			prepare: func() {
				srvc.EXPECT().GoogleCallback(gomock.Any(), gomock.Any()).
					Return("", errors.New("google callback error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			ctx := context.WithValue(t.Context(), oauth2.HTTPClient, &http.Client{Transport: tc.data.transport})
			target := "/auth/login/google/callback?state=" + tc.data.state + "&code=" + tc.data.code
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
			require.NoError(t, err)
			if tc.data.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "oauth_state", Value: tc.data.cookie})
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusTemporaryRedirect {
				assert.Equal(t, redirect_url+"oauth_code", w.Header().Get("Location"))
			}
		})
	}
}

func TestExchange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockGoogleAuthService(ctrl)

	google := NewGoogleAuthHandler(srvc, nil, zaptest.NewLogger(t))

	r := gin.New()
	r.POST("/auth/exchange", google.Exchange)

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func()
		status  int
	}{
		{
			caseName: "exchange",
			data: struct {
				body string
			}{body: `{"code":"123456"}`},
			prepare: func() {
				srvc.EXPECT().Exchange(gomock.Any(), gomock.Any()).
					Return(&models.Tokens{AccessToken: "access_token", RefreshToken: "refresh_token"}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data: struct {
				body string
			}{body: `{invalid`},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "exchange error",
			data: struct {
				body string
			}{body: `{"code":"123456"}`},
			prepare: func() {
				srvc.EXPECT().Exchange(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("exchange error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/exchange", strings.NewReader(tc.data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestGoogleLoginAndroid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockGoogleAuthService(ctrl)

	google := NewGoogleAuthHandler(srvc, nil, zaptest.NewLogger(t))

	r := gin.New()
	r.POST("/auth/login/google", google.GoogleLoginAndroid)

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func()
		status  int
	}{
		{
			caseName: "google login android",
			data: struct {
				body string
			}{body: `{"id_token":"id_token"}`},
			prepare: func() {
				srvc.EXPECT().GoogleLoginAndroid(gomock.Any(), gomock.Any()).
					Return(&models.Tokens{AccessToken: "access_token", RefreshToken: "refresh_token"}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data: struct {
				body string
			}{body: `{invalid`},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "invalid google token",
			data: struct {
				body string
			}{body: `{"id_token":"invalid_token"}`},
			prepare: func() {
				srvc.EXPECT().GoogleLoginAndroid(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("idtoken.Validate: %w", domain.ErrInvalidGoogleToken)).Times(1)
			},
			status: http.StatusUnauthorized,
		},
		{
			caseName: "google login android error",
			data: struct {
				body string
			}{body: `{"id_token":"id_token"}`},
			prepare: func() {
				srvc.EXPECT().GoogleLoginAndroid(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("google login android error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/login/google", strings.NewReader(tc.data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}
