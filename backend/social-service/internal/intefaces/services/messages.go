package services

import (
	"context"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"go.mongodb.org/mongo-driver/v2/bson"

	mongorepo "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/mongo"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/postgres"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/store"
)

type MessageService struct {
	repo  *postgres.MessagesRepo
	mrepo *mongorepo.MessagerRepo
	store *store.StreamStore
}

func NewMessageService(repo *postgres.MessagesRepo, mrepo *mongorepo.MessagerRepo, store *store.StreamStore) *MessageService {
	return &MessageService{repo: repo, mrepo: mrepo, store: store}
}

func (s *MessageService) SaveMessage(ctx context.Context, userID, chatID string, req *dto.MessageDTO) (*bson.ObjectID, error) {
	msg := models.Message{
		ChatID:    chatID,
		UserID:    userID,
		Text:      req.Text,
		CreatedAt: time.Now(),
	}
	blocked, err := s.repo.CheckBlock(ctx, userID, chatID)
	if err != nil {
		return nil, fmt.Errorf("s.repo.CheckBlock: %w", err)
	}
	if blocked {
		return nil, domain.ErrForbidden
	}

	objID, err := s.mrepo.SaveMessage(ctx, &msg)
	if err != nil {
		return nil, fmt.Errorf("s.mrepo.SaveMessage: %w", err)
	}

	return objID, nil
}

func (s *MessageService) GetMessages(ctx context.Context, chatID string) ([]models.Message, error) {
	msgs, err := s.mrepo.GetMessages(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("s.repo.SaveMessage: %w", err)
	}

	return msgs, nil
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
		Text:      req.Text,
		CreatedAt: time.Now(),
	}

	if err := s.store.SaveStreamMessage(ctx, streamID, &msg); err != nil {
		return fmt.Errorf("s.store.GetStreamMessages: %w", err)
	}

	return nil
}
