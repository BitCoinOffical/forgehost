package services

import (
	"errors"
	"testing"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	store_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSearchServiceSearch(t *testing.T) {
	const (
		q     = "mine"
		limit = 5
	)

	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockSearchRepo(ctrl)
	s := NewSearchService(repo, testPublicEndpoint)

	repoErr := errors.New("repo error")
	str := func(v string) *string { return &v }

	postID1, postID2 := uuid.New(), uuid.New()
	userID1, userID2 := uuid.New(), uuid.New()

	newPosts := func() []models.SearchPost {
		return []models.SearchPost{
			{ID: postID1, ImageID: str("posts/a.jpg")},
			{ID: postID2},
		}
	}
	newUsers := func() []models.SearchUser {
		return []models.SearchUser{
			{UserID: userID1, Username: "mike", AvatarURL: str("avatars/m.jpg")},
			{UserID: userID2, Username: "mila"},
		}
	}
	newStreams := func() []models.SearchStream {
		return []models.SearchStream{{ID: "s1", Title: "minecraft"}}
	}

	wantPosts := []models.SearchPost{
		{ID: postID1, ImageID: str("posts/a.jpg"), ImageURL: str(testPublicEndpoint + "/posts/a.jpg")},
		{ID: postID2},
	}
	wantUsers := []models.SearchUser{
		{UserID: userID1, Username: "mike", AvatarURL: str(testPublicEndpoint + "/avatars/m.jpg")},
		{UserID: userID2, Username: "mila"},
	}
	wantStreams := newStreams()

	testCases := []struct {
		caseName string
		data     struct {
			typ string
		}
		prepare  func()
		want     *models.SearchResponse
		wantErrs []error
		success  bool
	}{
		{
			caseName: "all types",
			data:     struct{ typ string }{typ: ""},
			prepare: func() {
				repo.EXPECT().SearchPosts(gomock.Any(), q, limit).Return(newPosts(), nil).Times(1)
				repo.EXPECT().SearchStreams(gomock.Any(), q, limit).Return(newStreams(), nil).Times(1)
				repo.EXPECT().SearchUsers(gomock.Any(), q, limit).Return(newUsers(), nil).Times(1)
			},
			want:    &models.SearchResponse{Posts: wantPosts, Streams: wantStreams, Users: wantUsers},
			success: true,
		},
		{
			caseName: "only posts, image urls are built",
			data:     struct{ typ string }{typ: SearchTypePosts},
			prepare: func() {
				repo.EXPECT().SearchPosts(gomock.Any(), q, limit).Return(newPosts(), nil).Times(1)
			},
			want:    &models.SearchResponse{Posts: wantPosts},
			success: true,
		},
		{
			caseName: "only streams",
			data:     struct{ typ string }{typ: SearchTypeStreams},
			prepare: func() {
				repo.EXPECT().SearchStreams(gomock.Any(), q, limit).Return(newStreams(), nil).Times(1)
			},
			want:    &models.SearchResponse{Streams: wantStreams},
			success: true,
		},
		{
			caseName: "only users, avatar urls are built",
			data:     struct{ typ string }{typ: SearchTypeUsers},
			prepare: func() {
				repo.EXPECT().SearchUsers(gomock.Any(), q, limit).Return(newUsers(), nil).Times(1)
			},
			want:    &models.SearchResponse{Users: wantUsers},
			success: true,
		},
		{
			caseName: "unknown type returns empty response",
			data:     struct{ typ string }{typ: "comments"},
			prepare:  func() {},
			want:     &models.SearchResponse{},
			success:  true,
		},
		{
			caseName: "search posts error",
			data:     struct{ typ string }{typ: SearchTypePosts},
			prepare: func() {
				repo.EXPECT().SearchPosts(gomock.Any(), q, limit).Return(nil, repoErr).Times(1)
			},
			wantErrs: []error{repoErr},
		},
		{
			caseName: "search streams error",
			data:     struct{ typ string }{typ: SearchTypeStreams},
			prepare: func() {
				repo.EXPECT().SearchStreams(gomock.Any(), q, limit).Return(nil, repoErr).Times(1)
			},
			wantErrs: []error{repoErr},
		},
		{
			caseName: "search users error",
			data:     struct{ typ string }{typ: SearchTypeUsers},
			prepare: func() {
				repo.EXPECT().SearchUsers(gomock.Any(), q, limit).Return(nil, repoErr).Times(1)
			},
			wantErrs: []error{repoErr},
		},
		{
			caseName: "one of three fails in all-types search",
			data:     struct{ typ string }{typ: ""},
			prepare: func() {
				repo.EXPECT().SearchPosts(gomock.Any(), q, limit).Return(nil, repoErr).Times(1)
				repo.EXPECT().SearchStreams(gomock.Any(), q, limit).Return(newStreams(), nil).AnyTimes()
				repo.EXPECT().SearchUsers(gomock.Any(), q, limit).Return(newUsers(), nil).AnyTimes()
			},
			wantErrs: []error{repoErr},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := s.Search(t.Context(), q, tc.data.typ, limit)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
				return
			}
			require.Error(t, err)
			for _, wantErr := range tc.wantErrs {
				require.ErrorIs(t, err, wantErr)
			}
			assert.Nil(t, got)
		})
	}
}

func TestSearchServiceSuggest(t *testing.T) {
	const q = "mi"

	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockSearchRepo(ctrl)
	s := NewSearchService(repo, testPublicEndpoint)

	repoErr := errors.New("repo error")
	userID := uuid.New()

	newUsers := func() []models.Suggestion {
		return []models.Suggestion{
			{Type: "user", ID: "u1", Text: "mike"},
			{Type: "user", ID: "u2", Text: "mila"},
		}
	}
	newStreams := func() []models.Suggestion {
		return []models.Suggestion{{Type: "stream", ID: "s1", Text: "minecraft"}}
	}

	testCases := []struct {
		caseName string
		data     struct {
			userID string
			limit  int
		}
		prepare  func()
		want     []models.Suggestion
		wantErrs []error
		success  bool
	}{
		{
			caseName: "users go first, then streams",
			data: struct {
				userID string
				limit  int
			}{userID: userID.String(), limit: 8},
			prepare: func() {
				repo.EXPECT().SuggestUsers(gomock.Any(), userID, q, 8).Return(newUsers(), nil).Times(1)
				repo.EXPECT().SuggestStreams(gomock.Any(), q, 8).Return(newStreams(), nil).Times(1)
			},
			want: []models.Suggestion{
				{Type: "user", ID: "u1", Text: "mike"},
				{Type: "user", ID: "u2", Text: "mila"},
				{Type: "stream", ID: "s1", Text: "minecraft"},
			},
			success: true,
		},
		{
			caseName: "result is trimmed to limit",
			data: struct {
				userID string
				limit  int
			}{userID: userID.String(), limit: 2},
			prepare: func() {
				repo.EXPECT().SuggestUsers(gomock.Any(), userID, q, 2).Return(newUsers(), nil).Times(1)
				repo.EXPECT().SuggestStreams(gomock.Any(), q, 2).Return(newStreams(), nil).Times(1)
			},
			want: []models.Suggestion{
				{Type: "user", ID: "u1", Text: "mike"},
				{Type: "user", ID: "u2", Text: "mila"},
			},
			success: true,
		},
		{
			caseName: "nothing found",
			data: struct {
				userID string
				limit  int
			}{userID: userID.String(), limit: 8},
			prepare: func() {
				repo.EXPECT().SuggestUsers(gomock.Any(), userID, q, 8).Return(nil, nil).Times(1)
				repo.EXPECT().SuggestStreams(gomock.Any(), q, 8).Return(nil, nil).Times(1)
			},
			want:    nil,
			success: true,
		},
		{
			caseName: "invalid user id",
			data: struct {
				userID string
				limit  int
			}{userID: "not-a-uuid", limit: 8},
			prepare: func() {},
		},
		{
			caseName: "suggest users error",
			data: struct {
				userID string
				limit  int
			}{userID: userID.String(), limit: 8},
			prepare: func() {
				repo.EXPECT().SuggestUsers(gomock.Any(), userID, q, 8).Return(nil, repoErr).Times(1)
				repo.EXPECT().SuggestStreams(gomock.Any(), q, 8).Return(newStreams(), nil).Times(1)
			},
			wantErrs: []error{repoErr},
		},
		{
			caseName: "suggest streams error",
			data: struct {
				userID string
				limit  int
			}{userID: userID.String(), limit: 8},
			prepare: func() {
				repo.EXPECT().SuggestUsers(gomock.Any(), userID, q, 8).Return(newUsers(), nil).Times(1)
				repo.EXPECT().SuggestStreams(gomock.Any(), q, 8).Return(nil, repoErr).Times(1)
			},
			wantErrs: []error{repoErr},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := s.Suggest(t.Context(), tc.data.userID, q, tc.data.limit)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
				return
			}
			require.Error(t, err)
			for _, wantErr := range tc.wantErrs {
				require.ErrorIs(t, err, wantErr)
			}
			assert.Nil(t, got)
		})
	}
}
