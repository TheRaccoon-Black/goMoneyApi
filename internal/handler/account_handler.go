package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/database"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/model"    
)

type AccountInput struct {
	Name    string  `json:"name" binding:"required"`
	Balance float64 `json:"balance"` 
}

// --- Handler untuk Membuat Akun Baru ---
func CreateAccount(c *gin.Context) {
	var input AccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ambil user dari context yang sudah di-set oleh middleware
	currentUser, _ := c.Get("currentUser")
	user := currentUser.(model.User)

	account := model.Account{
		UserID:  user.ID,
		Name:    input.Name,
		Balance: input.Balance,
	}

	if err := database.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// --- Handler untuk Mendapatkan Semua Akun ---
func GetAccounts(c *gin.Context) {
	currentUser, _ := c.Get("currentUser")
	user := currentUser.(model.User)

	var accounts []model.Account
	if err := database.DB.Preload("User").Where("user_id = ?", user.ID).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve accounts"})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// --- Handler untuk Mengupdate Akun ---
func UpdateAccount(c *gin.Context) {
    // Ambil ID dari URL
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    // Cari akun yang ada di DB
    var account model.Account
    if err := database.DB.First(&account, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
        return
    }

    // Cek kepemilikan
    currentUser, _ := c.Get("currentUser")
    user := currentUser.(model.User)
    if account.UserID != user.ID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this account"})
        return
    }

    // Bind data update
    var input AccountInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Update dan simpan
    account.Name = input.Name
    account.Balance = input.Balance
    database.DB.Save(&account)

    c.JSON(http.StatusOK, account)
}


// --- Handler untuk Menghapus Akun ---
func DeleteAccount(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    var account model.Account
    if err := database.DB.First(&account, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
        return
    }

    currentUser, _ := c.Get("currentUser")
    user := currentUser.(model.User)
    if account.UserID != user.ID {
        c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this account"})
        return
    }

    database.DB.Delete(&account)

    c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}