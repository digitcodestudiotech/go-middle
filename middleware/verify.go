package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/digitcodestudiotech/go-middle/crypto"
	"github.com/digitcodestudiotech/go-middle/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Options struct {
	PublicKeyURL string
	RefreshEvery time.Duration
}

func VerifyToken() gin.HandlerFunc {

	utils.LoadEnv()

	publicKeyURL := utils.GetEnv("PUBLIC_KEY_URL")
	if publicKeyURL == "" {
		panic("[go-middle] PUBLIC_KEY_URL is required in .env")
	}

	remoteKey, err := crypto.NewRemotePublicKey(publicKeyURL, 5*time.Minute)
	if err != nil {
		panic("[go-middle] failed loading remote public key: " + err.Error())
	}

	return func(c *gin.Context) {

		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		tokenStr := parts[1]

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return remoteKey.Get(), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("claims", claims)

		c.Next()
	}
}
