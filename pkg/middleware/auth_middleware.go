package middleware

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Masharah-Advisory/common/pkg/i18n"
	"github.com/Masharah-Advisory/common/pkg/response"
	"github.com/Masharah-Advisory/common/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthResponse struct {
	Success bool     `json:"success"`
	Data    AuthData `json:"data"`
	Message string   `json:"message"`
}

type AuthData struct {
	UserID uint `json:"user_id"`
}

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT token locally and adds user_id to header and context
func AuthMiddleware(jwtSecret ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, i18n.T(c, "missing_authorization_header"))
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			response.Unauthorized(c, i18n.T(c, "invalid_authorization_format"))
			c.Abort()
			return
		}

		// Use provided JWT secret or fallback to global one
		secret := utils.JWTSecret
		if len(jwtSecret) > 0 && jwtSecret[0] != "" {
			secret = jwtSecret[0]
		}

		if secret == "" {
			response.InternalError(c, i18n.T(c, "jwt_secret_not_configured"))
			c.Abort()
			return
		}

		// Parse and validate JWT token locally
		claims, err := parseJWTToken(tokenString, secret)
		if err != nil {
			response.Unauthorized(c, i18n.T(c, "invalid_or_expired_token"))
			c.Abort()
			return
		}

		// Set user ID in context and header for downstream services
		c.Set("user_id", claims.UserID)
		c.Request.Header.Set(utils.XUserIDHeader, strconv.FormatUint(uint64(claims.UserID), 10))
		fmt.Println("hello123", claims.UserID)
		c.Next()
	}
}

// parseJWTToken parses and validates JWT token locally
func parseJWTToken(tokenString, jwtSecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token's signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}
