package food

import (
	"time"

	"gorm.io/gorm"
)

type FoodEntry struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index"`
	FoodName  string    `gorm:"type:varchar(255);not null"`
	Calories  int       `gorm:"type:int;not null"`
	Price     float64   `gorm:"type:numeric(10,2);not null"`
	EntryDate time.Time `gorm:"not null"`
}

type FoodEntryWithUser struct {
	FoodEntry
	UserName string `gorm:"column:user_name"`
}
