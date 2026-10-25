package ws

import (
	"context"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	mongorepo "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/mongo"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/postgres"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/store"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type StreamService interface {
	GetStreamHistory(ctx context.Context, videoID string) ([]models.Message, error)
	SaveStreamMessage(ctx context.Context, userID, videoID string, req *dto.MessageDTO) error
	SaveStreamUser(ctx context.Context, userID, videoID string) error
	GetStreamUsers(ctx context.Context, videoID string) ([]string, error)
	SaveChannel(ctx context.Context, channelID, userID string) error
	GetChannel(ctx context.Context, userID string) (string, error)
	SaveStream(ctx context.Context, userID string, req *dto.StreamDTO) (string, error)
	BuildUpdateStream(ctx context.Context, userID, videoID string, req *dto.UpdateStreamDTO) (*models.Stream, error)
	DeleteStream(ctx context.Context, videoID, userID string) error
	DeleteComplatedStreams(ctx context.Context) error
	GetStream(ctx context.Context, videoID, userID string) (*models.Stream, error)
	GetCursorStreams(ctx context.Context, cursor string) ([]models.FeedStream, error)
}

type MessageService interface {
	SaveDirectMessage(ctx context.Context, userID, targetID string, req *dto.MessageDTO) (*bson.ObjectID, error)
	SaveGroupMessage(ctx context.Context, userID, chatID string, req *dto.MessageDTO) (*bson.ObjectID, error)
	GetMessages(ctx context.Context, chatID string) ([]models.Message, error)
	GetUsersFromChat(ctx context.Context, chatID string) ([]models.FeedChatMember, error)
}

type WebSockerServices struct {
	hub   *Hub
	msgs  MessageService
	strms StreamService
}

func NewWebSockerServices(coll *mongo.Collection, rdb *redis.Client, pool *pgxpool.Pool, logger *zap.Logger) *WebSockerServices {
	hub := NewHub(logger)
	store := store.NewStreamStore(rdb)
	msgRepo := postgres.NewMessagersRepo(pool)
	msgMrepo := mongorepo.NewMessagerRepo(coll)
	msgs := services.NewMessageService(msgRepo, msgMrepo)

	strmRepo := postgres.NewStreamRepo(pool)
	strms := services.NewStreamService(store, strmRepo)

	return &WebSockerServices{hub: hub, msgs: msgs, strms: strms}
}

type WebSockerHandlers struct {
	Msgs  *MessagerHandler
	Strms *StreamHandler
}

func NewWebSockerHandlers(srvc *WebSockerServices, logger *zap.Logger) *WebSockerHandlers {
	msgs := NewMessageHandler(srvc.hub, srvc.msgs, logger)
	strms := NewStreamHandler(srvc.hub, srvc.strms, logger)
	return &WebSockerHandlers{Msgs: msgs, Strms: strms}
}
