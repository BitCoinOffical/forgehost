package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type profileWant struct {
	subscribers   int
	subscriptions int
	posts         int
	postIDs       []string
}

func TestGetProfileByID(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewProfileRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (userID string, want profileWant)
		canceledCtx bool
		wantErr     error
		success     bool
	}{
		{
			caseName: "get profile with counters and posts",
			f: func(t *testing.T) (string, profileWant) {
				author, f1, f2, other := newUser(t, pool), newUser(t, pool), newUser(t, pool), newUser(t, pool)
				seedSubscription(t, pool, f1, &author, nil)
				seedSubscription(t, pool, f2, &author, nil)
				seedSubscription(t, pool, author, &other, nil)

				older, newer, deleted := newPost(t, pool, author), newPost(t, pool, author), newPost(t, pool, author)
				setPostCreatedAt(t, pool, older, time.Now().Add(-2*time.Hour))
				setPostCreatedAt(t, pool, newer, time.Now().Add(-1*time.Hour))
				softDeletePost(t, pool, deleted)
				newPost(t, pool, other)

				return author, profileWant{subscribers: 2, subscriptions: 1, posts: 2, postIDs: []string{newer, older}}
			},
			success: true,
		},
		{
			caseName: "get profile without posts and subscriptions",
			f: func(t *testing.T) (string, profileWant) {
				return newUser(t, pool), profileWant{postIDs: []string{}}
			},
			success: true,
		},
		{
			caseName: "get banned profile",
			f: func(t *testing.T) (string, profileWant) {
				userID := newUser(t, pool)
				banProfile(t, pool, userID)
				return userID, profileWant{}
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "get not found profile",
			f: func(t *testing.T) (string, profileWant) {
				return uuid.NewString(), profileWant{}
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "get profile with invalid id",
			f: func(t *testing.T) (string, profileWant) {
				return "invalid_uuid", profileWant{}
			},
		},
		{
			caseName: "get profile with canceled context",
			f: func(t *testing.T) (string, profileWant) {
				return newUser(t, pool), profileWant{}
			},
			canceledCtx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, want := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			profile, posts, err := repo.GetProfileByID(ctx, userID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				assert.Nil(t, profile)
				assert.Nil(t, posts)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, userID, profile.UserID)
			assert.Equal(t, "user_"+userID[:8], derefStr(profile.UserName))
			assert.False(t, profile.IsBanned)
			assert.Equal(t, want.subscribers, profile.Subscribers)
			assert.Equal(t, want.subscriptions, profile.Subscriptions)
			assert.Equal(t, want.posts, profile.Posts)
			assert.Equal(t, want.postIDs, postIDs(posts))
		})
	}
}

func TestSaveProfile(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewProfileRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) string
		wantErr  error
		success  bool
	}{
		{
			caseName: "save profile",
			f:        func(t *testing.T) string { return uuid.NewString() },
			success:  true,
		},
		{
			caseName: "save existing profile",
			f:        func(t *testing.T) string { return newUser(t, pool) },
			wantErr:  domain.ErrAlreadyExists,
		},
		{
			caseName: "save profile with invalid id",
			f:        func(t *testing.T) string { return "invalid_uuid" },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID := tc.f(t)

			err := repo.SaveProfile(context.Background(), userID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, countRows(t, pool, `SELECT COUNT(*) FROM profiles WHERE user_id = $1`, userID))
		})
	}
}

func TestBuildUpdateProfile(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewProfileRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (profile *models.Profile, wantName, wantBio, wantAvatar string, wantSubscribers int)
		wantErr  error
		success  bool
	}{
		{
			caseName: "update all fields",
			f: func(t *testing.T) (*models.Profile, string, string, string, int) {
				userID := newUser(t, pool)
				return &models.Profile{UserID: userID, UserName: strPtr("new_name_" + userID[:8]),
						Bio: strPtr("new bio"), AvatarUrl: strPtr("https://avatar.test/new")},
					"new_name_" + userID[:8], "new bio", "https://avatar.test/new", 0
			},
			success: true,
		},
		{
			caseName: "update only bio keeps other fields",
			f: func(t *testing.T) (*models.Profile, string, string, string, int) {
				userID := newUser(t, pool)
				return &models.Profile{UserID: userID, Bio: strPtr("only bio")},
					"user_" + userID[:8], "only bio", "https://avatar.test/" + userID[:8], 0
			},
			success: true,
		},
		{
			caseName: "update returns counters",
			f: func(t *testing.T) (*models.Profile, string, string, string, int) {
				userID, follower := newUser(t, pool), newUser(t, pool)
				seedSubscription(t, pool, follower, &userID, nil)
				return &models.Profile{UserID: userID, Bio: strPtr("bio")},
					"user_" + userID[:8], "bio", "https://avatar.test/" + userID[:8], 1
			},
			success: true,
		},
		{
			caseName: "update to taken username",
			f: func(t *testing.T) (*models.Profile, string, string, string, int) {
				userID, other := newUser(t, pool), newUser(t, pool)
				return &models.Profile{UserID: userID, UserName: strPtr("user_" + other[:8])}, "", "", "", 0
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			caseName: "update not found profile",
			f: func(t *testing.T) (*models.Profile, string, string, string, int) {
				return &models.Profile{UserID: uuid.NewString(), Bio: strPtr("bio")}, "", "", "", 0
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "update profile with invalid id",
			f: func(t *testing.T) (*models.Profile, string, string, string, int) {
				return &models.Profile{UserID: "invalid_uuid", Bio: strPtr("bio")}, "", "", "", 0
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			profile, wantName, wantBio, wantAvatar, wantSubscribers := tc.f(t)

			updated, err := repo.BuildUpdateProfile(context.Background(), profile)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				assert.Nil(t, updated)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, profile.UserID, updated.UserID)
			assert.Equal(t, wantName, derefStr(updated.UserName))
			assert.Equal(t, wantBio, derefStr(updated.Bio))
			assert.Equal(t, wantAvatar, derefStr(updated.AvatarUrl))
			assert.Equal(t, wantSubscribers, updated.Subscribers)
		})
	}
}

func TestGetSubscriptions(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewProfileRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (id string, wantIDs []string)
		success  bool
	}{
		{
			caseName: "get subscriptions",
			f: func(t *testing.T) (string, []string) {
				target, a, b, other := newUser(t, pool), newUser(t, pool), newUser(t, pool), newUser(t, pool)
				seedSubscription(t, pool, a, &target, nil)
				seedSubscription(t, pool, b, &target, nil)
				seedSubscription(t, pool, target, &other, nil)
				return target, []string{a, b}
			},
			success: true,
		},
		{
			caseName: "get subscriptions without any",
			f: func(t *testing.T) (string, []string) {
				return newUser(t, pool), []string{}
			},
			success: true,
		},
		{
			caseName: "get subscriptions with invalid id",
			f: func(t *testing.T) (string, []string) {
				return "invalid_uuid", nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			id, wantIDs := tc.f(t)

			subs, err := repo.GetSubscriptions(context.Background(), id)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, subs)
				return
			}
			require.NoError(t, err)
			got := make([]string, 0, len(subs))
			for _, s := range subs {
				got = append(got, s.UserID)
			}
			assert.ElementsMatch(t, wantIDs, got)
		})
	}
}

func TestGetSubscribers(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewProfileRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (id string, wantIDs []string)
		success  bool
	}{
		{
			caseName: "get subscribers",
			f: func(t *testing.T) (string, []string) {
				user, a, b, follower := newUser(t, pool), newUser(t, pool), newUser(t, pool), newUser(t, pool)
				seedSubscription(t, pool, user, &a, nil)
				seedSubscription(t, pool, user, &b, nil)
				seedSubscription(t, pool, follower, &user, nil)
				return user, []string{a, b}
			},
			success: true,
		},
		{
			caseName: "get subscribers ignores topic subscriptions",
			f: func(t *testing.T) (string, []string) {
				user, a := newUser(t, pool), newUser(t, pool)
				topicID, _ := seedTopic(t, pool, a)
				seedSubscription(t, pool, user, &a, nil)
				seedSubscription(t, pool, user, nil, &topicID)
				return user, []string{a}
			},
			success: true,
		},
		{
			caseName: "get subscribers without any",
			f: func(t *testing.T) (string, []string) {
				return newUser(t, pool), []string{}
			},
			success: true,
		},
		{
			caseName: "get subscribers with invalid id",
			f: func(t *testing.T) (string, []string) {
				return "invalid_uuid", nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			id, wantIDs := tc.f(t)

			subs, err := repo.GetSubscribers(context.Background(), id)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, subs)
				return
			}
			require.NoError(t, err)
			got := make([]string, 0, len(subs))
			for _, s := range subs {
				got = append(got, s.TargetID)
			}
			assert.ElementsMatch(t, wantIDs, got)
		})
	}
}

func TestSubscribe(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewProfileRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (userID, targetID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "subscribe",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), newUser(t, pool)
			},
			success: true,
		},
		{
			caseName: "subscribe twice",
			f: func(t *testing.T) (string, string) {
				userID, targetID := newUser(t, pool), newUser(t, pool)
				require.NoError(t, repo.Subscribe(context.Background(), userID, targetID))
				return userID, targetID
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			caseName: "subscribe to not found user",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), uuid.NewString()
			},
		},
		{
			caseName: "subscribe with invalid id",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), "invalid_uuid"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, targetID := tc.f(t)

			err := repo.Subscribe(context.Background(), userID, targetID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, countRows(t, pool,
				`SELECT COUNT(*) FROM subscriptions WHERE user_id = $1 AND target_user_id = $2`, userID, targetID))
		})
	}
}

func TestUnSubscribe(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewProfileRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (userID, targetID string)
		wantLeft int
		wantErr  error
		success  bool
	}{
		{
			caseName: "unsubscribe",
			f: func(t *testing.T) (string, string) {
				userID, targetID := newUser(t, pool), newUser(t, pool)
				seedSubscription(t, pool, userID, &targetID, nil)
				return userID, targetID
			},
			wantLeft: 0,
			success:  true,
		},
		{
			caseName: "unsubscribe keeps other subscriptions",
			f: func(t *testing.T) (string, string) {
				userID, targetID, other := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				seedSubscription(t, pool, userID, &targetID, nil)
				seedSubscription(t, pool, userID, &other, nil)
				return userID, targetID
			},
			wantLeft: 1,
			success:  true,
		},
		{
			caseName: "unsubscribe without subscription",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), newUser(t, pool)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "unsubscribe with invalid id",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), "invalid_uuid"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, targetID := tc.f(t)

			err := repo.UnSubscribe(context.Background(), userID, targetID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 0, countRows(t, pool,
				`SELECT COUNT(*) FROM subscriptions WHERE user_id = $1 AND target_user_id = $2`, userID, targetID))
			assert.Equal(t, tc.wantLeft, countRows(t, pool,
				`SELECT COUNT(*) FROM subscriptions WHERE user_id = $1`, userID))
		})
	}
}

func TestCreateProfileReport(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewProfileRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (userID, targetID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "create profile report",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), newUser(t, pool)
			},
			success: true,
		},
		{
			caseName: "create profile report twice",
			f: func(t *testing.T) (string, string) {
				userID, targetID := newUser(t, pool), newUser(t, pool)
				require.NoError(t, repo.CreateProfileReport(context.Background(), userID, targetID, "spam"))
				return userID, targetID
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			caseName: "create profile report with invalid id",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), "invalid_uuid"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, targetID := tc.f(t)

			err := repo.CreateProfileReport(context.Background(), userID, targetID, "spam")
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, countRows(t, pool,
				`SELECT COUNT(*) FROM profile_reports WHERE user_id = $1 AND target_id = $2`, userID, targetID))
		})
	}
}

func TestBlock(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewProfileRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (userID, targetID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "block",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), newUser(t, pool)
			},
			success: true,
		},
		{
			caseName: "block twice",
			f: func(t *testing.T) (string, string) {
				userID, targetID := newUser(t, pool), newUser(t, pool)
				seedBlock(t, pool, userID, targetID)
				return userID, targetID
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			caseName: "block with invalid id",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), "invalid_uuid"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, targetID := tc.f(t)

			err := repo.Block(context.Background(), userID, targetID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, countRows(t, pool,
				`SELECT COUNT(*) FROM blocks WHERE user_id = $1 AND target_id = $2`, userID, targetID))
		})
	}
}
