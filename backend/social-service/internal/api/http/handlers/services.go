package handlers

import (
	"context"
	"mime/multipart"

	"github.com/BitCoinOffical/forgehost/social-service/internal/adapters/seaweedfs"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/cache"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/postgres"
	s3store "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/s3"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/store"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type SearchService interface {
	Search(ctx context.Context, q, typ string, limit int) (*models.SearchResponse, error)
	Suggest(ctx context.Context, userIDStr string, q string, limit int) ([]models.Suggestion, error)
}

type UploaderService interface {
	UploadImage(ctx context.Context, prefix, userID, ext, contentType string, file multipart.File) (string, error)
}

type ChatService interface {
	CreateGroup(ctx context.Context, ownerID string, req *dto.CreateChatDTO) (string, error)
	CreateDirect(ctx context.Context, userID, targetID string) (string, error)
	GetChatList(ctx context.Context, userID string) ([]models.FeedDirects, error)
	MarkAsRead(ctx context.Context, chatID, userID string) error
	GetChatByID(ctx context.Context, chatID string) (*models.Chat, error)
	DeleteChat(ctx context.Context, chatID, ownerID string) error
	LeaveChat(ctx context.Context, chatID, userID string) error
	BuildUpdateChat(ctx context.Context, chatID, userID string, req *dto.UpdateChatDTO) (*models.Chat, error)
	SetUserRoleChat(ctx context.Context, chatID, ownerID string, req *dto.SetMemberRoleDTO) (*models.ChatMember, error)
	AddUserInChat(ctx context.Context, chatID, userID string) error
	KickUserFromChat(ctx context.Context, chatID, targetID, userID string) error
	GetUsersFromChat(ctx context.Context, chatID string) ([]models.FeedChatMember, error)
	BanUser(ctx context.Context, chatID, targetID, userID string) error
	UnbanUser(ctx context.Context, chatID, targetID, userID string) error
	CheckBanUser(ctx context.Context, chatID, targetID string) (bool, error)
}

type CommentsService interface {
	ListComments(ctx context.Context, postID string) ([]models.FeedComments, error)
	CreateComment(ctx context.Context, postID, userID string, comment *dto.CreateCommentDTO) error
	UpdateComment(ctx context.Context, postID, userID, commentID string, comment *dto.UpdateCommentDTO) (*models.Comments, error)
	DeleteComment(ctx context.Context, postID, userID, commentID string) error
	ReportComment(ctx context.Context, userID, commentID string, comment *dto.ReportCommentDTO) error
	LikeComment(ctx context.Context, userID, commentID string) error
	UnlikeComment(ctx context.Context, userID, commentID string) error
}

type PostsService interface {
	GetSubPosts(ctx context.Context, id string) ([]models.FeedPost, error)
	GetGlobalPosts(ctx context.Context, cursor string) ([]models.FeedPost, error)
	CreatePost(ctx context.Context, req *dto.CreatePostDTO, userID string) error
	UpdatePost(ctx context.Context, req *dto.UpdatePostDTO, postID, userID string) (*models.Post, error)
	GetPostByID(ctx context.Context, postID string) (*models.FeedPost, error)
	ViewPost(ctx context.Context, id string) error
	ReportPost(ctx context.Context, req *dto.ReportDTO, userID, postID string) error
	DeletePost(ctx context.Context, postID, userID string) error

	GetTopics(ctx context.Context) ([]models.Topics, error)
	GetTopicByID(ctx context.Context, topicID string) (*models.Topics, error)
	CreateTopic(ctx context.Context, userID string, req *dto.CreateTopicDTO) error
	DeleteTopic(ctx context.Context, topicID, userID string) error
	ReportTopic(ctx context.Context, userID, topicID string, req *dto.ReportTopicDTO) error

	LikePost(ctx context.Context, userID, postID string) error
	UnlikePost(ctx context.Context, userID, postID string) error
}

type ProfileService interface {
	GetProfileByID(ctx context.Context, id string) (*models.ProfileResponse, error)
	SaveProfile(ctx context.Context, event *dto.UserProfileDTO) error
	UpdateProfile(ctx context.Context, req *dto.UpdateProfileDTO, id string) (*models.Profile, error)
	GetSubscriptions(ctx context.Context, id string) ([]models.Subscriptions, error)
	GetSubscribers(ctx context.Context, id string) ([]models.Subscribes, error)
	Subscribe(ctx context.Context, userID, targetID string) error
	UnSubscribe(ctx context.Context, userID, targetID string) error
	Report(ctx context.Context, userID, targetID string, req *dto.ReportDTO) error
	Block(ctx context.Context, userID, targetID string) error
}

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

type Services struct {
	profile ProfileService
	post    PostsService
	coms    CommentsService
	chat    ChatService
	strm    StreamService
	img     UploaderService
	search  SearchService
}

func NewServices(pool *pgxpool.Pool, rdb *redis.Client, storefs *seaweedfs.Storage) *Services {
	profrepo := postgres.NewProfileRepo(pool)
	profile := services.NewProfileService(profrepo, storefs.PublicEndpoint)

	postrdb := cache.NewCache(rdb)
	postrepo := postgres.NewPostsRepo(pool)
	post := services.NewPostsService(postrepo, postrdb, storefs.PublicEndpoint)

	comrepo := postgres.NewCommentsRepo(pool)
	coms := services.NewCommentsService(comrepo)

	chatrepo := postgres.NewChatsRepo(pool)
	chat := services.NewChatService(chatrepo)

	store := store.NewStreamStore(rdb)
	stramrepo := postgres.NewStreamRepo(pool)
	strm := services.NewStreamService(store, stramrepo)

	imgstore := s3store.NewS3Loader(storefs)
	imgrepo := postgres.NewUploadRepo(pool)
	img := services.NewUploaderService(imgrepo, imgstore, storefs.PublicEndpoint)

	searchRepo := postgres.NewSearchRepo(pool)
	search := services.NewSearchService(searchRepo, storefs.PublicEndpoint)

	return &Services{profile: profile, post: post, coms: coms, chat: chat, strm: strm, img: img, search: search}
}

type Handlers struct {
	Profile  *ProfileHandler
	Posts    *PostHandler
	Comments *CommentHandler
	Chats    *ChatHandler
	Stream   *StreamHandler
	Uploader *UploaderHandler
	Search   *SearchHandler
}

func NewHandlers(srvc *Services, logger *zap.Logger, cfg *oauth2.Config) *Handlers {
	comments := NewCommentHandler(srvc.coms, logger)
	profile := NewProfileHandler(srvc.profile, logger)
	posts := NewPostHandler(srvc.post, logger)
	chat := NewChatHandler(srvc.chat, logger)
	stream := NewStreamHandler(logger, cfg, srvc.strm)
	uploader := NewUploaderHandler(srvc.img, logger)
	search := NewSearchHandler(srvc.search, logger)
	return &Handlers{Profile: profile, Posts: posts, Comments: comments, Chats: chat, Stream: stream, Uploader: uploader, Search: search}
}
