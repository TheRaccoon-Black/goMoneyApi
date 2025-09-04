package handler

import (
	"errors"
	"net/http"
	"time"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/database" 
	"github.com/TheRaccoon-Black/goMoneyApi/internal/model"   
	"gorm.io/gorm"
)
type TransactionInput struct {
	AccountID            uint      `json:"account_id" binding:"required"`
	SubCategoryID        *uint     `json:"sub_category_id"`
	Amount               float64   `json:"amount" binding:"required,gt=0"`
	Type                 string    `json:"type" binding:"required,oneof=expense income transfer"`
	Notes                string    `json:"notes"`
	TransactionDate      time.Time `json:"transaction_date" binding:"required"`
	DestinationAccountID *uint     `json:"destination_account_id"`
}

func CreateTransaction(c *gin.Context) {
	var input TransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currentUser := c.MustGet("currentUser").(model.User)
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var sourceAccount model.Account
		if err := tx.Where("id = ? AND user_id = ?", input.AccountID, currentUser.ID).First(&sourceAccount).Error; err != nil {
			return errors.New("source account not found")
		}
		transaction := model.Transaction{
			UserID:          currentUser.ID,
			AccountID:       input.AccountID,
			SubCategoryID:   input.SubCategoryID,
			Amount:          input.Amount,
			Type:            input.Type,
			Notes:           input.Notes,
			TransactionDate: input.TransactionDate,
		}
		switch input.Type {
		case "expense":
			if input.SubCategoryID == nil {
				return errors.New("sub_category_id is required for expenses")
			}
			sourceAccount.Balance -= input.Amount
			if err := tx.Save(&sourceAccount).Error; err != nil {
				return err
			}
		case "income":
			if input.SubCategoryID == nil {
				return errors.New("sub_category_id is required for income")
			}
			sourceAccount.Balance += input.Amount
			if err := tx.Save(&sourceAccount).Error; err != nil {
				return err
			}
		case "transfer":
			if input.DestinationAccountID == nil {
				return errors.New("destination_account_id is required for transfers")
			}
			if input.AccountID == *input.DestinationAccountID {
				return errors.New("source and destination accounts cannot be the same")
			}
			var destinationAccount model.Account
			if err := tx.Where("id = ? AND user_id = ?", *input.DestinationAccountID, currentUser.ID).First(&destinationAccount).Error; err != nil {
				return errors.New("destination account not found")
			}
			sourceAccount.Balance -= input.Amount
			destinationAccount.Balance += input.Amount
			transaction.DestinationAccountID = input.DestinationAccountID
			if err := tx.Save(&sourceAccount).Error; err != nil {
				return err
			}
			if err := tx.Save(&destinationAccount).Error; err != nil {
				return err
			}
		default:
			return errors.New("invalid transaction type")
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transaction created successfully"})
}

func GetTransactions(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(model.User)
	query := database.DB.Preload("Account").
		Preload("SubCategory").
		Preload("SubCategory.Category").
		Where("user_id = ?", currentUser.ID).
		Order("transaction_date desc")
	var transactions []model.Transaction
	if err := query.Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

func GetTransactionByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	currentUser := c.MustGet("currentUser").(model.User)
	var transaction model.Transaction
	query := database.DB.Preload("Account").
		Preload("SubCategory").
		Preload("SubCategory.Category").
		Where("id = ? AND user_id = ?", id, currentUser.ID).
		First(&transaction)
	if err := query.Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func DeleteTransaction(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	currentUser := c.MustGet("currentUser").(model.User)
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var transaction model.Transaction
		if err := tx.Where("id = ? AND user_id = ?", id, currentUser.ID).First(&transaction).Error; err != nil {
			return errors.New("transaction not found")
		}
		var sourceAccount model.Account
		if err := tx.First(&sourceAccount, transaction.AccountID).Error; err != nil {
			return errors.New("source account not found")
		}
		switch transaction.Type {
		case "expense":
			sourceAccount.Balance += transaction.Amount
		case "income":
			sourceAccount.Balance -= transaction.Amount
		case "transfer":
			var destAccount model.Account
			if err := tx.First(&destAccount, transaction.DestinationAccountID).Error; err != nil {
				return errors.New("destination account not found for reversal")
			}
			sourceAccount.Balance += transaction.Amount
			destAccount.Balance -= transaction.Amount
			if err := tx.Save(&destAccount).Error; err != nil {
				return err
			}
		}
		if err := tx.Save(&sourceAccount).Error; err != nil {
			return err
		}
		if err := tx.Delete(&transaction).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}


// --- FUNGSI BARU: UpdateTransaction ---
func UpdateTransaction(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    var input TransactionInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    currentUser := c.MustGet("currentUser").(model.User)

    err = database.DB.Transaction(func(tx *gorm.DB) error {
        // 1. Ambil data transaksi LAMA
        var oldTransaction model.Transaction
        if err := tx.Where("id = ? AND user_id = ?", id, currentUser.ID).First(&oldTransaction).Error; err != nil {
            return errors.New("transaction not found")
        }

        // 2. KEMBALIKAN SALDO berdasarkan transaksi LAMA (Revert)
        {
            var oldSourceAccount model.Account
            if err := tx.First(&oldSourceAccount, oldTransaction.AccountID).Error; err != nil {
                return errors.New("old source account not found")
            }
            switch oldTransaction.Type {
            case "expense":
                oldSourceAccount.Balance += oldTransaction.Amount
            case "income":
                oldSourceAccount.Balance -= oldTransaction.Amount
            case "transfer":
                var oldDestAccount model.Account
                if err := tx.First(&oldDestAccount, oldTransaction.DestinationAccountID).Error; err != nil {
                    return errors.New("old destination account not found")
                }
                oldSourceAccount.Balance += oldTransaction.Amount
                oldDestAccount.Balance -= oldTransaction.Amount
                if err := tx.Save(&oldDestAccount).Error; err != nil {
                    return err
                }
            }
            if err := tx.Save(&oldSourceAccount).Error; err != nil {
                return err
            }
        }

        // 3. TERAPKAN SALDO berdasarkan input BARU (Apply)
        {
            var newSourceAccount model.Account
            if err := tx.Where("id = ? AND user_id = ?", input.AccountID, currentUser.ID).First(&newSourceAccount).Error; err != nil {
                return errors.New("new source account not found")
            }
            switch input.Type {
            case "expense":
                newSourceAccount.Balance -= input.Amount
            case "income":
                newSourceAccount.Balance += input.Amount
            case "transfer":
                if input.DestinationAccountID == nil {
                    return errors.New("destination_account_id is required")
                }
                var newDestAccount model.Account
                if err := tx.Where("id = ? AND user_id = ?", *input.DestinationAccountID, currentUser.ID).First(&newDestAccount).Error; err != nil {
                    return errors.New("new destination account not found")
                }
                newSourceAccount.Balance -= input.Amount
                newDestAccount.Balance += input.Amount
                if err := tx.Save(&newDestAccount).Error; err != nil {
                    return err
                }
            }
            if err := tx.Save(&newSourceAccount).Error; err != nil {
                return err
            }
        }

        // 4. Update data transaksi di database dengan data BARU
        oldTransaction.AccountID = input.AccountID
        oldTransaction.SubCategoryID = input.SubCategoryID
        oldTransaction.Amount = input.Amount
        oldTransaction.Type = input.Type
        oldTransaction.Notes = input.Notes
        oldTransaction.TransactionDate = input.TransactionDate
        oldTransaction.DestinationAccountID = input.DestinationAccountID
        if err := tx.Save(&oldTransaction).Error; err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Transaction updated successfully"})
}