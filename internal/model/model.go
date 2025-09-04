package model

import (
	"time"
	"gorm.io/gorm"
)


type User struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"size:255;not null"`
	Email        string `gorm:"size:255;not null;unique"`
	PasswordHash string `gorm:"size:255;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Account struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	User      User      `gorm:"foreignKey:UserID"`
	Name      string    `gorm:"size:255;not null"`
	Balance   float64   `gorm:"type:decimal(15,2);not null;default:0.00"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Category struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	User      User      `gorm:"foreignKey:UserID"`
	Name      string    `gorm:"size:255;not null"`
	Type      string    `gorm:"size:50;not null"` 
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SubCategory struct {
	ID         uint     `gorm:"primaryKey"`
	UserID     uint     `gorm:"not null"`
	User       User     `gorm:"foreignKey:UserID"`
	CategoryID uint     `gorm:"not null"`
	Category   Category `gorm:"foreignKey:CategoryID"`
	Name       string   `gorm:"size:255;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Transaction struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"not null"`
	User          User      `gorm:"foreignKey:UserID"`
	AccountID     uint      `gorm:"not null"`
	Account       Account   `gorm:"foreignKey:AccountID"`
	SubCategoryID *uint     
	SubCategory   SubCategory `gorm:"foreignKey:SubCategoryID"`
	Amount        float64   `gorm:"type:decimal(15,2);not null"`
	Type          string    `gorm:"size:50;not null"` 
	Notes         string    `gorm:"type:text"`
	TransactionDate time.Time `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}