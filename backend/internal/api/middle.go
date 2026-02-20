package api

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"nagare/internal/service"
	"nagare/internal/model"
)

func isQueryTokenAllowed(path string) bool {
	return strings.HasSuffix(path, "/ws") ||
		strings.HasSuffix(path, "/ssh") ||
		strings.HasSuffix(path, "/download")
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := ""
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else if isQueryTokenAllowed(c.Request.URL.Path) {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			respondError(c, model.ErrUnauthorized)
			c.Abort()
			return
		}
		key := []byte(viper.GetString("jwt.secret_key"))
		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})
		if err != nil {
			service.LogService("error", "jwt parsing error", map[string]interface{}{"error": err.Error(), "path": c.FullPath()}, nil, c.ClientIP())
			respondError(c, model.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

func PrivilegesMiddleware(requiredPrivileges int) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := ""
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else if isQueryTokenAllowed(c.Request.URL.Path) {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			service.LogService("warn", "missing authorization header or token query", map[string]interface{}{"path": c.FullPath()}, nil, c.ClientIP())
			respondError(c, model.ErrUnauthorized)
			c.Abort()
			return
		}
		key := []byte(viper.GetString("jwt.secret_key"))
		token, err := jwt.ParseWithClaims(tokenString, &service.CustomedClaims{}, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})
		if err != nil {
			service.LogService("error", "jwt parsing error", map[string]interface{}{"error": err.Error(), "path": c.FullPath(), "token_len": len(tokenString)}, nil, c.ClientIP())
			respondError(c, model.ErrUnauthorized)
			c.Abort()
			return
		}
		if claims, ok := token.Claims.(*service.CustomedClaims); ok && token.Valid {
			if claims.Privileges < requiredPrivileges {
				service.LogService("warn", "insufficient privileges", map[string]interface{}{"path": c.FullPath(), "required": requiredPrivileges, "privileges": claims.Privileges}, nil, c.ClientIP())
				respondError(c, model.ErrForbidden)
				c.Abort()
				return
			}
			c.Set("uid", claims.UID)
			c.Set("username", claims.Username)
			c.Set("privileges", claims.Privileges)
			c.Next()
		} else {
			service.LogService("error", "invalid jwt claims", map[string]interface{}{"path": c.FullPath()}, nil, c.ClientIP())
			respondError(c, model.ErrUnauthorized)
			c.Abort()
			return
		}
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" {
			requestID = newRequestID()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func AccessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start).Milliseconds()
		requestID, _ := c.Get("request_id")
		username, _ := c.Get("username")

		context := map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.FullPath(),
			"status":      c.Writer.Status(),
			"duration_ms": duration,
			"request_id":  requestID,
			"query":       c.Request.URL.RawQuery,
		}
		if username != nil {
			context["username"] = username
		}

		service.LogService("info", "api access", context, nil, c.ClientIP())
	}
}

func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		// Only log data-modifying operations
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		latency := time.Since(start).Microseconds()

		uidValue, existsUid := c.Get("uid")
		usernameValue, existsUsername := c.Get("username")

		var uid uint
		var username string
		if existsUid {
			uid = uidValue.(uint)
		}
		if existsUsername {
			username = usernameValue.(string)
		}

		entry := model.AuditLog{
			UserID:    uid,
			Username:  username,
			Action:    getActionDescription(c),
			Method:    method,
			Path:      c.FullPath(),
			IP:        c.ClientIP(),
			Status:    c.Writer.Status(),
			Latency:   latency,
			UserAgent: c.Request.UserAgent(),
		}

		// Save asynchronously to not block the response
		go func(e model.AuditLog) {
			_ = service.AddAuditLogServ(e)
		}(entry)
	}
}

func getActionDescription(c *gin.Context) string {
	path := c.FullPath()
	method := c.Request.Method

	// Simple heuristic for action description
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 0 {
		resource := parts[len(parts)-1]
		// If the last part is a parameter (like :id), use the second to last
		if strings.HasPrefix(resource, ":") && len(parts) > 1 {
			resource = parts[len(parts)-2]
		}

		switch method {
		case "POST":
			return "Create " + resource
		case "PUT", "PATCH":
			return "Update " + resource
		case "DELETE":
			return "Delete " + resource
		}
	}
	return method + " " + path
}

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b)
}
