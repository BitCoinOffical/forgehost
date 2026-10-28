package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	jwtpkg "github.com/BitCoinOffical/forgehost/social-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

const (
	testChatID     = "chat-id"
	testUserID     = "user-id"
	testTargetID   = "target-id"
	userContextKey = "user_claims"
)

type chatInput struct {
	body   string
	userID any
}

type chatCase struct {
	caseName string
	data     chatInput
	prepare  func()
	status   int
}

func userClaims(id string) any {
	return &jwtpkg.Claims{UserID: id}
}

func authCases(body string) []chatCase {
	return []chatCase{
		{
			caseName: "user id not found in context",
			data:     chatInput{body: body},
			prepare:  func() {},
			status:   http.StatusUnauthorized,
		},
		{
			caseName: "incorrect user id type",
			data:     chatInput{body: body, userID: "not_claims"},
			prepare:  func() {},
			status:   http.StatusUnauthorized,
		},
	}
}

func serveChat(t *testing.T, method, route, url string, in chatInput, h gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	r := gin.New()
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
			return usernameRe.MatchString(fl.Field().String())
		})
	}
	r.Handle(method, route, func(c *gin.Context) {
		if in.userID != nil {
			c.Set(userContextKey, in.userID)
		}
	}, h)

	req, err := http.NewRequestWithContext(t.Context(), method, url, strings.NewReader(in.body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func strPtr(s string) *string {
	return &s
}

type uploadImageInput struct {
	userID   any
	filename string
	file     []byte
}

func serveUploadPostImage(
	t *testing.T,
	handler *UploaderHandler,
	data uploadImageInput,
) *httptest.ResponseRecorder {
	t.Helper()

	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	if data.file != nil {
		filename := data.filename
		if filename == "" {
			filename = "image.png"
		}

		part, err := form.CreateFormFile("image", filename)
		require.NoError(t, err)

		_, err = part.Write(data.file)
		require.NoError(t, err)
	}

	require.NoError(t, form.Close())

	req := httptest.NewRequest(http.MethodPost, "/posts/image", body)
	req.Header.Set("Content-Type", form.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	if data.userID != nil {
		c.Set("user_claims", data.userID)
	}

	handler.UploadPostImage(c)

	return w
}
func serveAvatarPostImage(
	t *testing.T,
	handler *UploaderHandler,
	data uploadImageInput,
) *httptest.ResponseRecorder {
	t.Helper()

	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	if data.file != nil {
		filename := data.filename
		if filename == "" {
			filename = "image.png"
		}

		part, err := form.CreateFormFile("image", filename)
		require.NoError(t, err)

		_, err = part.Write(data.file)
		require.NoError(t, err)
	}

	require.NoError(t, form.Close())

	req := httptest.NewRequest(http.MethodPost, "/posts/image", body)
	req.Header.Set("Content-Type", form.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	if data.userID != nil {
		c.Set("user_claims", data.userID)
	}

	handler.UploadAvatarImage(c)

	return w
}
