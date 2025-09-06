// file: internal/handler/budget_handler.go
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/database"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/model"
	"gorm.io/gorm/clause"
)

type BudgetInput struct {
	CategoryID uint    `json:"category_id" binding:"required"`
	Amount     float64 `json:"amount"`
	Month      int     `json:"month" binding:"required"`
	Year       int     `json:"year" binding:"required"`
}

// Handler untuk mendapatkan semua budget di bulan tertentu
func GetBudgets(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(model.User)
	year, _ := strconv.Atoi(c.Query("year"))
	month, _ := strconv.Atoi(c.Query("month"))

	if year == 0 || month == 0 {
		// Jika tidak ada parameter, gunakan bulan saat ini
		now := time.Now()
		year = now.Year()
		month = int(now.Month())
	}

	var budgets []model.Budget
	database.DB.Preload("Category").
		Where("user_id = ? AND year = ? AND month = ?", currentUser.ID, year, month).
		Find(&budgets)

	c.JSON(http.StatusOK, budgets)
}

// Handler untuk menyimpan/memperbarui beberapa budget sekaligus
func SetBudgets(c *gin.Context) {
	var inputs []BudgetInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUser := c.MustGet("currentUser").(model.User)
	var budgetsToUpsert []model.Budget

	for _, input := range inputs {
		budgetsToUpsert = append(budgetsToUpsert, model.Budget{
			UserID:     currentUser.ID,
			CategoryID: input.CategoryID,
			Amount:     input.Amount,
			Month:      input.Month,
			Year:       input.Year,
		})
	}

	// GORM "Upsert": Jika ada, update. Jika tidak ada, buat baru.
	// Kita cocokkan berdasarkan unique index yang kita buat di model.
	err := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "category_id"}, {Name: "month"}, {Name: "year"}},
		DoUpdates: clause.AssignmentColumns([]string{"amount"}),
	}).Create(&budgetsToUpsert).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set budgets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Budgets set successfully"})
}