package ws

import (
	mongorepo "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/mongo"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/postgres"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/store"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type WebSockerServices struct {
	hub   *Hub
	msgs  *services.MessageService
	strms *services.StreamService
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
