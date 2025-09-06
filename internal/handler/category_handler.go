package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/database"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/model"
)


type CategoryInput struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required"` // "expense" or "income"
}

type SubCategoryInput struct {
	Name string `json:"name" binding:"required"`
}

// --- HANDLER UNTUK KATEGORI ---

func CreateCategory(c *gin.Context) {
	var input CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currentUser := c.MustGet("currentUser").(model.User)
	category := model.Category{
		UserID: currentUser.ID,
		Name:   input.Name,
		Type:   input.Type,
	}
	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	// <-- PERUBAHAN: Muat relasi User sebelum mengirim respons
	database.DB.Preload("User").First(&category, category.ID)

	c.JSON(http.StatusOK, category)
}

func GetCategories(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(model.User)
	categoryType := c.Query("type")
	
	// <-- PERUBAHAN UTAMA: Tambahkan .Preload("SubCategories")
	// Ini akan mengambil semua sub-kategori yang terkait dengan setiap kategori.
	query := database.DB.Preload("SubCategories").Where("user_id = ?", currentUser.ID)
	
	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}

	var categories []model.Category
	if err := query.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func UpdateCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var category model.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	currentUser := c.MustGet("currentUser").(model.User)
	if category.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this category"})
		return
	}
	var input CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	category.Name = input.Name
	category.Type = input.Type
	database.DB.Save(&category)
	
    // <-- PERUBAHAN: Muat relasi User sebelum mengirim respons
	database.DB.Preload("User").First(&category, category.ID)
	
	c.JSON(http.StatusOK, category)
}

func DeleteCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var category model.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	currentUser := c.MustGet("currentUser").(model.User)
	if category.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this category"})
		return
	}
	database.DB.Delete(&category)
	c.JSON(http.StatusOK, gin.H{"message": "Category and its sub-categories deleted successfully"})
}


// --- HANDLER UNTUK SUB-KATEGORI ---

func CreateSubCategory(c *gin.Context) {
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Category ID"})
		return
	}
	var input SubCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currentUser := c.MustGet("currentUser").(model.User)
	var category model.Category
	if err := database.DB.Where("id = ? AND user_id = ?", categoryID, currentUser.ID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parent category not found"})
		return
	}
	subCategory := model.SubCategory{
		UserID:     currentUser.ID,
		CategoryID: uint(categoryID),
		Name:       input.Name,
	}
	if err := database.DB.Create(&subCategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sub-category"})
		return
	}
    
    // <-- PERUBAHAN: Muat relasi User dan Category sebelum mengirim respons
    database.DB.Preload("User").Preload("Category").First(&subCategory, subCategory.ID)
	
	c.JSON(http.StatusOK, subCategory)
}

func GetSubCategoriesForCategory(c *gin.Context) {
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Category ID"})
		return
	}
	currentUser := c.MustGet("currentUser").(model.User)
	var subCategories []model.SubCategory
	
    // <-- PERUBAHAN: Tambahkan .Preload untuk User dan Category
	if err := database.DB.Preload("User").Preload("Category").Where("category_id = ? AND user_id = ?", categoryID, currentUser.ID).Find(&subCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sub-categories"})
		return
	}
	
	c.JSON(http.StatusOK, subCategories)
}

func UpdateSubCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var subCategory model.SubCategory
	if err := database.DB.First(&subCategory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sub-category not found"})
		return
	}
	currentUser := c.MustGet("currentUser").(model.User)
	if subCategory.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this sub-category"})
		return
	}
	var input SubCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	subCategory.Name = input.Name
	database.DB.Save(&subCategory)
	
    // <-- PERUBAHAN: Muat relasi User dan Category sebelum mengirim respons
    database.DB.Preload("User").Preload("Category").First(&subCategory, subCategory.ID)

	c.JSON(http.StatusOK, subCategory)
}

func DeleteSubCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var subCategory model.SubCategory
	if err := database.DB.First(&subCategory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sub-category not found"})
		return
	}
	currentUser := c.MustGet("currentUser").(model.User)
	if subCategory.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this sub-category"})
		return
	}
	database.DB.Delete(&subCategory)
	c.JSON(http.StatusOK, gin.H{"message": "Sub-category deleted successfully"})
}

func GetAllSubCategoriesForCategory(c *gin.Context) {
    // Mengambil ID kategori dari URL, contoh: /api/categories/1/... -> "1"
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Category ID"})
		return
	}

    // Mengambil data user yang sedang login
	currentUser := c.MustGet("currentUser").(model.User)

    // Mencari semua sub-kategori di database
    // yang category_id-nya cocok DAN user_id-nya cocok
	var subCategories []model.SubCategory
	if err := database.DB.Where("category_id = ? AND user_id = ?", categoryID, currentUser.ID).Find(&subCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sub-categories"})
		return
	}

    // Mengembalikan hasilnya sebagai JSON
	c.JSON(http.StatusOK, subCategories)
}