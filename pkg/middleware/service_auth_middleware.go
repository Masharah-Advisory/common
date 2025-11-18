package middleware

import (
	"net/http"

	"github.com/Masharah-Advisory/common/pkg/i18n"
	"github.com/Masharah-Advisory/common/pkg/response"
	"github.com/Masharah-Advisory/common/pkg/utils"
	"github.com/gin-gonic/gin"
)

// This middleware validates requests from other internal services.
func ServiceAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceSecret := c.GetHeader(utils.XServiceSecretHeader)

		if serviceSecret == "" {
			response.Error(c, http.StatusUnauthorized, i18n.T(c, "missing_service_headers"))
			c.Abort()
			return
		}

		if serviceSecret != utils.ServiceSecret {
			response.Error(c, http.StatusUnauthorized, i18n.T(c, "invalid_service_credentials"))
			c.Abort()
			return
		}

		c.Next()
	}
}
