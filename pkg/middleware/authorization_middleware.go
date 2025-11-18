package middleware

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Masharah-Advisory/common/pkg/httpclient"
	"github.com/Masharah-Advisory/common/pkg/i18n"
	"github.com/Masharah-Advisory/common/pkg/response"
	"github.com/Masharah-Advisory/common/pkg/utils"
	"github.com/gin-gonic/gin"
)

type AccessResponse struct {
	Success bool       `json:"success"`
	Data    AccessData `json:"data"`
	Message string     `json:"message"`
}

type AccessData struct {
	Allowed bool `json:"allowed"`
}

// RequirePermission validates that user has a specific permission (user-only middleware)
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (should be set by AuthMiddleware)
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

		// Call auth service to check access
		allowed, err := checkUserPermission(uid, permission)
		fmt.Println(err.Error())
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
	}
}

// RequirePermissions validates that user has all specified permissions (user-only middleware)
func RequirePermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
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
			allowed, err := checkUserPermission(uid, permission)
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
	}
}

// checkUserPermission calls auth service to validate user permission
func checkUserPermission(userID uint, permission string) (bool, error) {
	payload := map[string]interface{}{
		"user_id":    userID,
		"permission": permission,
	}

	headers := map[string]string{
		utils.XServiceIDHeader:     utils.ServiceID,
		utils.XServiceSecretHeader: utils.ServiceSecret,
	}

	resp, err := httpclient.PostJSON(utils.AuthServiceURL+"/api/v1/auth/access", payload, headers)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var accessResp AccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&accessResp); err != nil {
		return false, err
	}

	if !accessResp.Success {
		return false, nil
	}

	return accessResp.Data.Allowed, nil
}
