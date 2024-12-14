package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// User model
type User struct {
	gorm.Model
	Username string `json:"username"`
	Email    string `json:"email" gorm:"unique"`
}

// Product model
type Product struct {
	gorm.Model
	UserID                  uint           `json:"user_id"`
	ProductName             string         `json:"product_name"`
	ProductDescription      string         `json:"product_description"`
	ProductImages           pq.StringArray `gorm:"type:text[]" json:"product_images"`
	CompressedProductImages pq.StringArray `gorm:"type:text[]" json:"compressed_product_images"`
	ProductPrice            float64        `json:"product_price"`
}
