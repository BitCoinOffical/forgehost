package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchPosts(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewSearchRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (q string, limit int, want []models.SearchPost)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "find by substring",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedPostText(t, pool, author, strPtr("hello "+token+" world"), nil)
				seedPostText(t, pool, author, strPtr("something else"), nil)
				return token, 10, []models.SearchPost{wantPost(match, author, strPtr("hello "+token+" world"), nil)}
			},
			success: true,
		},
		{
			caseName: "case insensitive",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedPostText(t, pool, author, strPtr("Mixed Case "+token), nil)
				return strings.ToUpper(token), 10, []models.SearchPost{wantPost(match, author, strPtr("Mixed Case "+token), nil)}
			},
			success: true,
		},
		{
			caseName: "cyrillic case insensitive",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedPostText(t, pool, author, strPtr("Привет мир "+token), nil)
				seedPostText(t, pool, author, strPtr("Другой текст "+token), nil)
				return "ПРИВЕТ МИР " + token, 10, []models.SearchPost{wantPost(match, author, strPtr("Привет мир "+token), nil)}
			},
			success: true,
		},
		{
			caseName: "newest first",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				author := newUser(t, pool)
				ids := seedPostsOldestFirst(t, pool, author, token, 3)
				return token, 10, []models.SearchPost{
					wantPost(ids[2], author, strPtr(token+"2"), nil),
					wantPost(ids[1], author, strPtr(token+"1"), nil),
					wantPost(ids[0], author, strPtr(token+"0"), nil),
				}
			},
			success: true,
		},
		{
			caseName: "limit",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				author := newUser(t, pool)
				ids := seedPostsOldestFirst(t, pool, author, token, 3)
				return token, 2, []models.SearchPost{
					wantPost(ids[2], author, strPtr(token+"2"), nil),
					wantPost(ids[1], author, strPtr(token+"1"), nil),
				}
			},
			success: true,
		},
		{
			caseName: "limit zero",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				seedPostsOldestFirst(t, pool, newUser(t, pool), token, 3)
				return token, 0, nil
			},
			success: true,
		},
		{
			caseName: "nothing found",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				return searchToken(), 10, nil
			},
			success: true,
		},
		{
			caseName: "returns image id",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				author := newUser(t, pool)
				imageID := uuid.NewString()
				seedUpload(t, pool, imageID, author)
				match := seedPostText(t, pool, author, strPtr(token), &imageID)
				return token, 10, []models.SearchPost{wantPost(match, author, strPtr(token), &imageID)}
			},
			success: true,
		},
		{
			caseName: "excludes deleted posts",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				author := newUser(t, pool)
				kept := seedPostText(t, pool, author, strPtr(token+"keep"), nil)
				deleted := seedPostText(t, pool, author, strPtr(token+"gone"), nil)
				softDeletePost(t, pool, deleted)
				return token, 10, []models.SearchPost{wantPost(kept, author, strPtr(token+"keep"), nil)}
			},
			success: true,
		},
		{
			caseName: "percent is matched literally",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedPostText(t, pool, author, strPtr(token+"100%"), nil)
				seedPostText(t, pool, author, strPtr(token+"1000"), nil)
				return token + "100%", 10, []models.SearchPost{wantPost(match, author, strPtr(token+"100%"), nil)}
			},
			success: true,
		},
		{
			caseName: "underscore is matched literally",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedPostText(t, pool, author, strPtr(token+"a_b"), nil)
				seedPostText(t, pool, author, strPtr(token+"axb"), nil)
				return token + "a_b", 10, []models.SearchPost{wantPost(match, author, strPtr(token+"a_b"), nil)}
			},
			success: true,
		},
		{
			caseName: "backslash is matched literally",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedPostText(t, pool, author, strPtr(token+"a\\b"), nil)
				seedPostText(t, pool, author, strPtr(token+"ab"), nil)
				return token + "a\\b", 10, []models.SearchPost{wantPost(match, author, strPtr(token+"a\\b"), nil)}
			},
			success: true,
		},
		{
			caseName: "negative limit",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				return searchToken(), -1, nil
			},
		},
		{
			caseName: "canceled context",
			f: func(t *testing.T) (string, int, []models.SearchPost) {
				return searchToken(), 10, nil
			},
			canceledCtx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			q, limit, want := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			got, err := repo.SearchPosts(ctx, q, limit)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Len(t, got, len(want))
			if len(want) > 0 {
				assert.Equal(t, want, got)
			}
		})
	}
}

func TestSearchStreams(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewSearchRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (q string, limit int, want []models.SearchStream)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "find by substring",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedStreamTitle(t, pool, author, "live "+token+" now")
				seedStreamTitle(t, pool, author, "another stream")
				return token, 10, []models.SearchStream{wantStream(match, author, "live "+token+" now")}
			},
			success: true,
		},
		{
			caseName: "case insensitive",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedStreamTitle(t, pool, author, "Mixed Case "+token)
				return strings.ToUpper(token), 10, []models.SearchStream{wantStream(match, author, "Mixed Case "+token)}
			},
			success: true,
		},
		{
			caseName: "cyrillic case insensitive",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedStreamTitle(t, pool, author, "Прямой эфир "+token)
				seedStreamTitle(t, pool, author, "Другой эфир "+token)
				return "ПРЯМОЙ ЭФИР " + token, 10, []models.SearchStream{wantStream(match, author, "Прямой эфир "+token)}
			},
			success: true,
		},
		{
			caseName: "ordered by title",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				token := searchToken()
				author := newUser(t, pool)
				ids := seedStreamsByTitle(t, pool, author, token, 3)
				return token, 10, []models.SearchStream{
					wantStream(ids[0], author, token+"0"),
					wantStream(ids[1], author, token+"1"),
					wantStream(ids[2], author, token+"2"),
				}
			},
			success: true,
		},
		{
			caseName: "streams of different users",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				token := searchToken()
				first, second := newUser(t, pool), newUser(t, pool)
				secondID := seedStreamTitle(t, pool, second, token+"b")
				firstID := seedStreamTitle(t, pool, first, token+"a")
				return token, 10, []models.SearchStream{
					wantStream(firstID, first, token+"a"),
					wantStream(secondID, second, token+"b"),
				}
			},
			success: true,
		},
		{
			caseName: "limit",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				token := searchToken()
				author := newUser(t, pool)
				ids := seedStreamsByTitle(t, pool, author, token, 3)
				return token, 2, []models.SearchStream{
					wantStream(ids[0], author, token+"0"),
					wantStream(ids[1], author, token+"1"),
				}
			},
			success: true,
		},
		{
			caseName: "limit zero",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				token := searchToken()
				seedStreamsByTitle(t, pool, newUser(t, pool), token, 3)
				return token, 0, nil
			},
			success: true,
		},
		{
			caseName: "nothing found",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				return searchToken(), 10, nil
			},
			success: true,
		},
		{
			caseName: "percent is matched literally",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedStreamTitle(t, pool, author, token+"100%")
				seedStreamTitle(t, pool, author, token+"1000")
				return token + "100%", 10, []models.SearchStream{wantStream(match, author, token+"100%")}
			},
			success: true,
		},
		{
			caseName: "underscore is matched literally",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedStreamTitle(t, pool, author, token+"a_b")
				seedStreamTitle(t, pool, author, token+"axb")
				return token + "a_b", 10, []models.SearchStream{wantStream(match, author, token+"a_b")}
			},
			success: true,
		},
		{
			caseName: "backslash is matched literally",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedStreamTitle(t, pool, author, token+"a\\b")
				seedStreamTitle(t, pool, author, token+"ab")
				return token + "a\\b", 10, []models.SearchStream{wantStream(match, author, token+"a\\b")}
			},
			success: true,
		},
		{
			caseName: "negative limit",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				return searchToken(), -1, nil
			},
		},
		{
			caseName: "canceled context",
			f: func(t *testing.T) (string, int, []models.SearchStream) {
				return searchToken(), 10, nil
			},
			canceledCtx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			q, limit, want := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			got, err := repo.SearchStreams(ctx, q, limit)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Len(t, got, len(want))
			if len(want) > 0 {
				assert.Equal(t, want, got)
			}
		})
	}
}

func TestSearchUsers(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewSearchRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (q string, limit int, want []models.SearchUser)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "find by substring",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				match := newUser(t, pool)
				renameUser(t, pool, match, "bob"+token+"z")
				newUser(t, pool)
				return token, 10, []models.SearchUser{wantUser(match, "bob"+token+"z", defaultAvatarURL(match))}
			},
			success: true,
		},
		{
			caseName: "case insensitive",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				match := newUser(t, pool)
				renameUser(t, pool, match, "Alice"+token)
				return strings.ToLower("ALICE" + token), 10, []models.SearchUser{wantUser(match, "Alice"+token, defaultAvatarURL(match))}
			},
			success: true,
		},
		{
			caseName: "cyrillic case insensitive",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				match, decoy := newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, match, "Анна"+token)
				renameUser(t, pool, decoy, "Борис"+token)
				return "АННА" + token, 10, []models.SearchUser{wantUser(match, "Анна"+token, defaultAvatarURL(match))}
			},
			success: true,
		},
		{
			caseName: "ordered by username",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				ids := seedUsersByName(t, pool, token, 3)
				return token, 10, []models.SearchUser{
					wantUser(ids[0], token+"0", defaultAvatarURL(ids[0])),
					wantUser(ids[1], token+"1", defaultAvatarURL(ids[1])),
					wantUser(ids[2], token+"2", defaultAvatarURL(ids[2])),
				}
			},
			success: true,
		},
		{
			caseName: "limit",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				ids := seedUsersByName(t, pool, token, 3)
				return token, 2, []models.SearchUser{
					wantUser(ids[0], token+"0", defaultAvatarURL(ids[0])),
					wantUser(ids[1], token+"1", defaultAvatarURL(ids[1])),
				}
			},
			success: true,
		},
		{
			caseName: "limit zero",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				seedUsersByName(t, pool, token, 3)
				return token, 0, nil
			},
			success: true,
		},
		{
			caseName: "nothing found",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				return searchToken(), 10, nil
			},
			success: true,
		},
		{
			caseName: "returns custom avatar url",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				match := newUser(t, pool)
				renameUser(t, pool, match, token)
				avatarURL := "https://cdn.test/avatars/" + uuid.NewString() + ".png"
				setAvatarURL(t, pool, match, avatarURL)
				return token, 10, []models.SearchUser{wantUser(match, token, &avatarURL)}
			},
			success: true,
		},
		{
			caseName: "user without avatar has nil avatar url",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				match := newUser(t, pool)
				renameUser(t, pool, match, token)
				clearAvatarURL(t, pool, match)
				return token, 10, []models.SearchUser{wantUser(match, token, nil)}
			},
			success: true,
		},
		{
			caseName: "excludes banned users",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				kept, banned := newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, kept, token+"keep")
				renameUser(t, pool, banned, token+"ban")
				banProfile(t, pool, banned)
				return token, 10, []models.SearchUser{wantUser(kept, token+"keep", defaultAvatarURL(kept))}
			},
			success: true,
		},
		{
			caseName: "percent is matched literally",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				match, decoy := newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, match, token+"100%")
				renameUser(t, pool, decoy, token+"1000")
				return token + "100%", 10, []models.SearchUser{wantUser(match, token+"100%", defaultAvatarURL(match))}
			},
			success: true,
		},
		{
			caseName: "underscore is matched literally",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				match, decoy := newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, match, token+"a_b")
				renameUser(t, pool, decoy, token+"axb")
				return token + "a_b", 10, []models.SearchUser{wantUser(match, token+"a_b", defaultAvatarURL(match))}
			},
			success: true,
		},
		{
			caseName: "backslash is matched literally",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				token := searchToken()
				match, decoy := newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, match, token+"a\\b")
				renameUser(t, pool, decoy, token+"ab")
				return token + "a\\b", 10, []models.SearchUser{wantUser(match, token+"a\\b", defaultAvatarURL(match))}
			},
			success: true,
		},
		{
			caseName: "negative limit",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				return searchToken(), -1, nil
			},
		},
		{
			caseName: "canceled context",
			f: func(t *testing.T) (string, int, []models.SearchUser) {
				return searchToken(), 10, nil
			},
			canceledCtx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			q, limit, want := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			got, err := repo.SearchUsers(ctx, q, limit)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Len(t, got, len(want))
			if len(want) > 0 {
				assert.Equal(t, want, got)
			}
		})
	}
}

func TestSuggestUsers(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewSearchRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (userID uuid.UUID, prefix string, limit int, want []models.Suggestion)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "prefix match only",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer, match, decoy := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, match, token+"x")
				renameUser(t, pool, decoy, "z"+token)
				return uuid.MustParse(viewer), token, 10, []models.Suggestion{wantSuggestion("user", match, token+"x")}
			},
			success: true,
		},
		{
			caseName: "case insensitive",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer, match := newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, match, token+"Alice")
				return uuid.MustParse(viewer), strings.ToUpper(token), 10, []models.Suggestion{wantSuggestion("user", match, token+"Alice")}
			},
			success: true,
		},
		{
			caseName: "ordered by username",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer := newUser(t, pool)
				ids := seedUsersByName(t, pool, token, 3)
				return uuid.MustParse(viewer), token, 10, []models.Suggestion{
					wantSuggestion("user", ids[0], token+"0"),
					wantSuggestion("user", ids[1], token+"1"),
					wantSuggestion("user", ids[2], token+"2"),
				}
			},
			success: true,
		},
		{
			caseName: "limit",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer := newUser(t, pool)
				ids := seedUsersByName(t, pool, token, 3)
				return uuid.MustParse(viewer), token, 2, []models.Suggestion{
					wantSuggestion("user", ids[0], token+"0"),
					wantSuggestion("user", ids[1], token+"1"),
				}
			},
			success: true,
		},
		{
			caseName: "limit zero",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer := newUser(t, pool)
				seedUsersByName(t, pool, token, 3)
				return uuid.MustParse(viewer), token, 0, nil
			},
			success: true,
		},
		{
			caseName: "nothing found",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				return uuid.MustParse(newUser(t, pool)), searchToken(), 10, nil
			},
			success: true,
		},
		{
			caseName: "unknown requester is not filtered",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				match := newUser(t, pool)
				renameUser(t, pool, match, token)
				return uuid.New(), token, 10, []models.Suggestion{wantSuggestion("user", match, token)}
			},
			success: true,
		},
		{
			caseName: "excludes requester",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer, other := newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, viewer, token+"self")
				renameUser(t, pool, other, token+"other")
				return uuid.MustParse(viewer), token, 10, []models.Suggestion{wantSuggestion("user", other, token+"other")}
			},
			success: true,
		},
		{
			caseName: "excludes users blocked by requester",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer, blocked, kept := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, blocked, token+"blocked")
				renameUser(t, pool, kept, token+"kept")
				seedBlock(t, pool, viewer, blocked)
				return uuid.MustParse(viewer), token, 10, []models.Suggestion{wantSuggestion("user", kept, token+"kept")}
			},
			success: true,
		},
		{
			caseName: "excludes users who blocked requester",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer, blocker, kept := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, blocker, token+"blocker")
				renameUser(t, pool, kept, token+"kept")
				seedBlock(t, pool, blocker, viewer)
				return uuid.MustParse(viewer), token, 10, []models.Suggestion{wantSuggestion("user", kept, token+"kept")}
			},
			success: true,
		},
		{
			caseName: "block between other users does not hide them",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer, first, second := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, first, token+"a")
				renameUser(t, pool, second, token+"b")
				seedBlock(t, pool, first, second)
				return uuid.MustParse(viewer), token, 10, []models.Suggestion{
					wantSuggestion("user", first, token+"a"),
					wantSuggestion("user", second, token+"b"),
				}
			},
			success: true,
		},
		{
			caseName: "percent is matched literally",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer, match, decoy := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, match, token+"1%x")
				renameUser(t, pool, decoy, token+"12")
				return uuid.MustParse(viewer), token + "1%", 10, []models.Suggestion{wantSuggestion("user", match, token+"1%x")}
			},
			success: true,
		},
		{
			caseName: "underscore is matched literally",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				token := searchToken()
				viewer, match, decoy := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				renameUser(t, pool, match, token+"a_b")
				renameUser(t, pool, decoy, token+"axb")
				return uuid.MustParse(viewer), token + "a_", 10, []models.Suggestion{wantSuggestion("user", match, token+"a_b")}
			},
			success: true,
		},
		{
			caseName: "negative limit",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				return uuid.MustParse(newUser(t, pool)), searchToken(), -1, nil
			},
		},
		{
			caseName: "canceled context",
			f: func(t *testing.T) (uuid.UUID, string, int, []models.Suggestion) {
				return uuid.MustParse(newUser(t, pool)), searchToken(), 10, nil
			},
			canceledCtx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, prefix, limit, want := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			got, err := repo.SuggestUsers(ctx, userID, prefix, limit)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Len(t, got, len(want))
			if len(want) > 0 {
				assert.Equal(t, want, got)
			}
		})
	}
}

func TestSuggestStreams(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewSearchRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (prefix string, limit int, want []models.Suggestion)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "prefix match only",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedStreamTitle(t, pool, author, token+" live")
				seedStreamTitle(t, pool, author, "z "+token)
				return token, 10, []models.Suggestion{wantSuggestion("stream", match, token+" live")}
			},
			success: true,
		},
		{
			caseName: "case insensitive",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				token := searchToken()
				match := seedStreamTitle(t, pool, newUser(t, pool), token+"Live")
				return strings.ToUpper(token), 10, []models.Suggestion{wantSuggestion("stream", match, token+"Live")}
			},
			success: true,
		},
		{
			caseName: "cyrillic case insensitive",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedStreamTitle(t, pool, author, token+"Эфир")
				return strings.ToUpper(token) + "ЭФИР", 10, []models.Suggestion{wantSuggestion("stream", match, token+"Эфир")}
			},
			success: true,
		},
		{
			caseName: "ordered by title",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				token := searchToken()
				ids := seedStreamsByTitle(t, pool, newUser(t, pool), token, 3)
				return token, 10, []models.Suggestion{
					wantSuggestion("stream", ids[0], token+"0"),
					wantSuggestion("stream", ids[1], token+"1"),
					wantSuggestion("stream", ids[2], token+"2"),
				}
			},
			success: true,
		},
		{
			caseName: "limit",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				token := searchToken()
				ids := seedStreamsByTitle(t, pool, newUser(t, pool), token, 3)
				return token, 2, []models.Suggestion{
					wantSuggestion("stream", ids[0], token+"0"),
					wantSuggestion("stream", ids[1], token+"1"),
				}
			},
			success: true,
		},
		{
			caseName: "limit zero",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				token := searchToken()
				seedStreamsByTitle(t, pool, newUser(t, pool), token, 3)
				return token, 0, nil
			},
			success: true,
		},
		{
			caseName: "nothing found",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				return searchToken(), 10, nil
			},
			success: true,
		},
		{
			caseName: "percent is matched literally",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedStreamTitle(t, pool, author, token+"1%x")
				seedStreamTitle(t, pool, author, token+"12")
				return token + "1%", 10, []models.Suggestion{wantSuggestion("stream", match, token+"1%x")}
			},
			success: true,
		},
		{
			caseName: "underscore is matched literally",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				token := searchToken()
				author := newUser(t, pool)
				match := seedStreamTitle(t, pool, author, token+"a_b")
				seedStreamTitle(t, pool, author, token+"axb")
				return token + "a_", 10, []models.Suggestion{wantSuggestion("stream", match, token+"a_b")}
			},
			success: true,
		},
		{
			caseName: "negative limit",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				return searchToken(), -1, nil
			},
		},
		{
			caseName: "canceled context",
			f: func(t *testing.T) (string, int, []models.Suggestion) {
				return searchToken(), 10, nil
			},
			canceledCtx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			prefix, limit, want := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			got, err := repo.SuggestStreams(ctx, prefix, limit)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Len(t, got, len(want))
			if len(want) > 0 {
				assert.Equal(t, want, got)
			}
		})
	}
}
