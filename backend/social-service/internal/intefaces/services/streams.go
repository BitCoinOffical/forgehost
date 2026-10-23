package services

import (
	"context"
	"encoding/base64"
	v2 "encoding/json/v2"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
)

type StreamService struct {
	store StreamStore
	repo  StreamRepo
}

func NewStreamService(store StreamStore, repo StreamRepo) *StreamService {
	return &StreamService{store: store, repo: repo}
}

func (s *StreamService) GetStreamHistory(ctx context.Context, videoID string) ([]models.Message, error) {
	msgs, err := s.store.GetStreamMessages(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("s.store.GetStreamMessages: %w", err)
	}

	return msgs, nil
}

func (s *StreamService) SaveStreamMessage(ctx context.Context, userID, videoID string, req *dto.MessageDTO) error {
	msg := models.Message{
		ChatID:    videoID,
		UserID:    userID,
		Text:      req.Text,
		CreatedAt: time.Now(),
	}

	if err := s.store.SaveStreamMessage(ctx, videoID, &msg); err != nil {
		return fmt.Errorf("s.store.GetStreamMessages: %w", err)
	}

	return nil
}

func (s *StreamService) SaveStreamUser(ctx context.Context, userID, videoID string) error {
	if err := s.store.SaveStreamUser(ctx, userID, videoID); err != nil {
		return fmt.Errorf("s.store.SaveStreamUser: %w", err)
	}
	return nil
}

func (s *StreamService) GetStreamUsers(ctx context.Context, videoID string) ([]string, error) {
	users, err := s.store.GetStreamUsers(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("s.store.GetStreamUsers: %w", err)
	}
	return users, nil
}

func (s *StreamService) SaveChannel(ctx context.Context, channelID, userID string) error {
	if err := s.repo.SaveChannel(ctx, channelID, userID); err != nil {
		return fmt.Errorf("s.repo.SaveChannel: %w", err)
	}
	return nil
}

func (s *StreamService) GetChannel(ctx context.Context, userID string) (string, error) {
	channelID, err := s.repo.GetChannel(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("s.repo.GetChannel: %w", err)
	}
	return channelID, nil
}

func (s *StreamService) SaveStream(ctx context.Context, userID string, req *dto.StreamDTO) (string, error) {
	stream := models.Stream{
		UserID:   userID,
		VideoID:  req.VideoID,
		Title:    &req.Title,
		Status:   req.Status,
		EmbedURL: &req.EmbedURL,
		Viewers:  req.ConcurrentViewers,
		Likes:    req.LikeCount,
	}
	streamID, err := s.repo.SaveStream(ctx, &stream)
	if err != nil {
		return "", fmt.Errorf("s.repo.SaveStream: %w", err)
	}

	return streamID, nil
}

func (s *StreamService) BuildUpdateStream(ctx context.Context, userID string, req *dto.UpdateStreamDTO) (*models.Stream, error) {
	stream := models.Stream{
		UserID:   userID,
		Title:    req.Title,
		Status:   req.Status,
		EmbedURL: req.EmbedURL,
		Viewers:  req.ConcurrentViewers,
		Likes:    req.LikeCount,
	}
	newStream, err := s.repo.BuildUpdateStream(ctx, &stream)
	if err != nil {
		return nil, fmt.Errorf("s.repo.BuildUpdateStream: %w", err)
	}
	return newStream, nil
}

func (s *StreamService) DeleteStream(ctx context.Context, videoID, userID string) error {
	if err := s.repo.DeleteStream(ctx, videoID, userID); err != nil {
		return fmt.Errorf(" s.repo.DeleteStream: %w", err)
	}
	return nil
}

func (s *StreamService) DeleteComplatedStreams(ctx context.Context) error {
	if err := s.repo.DeleteComplatedStreams(ctx); err != nil {
		return fmt.Errorf("s.repo.DeleteComplatedStreams: %w", err)
	}
	return nil
}

func (s *StreamService) GetStream(ctx context.Context, videoID, userID string) (*models.Stream, error) {
	stream, err := s.repo.GetStream(ctx, videoID, userID)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetStream: %w", err)
	}
	return stream, nil
}

func (s *StreamService) GetCursorStreams(ctx context.Context, cursor string) ([]models.FeedStream, error) {
	var req dto.CursorStreamDTO
	if cursor != "" {
		decoded, err := base64.URLEncoding.DecodeString(cursor)
		if err != nil {
			return nil, fmt.Errorf("base64.URLEncoding.DecodeString: %w", err)
		}

		if err := v2.Unmarshal(decoded, req); err != nil {
			return nil, fmt.Errorf("v2.Unmarshal: %w", err)
		}

	}
	streams, err := s.repo.GetCursorStreams(ctx, req.StreamID, req.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetCursorStreams: %w", err)
	}
	return streams, nil
}
