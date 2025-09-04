package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/TheRaccoon-Black/goMoneyApi/internal/database" // Ganti dengan path modul Anda
	"github.com/TheRaccoon-Black/goMoneyApi/internal/model"    // Ganti dengan path modul Anda
)

func main() {
	// Memuat konfigurasi (akan kita kembangkan nanti)

	// Menghubungkan ke database
	database.ConnectDatabase()

	// Menjalankan Auto Migration
	err := database.DB.AutoMigrate(&model.User{}, &model.Account{}, &model.Category{}, &model.SubCategory{}, &model.Transaction{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Inisialisasi Gin Router
	router := gin.Default()

	// Rute percobaan
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Menjalankan server
	router.Run(":8080") // Server akan berjalan di port 8080
}