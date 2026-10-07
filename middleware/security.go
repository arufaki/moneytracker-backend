package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MaxBodySize limits the request body size.
func MaxBodySize(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}

type ipLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientLimiter
}

type clientLimiter struct {
	count     int
	lastReset time.Time
}

// RateLimiter limits requests per IP for a given window duration.
func RateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	l := &ipLimiter{
		clients: make(map[string]*clientLimiter),
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		l.mu.Lock()
		cli, exists := l.clients[ip]
		if !exists || now.Sub(cli.lastReset) > window {
			l.clients[ip] = &clientLimiter{count: 1, lastReset: now}
			l.mu.Unlock()
			c.Next()
			return
		}

		if cli.count >= limit {
			l.mu.Unlock()
			log.Printf("[SECURITY] Rate limit exceeded for IP: %s on path: %s", ip, c.Request.URL.Path)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		cli.count++
		l.mu.Unlock()
		c.Next()
	}
}
