package controllers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ─── JWKS cache ───────────────────────────────────────────────

type jwksKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

var (
	jwksCache map[string]*ecdsa.PublicKey
	jwksMu    sync.RWMutex
)

func fetchJWKS(supabaseURL string) (map[string]*ecdsa.PublicKey, error) {
	url := strings.TrimRight(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JWKS: %w", err)
	}

	keys := make(map[string]*ecdsa.PublicKey)
	for _, k := range jwks.Keys {
		if k.Kty != "EC" {
			continue
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil {
			continue
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(k.Y)
		if err != nil {
			continue
		}

		var curve elliptic.Curve
		switch k.Crv {
		case "P-256":
			curve = elliptic.P256()
		case "P-384":
			curve = elliptic.P384()
		case "P-521":
			curve = elliptic.P521()
		default:
			continue
		}

		pub := &ecdsa.PublicKey{
			Curve: curve,
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
		keys[k.Kid] = pub
	}

	return keys, nil
}

// RequireAuth valida o token JWT usando as chaves públicas JWKS do Supabase (ES256).
// As chaves são buscadas uma vez e cacheadas em memória.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extrai o header Authorization
		auth := strings.TrimSpace(c.GetHeader("Authorization"))
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// 2. Extrai o token Bearer
		tokenStr := ""
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			tokenStr = strings.TrimSpace(auth[7:])
		}
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		// 3. Obtém chaves JWKS (com cache em memória)
		jwksMu.RLock()
		keys := jwksCache
		jwksMu.RUnlock()

		if keys == nil {
			supabaseURL := os.Getenv("SUPABASE_URL")
			if supabaseURL == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "SUPABASE_URL not configured"})
				c.Abort()
				return
			}
			var err error
			keys, err = fetchJWKS(supabaseURL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch signing keys"})
				c.Abort()
				return
			}
			jwksMu.Lock()
			jwksCache = keys
			jwksMu.Unlock()
		}

		// 4. Valida o token com a chave correta (pelo kid no header do JWT)
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			kid, _ := t.Header["kid"].(string)
			pub, ok := keys[kid]
			if !ok {
				// kid não encontrado — tenta recarregar o JWKS (chave pode ter rotacionado)
				supabaseURL := os.Getenv("SUPABASE_URL")
				newKeys, err := fetchJWKS(supabaseURL)
				if err != nil {
					return nil, fmt.Errorf("falha ao recarregar JWKS: %w", err)
				}
				jwksMu.Lock()
				jwksCache = newKeys
				jwksMu.Unlock()
				pub, ok = newKeys[kid]
				if !ok {
					return nil, fmt.Errorf("chave não encontrada para kid: %s", kid)
				}
			}
			return pub, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// 5. Extrai as claims e salva no context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		// Serializa os claims para JSON — é o formato que o Postgres espera
		// em request.jwt.claims para que auth.uid() funcione corretamente.
		// Antes estava salvando o JWT bruto (base64), que o banco não consegue ler.
		claimsJSON, err := json.Marshal(claims)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize claims"})
			c.Abort()
			return
		}

		c.Set("user_id", claims["sub"])      // UUID do usuário
		c.Set("token", string(claimsJSON))   // JSON dos claims → ativa auth.uid() no RLS
		c.Set("user_token", tokenStr)        // JWT bruto → chamadas à Supabase Auth API
		c.Set("claims", claims)              // claims completas para uso interno

		c.Next()
	}
}