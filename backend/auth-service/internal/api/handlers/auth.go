package handlers

import (
	"github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/services"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var authRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "auth_requests_total",
	Help: "Total number of auth requests",
})

type AuthHandler struct {
	logger   *zap.Logger
	oauthCfg *oauth2.Config
	authsrvc *services.AuthService
}

func NewAuthHandler(logger *zap.Logger, authsrvc *services.AuthService, oauthCfg *oauth2.Config) *AuthHandler {
	return &AuthHandler{logger: logger, authsrvc: authsrvc, oauthCfg: oauthCfg}
}
