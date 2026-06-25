package migrations

import (
	"time"

	"github.com/captainthx/calorie/backend/internal/food"
	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func All(seed bool) []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "20260625000001",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.Exec(`CREATE TYPE user_role_enum AS ENUM ('ADMIN', 'USER')`).Error; err != nil {
					return err
				}
				return tx.AutoMigrate(&user.Users{})
			},
			Rollback: func(tx *gorm.DB) error {
				if err := tx.Migrator().DropTable(&user.Users{}); err != nil {
					return err
				}
				return tx.Exec(`DROP TYPE IF EXISTS user_role_enum`).Error
			},
		},
		{
			ID: "20260625000002",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&food.FoodEntry{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(&food.FoodEntry{})
			},
		},
		{
			ID: "20260625000003",
			Migrate: func(tx *gorm.DB) error {
				if !seed {
					return nil
				}
				seeds := []user.Users{
					{Name: "John", Role: user.User, Token: "user-token-123", DailyCalorieLimit: 2100, MonthlyPriceLimit: 1000},
					{Name: "Jane", Role: user.User, Token: "user-token-456", DailyCalorieLimit: 2100, MonthlyPriceLimit: 1000},
					{Name: "Admin", Role: user.Admin, Token: "admin-token-789"},
				}
				if err := tx.Create(&seeds).Error; err != nil {
					return err
				}
				return SeedFoodEntries(tx, seeds[0].ID, seeds[1].ID)
			},
			Rollback: func(tx *gorm.DB) error {
				return nil
			},
		},
	}
}

// SeedFoodEntries inserts dev food entries for the given user IDs.
// Exported so integration tests can re-seed after truncation.
func SeedFoodEntries(db *gorm.DB, johnID, janeID uint) error {
	entries := []food.FoodEntry{
		// John - yesterday 2200 cal (exceeds 2100 limit)
		{UserID: johnID, FoodName: "Breakfast", Calories: 700, Price: 50, EntryDate: ago(1, 8)},
		{UserID: johnID, FoodName: "Lunch", Calories: 800, Price: 75, EntryDate: ago(1, 12)},
		{UserID: johnID, FoodName: "Dinner", Calories: 700, Price: 80, EntryDate: ago(1, 18)},
		// John - today
		{UserID: johnID, FoodName: "Breakfast", Calories: 450, Price: 35, EntryDate: ago(0, 8)},
		{UserID: johnID, FoodName: "Lunch", Calories: 450, Price: 60, EntryDate: ago(0, 12)},
		// John - last 7 days
		{UserID: johnID, FoodName: "Rice & Curry", Calories: 600, Price: 45, EntryDate: ago(2, 12)},
		{UserID: johnID, FoodName: "Salad", Calories: 300, Price: 30, EntryDate: ago(3, 12)},
		{UserID: johnID, FoodName: "Pad Thai", Calories: 700, Price: 55, EntryDate: ago(4, 12)},
		{UserID: johnID, FoodName: "Tom Yum Soup", Calories: 400, Price: 65, EntryDate: ago(5, 12)},
		{UserID: johnID, FoodName: "Stir Fried Rice", Calories: 650, Price: 40, EntryDate: ago(6, 12)},
		// John - previous 7 days
		{UserID: johnID, FoodName: "Noodles", Calories: 700, Price: 50, EntryDate: ago(7, 12)},
		{UserID: johnID, FoodName: "Smoothie", Calories: 250, Price: 80, EntryDate: ago(8, 9)},
		{UserID: johnID, FoodName: "Sandwich", Calories: 500, Price: 45, EntryDate: ago(10, 12)},
		{UserID: johnID, FoodName: "Chicken Rice", Calories: 550, Price: 45, EntryDate: ago(12, 12)},
		{UserID: johnID, FoodName: "Pizza", Calories: 800, Price: 90, EntryDate: ago(13, 12)},
		// Jane - monthly price 1565 (exceeds 1000 limit)
		{UserID: janeID, FoodName: "Sushi", Calories: 600, Price: 280, EntryDate: ago(0, 13)},
		{UserID: janeID, FoodName: "Steak", Calories: 900, Price: 320, EntryDate: ago(1, 19)},
		{UserID: janeID, FoodName: "Lobster", Calories: 700, Price: 450, EntryDate: ago(2, 13)},
		{UserID: janeID, FoodName: "Salad", Calories: 200, Price: 35, EntryDate: ago(3, 12)},
		{UserID: janeID, FoodName: "Pasta", Calories: 500, Price: 75, EntryDate: ago(4, 12)},
		{UserID: janeID, FoodName: "Breakfast", Calories: 350, Price: 40, EntryDate: ago(5, 8)},
		{UserID: janeID, FoodName: "Thai food", Calories: 550, Price: 65, EntryDate: ago(6, 12)},
		{UserID: janeID, FoodName: "Dim sum", Calories: 600, Price: 120, EntryDate: ago(8, 12)},
		{UserID: janeID, FoodName: "Burger", Calories: 700, Price: 85, EntryDate: ago(11, 12)},
		{UserID: janeID, FoodName: "Ramen", Calories: 650, Price: 95, EntryDate: ago(12, 12)},
	}
	return db.Create(&entries).Error
}

func ago(n, hour int) time.Time {
	now := time.Now()
	d := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	return d.AddDate(0, 0, -n)
}
