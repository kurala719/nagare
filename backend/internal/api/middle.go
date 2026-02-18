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

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := ""
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
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
		} else {
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
			service.LogService("error", "jwt parsing error", map[string]interface{}{"error": err.Error(), "path": c.FullPath()}, nil, c.ClientIP())
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

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b)
}
