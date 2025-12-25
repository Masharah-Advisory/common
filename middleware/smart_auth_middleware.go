package middleware

import (
	"strconv"
	"strings"

	"github.com/Masharah-Advisory/common/i18n"
	"github.com/Masharah-Advisory/common/response"
	"github.com/Masharah-Advisory/common/utils"
	"github.com/gin-gonic/gin"
)

// SmartAuthMiddleware automatically detects request source and applies appropriate authentication
func SmartAuthMiddleware(jwtSecret ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if this is an internal service request (has service headers)
		serviceSecret := c.GetHeader(utils.XServiceSecretHeader)

		if serviceSecret != "" {
			// This is an internal service request - validate service auth
			if serviceSecret == utils.ServiceSecret {
				c.Set("authType", "service")
				c.Next()
				return
			} else {
				response.Unauthorized(c, i18n.T(c, "invalid_service_credentials"))
				c.Abort()
				return
			}
		}

		// Check if this has Authorization header (external user request)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// This is an external user request - validate JWT token directly

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
			c.Request.Header.Set(utils.XUserIDHeader, strconv.FormatUint(claims.UserID, 10))
			c.Set("authType", "user")
			c.Next()
			return
		}

		// No authentication headers found
		response.Unauthorized(c, i18n.T(c, "missing_authentication"))
		c.Abort()
	}
}
