package auth

import (
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/lib/structure"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GenerateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil // без '='
}

func ValidateToken(c *gin.Context, cfg *structure.Config) (bool, error) {

	token := c.GetHeader("Authorization")
	if token == "" {
		return false, status.Error(codes.Unauthenticated, "token is empty")
	}

	if !strings.HasPrefix(token, "Bearer ") {
		return false, status.Error(codes.Unauthenticated, "token is invalid")
	}

	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	if token != cfg.TokenSecret {
		return false, status.Error(codes.Unauthenticated, "token is invalid")
	}

	return true, nil
}
