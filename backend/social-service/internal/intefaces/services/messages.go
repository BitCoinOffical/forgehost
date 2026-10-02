package services

import (
	"context"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"

	mongorepo "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/mongo"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/store"
)

type MessageService struct {
	repo  *mongorepo.MessagerRepo
	store *store.StreamStore
}

func NewMessageService(repo *mongorepo.MessagerRepo, store *store.StreamStore) *MessageService {
	return &MessageService{repo: repo, store: store}
}

func (s *MessageService) GetStreamHistory(ctx context.Context, streamID string) ([]models.Message, error) {
	msgs, err := s.store.GetStreamMessages(ctx, streamID)
	if err != nil {
		return nil, fmt.Errorf("s.store.GetStreamMessages: %w", err)
	}

	return msgs, nil
}

func (s *MessageService) SaveStreamMessage(ctx context.Context, userID, streamID string, req *dto.MessageDTO) error {
	msg := models.Message{
		ChatID:    streamID,
		UserID:    userID,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := s.store.SaveStreamMessage(ctx, streamID, &msg); err != nil {
		return fmt.Errorf("s.store.GetStreamMessages: %w", err)
	}

	return nil
}
