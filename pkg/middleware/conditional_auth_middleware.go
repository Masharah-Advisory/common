package middleware

import (
	"strconv"

	"github.com/Masharah-Advisory/common/pkg/i18n"
	"github.com/Masharah-Advisory/common/pkg/response"
	"github.com/gin-gonic/gin"
)

// PermissionMiddleware checks permissions only for user requests,
// allows service requests to bypass permission checks
func PermissionMiddleware(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authType, exists := c.Get("authType")
		if !exists {
			response.Unauthorized(c, i18n.T(c, "authentication_required"))
			c.Abort()
			return
		}

		// If service request, allow access without permission check
		if authType == "service" {
			c.Next()
			return
		}

		// If user request, check permission
		if authType == "user" {
			userID, exists := c.Get("user_id")
			if !exists {
				response.Unauthorized(c, i18n.T(c, "user_id_not_found"))
				c.Abort()
				return
			}

			// Convert userID to uint
			var uid uint
			switch v := userID.(type) {
			case uint:
				uid = v
			case int:
				uid = uint(v)
			case string:
				parsed, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					response.Unauthorized(c, i18n.T(c, "invalid_user_id_format"))
					c.Abort()
					return
				}
				uid = uint(parsed)
			default:
				response.Unauthorized(c, i18n.T(c, "invalid_user_id_type"))
				c.Abort()
				return
			}

			// Check permission via auth service
			allowed, err := checkUserPermission(c, uid, permission)
			if err != nil {
				response.InternalError(c, i18n.T(c, "failed_to_validate_permissions"))
				c.Abort()
				return
			}

			if !allowed {
				response.Forbidden(c, i18n.T(c, "insufficient_permissions")+": "+permission)
				c.Abort()
				return
			}

			c.Next()
			return
		}

		response.Unauthorized(c, i18n.T(c, "invalid_authentication_type"))
		c.Abort()
	}
}

// PermissionAnyMiddleware checks multiple permissions only for user requests
// For future use when multiple permissions are needed
func PermissionAnyMiddleware(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authType, exists := c.Get("authType")
		if !exists {
			response.Unauthorized(c, i18n.T(c, "authentication_required"))
			c.Abort()
			return
		}

		// If service request, allow access without permission check
		if authType == "service" {
			c.Next()
			return
		}

		// If user request, check all permissions
		if authType == "user" {
			userID, exists := c.Get("user_id")
			if !exists {
				response.Unauthorized(c, i18n.T(c, "user_id_not_found"))
				c.Abort()
				return
			}

			// Convert userID to uint
			var uid uint
			switch v := userID.(type) {
			case uint:
				uid = v
			case int:
				uid = uint(v)
			case string:
				parsed, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					response.Unauthorized(c, i18n.T(c, "invalid_user_id_format"))
					c.Abort()
					return
				}
				uid = uint(parsed)
			default:
				response.Unauthorized(c, i18n.T(c, "invalid_user_id_type"))
				c.Abort()
				return
			}

			// Check all permissions
			for _, permission := range permissions {
				allowed, err := checkUserPermission(c, uid, permission)
				if err != nil {
					response.InternalError(c, i18n.T(c, "failed_to_validate_permissions"))
					c.Abort()
					return
				}

				if !allowed {
					response.Forbidden(c, i18n.T(c, "insufficient_permissions")+": "+permission)
					c.Abort()
					return
				}
			}

			c.Next()
			return
		}

		response.Unauthorized(c, i18n.T(c, "invalid_authentication_type"))
		c.Abort()
	}
}
