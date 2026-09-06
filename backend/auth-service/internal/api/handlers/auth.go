package handlers

import (
	"github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/services"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	logger   *zap.Logger
	oauthCfg *oauth2.Config
	authsrvc *services.AuthService
}

func NewAuthHandler(logger *zap.Logger, authsrvc *services.AuthService, oauthCfg *oauth2.Config) *AuthHandler {
	return &AuthHandler{logger: logger, authsrvc: authsrvc, oauthCfg: oauthCfg}
}
