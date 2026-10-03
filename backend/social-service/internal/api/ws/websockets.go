package ws

import (
	"context"
	"sync"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Hub struct {
	mu    sync.RWMutex
	conns map[string]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{conns: make(map[string]*websocket.Conn)}
}

func (h *Hub) Register(userID string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[userID] = c
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, userID)
}

func (h *Hub) SendTo(ctx context.Context, targetID string, msg any) error {
	h.mu.RLock()
	conn, ok := h.conns[targetID]
	h.mu.RUnlock()
	if !ok {
		return domain.ErrUserOffline
	}
	return wsjson.Write(ctx, conn, msg)
}
