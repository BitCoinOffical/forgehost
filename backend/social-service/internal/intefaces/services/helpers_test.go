package services

import (
	"encoding/base64"
	v2 "encoding/json/v2"
	"strings"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/stretchr/testify/require"
)

const (
	testPublicEndpoint = "https://cdn.test"
	PrefixPosts        = "posts/"
	PrefixAvatars      = "avatars/"
)

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func uintPtr(u uint64) *uint64 {
	return &u
}

func postImageID(n string) *string {
	return strPtr("post/" + n + ".png")
}

func postImageURL(n string) *string {
	return strPtr(testPublicEndpoint + "/post/" + n + ".png")
}

func globalFeed() []models.FeedPost {
	return []models.FeedPost{{PostID: "1", ImageID: postImageID("1")}, {PostID: "2"}}
}

func globalFeedWithURLs() []models.FeedPost {
	return []models.FeedPost{
		{PostID: "1", ImageID: postImageID("1"), ImageURL: postImageURL("1")},
		{PostID: "2"},
	}
}

type testUploadFile struct {
	*strings.Reader
}

func (testUploadFile) Close() error {
	return nil
}
func makeCursor(t *testing.T, postID *string, createdAt *time.Time) string {
	t.Helper()
	raw, err := v2.Marshal(dto.CursorDTO{PostID: postID, CreatedAt: createdAt})
	require.NoError(t, err)
	return base64.URLEncoding.EncodeToString(raw)
}
