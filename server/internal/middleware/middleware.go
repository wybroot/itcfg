package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Env-Key")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// LoggerMiddleware 请求日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		log.Printf("[%s] %s %s %d %v", method, path, c.ClientIP(), statusCode, latency)
	}
}

// EnvAuthMiddleware 环境认证中间件（用于 Agent 调用）
func EnvAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		envKey := c.GetHeader("X-Env-Key")
		if envKey == "" {
			envKey = c.Query("env_key")
		}
		if envKey != "" {
			c.Set("env_key", envKey)
		}
		c.Next()
	}
}