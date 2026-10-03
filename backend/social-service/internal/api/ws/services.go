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
	hub  *Hub
	srvc *services.MessageService
}

func NewWebSockerServices(coll *mongo.Collection, rdb *redis.Client, pool *pgxpool.Pool) *WebSockerServices {
	hub := NewHub()
	store := store.NewStreamStore(rdb)
	repo := postgres.NewMessagersRepo(pool)
	mrepo := mongorepo.NewMessagerRepo(coll)
	srvc := services.NewMessageService(repo, mrepo, store)

	return &WebSockerServices{hub: hub, srvc: srvc}
}

type WebSockerHandlers struct {
	Msgs *MessagerHandler
}

func NewWebSockerHandlers(srvc *WebSockerServices, logger *zap.Logger) *WebSockerHandlers {
	msgs := NewMessageHandler(srvc.hub, srvc.srvc, logger)
	return &WebSockerHandlers{Msgs: msgs}
}
