package food

import (
	"errors"
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
		return nil, err
	}
	return toFoodEntryResponse(entry), nil
}

func (s *FoodService) List(userID uint, dateFrom, dateTo *time.Time) ([]FoodEntryResponse, error) {
	entries, err := s.repo.FindByUserID(userID, dateFrom, dateTo)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	if entry.UserID != userID {
		return nil, ErrForbidden
	}
	return toFoodEntryResponse(entry), nil
}

func (s *FoodService) Update(id, userID uint, req UpdateFoodEntryRequest) (*FoodEntryResponse, error) {
	entry, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return toFoodEntryResponse(entry), nil
}

func (s *FoodService) Delete(id, userID uint) error {
	entry, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if entry.UserID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(id)
}

func (s *FoodService) DailySummary(u *user.Users, date time.Time) (*DailySummaryResponse, error) {
	totalCal, err := s.repo.SumCaloriesOnDay(u.ID, date)
	if err != nil {
		return nil, err
	}
	totalPrice, err := s.repo.SumPriceInMonth(u.ID, date.Year(), int(date.Month()))
	if err != nil {
		return nil, err
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

func (s *FoodService) ListAll(dateFrom, dateTo *time.Time) ([]AdminFoodEntryResponse, error) {
	entries, err := s.repo.FindAll(dateFrom, dateTo)
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, err
	}
	entry := &FoodEntry{
		UserID:    req.UserID,
		FoodName:  req.FoodName,
		Calories:  *req.Calories,
		Price:     *req.Price,
		EntryDate: req.EntryDate,
	}
	if err := s.repo.Create(entry); err != nil {
		return nil, err
	}
	result, err := s.repo.FindByIDWithUser(entry.ID)
	if err != nil {
		return nil, err
	}
	return toAdminFoodEntryResponse(result), nil
}

func (s *FoodService) AdminUpdate(id uint, req UpdateFoodEntryRequest) (*AdminFoodEntryResponse, error) {
	entryWithUser, err := s.repo.FindByIDWithUser(id)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return toAdminFoodEntryResponse(entryWithUser), nil
}

func (s *FoodService) AdminDelete(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
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
		return nil, err
	}
	prev7, err := s.repo.CountEntriesInRange(prev7Start, prev7End)
	if err != nil {
		return nil, err
	}
	totalCalLast7, err := s.repo.SumCaloriesInRange(last7Start, last7End)
	if err != nil {
		return nil, err
	}
	usersCount, err := s.repo.CountUsers()
	if err != nil {
		return nil, err
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

func (s *FoodService) ListDailySummaries(u *user.Users, dateFrom, dateTo time.Time) ([]DailySummaryResponse, error) {
	calPerDay, err := s.repo.SumCaloriesPerDay(u.ID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	pricePerMonth, err := s.repo.SumPriceInMonths(u.ID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	var result []DailySummaryResponse
	for d := dateFrom; !d.After(dateTo); d = d.AddDate(0, 0, 1) {
		dayKey := d.Format("2006-01-02")
		monthKey := d.Format("2006-01")
		cal := calPerDay[dayKey]
		price := pricePerMonth[monthKey]
		result = append(result, DailySummaryResponse{
			Date:            dayKey,
			TotalCalories:   cal,
			TotalPrice:      price,
			CalorieLimit:    u.DailyCalorieLimit,
			CalorieExceeded: cal > u.DailyCalorieLimit,
			PriceLimit:      float64(u.MonthlyPriceLimit),
			PriceExceeded:   price > float64(u.MonthlyPriceLimit),
		})
	}
	return result, nil
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
