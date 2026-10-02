package mongorepo

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MessagerRepo struct {
	coll *mongo.Collection
}

func NewMessagerRepo(coll *mongo.Collection) *MessagerRepo {
	return &MessagerRepo{coll: coll}
}

// for group and room
func (r *MessagerRepo) SaveMessage(ctx context.Context, msg *models.Message) (*bson.ObjectID, error) {
	res, err := r.coll.InsertOne(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("r.coll.InsertOne: %w", err)
	}

	objID, ok := res.InsertedID.(bson.ObjectID)
	if !ok {
		return nil, domain.ErrIncorrectType
	}

	return &objID, nil
}

func (r *MessagerRepo) GetMessages(ctx context.Context, id string) ([]models.Message, error) {
	filter := bson.M{"chat_id": id}
	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("r.coll.Find: %w", err)
	}
	defer cursor.Close(ctx)

	var msgs []models.Message
	if err := cursor.All(ctx, &msgs); err != nil {
		return nil, fmt.Errorf("cursor.All: %w", err)
	}

	return msgs, nil
}
