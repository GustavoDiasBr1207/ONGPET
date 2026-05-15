package controllers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipEntry struct {
	count     int
	windowEnd time.Time
}

type ipRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
	limit   int
	window  time.Duration
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	rl := &ipRateLimiter{
		entries: make(map[string]*ipEntry),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	e, ok := rl.entries[ip]
	if !ok || now.After(e.windowEnd) {
		rl.entries[ip] = &ipEntry{count: 1, windowEnd: now.Add(rl.window)}
		return true
	}
	if e.count >= rl.limit {
		return false
	}
	e.count++
	return true
}

// cleanup remove entradas expiradas a cada 5 minutos.
func (rl *ipRateLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		rl.mu.Lock()
		now := time.Now()
		for ip, e := range rl.entries {
			if now.After(e.windowEnd) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit retorna um middleware Gin que limita `limit` requisições por `window` por IP.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newIPRateLimiter(limit, window)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "muitas requisições, tente novamente em instantes",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
