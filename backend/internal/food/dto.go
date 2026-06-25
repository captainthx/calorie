package food

import "time"

type CreateFoodEntryRequest struct {
	FoodName  string    `json:"food_name"  binding:"required"`
	Calories  *int      `json:"calories"   binding:"required,gte=0,lte=10000"`
	Price     *float64  `json:"price"      binding:"required,gte=0,lte=10000"`
	EntryDate time.Time `json:"entry_date" binding:"required"`
}

type AdminCreateFoodEntryRequest struct {
	UserID    uint      `json:"user_id"    binding:"required"`
	FoodName  string    `json:"food_name"  binding:"required"`
	Calories  *int      `json:"calories"   binding:"required,gte=0,lte=10000"`
	Price     *float64  `json:"price"      binding:"required,gte=0,lte=10000"`
	EntryDate time.Time `json:"entry_date" binding:"required"`
}

type UpdateFoodEntryRequest struct {
	FoodName  *string    `json:"food_name"`
	Calories  *int       `json:"calories"   binding:"omitempty,gte=0,lte=10000"`
	Price     *float64   `json:"price"      binding:"omitempty,gte=0,lte=10000"`
	EntryDate *time.Time `json:"entry_date"`
}

type PutFoodEntryRequest struct {
	FoodName  string    `json:"food_name"  binding:"required"`
	Calories  *int      `json:"calories"   binding:"required,gte=0,lte=10000"`
	Price     *float64  `json:"price"      binding:"required,gte=0,lte=10000"`
	EntryDate time.Time `json:"entry_date" binding:"required"`
}

type FoodEntryResponse struct {
	ID        uint      `json:"id"`
	FoodName  string    `json:"food_name"`
	Calories  int       `json:"calories"`
	Price     float64   `json:"price"`
	EntryDate time.Time `json:"entry_date"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminFoodEntryResponse struct {
	FoodEntryResponse
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
}

type DailySummaryResponse struct {
	Date            string  `json:"date"`
	TotalCalories   int     `json:"total_calories"`
	TotalPrice      float64 `json:"total_price"`
	CalorieLimit    int     `json:"calorie_limit"`
	CalorieExceeded bool    `json:"calorie_exceeded"`
	PriceLimit      float64 `json:"price_limit"`
	PriceExceeded   bool    `json:"price_exceeded"`
}

type DailySummaryRangeItem struct {
	Date            string `json:"date"`
	TotalCalories   int    `json:"total_calories"`
	CalorieLimit    int    `json:"calorie_limit"`
	CalorieExceeded bool   `json:"calorie_exceeded"`
}

type ReportResponse struct {
	EntriesLast7Days         int64          `json:"entries_last_7_days"`
	EntriesPrevious7Days     int64          `json:"entries_previous_7_days"`
	AvgCaloriesPerUserLast7D float64        `json:"average_calories_per_user_last_7_days"`
	UsersCount               int64          `json:"users_count"`
	Comparison               ComparisonData `json:"entries_comparison"`
}

type ComparisonData struct {
	CurrentWeek  int64 `json:"current_week"`
	PreviousWeek int64 `json:"previous_week"`
	Difference   int64 `json:"difference"`
}

type DailyCalorieRow struct {
	Date          string `gorm:"column:date"`
	TotalCalories int    `gorm:"column:total_calories"`
}
