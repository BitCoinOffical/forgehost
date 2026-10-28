package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	store_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUploadImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockUploadRepo(ctrl)
	s3 := store_mocks.NewMockS3(ctrl)

	post := NewUploaderService(repo, s3, testPublicEndpoint)

	file := testUploadFile{strings.NewReader("image")}

	uploadErr := errors.New("upload error")
	createErr := errors.New("create upload error")
	deleteErr := errors.New("delete error")

	var uploadedKey string
	rememberKey := func(_ context.Context, key, _ string, _ io.ReadSeeker) error {
		uploadedKey = key
		return nil
	}
	sameKey := gomock.Cond(func(key string) bool { return key == uploadedKey })

	newKey := func(prefix, ext string) gomock.Matcher {
		return gomock.Cond(func(key string) bool {
			return strings.HasPrefix(key, prefix) && strings.HasSuffix(key, ext)
		})
	}

	type args struct {
		prefix      string
		userID      string
		ext         string
		contentType string
	}

	testCases := []struct {
		caseName string
		data     args
		prepare  func(key gomock.Matcher)
		wantErrs []error
		success  bool
	}{
		{
			caseName: "upload post image",
			data:     args{prefix: PrefixPosts, userID: "user_id", ext: ".png", contentType: "image/png"},
			prepare: func(key gomock.Matcher) {
				gomock.InOrder(
					s3.EXPECT().Upload(gomock.Any(), key, "image/png", gomock.Eq(file)).
						DoAndReturn(rememberKey).Times(1),
					repo.EXPECT().CreateUpload(gomock.Any(), sameKey, "user_id").
						Return(nil).Times(1),
				)
			},
			success: true,
		},
		{
			caseName: "upload avatar image",
			data:     args{prefix: PrefixAvatars, userID: "user_id", ext: ".jpg", contentType: "image/jpeg"},
			prepare: func(key gomock.Matcher) {
				gomock.InOrder(
					s3.EXPECT().Upload(gomock.Any(), key, "image/jpeg", gomock.Eq(file)).
						DoAndReturn(rememberKey).Times(1),
					repo.EXPECT().CreateUpload(gomock.Any(), sameKey, "user_id").
						Return(nil).Times(1),
				)
			},
			success: true,
		},
		{
			caseName: "upload to storage error",
			data:     args{prefix: PrefixPosts, userID: "user_id", ext: ".png", contentType: "image/png"},
			prepare: func(key gomock.Matcher) {
				s3.EXPECT().Upload(gomock.Any(), key, "image/png", gomock.Eq(file)).
					Return(uploadErr).Times(1)
			},
			wantErrs: []error{uploadErr},
		},
		{
			caseName: "create upload error removes uploaded file",
			data:     args{prefix: PrefixPosts, userID: "user_id", ext: ".png", contentType: "image/png"},
			prepare: func(key gomock.Matcher) {
				gomock.InOrder(
					s3.EXPECT().Upload(gomock.Any(), key, "image/png", gomock.Eq(file)).
						DoAndReturn(rememberKey).Times(1),
					repo.EXPECT().CreateUpload(gomock.Any(), sameKey, "user_id").
						Return(createErr).Times(1),
					s3.EXPECT().Delete(gomock.Any(), sameKey).
						Return(nil).Times(1),
				)
			},
			wantErrs: []error{createErr},
		},
		{
			caseName: "create upload and delete errors",
			data:     args{prefix: PrefixPosts, userID: "user_id", ext: ".png", contentType: "image/png"},
			prepare: func(key gomock.Matcher) {
				gomock.InOrder(
					s3.EXPECT().Upload(gomock.Any(), key, "image/png", gomock.Eq(file)).
						DoAndReturn(rememberKey).Times(1),
					repo.EXPECT().CreateUpload(gomock.Any(), sameKey, "user_id").
						Return(createErr).Times(1),
					s3.EXPECT().Delete(gomock.Any(), sameKey).
						Return(deleteErr).Times(1),
				)
			},
			wantErrs: []error{createErr, deleteErr},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			uploadedKey = ""
			tc.prepare(newKey(tc.data.prefix, tc.data.ext))

			got, err := post.UploadImage(t.Context(), tc.data.prefix, tc.data.userID, tc.data.ext, tc.data.contentType, file)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, uploadedKey, got)

				require.True(t, strings.HasPrefix(got, tc.data.prefix))
				require.True(t, strings.HasSuffix(got, tc.data.ext))
				id := strings.TrimSuffix(strings.TrimPrefix(got, tc.data.prefix), tc.data.ext)
				_, parseErr := uuid.Parse(id)
				assert.NoError(t, parseErr)
			} else {
				require.Error(t, err)
				for _, wantErr := range tc.wantErrs {
					require.ErrorIs(t, err, wantErr)
				}
				assert.Empty(t, got)
			}
		})
	}
}
