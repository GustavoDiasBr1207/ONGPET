package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RequireAuth valida o token JWT usando o endpoint /auth/v1/user do Supabase
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := strings.TrimSpace(c.GetHeader("Authorization"))
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		token := ""
		lower := strings.ToLower(auth)
		if strings.HasPrefix(lower, "bearer ") {
			token = strings.TrimSpace(auth[7:])
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		supabaseURL := os.Getenv("SUPABASE_URL")
		supabaseKey := os.Getenv("SUPABASE_KEY")
		if supabaseURL == "" || supabaseKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "supabase configuration missing"})
			c.Abort()
			return
		}

		req, err := http.NewRequest("GET", strings.TrimRight(supabaseURL, "/")+"/auth/v1/user", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
			c.Abort()
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("apikey", supabaseKey)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate token"})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// try to decode error body if possible
			var body interface{}
			if err := json.NewDecoder(resp.Body).Decode(&body); err == nil {
				c.JSON(resp.StatusCode, gin.H{"error": "invalid or expired token", "detail": body})
			} else {
				c.JSON(resp.StatusCode, gin.H{"error": "invalid or expired token"})
			}
			c.Abort()
			return
		}

		var user map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user response"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
