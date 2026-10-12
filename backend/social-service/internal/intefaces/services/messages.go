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
)

type MessageService struct {
	repo  *postgres.MessagesRepo
	mrepo *mongorepo.MessagerRepo
}

func NewMessageService(repo *postgres.MessagesRepo, mrepo *mongorepo.MessagerRepo) *MessageService {
	return &MessageService{repo: repo, mrepo: mrepo}
}

func (s *MessageService) SaveDirectMessage(ctx context.Context, userID, targetID string, req *dto.MessageDTO) (*bson.ObjectID, error) {
	msg := models.Message{
		ChatID:    targetID,
		UserID:    userID,
		Text:      req.Text,
		CreatedAt: time.Now(),
	}
	blocked, err := s.repo.CheckBlock(ctx, userID, targetID)
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

func (s *MessageService) SaveGroupMessage(ctx context.Context, userID, chatID string, req *dto.MessageDTO) (*bson.ObjectID, error) {
	msg := models.Message{
		ChatID:    chatID,
		UserID:    userID,
		Text:      req.Text,
		CreatedAt: time.Now(),
	}
	blocked, err := s.repo.CheckBanUser(ctx, userID, chatID)
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

func (s *MessageService) GetUsersFromChat(ctx context.Context, chatID string) ([]models.FeedChatMember, error) {
	resp, err := s.repo.GetUsersFromChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetUsersFromChat: %w", err)
	}
	return resp, nil
}
