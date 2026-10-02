package services

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/postgres"
)

type ChatService struct {
	repo *postgres.ChatsRepo
}

func NewChatService(repo *postgres.ChatsRepo) *ChatService {
	return &ChatService{repo: repo}
}
func (s *ChatService) CreateChat(ctx context.Context, ownerID string, req *dto.CreateChatDTO) (string, error) {
	chat := models.Chat{
		Title:       &req.Title,
		Description: req.Description,
	}

	chatID, err := s.repo.CreateChat(ctx, ownerID, &chat)
	if err != nil {
		return "", fmt.Errorf("s.repo.CreateChat: %w", err)
	}

	return chatID, nil
}

func (s *ChatService) GetChatByID(ctx context.Context, chatID string) (*models.Chat, error) {
	chat, err := s.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetChatByID: %w", err)
	}
	return chat, nil
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID, ownerID string) error {
	if err := s.repo.DeleteChat(ctx, chatID, ownerID); err != nil {
		return fmt.Errorf("s.repo.DeleteChat: %w", err)
	}
	return nil
}

func (s *ChatService) LeaveChat(ctx context.Context, chatID, userID string) error {
	if err := s.repo.LeaveChat(ctx, chatID, userID); err != nil {
		return fmt.Errorf("s.repo.LeaveChat: %w", err)
	}
	return nil
}

func (s *ChatService) BuildUpdateChat(ctx context.Context, chatID, userID string, req *dto.UpdateChatDTO) (*models.Chat, error) {
	chat := models.Chat{
		Title:       req.Title,
		Description: req.Description,
	}
	resp, err := s.repo.BuildUpdateChat(ctx, chatID, userID, &chat)
	if err != nil {
		return nil, fmt.Errorf("s.repo.BuildUpdateChat: %w", err)
	}
	return resp, nil
}

func (s *ChatService) SetUserRoleChat(ctx context.Context, chatID, ownerID string, req *dto.SetMemberRoleDTO) (*models.ChatMember, error) {
	members := models.ChatMember{
		ChatID: chatID,
		UserID: req.UserID,
		Role:   models.MemberRole(req.Role),
	}
	resp, err := s.repo.SetUserRoleChat(ctx, ownerID, &members)
	if err != nil {
		return nil, fmt.Errorf("s.repo.SetUserRoleChat: %w", err)
	}
	return resp, nil
}

func (s *ChatService) AddUserInChat(ctx context.Context, chatID, userID string) error {
	banned, err := s.repo.CheckBanUser(ctx, chatID, userID)
	if err != nil {
		return fmt.Errorf("s.repo.CheckBanUser: %w", err)
	}
	if banned {
		return domain.ErrForbidden
	}
	if err := s.repo.AddUserInChat(ctx, chatID, userID); err != nil {
		return fmt.Errorf("s.repo.AddUserInChat: %w", err)
	}
	return nil
}

func (s *ChatService) KickUserFromChat(ctx context.Context, chatID, targetID, userID string) error {
	if err := s.repo.KickUserFromChat(ctx, chatID, targetID, userID); err != nil {
		return fmt.Errorf("s.repo.KickUserFromChat: %w", err)
	}
	return nil
}

func (s *ChatService) GetUsersFromChat(ctx context.Context, chatID string) ([]models.FeedChatMember, error) {
	resp, err := s.repo.GetUsersFromChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetUsersFromChat: %w", err)
	}
	return resp, nil
}

func (s *ChatService) BanUser(ctx context.Context, chatID, targetID, userID string) error {
	if err := s.repo.BanUser(ctx, chatID, targetID, userID); err != nil {
		return fmt.Errorf("s.repo.BanUser: %w", err)
	}
	return nil
}

func (s *ChatService) UnbanUser(ctx context.Context, chatID, targetID, userID string) error {
	if err := s.repo.UnbanUser(ctx, chatID, targetID, userID); err != nil {
		return fmt.Errorf("s.repo.UnbanUser: %w", err)
	}
	return nil
}

func (s *ChatService) CheckBanUser(ctx context.Context, chatID, targetID string) (bool, error) {
	ok, err := s.repo.CheckBanUser(ctx, chatID, targetID)
	if err != nil {
		return false, fmt.Errorf("s.repo.CheckBanUser: %w", err)
	}
	return ok, nil
}
