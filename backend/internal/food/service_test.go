package food

import (
	"testing"
	"time"

	"github.com/captainthx/calorie/backend/internal/user"
	"gorm.io/gorm"
)

// --- mock food repository ---

type mockRepo struct {
	entry       *FoodEntry
	calSum      int
	priceSum    float64
	last7Count  int64
	prev7Count  int64
	usersCount  int64
	last7CalSum int64
	capturedTo  time.Time
}

func (m *mockRepo) Create(e *FoodEntry) error                { e.ID = 1; return nil }
func (m *mockRepo) FindByID(id uint) (*FoodEntry, error) {
	if m.entry == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return m.entry, nil
}
func (m *mockRepo) FindByIDWithUser(id uint) (*FoodEntryWithUser, error) {
	if m.entry == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return &FoodEntryWithUser{FoodEntry: *m.entry}, nil
}
func (m *mockRepo) FindByUserID(userID uint, df, dt *time.Time) ([]FoodEntry, error) {
	return nil, nil
}
func (m *mockRepo) FindAll(df, dt *time.Time) ([]FoodEntryWithUser, error) { return nil, nil }
func (m *mockRepo) Update(e *FoodEntry) error                              { return nil }
func (m *mockRepo) Delete(id uint) error                                   { return nil }
func (m *mockRepo) SumCaloriesOnDay(userID uint, date time.Time) (int, error) {
	return m.calSum, nil
}
func (m *mockRepo) SumPriceInMonth(userID uint, year, month int) (float64, error) {
	return m.priceSum, nil
}
func (m *mockRepo) CountEntriesInRange(from, to time.Time) (int64, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if to.After(today) {
		// last-7 call: to = tomorrow
		m.capturedTo = to
		return m.last7Count, nil
	}
	return m.prev7Count, nil
}
func (m *mockRepo) AvgCaloriesPerUserInRange(from, to time.Time) (float64, error) { return 0, nil }
func (m *mockRepo) CountUsers() (int64, error)                                     { return m.usersCount, nil }
func (m *mockRepo) SumCaloriesInRange(from, to time.Time) (int64, error) {
	return m.last7CalSum, nil
}
func (m *mockRepo) SumCaloriesPerDay(userID uint, from, to time.Time) (map[string]int, error) {
	return map[string]int{}, nil
}
func (m *mockRepo) SumPriceInMonths(userID uint, from, to time.Time) (map[string]float64, error) {
	return map[string]float64{}, nil
}

// --- mock user repository ---

type mockUserRepo struct{}

func (m *mockUserRepo) GetUserByToken(token string) (*user.Users, error) { return nil, nil }
func (m *mockUserRepo) GetUserByID(id uint) (*user.Users, error)         { return nil, nil }

func newTestSvc(repo *mockRepo) *FoodService {
	return NewFoodService(repo, &mockUserRepo{})
}

// Test 1: user sees/edits/deletes only their own entries
func TestOwnership(t *testing.T) {
	entry := &FoodEntry{UserID: 1}
	entry.ID = 10
	svc := newTestSvc(&mockRepo{entry: entry})

	if _, err := svc.GetByID(10, 2); err != ErrForbidden {
		t.Fatalf("GetByID wrong user: want ErrForbidden, got %v", err)
	}
	if _, err := svc.Update(10, 2, UpdateFoodEntryRequest{}); err != ErrForbidden {
		t.Fatalf("Update wrong user: want ErrForbidden, got %v", err)
	}
	if err := svc.Delete(10, 2); err != ErrForbidden {
		t.Fatalf("Delete wrong user: want ErrForbidden, got %v", err)
	}
	if _, err := svc.GetByID(10, 1); err != nil {
		t.Fatalf("GetByID owner: unexpected error %v", err)
	}
}

// Test 2: calorie daily, price monthly, exceeded uses strict >
func TestDailySummaryCalcsAndExceededFlag(t *testing.T) {
	u := &user.Users{DailyCalorieLimit: 2100, MonthlyPriceLimit: 1000}
	u.ID = 1

	s, err := newTestSvc(&mockRepo{calSum: 2200, priceSum: 1100}).DailySummary(u, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalCalories != 2200 {
		t.Errorf("calories: want 2200, got %d", s.TotalCalories)
	}
	if !s.CalorieExceeded {
		t.Error("calorie_exceeded should be true (2200 > 2100)")
	}
	if !s.PriceExceeded {
		t.Error("price_exceeded should be true (1100 > 1000)")
	}

	// exactly at limit — must NOT be exceeded (strict >)
	s2, _ := newTestSvc(&mockRepo{calSum: 2100, priceSum: 1000}).DailySummary(u, time.Now())
	if s2.CalorieExceeded {
		t.Error("calorie_exceeded should be false at exact limit (2100 == 2100)")
	}
	if s2.PriceExceeded {
		t.Error("price_exceeded should be false at exact limit (1000 == 1000)")
	}
}

// Test 3: GetReport last-7 window includes today
func TestGetReportWindowIncludesToday(t *testing.T) {
	repo := &mockRepo{last7Count: 5, prev7Count: 3, usersCount: 2, last7CalSum: 1000}
	report, err := newTestSvc(repo).GetReport()
	if err != nil {
		t.Fatal(err)
	}
	if report.EntriesLast7Days != 5 {
		t.Errorf("last7: want 5, got %d", report.EntriesLast7Days)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	expectedTo := today.AddDate(0, 0, 1)
	if !repo.capturedTo.Equal(expectedTo) {
		t.Errorf("last7 end boundary: want %v (tomorrow), got %v — today must be included", expectedTo, repo.capturedTo)
	}
}
