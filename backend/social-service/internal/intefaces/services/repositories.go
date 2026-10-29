package services

import (
	"context"
	"io"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type UploadRepo interface {
	CreateUpload(ctx context.Context, imageID string, userID string) error
}

type S3 interface {
	Delete(ctx context.Context, imageID string) error
	GetURL(ctx context.Context, imageID string, ttl time.Duration) (string, error)
	Upload(ctx context.Context, key string, contentType string, body io.ReadSeeker) error
}

type SearchRepo interface {
	SearchPosts(ctx context.Context, q string, limit int) ([]models.SearchPost, error)
	SearchStreams(ctx context.Context, q string, limit int) ([]models.SearchStream, error)
	SearchUsers(ctx context.Context, q string, limit int) ([]models.SearchUser, error)
	SuggestUsers(ctx context.Context, userID uuid.UUID, prefix string, limit int) ([]models.Suggestion, error)
	SuggestStreams(ctx context.Context, prefix string, limit int) ([]models.Suggestion, error)
}

type CommentsRepo interface {
	ListComments(ctx context.Context, postId string) ([]models.FeedComments, error)
	CreateComment(ctx context.Context, comment *models.Comments) error
	UpdateComment(ctx context.Context, comment *models.Comments) (*models.Comments, error)
	DeleteComment(ctx context.Context, comment *models.Comments) error
	ReportComment(ctx context.Context, report *models.CommentReport) error
	LikeComment(ctx context.Context, userId, commentId string) error
	UnlikeComment(ctx context.Context, userId, commentId string) error
}

type PostsRepo interface {
	GetPostByID(ctx context.Context, id string) (*models.FeedPost, error)
	ViewPost(ctx context.Context, id string) error
	GetSubPosts(ctx context.Context, id string) ([]models.FeedPost, []models.FeedPost, error)
	GetGlobalPosts(ctx context.Context, postId *uuid.UUID, createdAt *time.Time) ([]models.FeedPost, error)
	CreatePost(ctx context.Context, post *models.Post) error
	BuildUpdatePost(ctx context.Context, post *models.Post) (*models.Post, error)
	DeletePost(ctx context.Context, postId string, userId string) error
	ReportPost(ctx context.Context, report *models.PostReport) error

	GetTopics(ctx context.Context) ([]models.Topics, error)
	GetTopicsByID(ctx context.Context, id string) (*models.Topics, error)
	CreateTopic(ctx context.Context, topic *models.Topics) error
	DeleteTopic(ctx context.Context, topic *models.Topics) error
	ReportTopic(ctx context.Context, topic *models.TopicsReport) error

	LikePost(ctx context.Context, userId, postId string) error
	UnlikePost(ctx context.Context, userId, postId string) error
}

type PostsCache interface {
	GetGlobal(ctx context.Context, cursor string) ([]models.FeedPost, error)
	SetGlobal(ctx context.Context, posts []models.FeedPost, cursor string) error
	DeleteGlobal(ctx context.Context, cursor string) error
}

type ProfileRepo interface {
	GetProfileByID(ctx context.Context, id string) (*models.Profile, []models.FeedPost, error)
	SaveProfile(ctx context.Context, userId string) error
	BuildUpdateProfile(ctx context.Context, profile *models.Profile) (*models.Profile, error)
	GetSubscriptions(ctx context.Context, id string) ([]models.Subscriptions, error)
	GetSubscribers(ctx context.Context, id string) ([]models.Subscribes, error)
	Subscribe(ctx context.Context, userId, targetId string) error
	UnSubscribe(ctx context.Context, userId, targetId string) error
	CreateProfileReport(ctx context.Context, userId, targetId, cause string) error
	Block(ctx context.Context, userID, targetID string) error
}

type ChatsRepo interface {
	CreateGroup(ctx context.Context, ownerID string, chat *models.Chat) (string, error)
	CreateDirect(ctx context.Context, userID, targetID string) (string, error)
	GetChatByID(ctx context.Context, chatID string) (*models.Chat, error)
	DeleteChat(ctx context.Context, chatID, ownerID string) error
	LeaveChat(ctx context.Context, chatID, userID string) error
	BuildUpdateChat(ctx context.Context, chatID, userID string, chat *models.Chat) (*models.Chat, error)
	SetUserRoleChat(ctx context.Context, ownerID string, members *models.ChatMember) (*models.ChatMember, error)
	AddUserInChat(ctx context.Context, chatID, userID string) error
	KickUserFromChat(ctx context.Context, chatID, targetID, userID string) error
	GetUsersFromChat(ctx context.Context, chatID string) ([]models.FeedChatMember, error)
	BanUser(ctx context.Context, chatID, targetID, userID string) error
	UnbanUser(ctx context.Context, chatID, targetID, userID string) error
	CheckBanUser(ctx context.Context, chatID, targetID string) (bool, error)
	CheckBlock(ctx context.Context, userID, targetID string) (bool, error)
	GetChatList(ctx context.Context, userID string) ([]models.FeedDirects, error)
	TouchChat(ctx context.Context, chatID string) error
	MarkAsRead(ctx context.Context, chatID, userID string) error
}

type MessagesRepo interface {
	CheckBlock(ctx context.Context, userID, targetID string) (bool, error)
	CheckBanUser(ctx context.Context, chatID, targetID string) (bool, error)
	GetUsersFromChat(ctx context.Context, chatID string) ([]models.FeedChatMember, error)
}

type MessagerRepo interface {
	SaveMessage(ctx context.Context, msg *models.Message) (*bson.ObjectID, error)
	GetMessages(ctx context.Context, chatID string) ([]models.Message, error)
}

type StreamRepo interface {
	SaveChannel(ctx context.Context, channelID, userID string) error
	GetChannel(ctx context.Context, userID string) (string, error)
	SaveStream(ctx context.Context, stream *models.Stream) (string, error)
	BuildUpdateStream(ctx context.Context, stream *models.Stream) (*models.Stream, error)
	DeleteStream(ctx context.Context, videoID, userID string) error
	DeleteComplatedStreams(ctx context.Context) error
	GetStream(ctx context.Context, videoID, userID string) (*models.Stream, error)
	GetCursorStreams(ctx context.Context, streamID string, createdAt time.Time) ([]models.FeedStream, error)
}

type StreamStore interface {
	SaveStreamMessage(ctx context.Context, videoID string, msg *models.Message) error
	GetStreamMessages(ctx context.Context, videoID string) ([]models.Message, error)
	SaveStreamUser(ctx context.Context, userID, videoID string) error
	GetStreamUsers(ctx context.Context, videoID string) ([]string, error)
}
