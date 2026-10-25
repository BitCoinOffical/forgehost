// profile_test.go
package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers/mock"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestMe(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockProfileService(ctrl)

	profile := NewProfileHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "get my profile",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetProfileByID(gomock.Any(), testUserID).
					Return(&models.ProfileResponse{Profile: models.Profile{UserID: testUserID}}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "my profile not found",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetProfileByID(gomock.Any(), testUserID).
					Return(nil, fmt.Errorf("s.repo.GetProfileByID: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "get my profile error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().GetProfileByID(gomock.Any(), testUserID).
					Return(nil, errors.New("get profile error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/profile/me", "/profile/me", tc.data, profile.Me)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestGetProfileByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockProfileService(ctrl)

	profile := NewProfileHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "get profile by id",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetProfileByID(gomock.Any(), testTargetID).
					Return(&models.ProfileResponse{Profile: models.Profile{UserID: testTargetID}}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "get not found profile",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetProfileByID(gomock.Any(), testTargetID).
					Return(nil, fmt.Errorf("s.repo.GetProfileByID: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "get profile by id error",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetProfileByID(gomock.Any(), testTargetID).
					Return(nil, errors.New("get profile error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/profile/:user_id", "/profile/"+testTargetID, tc.data, profile.GetProfileByID)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockProfileService(ctrl)

	profile := NewProfileHandler(srvc, zaptest.NewLogger(t))

	body := `{"username":"name","bio":"bio","avatar_url":"https://avatar.test/1.png"}`

	testCases := []chatCase{
		{
			caseName: "update profile",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().UpdateProfile(gomock.Any(),
					&dto.UpdateProfileDTO{UserName: strPtr("name"), Bio: strPtr("bio"), AvatarUrl: strPtr("https://avatar.test/1.png")}, testUserID).
					Return(&models.Profile{UserID: testUserID}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{userID: userClaims(testUserID), body: `{invalid`},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "update profile with taken username",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().UpdateProfile(gomock.Any(), gomock.Any(), testUserID).
					Return(nil, fmt.Errorf("s.repo.BuildUpdateProfile: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "update not found profile",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().UpdateProfile(gomock.Any(), gomock.Any(), testUserID).
					Return(nil, fmt.Errorf("s.repo.BuildUpdateProfile: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "update profile error",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().UpdateProfile(gomock.Any(), gomock.Any(), testUserID).
					Return(nil, errors.New("update profile error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPatch, "/profile", "/profile", tc.data, profile.UpdateProfile)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestGetSubscribers(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockProfileService(ctrl)

	profile := NewProfileHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "get subscribers",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetSubscribers(gomock.Any(), testTargetID).
					Return([]models.Subscribes{{TargetID: "target-1"}, {TargetID: "target-2"}}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "get subscribers error",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetSubscribers(gomock.Any(), testTargetID).
					Return(nil, errors.New("get subscribers error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/profile/:user_id/subscribers", "/profile/"+testTargetID+"/subscribers", tc.data, profile.GetSubscribers)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestGetSubscriptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockProfileService(ctrl)

	profile := NewProfileHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "get subscriptions",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetSubscriptions(gomock.Any(), testTargetID).
					Return([]models.Subscriptions{{UserID: "user-1"}, {UserID: "user-2"}}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "get subscriptions error",
			data:     chatInput{},
			prepare: func() {
				srvc.EXPECT().GetSubscriptions(gomock.Any(), testTargetID).
					Return(nil, errors.New("get subscriptions error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/profile/:user_id/subscriptions", "/profile/"+testTargetID+"/subscriptions", tc.data, profile.GetSubscriptions)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockProfileService(ctrl)

	profile := NewProfileHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "subscribe",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().Subscribe(gomock.Any(), testUserID, testTargetID).
					Return(nil).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "subscribe to self",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().Subscribe(gomock.Any(), testUserID, testTargetID).
					Return(domain.ErrSubscribeToSelf).Times(1)
			},
			status: http.StatusBadRequest,
		},
		{
			caseName: "subscribe twice",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().Subscribe(gomock.Any(), testUserID, testTargetID).
					Return(fmt.Errorf("s.repo.Subscribe: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "subscribe error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().Subscribe(gomock.Any(), testUserID, testTargetID).
					Return(errors.New("subscribe error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, "/profile/:user_id/subscribe", "/profile/"+testTargetID+"/subscribe", tc.data, profile.Subscribe)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUnsubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockProfileService(ctrl)

	profile := NewProfileHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "unsubscribe",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UnSubscribe(gomock.Any(), testUserID, testTargetID).
					Return(nil).Times(1)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "unsubscribe without subscription",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UnSubscribe(gomock.Any(), testUserID, testTargetID).
					Return(fmt.Errorf("s.repo.UnSubscribe: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "unsubscribe error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UnSubscribe(gomock.Any(), testUserID, testTargetID).
					Return(errors.New("unsubscribe error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodDelete, "/profile/:user_id/unsubscribe", "/profile/"+testTargetID+"/unsubscribe", tc.data, profile.Unsubscribe)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestReportProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockProfileService(ctrl)

	profile := NewProfileHandler(srvc, zaptest.NewLogger(t))

	body := `{"cause":"spam"}`

	testCases := []chatCase{
		{
			caseName: "report profile",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().Report(gomock.Any(), testUserID, testTargetID, &dto.ReportDTO{Cause: "spam"}).
					Return(nil).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{userID: userClaims(testUserID), body: `{invalid`},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "report profile twice",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().Report(gomock.Any(), testUserID, testTargetID, gomock.Any()).
					Return(fmt.Errorf("r.repo.CreateProfileReport: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "report profile error",
			data:     chatInput{userID: userClaims(testUserID), body: body},
			prepare: func() {
				srvc.EXPECT().Report(gomock.Any(), testUserID, testTargetID, gomock.Any()).
					Return(errors.New("report profile error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, "/profile/:user_id/report", "/profile/"+testTargetID+"/report", tc.data, profile.Report)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestBlockUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockProfileService(ctrl)

	profile := NewProfileHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "block user",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().Block(gomock.Any(), testUserID, testTargetID).
					Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "block user twice",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().Block(gomock.Any(), testUserID, testTargetID).
					Return(fmt.Errorf("r.repo.Block: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "block user error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().Block(gomock.Any(), testUserID, testTargetID).
					Return(errors.New("block user error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, "/profile/:user_id/block", "/profile/"+testTargetID+"/block", tc.data, profile.Block)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}
