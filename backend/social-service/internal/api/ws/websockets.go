package ws

import (
	"context"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"go.uber.org/zap"
)

type Hub struct {
	mu     sync.RWMutex
	conns  map[string]map[*websocket.Conn]struct{}
	logger *zap.Logger
}

func NewHub(logger *zap.Logger) *Hub {
	return &Hub{conns: make(map[string]map[*websocket.Conn]struct{}), logger: logger}
}

func (h *Hub) Register(userID string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[userID] == nil {
		h.conns[userID] = make(map[*websocket.Conn]struct{})
	}
	h.conns[userID][c] = struct{}{}
}

func (h *Hub) Unregister(userID string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns[userID], c)
	if len(h.conns[userID]) == 0 {
		delete(h.conns, userID)
	}
}

func (h *Hub) Broadcast(ctx context.Context, targetIDs []string, msg any) {
	for _, targetID := range targetIDs {
		h.mu.RLock()
		conns := make([]*websocket.Conn, 0, len(h.conns[targetID]))
		for conn := range h.conns[targetID] {
			conns = append(conns, conn)
		}
		h.mu.RUnlock()
		for _, c := range conns {
			if err := wsjson.Write(ctx, c, msg); err != nil {
				h.logger.Error("ws write", zap.String("target id", targetID), zap.Error(err))
			}
		}
	}
}
