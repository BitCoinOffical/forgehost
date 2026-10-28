package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestUploadPostImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pngData := []byte("\x89PNG\r\n\x1a\n")
	jpegData := []byte{0xff, 0xd8, 0xff}
	webpData := []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")

	largeImage := make([]byte, maxImageSize+1)
	copy(largeImage, pngData)

	testCases := []struct {
		caseName        string
		data            uploadImageInput
		wantStatus      int
		wantImageID     string
		wantExt         string
		wantContentType string
		serviceError    bool
	}{
		{
			caseName: "upload png",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.png",
				file:     pngData,
			},
			wantStatus:      http.StatusCreated,
			wantImageID:     "post/123.png",
			wantExt:         ".png",
			wantContentType: "image/png",
		},
		{
			caseName: "upload jpeg",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.jpg",
				file:     jpegData,
			},
			wantStatus:      http.StatusCreated,
			wantImageID:     "post/123.png",
			wantExt:         ".jpg",
			wantContentType: "image/jpeg",
		},
		{
			caseName: "upload webp",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.webp",
				file:     webpData,
			},
			wantStatus:      http.StatusCreated,
			wantImageID:     "post/123.png",
			wantExt:         ".webp",
			wantContentType: "image/webp",
		},
		{
			caseName: "missing image",
			data: uploadImageInput{
				userID: userClaims("user-id"),
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			caseName: "unsupported image format",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.gif",
				file:     []byte("GIF89a"),
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			caseName: "image too large",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.png",
				file:     largeImage,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			caseName: "upload service error",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.png",
				file:     pngData,
			},
			wantStatus:      http.StatusInternalServerError,
			wantExt:         ".png",
			wantContentType: "image/png",
			serviceError:    true,
		},
		{
			caseName: "unauthorized",
			data: uploadImageInput{
				file: pngData,
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockUploaderService(ctrl)

			if tc.wantExt != "" {
				call := srvc.EXPECT().
					UploadImage(
						gomock.Any(),
						"posts/",
						"user-id",
						tc.wantExt,
						tc.wantContentType,
						gomock.Any(),
					).
					Times(1)

				if tc.serviceError {
					call.Return("", errors.New("upload image error"))
				} else {
					call.DoAndReturn(func(
						_ context.Context,
						prefix string,
						userID string,
						ext string,
						contentType string,
						file io.Reader,
					) (string, error) {
						assert.Equal(t, "user-id", userID)
						assert.Equal(t, tc.wantExt, ext)
						assert.Equal(t, tc.wantContentType, contentType)

						uploaded, err := io.ReadAll(file)
						require.NoError(t, err)
						assert.Equal(t, tc.data.file, uploaded)

						return tc.wantImageID, nil
					})
				}
			}

			handler := NewUploaderHandler(srvc, zaptest.NewLogger(t))
			w := serveUploadPostImage(t, handler, tc.data)

			assert.Equal(t, tc.wantStatus, w.Code)

			if tc.wantImageID != "" {
				assert.Contains(t, w.Body.String(), tc.wantImageID)
			}
		})
	}
}

func TestUploadAvatarImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pngData := []byte("\x89PNG\r\n\x1a\n")
	jpegData := []byte{0xff, 0xd8, 0xff}
	webpData := []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")

	largeImage := make([]byte, maxImageSize+1)
	copy(largeImage, pngData)

	testCases := []struct {
		caseName        string
		data            uploadImageInput
		wantStatus      int
		wantImageID     string
		wantExt         string
		wantContentType string
		serviceError    bool
	}{
		{
			caseName: "upload png",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.png",
				file:     pngData,
			},
			wantStatus:      http.StatusCreated,
			wantImageID:     "avatar/123.png",
			wantExt:         ".png",
			wantContentType: "image/png",
		},
		{
			caseName: "upload jpeg",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.jpg",
				file:     jpegData,
			},
			wantStatus:      http.StatusCreated,
			wantImageID:     "post/123.png",
			wantExt:         ".jpg",
			wantContentType: "image/jpeg",
		},
		{
			caseName: "upload webp",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.webp",
				file:     webpData,
			},
			wantStatus:      http.StatusCreated,
			wantImageID:     "post/123.png",
			wantExt:         ".webp",
			wantContentType: "image/webp",
		},
		{
			caseName: "missing image",
			data: uploadImageInput{
				userID: userClaims("user-id"),
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			caseName: "unsupported image format",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.gif",
				file:     []byte("GIF89a"),
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			caseName: "image too large",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.png",
				file:     largeImage,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			caseName: "upload service error",
			data: uploadImageInput{
				userID:   userClaims("user-id"),
				filename: "image.png",
				file:     pngData,
			},
			wantStatus:      http.StatusInternalServerError,
			wantExt:         ".png",
			wantContentType: "image/png",
			serviceError:    true,
		},
		{
			caseName: "unauthorized",
			data: uploadImageInput{
				file: pngData,
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockUploaderService(ctrl)

			if tc.wantExt != "" {
				call := srvc.EXPECT().
					UploadImage(
						gomock.Any(),
						"avatars/",
						"user-id",
						tc.wantExt,
						tc.wantContentType,
						gomock.Any(),
					).
					Times(1)

				if tc.serviceError {
					call.Return("", errors.New("upload image error"))
				} else {
					call.DoAndReturn(func(
						_ context.Context,
						prefix string,
						userID string,
						ext string,
						contentType string,
						file io.Reader,
					) (string, error) {
						assert.Equal(t, "user-id", userID)
						assert.Equal(t, tc.wantExt, ext)
						assert.Equal(t, tc.wantContentType, contentType)

						uploaded, err := io.ReadAll(file)
						require.NoError(t, err)
						assert.Equal(t, tc.data.file, uploaded)

						return tc.wantImageID, nil
					})
				}
			}

			handler := NewUploaderHandler(srvc, zaptest.NewLogger(t))
			w := serveAvatarPostImage(t, handler, tc.data)

			assert.Equal(t, tc.wantStatus, w.Code)

			if tc.wantImageID != "" {
				assert.Contains(t, w.Body.String(), tc.wantImageID)
			}
		})
	}
}
