package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwtpkg "github.com/BitCoinOffical/forgehost/social-service/pkg/jwt"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type msgInput struct {
	body   string
	userID any
}

type msgCase struct {
	caseName string
	data     msgInput
	prepare  func()
	status   int
}

func userClaims(id string) any {
	return &jwtpkg.Claims{UserID: id}
}

func withUser(userID any) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID != nil {
			c.Set("user_claims", userID)
		}
	}
}

func authCases(body string) []msgCase {
	return []msgCase{
		{
			caseName: "user id not found in context",
			data:     msgInput{body: body},
			prepare:  func() {},
			status:   http.StatusUnauthorized,
		},
		{
			caseName: "incorrect user id type",
			data:     msgInput{body: body, userID: "not_claims"},
			prepare:  func() {},
			status:   http.StatusUnauthorized,
		},
	}
}

func serveMsg(t *testing.T, method, route, url string, in msgInput, h gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	r := gin.New()
	r.Handle(method, route, withUser(in.userID), h)

	req, err := http.NewRequestWithContext(t.Context(), method, url, strings.NewReader(in.body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func newConnPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()

	accepted := make(chan *websocket.Conn, 1)
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		accepted <- conn
		<-release
	}))

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	client, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	require.NoError(t, err)

	var server *websocket.Conn
	select {
	case server = <-accepted:
	case <-ctx.Done():
		t.Fatal("server side connection was not accepted")
	}

	t.Cleanup(func() {
		close(release)
		require.NoError(t, client.CloseNow())
		require.NoError(t, server.CloseNow())
		srv.Close()
	})

	return server, client
}

func connect(t *testing.T, hub *Hub, userID string) *websocket.Conn {
	t.Helper()
	server, client := newConnPair(t)
	hub.Register(userID, server)
	return client
}
