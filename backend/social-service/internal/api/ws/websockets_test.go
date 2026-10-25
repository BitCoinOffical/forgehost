package ws

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestHubRegister(t *testing.T) {
	testCases := []struct {
		caseName  string
		f         func(t *testing.T, hub *Hub) (userID string, conn *websocket.Conn)
		wantConns int
		wantUsers int
	}{
		{
			caseName: "first connection of user",
			f: func(t *testing.T, hub *Hub) (string, *websocket.Conn) {
				server, _ := newConnPair(t)
				return "user-id", server
			},
			wantConns: 1,
			wantUsers: 1,
		},
		{
			caseName: "second connection of same user",
			f: func(t *testing.T, hub *Hub) (string, *websocket.Conn) {
				first, _ := newConnPair(t)
				hub.Register("user-id", first)
				second, _ := newConnPair(t)
				return "user-id", second
			},
			wantConns: 2,
			wantUsers: 1,
		},
		{
			caseName: "connection of another user",
			f: func(t *testing.T, hub *Hub) (string, *websocket.Conn) {
				other, _ := newConnPair(t)
				hub.Register("other-id", other)
				server, _ := newConnPair(t)
				return "user-id", server
			},
			wantConns: 1,
			wantUsers: 2,
		},
		{
			caseName: "same connection twice",
			f: func(t *testing.T, hub *Hub) (string, *websocket.Conn) {
				server, _ := newConnPair(t)
				hub.Register("user-id", server)
				return "user-id", server
			},
			wantConns: 1,
			wantUsers: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			hub := NewHub(zaptest.NewLogger(t))
			userID, conn := tc.f(t, hub)

			hub.Register(userID, conn)

			assert.Len(t, hub.conns[userID], tc.wantConns)
			assert.Len(t, hub.conns, tc.wantUsers)
			assert.Contains(t, hub.conns[userID], conn)
		})
	}
}

func TestHubUnregister(t *testing.T) {
	testCases := []struct {
		caseName  string
		f         func(t *testing.T, hub *Hub) (userID string, conn *websocket.Conn)
		wantConns int
		wantUsers int
	}{
		{
			caseName: "one of two connections",
			f: func(t *testing.T, hub *Hub) (string, *websocket.Conn) {
				first, _ := newConnPair(t)
				second, _ := newConnPair(t)
				hub.Register("user-id", first)
				hub.Register("user-id", second)
				return "user-id", first
			},
			wantConns: 1,
			wantUsers: 1,
		},
		{
			caseName: "last connection removes user",
			f: func(t *testing.T, hub *Hub) (string, *websocket.Conn) {
				server, _ := newConnPair(t)
				hub.Register("user-id", server)
				return "user-id", server
			},
			wantConns: 0,
			wantUsers: 0,
		},
		{
			caseName: "unknown connection of known user",
			f: func(t *testing.T, hub *Hub) (string, *websocket.Conn) {
				known, _ := newConnPair(t)
				hub.Register("user-id", known)
				unknown, _ := newConnPair(t)
				return "user-id", unknown
			},
			wantConns: 1,
			wantUsers: 1,
		},
		{
			caseName: "unknown user",
			f: func(t *testing.T, hub *Hub) (string, *websocket.Conn) {
				known, _ := newConnPair(t)
				hub.Register("other-id", known)
				unknown, _ := newConnPair(t)
				return "user-id", unknown
			},
			wantConns: 0,
			wantUsers: 1,
		},
		{
			caseName: "empty hub",
			f: func(t *testing.T, hub *Hub) (string, *websocket.Conn) {
				server, _ := newConnPair(t)
				return "user-id", server
			},
			wantConns: 0,
			wantUsers: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			hub := NewHub(zaptest.NewLogger(t))
			userID, conn := tc.f(t, hub)

			hub.Unregister(userID, conn)

			assert.Len(t, hub.conns[userID], tc.wantConns)
			assert.Len(t, hub.conns, tc.wantUsers)
			assert.NotContains(t, hub.conns[userID], conn)
		})
	}
}

func TestHubBroadcast(t *testing.T) {
	msg := map[string]any{"text": "hello"}

	testCases := []struct {
		caseName    string
		payload     any
		canceledCtx bool
		f           func(t *testing.T, hub *Hub) (targets []string, wantReceive, wantSilent []*websocket.Conn)
	}{
		{
			caseName: "single user single connection",
			payload:  msg,
			f: func(t *testing.T, hub *Hub) ([]string, []*websocket.Conn, []*websocket.Conn) {
				client := connect(t, hub, "user-id")
				return []string{"user-id"}, []*websocket.Conn{client}, nil
			},
		},
		{
			caseName: "user with two connections gets message on both",
			payload:  msg,
			f: func(t *testing.T, hub *Hub) ([]string, []*websocket.Conn, []*websocket.Conn) {
				first := connect(t, hub, "user-id")
				second := connect(t, hub, "user-id")
				return []string{"user-id"}, []*websocket.Conn{first, second}, nil
			},
		},
		{
			caseName: "several targets",
			payload:  msg,
			f: func(t *testing.T, hub *Hub) ([]string, []*websocket.Conn, []*websocket.Conn) {
				first := connect(t, hub, "first-id")
				second := connect(t, hub, "second-id")
				return []string{"first-id", "second-id"}, []*websocket.Conn{first, second}, nil
			},
		},
		{
			caseName: "not targeted user gets nothing",
			payload:  msg,
			f: func(t *testing.T, hub *Hub) ([]string, []*websocket.Conn, []*websocket.Conn) {
				target := connect(t, hub, "target-id")
				bystander := connect(t, hub, "bystander-id")
				return []string{"target-id"}, []*websocket.Conn{target}, []*websocket.Conn{bystander}
			},
		},
		{
			caseName: "unknown target is skipped",
			payload:  msg,
			f: func(t *testing.T, hub *Hub) ([]string, []*websocket.Conn, []*websocket.Conn) {
				known := connect(t, hub, "known-id")
				return []string{"unknown-id", "known-id"}, []*websocket.Conn{known}, nil
			},
		},
		{
			caseName: "empty targets",
			payload:  msg,
			f: func(t *testing.T, hub *Hub) ([]string, []*websocket.Conn, []*websocket.Conn) {
				client := connect(t, hub, "user-id")
				return nil, nil, []*websocket.Conn{client}
			},
		},
		{
			caseName: "closed connection does not stop delivery to others",
			payload:  msg,
			f: func(t *testing.T, hub *Hub) ([]string, []*websocket.Conn, []*websocket.Conn) {
				deadServer, _ := newConnPair(t)
				hub.Register("dead-id", deadServer)
				require.NoError(t, deadServer.CloseNow())
				alive := connect(t, hub, "alive-id")
				return []string{"dead-id", "alive-id"}, []*websocket.Conn{alive}, nil
			},
		},
		{
			caseName:    "canceled context delivers nothing",
			payload:     msg,
			canceledCtx: true,
			f: func(t *testing.T, hub *Hub) ([]string, []*websocket.Conn, []*websocket.Conn) {
				client := connect(t, hub, "user-id")
				return []string{"user-id"}, nil, []*websocket.Conn{client}
			},
		},
		{
			caseName: "unmarshalable payload delivers nothing",
			payload:  make(chan int),
			f: func(t *testing.T, hub *Hub) ([]string, []*websocket.Conn, []*websocket.Conn) {
				client := connect(t, hub, "user-id")
				return []string{"user-id"}, nil, []*websocket.Conn{client}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			hub := NewHub(zaptest.NewLogger(t))
			targets, wantReceive, wantSilent := tc.f(t, hub)

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			if tc.canceledCtx {
				canceled, stop := context.WithCancel(ctx)
				stop()
				ctx = canceled
			}

			require.NotPanics(t, func() {
				hub.Broadcast(ctx, targets, tc.payload)
			})

			for _, client := range wantReceive {
				readCtx, readCancel := context.WithTimeout(t.Context(), 5*time.Second)
				var got map[string]any
				require.NoError(t, wsjson.Read(readCtx, client, &got))
				readCancel()
				assert.Equal(t, msg, got)
			}

			for _, client := range wantSilent {
				readCtx, readCancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
				_, _, err := client.Read(readCtx)
				readCancel()
				require.Error(t, err)
			}
		})
	}
}

func TestHubConcurrentUse(t *testing.T) {
	hub := NewHub(zaptest.NewLogger(t))

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	ids := make([]string, 0, 5)
	servers := make([]*websocket.Conn, 0, 5)
	for i := range 5 {
		server, _ := newConnPair(t)
		servers = append(servers, server)
		ids = append(ids, "user-"+strconv.Itoa(i))
	}

	var wg sync.WaitGroup
	for i := range 5 {
		wg.Go(func() {
			for range 50 {
				hub.Register(ids[i], servers[i])
				hub.Unregister(ids[i], servers[i])
			}
		})
	}
	wg.Go(func() {
		for range 50 {
			hub.Broadcast(ctx, ids, map[string]any{"text": "x"})
		}
	})
	wg.Wait()

	assert.Empty(t, hub.conns)
}
