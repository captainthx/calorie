package food

import (
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(entry *FoodEntry) error
	FindByID(id uint) (*FoodEntry, error)
	FindByIDWithUser(id uint) (*FoodEntryWithUser, error)
	FindByUserID(userID uint, dateFrom, dateTo *time.Time) ([]FoodEntry, error)
	FindAll(dateFrom, dateTo *time.Time) ([]FoodEntryWithUser, error)
	Update(entry *FoodEntry) error
	Delete(id uint) error
	SumCaloriesOnDay(userID uint, date time.Time) (int, error)
	SumPriceInMonth(userID uint, year, month int) (float64, error)
	CountEntriesInRange(from, to time.Time) (int64, error)
	AvgCaloriesPerUserInRange(from, to time.Time) (float64, error)
	CountUsers() (int64, error)
	SumCaloriesInRange(from, to time.Time) (int64, error)
	SumCaloriesPerDay(userID uint, from, to time.Time) (map[string]int, error)
	SumPriceInMonths(userID uint, from, to time.Time) (map[string]float64, error)
}

type repository struct {
	db *gorm.DB
}

func NewFoodRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(entry *FoodEntry) error {
	return r.db.Create(entry).Error
}

func (r *repository) FindByID(id uint) (*FoodEntry, error) {
	var entry FoodEntry
	if err := r.db.First(&entry, id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *repository) FindByIDWithUser(id uint) (*FoodEntryWithUser, error) {
	var result FoodEntryWithUser
	err := r.db.Model(&FoodEntry{}).
		Select("food_entries.*, users.name as user_name").
		Joins("JOIN users ON users.id = food_entries.user_id").
		Where("food_entries.id = ?", id).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *repository) FindByUserID(userID uint, dateFrom, dateTo *time.Time) ([]FoodEntry, error) {
	var entries []FoodEntry
	q := r.db.Where("user_id = ?", userID)
	if dateFrom != nil {
		q = q.Where("entry_date >= ?", dateFrom)
	}
	if dateTo != nil {
		q = q.Where("entry_date < ?", dateTo.AddDate(0, 0, 1))
	}
	return entries, q.Order("entry_date DESC").Find(&entries).Error
}

func (r *repository) FindAll(dateFrom, dateTo *time.Time) ([]FoodEntryWithUser, error) {
	var results []FoodEntryWithUser
	q := r.db.Model(&FoodEntry{}).
		Select("food_entries.*, users.name as user_name").
		Joins("JOIN users ON users.id = food_entries.user_id")
	if dateFrom != nil {
		q = q.Where("food_entries.entry_date >= ?", dateFrom)
	}
	if dateTo != nil {
		q = q.Where("food_entries.entry_date < ?", dateTo.AddDate(0, 0, 1))
	}
	return results, q.Order("food_entries.entry_date DESC").Find(&results).Error
}

func (r *repository) Update(entry *FoodEntry) error {
	return r.db.Save(entry).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&FoodEntry{}, id).Error
}

func (r *repository) SumCaloriesOnDay(userID uint, date time.Time) (int, error) {
	var total int
	err := r.db.Model(&FoodEntry{}).
		Select("COALESCE(SUM(calories), 0)").
		Where("user_id = ? AND DATE(entry_date) = DATE(?)", userID, date).
		Scan(&total).Error
	return total, err
}

func (r *repository) SumPriceInMonth(userID uint, year, month int) (float64, error) {
	var total float64
	err := r.db.Model(&FoodEntry{}).
		Select("COALESCE(SUM(price), 0)").
		Where("user_id = ? AND EXTRACT(YEAR FROM entry_date) = ? AND EXTRACT(MONTH FROM entry_date) = ?", userID, year, month).
		Scan(&total).Error
	return total, err
}

func (r *repository) CountEntriesInRange(from, to time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&FoodEntry{}).
		Where("entry_date >= ? AND entry_date < ?", from, to).
		Count(&count).Error
	return count, err
}

func (r *repository) AvgCaloriesPerUserInRange(from, to time.Time) (float64, error) {
	var result struct {
		Total     int64
		UserCount int64
	}
	err := r.db.Model(&FoodEntry{}).
		Select("COALESCE(SUM(calories), 0) as total, COUNT(DISTINCT user_id) as user_count").
		Where("entry_date >= ? AND entry_date < ?", from, to).
		Scan(&result).Error
	if err != nil || result.UserCount == 0 {
		return 0, err
	}
	return float64(result.Total) / float64(result.UserCount), nil
}

func (r *repository) CountUsers() (int64, error) {
	var count int64
	err := r.db.Table("users").Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

func (r *repository) SumCaloriesInRange(from, to time.Time) (int64, error) {
	var total int64
	err := r.db.Model(&FoodEntry{}).
		Select("COALESCE(SUM(calories), 0)").
		Where("entry_date >= ? AND entry_date < ?", from, to).
		Scan(&total).Error
	return total, err
}

func (r *repository) SumCaloriesPerDay(userID uint, from, to time.Time) (map[string]int, error) {
	type row struct {
		Day   string `gorm:"column:day"`
		Total int    `gorm:"column:total"`
	}
	var rows []row
	err := r.db.Model(&FoodEntry{}).
		Select("TO_CHAR(entry_date, 'YYYY-MM-DD') as day, COALESCE(SUM(calories), 0) as total").
		Where("user_id = ? AND entry_date >= ? AND entry_date < ?", userID, from, to.AddDate(0, 0, 1)).
		Group("day").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]int, len(rows))
	for _, row := range rows {
		result[row.Day] = row.Total
	}
	return result, nil
}

func (r *repository) SumPriceInMonths(userID uint, from, to time.Time) (map[string]float64, error) {
	type row struct {
		YearMonth string  `gorm:"column:ym"`
		Total     float64 `gorm:"column:total"`
	}
	var rows []row
	err := r.db.Model(&FoodEntry{}).
		Select("TO_CHAR(entry_date, 'YYYY-MM') as ym, COALESCE(SUM(price), 0) as total").
		Where("user_id = ? AND entry_date >= ? AND entry_date < ?", userID, from, to.AddDate(0, 0, 1)).
		Group("ym").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]float64, len(rows))
	for _, row := range rows {
		result[row.YearMonth] = row.Total
	}
	return result, nil
}
