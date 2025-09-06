package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/database" // Ganti dengan path modul Anda
	"github.com/TheRaccoon-Black/goMoneyApi/internal/model"    
	"github.com/TheRaccoon-Black/goMoneyApi/internal/handler"    
	"github.com/TheRaccoon-Black/goMoneyApi/internal/middleware"    
)

func main() {
	// Memuat konfigurasi (akan kita kembangkan nanti)

	// Menghubungkan ke database
	database.ConnectDatabase()

	// Menjalankan Auto Migration
	err := database.DB.AutoMigrate(&model.User{}, &model.Account{}, &model.Category{}, &model.SubCategory{}, &model.Transaction{}, &model.Budget{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Inisialisasi Gin Router
	router := gin.Default()


	// Konfigurasi CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://10.74.197.27:3000"} // <-- Tambahkan alamat IP frontend Anda
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", handler.Register)
		authRoutes.POST("/login", handler.Login)
	}


	apiRoutes := router.Group("/api")
	apiRoutes.Use(middleware.AuthMiddleware()) 
	{
		apiRoutes.GET("/profile", handler.GetCurrentUserProfile)

		//account routes
		apiRoutes.POST("/accounts", handler.CreateAccount)
        apiRoutes.GET("/accounts", handler.GetAccounts)
        apiRoutes.PUT("/accounts/:id", handler.UpdateAccount)
        apiRoutes.DELETE("/accounts/:id", handler.DeleteAccount)

		// Rute Kategori
		apiRoutes.POST("/categories", handler.CreateCategory)
		apiRoutes.GET("/categories", handler.GetCategories)
		apiRoutes.PUT("/categories/:id", handler.UpdateCategory)      
		apiRoutes.DELETE("/categories/:id", handler.DeleteCategory)

		// Rute Sub-Kategori
		apiRoutes.POST("/categories/:id/subcategories", handler.CreateSubCategory)
		apiRoutes.GET("/categories/:id/subcategories", handler.GetSubCategoriesForCategory)
		apiRoutes.PUT("/subcategories/:id", handler.UpdateSubCategory)    
		apiRoutes.DELETE("/subcategories/:id", handler.DeleteSubCategory)
		apiRoutes.GET("/categories/:id/allsubcategories", handler.GetAllSubCategoriesForCategory)

		// Rute Transaksi
		apiRoutes.POST("/transactions", handler.CreateTransaction)
        apiRoutes.GET("/transactions", handler.GetTransactions)
        apiRoutes.GET("/transactions/:id", handler.GetTransactionByID)
        apiRoutes.DELETE("/transactions/:id", handler.DeleteTransaction)
        apiRoutes.PUT("/transactions/:id", handler.UpdateTransaction)

		// Rute Budget
    	apiRoutes.GET("/budgets", handler.GetBudgets)
    	apiRoutes.POST("/budgets", handler.SetBudgets)
		apiRoutes.GET("/budgets/suggestions", handler.GetBudgetSuggestions)
	}
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Menjalankan server
	router.Run(":8080") // Server akan berjalan di port 8080
}