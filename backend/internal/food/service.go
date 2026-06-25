package food

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/captainthx/calorie/backend/internal/user"
	"gorm.io/gorm"
)

var ErrForbidden = errors.New("forbidden")
var ErrUserNotFound = errors.New("user not found")

type FoodService struct {
	repo     Repository
	userRepo user.Repository
}

func NewFoodService(repo Repository, userRepo user.Repository) *FoodService {
	return &FoodService{repo: repo, userRepo: userRepo}
}

func (s *FoodService) Create(u *user.Users, req CreateFoodEntryRequest) (*FoodEntryResponse, error) {
	req.FoodName = strings.TrimSpace(req.FoodName)
	if req.FoodName == "" {
		return nil, errors.New("food_name cannot be empty")
	}
	entry := &FoodEntry{
		UserID:    u.ID,
		FoodName:  req.FoodName,
		Calories:  *req.Calories,
		Price:     *req.Price,
		EntryDate: req.EntryDate,
	}
	if err := s.repo.Create(entry); err != nil {
		return nil, fmt.Errorf("create food entry: %w", err)
	}
	return toFoodEntryResponse(entry), nil
}

func (s *FoodService) List(userID uint, dateFrom, dateTo *time.Time) ([]FoodEntryResponse, error) {
	entries, err := s.repo.FindByUserID(userID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("list food entries for user %d: %w", userID, err)
	}
	result := make([]FoodEntryResponse, len(entries))
	for i := range entries {
		result[i] = *toFoodEntryResponse(&entries[i])
	}
	return result, nil
}

func (s *FoodService) GetByID(id, userID uint) (*FoodEntryResponse, error) {
	entry, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get food entry %d: %w", id, err)
	}
	if entry.UserID != userID {
		return nil, ErrForbidden
	}
	return toFoodEntryResponse(entry), nil
}

func (s *FoodService) Update(id, userID uint, req UpdateFoodEntryRequest) (*FoodEntryResponse, error) {
	entry, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get food entry %d for update: %w", id, err)
	}
	if entry.UserID != userID {
		return nil, ErrForbidden
	}
	if req.FoodName != nil {
		trimmed := strings.TrimSpace(*req.FoodName)
		if trimmed == "" {
			return nil, errors.New("food_name cannot be empty")
		}
		req.FoodName = &trimmed
	}
	applyUpdate(entry, req)
	if err := s.repo.Update(entry); err != nil {
		return nil, fmt.Errorf("update food entry %d: %w", id, err)
	}
	return toFoodEntryResponse(entry), nil
}

func (s *FoodService) Delete(id, userID uint) error {
	entry, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("get food entry %d for delete: %w", id, err)
	}
	if entry.UserID != userID {
		return ErrForbidden
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete food entry %d: %w", id, err)
	}
	return nil
}

func (s *FoodService) DailySummary(u *user.Users, date time.Time) (*DailySummaryResponse, error) {
	totalCal, err := s.repo.SumCaloriesOnDay(u.ID, date)
	if err != nil {
		return nil, fmt.Errorf("sum daily calories for user %d on %s: %w", u.ID, date.Format("2006-01-02"), err)
	}
	totalPrice, err := s.repo.SumPriceInMonth(u.ID, date.Year(), int(date.Month()))
	if err != nil {
		return nil, fmt.Errorf("sum monthly price for user %d on %s: %w", u.ID, date.Format("2006-01-02"), err)
	}
	return &DailySummaryResponse{
		Date:            date.Format("2006-01-02"),
		TotalCalories:   totalCal,
		TotalPrice:      totalPrice,
		CalorieLimit:    u.DailyCalorieLimit,
		CalorieExceeded: totalCal > u.DailyCalorieLimit,
		PriceLimit:      float64(u.MonthlyPriceLimit),
		PriceExceeded:   totalPrice > float64(u.MonthlyPriceLimit),
	}, nil
}

func (s *FoodService) DailySummaryRange(u *user.Users, from, to time.Time) ([]DailySummaryRangeItem, error) {
	rows, err := s.repo.SumCaloriesPerDayInRange(u.ID, from, to)
	if err != nil {
		return nil, fmt.Errorf("sum calories per day for user %d: %w", u.ID, err)
	}
	result := make([]DailySummaryRangeItem, len(rows))
	for i, row := range rows {
		result[i] = DailySummaryRangeItem{
			Date:            row.Date,
			TotalCalories:   row.TotalCalories,
			CalorieLimit:    u.DailyCalorieLimit,
			CalorieExceeded: row.TotalCalories > u.DailyCalorieLimit,
		}
	}
	return result, nil
}

func (s *FoodService) ListAll(dateFrom, dateTo *time.Time) ([]AdminFoodEntryResponse, error) {
	entries, err := s.repo.FindAll(dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("list admin food entries: %w", err)
	}
	result := make([]AdminFoodEntryResponse, len(entries))
	for i := range entries {
		result[i] = *toAdminFoodEntryResponse(&entries[i])
	}
	return result, nil
}

func (s *FoodService) AdminGetByID(id uint) (*AdminFoodEntryResponse, error) {
	entry, err := s.repo.FindByIDWithUser(id)
	if err != nil {
		return nil, fmt.Errorf("get admin food entry %d: %w", id, err)
	}
	return toAdminFoodEntryResponse(entry), nil
}

func (s *FoodService) AdminCreate(req AdminCreateFoodEntryRequest) (*AdminFoodEntryResponse, error) {
	req.FoodName = strings.TrimSpace(req.FoodName)
	if req.FoodName == "" {
		return nil, errors.New("food_name cannot be empty")
	}
	if _, err := s.userRepo.GetUserByID(req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user %d for admin create: %w", req.UserID, err)
	}
	entry := &FoodEntry{
		UserID:    req.UserID,
		FoodName:  req.FoodName,
		Calories:  *req.Calories,
		Price:     *req.Price,
		EntryDate: req.EntryDate,
	}
	if err := s.repo.Create(entry); err != nil {
		return nil, fmt.Errorf("create admin food entry for user %d: %w", req.UserID, err)
	}
	result, err := s.repo.FindByIDWithUser(entry.ID)
	if err != nil {
		return nil, fmt.Errorf("load created admin food entry %d: %w", entry.ID, err)
	}
	return toAdminFoodEntryResponse(result), nil
}

func (s *FoodService) AdminUpdate(id uint, req UpdateFoodEntryRequest) (*AdminFoodEntryResponse, error) {
	entryWithUser, err := s.repo.FindByIDWithUser(id)
	if err != nil {
		return nil, fmt.Errorf("get admin food entry %d for update: %w", id, err)
	}
	if req.FoodName != nil {
		trimmed := strings.TrimSpace(*req.FoodName)
		if trimmed == "" {
			return nil, errors.New("food_name cannot be empty")
		}
		req.FoodName = &trimmed
	}
	applyUpdate(&entryWithUser.FoodEntry, req)
	if err := s.repo.Update(&entryWithUser.FoodEntry); err != nil {
		return nil, fmt.Errorf("update admin food entry %d: %w", id, err)
	}
	return toAdminFoodEntryResponse(entryWithUser), nil
}

func (s *FoodService) AdminDelete(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("get admin food entry %d for delete: %w", id, err)
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete admin food entry %d: %w", id, err)
	}
	return nil
}

func (s *FoodService) GetReport() (*ReportResponse, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	last7Start := today.AddDate(0, 0, -6)
	last7End := today.AddDate(0, 0, 1)
	prev7Start := today.AddDate(0, 0, -13)
	prev7End := last7Start

	last7, err := s.repo.CountEntriesInRange(last7Start, last7End)
	if err != nil {
		return nil, fmt.Errorf("count entries in last 7 days: %w", err)
	}
	prev7, err := s.repo.CountEntriesInRange(prev7Start, prev7End)
	if err != nil {
		return nil, fmt.Errorf("count entries in previous 7 days: %w", err)
	}
	totalCalLast7, err := s.repo.SumCaloriesInRange(last7Start, last7End)
	if err != nil {
		return nil, fmt.Errorf("sum calories in last 7 days: %w", err)
	}
	usersCount, err := s.repo.CountUsers()
	if err != nil {
		return nil, fmt.Errorf("count users for report: %w", err)
	}
	var avgCal float64
	if usersCount > 0 {
		avgCal = float64(totalCalLast7) / float64(usersCount)
	}
	return &ReportResponse{
		EntriesLast7Days:         last7,
		EntriesPrevious7Days:     prev7,
		AvgCaloriesPerUserLast7D: avgCal,
		UsersCount:               usersCount,
		Comparison: ComparisonData{
			CurrentWeek:  last7,
			PreviousWeek: prev7,
			Difference:   last7 - prev7,
		},
	}, nil
}

func applyUpdate(entry *FoodEntry, req UpdateFoodEntryRequest) {
	if req.FoodName != nil {
		entry.FoodName = *req.FoodName
	}
	if req.Calories != nil {
		entry.Calories = *req.Calories
	}
	if req.Price != nil {
		entry.Price = *req.Price
	}
	if req.EntryDate != nil {
		entry.EntryDate = *req.EntryDate
	}
}

func toFoodEntryResponse(e *FoodEntry) *FoodEntryResponse {
	return &FoodEntryResponse{
		ID:        e.ID,
		FoodName:  e.FoodName,
		Calories:  e.Calories,
		Price:     e.Price,
		EntryDate: e.EntryDate,
		CreatedAt: e.CreatedAt,
	}
}

func toAdminFoodEntryResponse(e *FoodEntryWithUser) *AdminFoodEntryResponse {
	return &AdminFoodEntryResponse{
		FoodEntryResponse: *toFoodEntryResponse(&e.FoodEntry),
		UserID:            e.UserID,
		UserName:          e.UserName,
	}
}
