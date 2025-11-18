package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.GetHeader("X-Request-ID")
		if requestId == "" {
			requestId = generateRequestID()
		}

		c.Set("request_id", requestId)
		c.Header("X-Request-ID", requestId)
		c.Next()
	}
}

// generateRequestID creates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102-150405") + "-" + uuid.New().String()
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// CorsMiddleware configures CORS middleware
func CorsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(corsConfig)
}

// APIKeyAuthMiddleware validates API key for protected endpoints
func APIKeyAuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" {
			c.Next()
			return
		}

		providedKey := c.GetHeader("X-API-Key")
		if providedKey == "" {
			providedKey = c.Query("api_key")
		}

		if providedKey != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or missing API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting per IP
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.RWMutex
		clients = make(map[string]*client)
	)

	// Cleanup old clients every 5 minutes
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 10*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		if _, exists := clients[ip]; !exists {
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMinute)), requestsPerMinute),
			}
		}
		clients[ip].lastSeen = time.Now()
		limiter := clients[ip].limiter
		mu.Unlock()

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TrustedIPMiddleware restricts access to trusted IPs for sensitive endpoints
func TrustedIPMiddleware(trustedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(trustedIPs) == 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()

		// Check if client IP is in trusted list
		trusted := false
		for _, trustedIP := range trustedIPs {
			if trustedIP == clientIP {
				trusted = true
				break
			}
		}

		if !trusted {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: IP not in trusted list",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestSizeLimitMiddleware limits request body size
func RequestSizeLimitMiddleware(maxSizeBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSizeBytes {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request too large",
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSizeBytes)
		c.Next()
	}
}

// ContentTypeValidationMiddleware validates content type for POST/PUT requests
func ContentTypeValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "Content-Type must be application/json",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
