package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/database" 
	"github.com/TheRaccoon-Black/goMoneyApi/internal/model"    
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := model.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID, // Subject (siapa pemilik token)
		"exp": time.Now().Add(time.Hour * 2 * 1).Unix(), //kali 2 jam kali satu hari
	})

	// Ambil secret key dari environment variable (untuk keamanan)
	// Untuk sekarang, kita hardcode dulu
	jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "RAHASIA_BANGET_INI_JANGAN_SAMPAI_BOCOR" // Ganti di produksi nanti!
    }

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// Kirim token sebagai jawaban
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": tokenString})
}

func GetCurrentUserProfile(c *gin.Context) {
	// Ambil data user yang sudah disimpan oleh middleware di context
	user, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Kirim data user sebagai response (pastikan tidak mengirim password hash)
	// Kita bisa membuat struct terpisah untuk response agar lebih aman
	userResponse := gin.H{
		"id":    user.(model.User).ID,
		"name":  user.(model.User).Name,
		"email": user.(model.User).Email,
	}

	c.JSON(http.StatusOK, userResponse)
}