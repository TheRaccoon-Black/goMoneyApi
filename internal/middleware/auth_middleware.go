package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/database" 
	"github.com/TheRaccoon-Black/goMoneyApi/internal/model"   
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// 2. Cek format "Bearer <token>"
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		tokenString := headerParts[1]

		// Ambil secret key
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "RAHASIA_BANGET_INI_JANGAN_SAMPAI_BOCOR"
		}

		// 3. Parse dan validasi token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 4. Ambil claims dari token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Cek apakah token sudah expired
		exp, ok := claims["exp"].(float64)
		if !ok || float64(time.Now().Unix()) > exp {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			return
		}

		// 5. Cari user di database berdasarkan ID dari token
		userID := claims["sub"]
		var user model.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// 6. Simpan informasi user di context untuk digunakan di handler selanjutnya
		c.Set("currentUser", user)

		// Lanjutkan ke handler berikutnya
		c.Next()
	}
}