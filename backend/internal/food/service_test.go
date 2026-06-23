package food

import (
	"errors"
	"testing"
	"time"

	"github.com/captainthx/calorie/backend/internal/user"
	"gorm.io/gorm"
)

// --- helpers ---

func intPtr(n int) *int           { return &n }
func floatPtr(f float64) *float64 { return &f }
func strPtr(s string) *string     { return &s }

// --- spy mock food repository ---

type mockRepo struct {
	// configurable return values
	entry         *FoodEntry
	entries       []FoodEntry
	allEntries    []FoodEntryWithUser
	calSum        int
	priceSum      float64
	calPerDay     map[string]int
	pricePerMonth map[string]float64
	last7Count    int64
	prev7Count    int64
	usersCount    int64
	last7CalSum   int64
	capturedTo    time.Time

	// error injection
	createErr           error
	updateErr           error
	deleteErr           error
	findByUserIDErr     error
	findAllErr          error
	sumCalOnDayErr      error
	sumPriceInMonthErr  error
	sumCalPerDayErr     error
	sumPriceInMonthsErr error

	// spy: call tracking
	createCalled          bool
	createdEntry          *FoodEntry
	updateCalled          bool
	updatedEntry          *FoodEntry
	deleteCalled          bool
	deletedID             uint
	findByIDArg           uint
	findByIDCount         int
	findByIDWithUserArg   uint
	findByIDWithUserCount int
}

func (m *mockRepo) Create(e *FoodEntry) error {
	m.createCalled = true
	m.createdEntry = e
	e.ID = 1
	return m.createErr
}
func (m *mockRepo) FindByID(id uint) (*FoodEntry, error) {
	m.findByIDArg = id
	m.findByIDCount++
	if m.entry == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return m.entry, nil
}
func (m *mockRepo) FindByIDWithUser(id uint) (*FoodEntryWithUser, error) {
	m.findByIDWithUserArg = id
	m.findByIDWithUserCount++
	if m.entry == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return &FoodEntryWithUser{FoodEntry: *m.entry}, nil
}
func (m *mockRepo) FindByUserID(userID uint, df, dt *time.Time) ([]FoodEntry, error) {
	return m.entries, m.findByUserIDErr
}
func (m *mockRepo) FindAll(df, dt *time.Time) ([]FoodEntryWithUser, error) {
	return m.allEntries, m.findAllErr
}
func (m *mockRepo) Update(e *FoodEntry) error {
	m.updateCalled = true
	m.updatedEntry = e
	return m.updateErr
}
func (m *mockRepo) Delete(id uint) error {
	m.deleteCalled = true
	m.deletedID = id
	return m.deleteErr
}
func (m *mockRepo) SumCaloriesOnDay(userID uint, date time.Time) (int, error) {
	return m.calSum, m.sumCalOnDayErr
}
func (m *mockRepo) SumPriceInMonth(userID uint, year, month int) (float64, error) {
	return m.priceSum, m.sumPriceInMonthErr
}
func (m *mockRepo) CountEntriesInRange(from, to time.Time) (int64, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if to.After(today) {
		m.capturedTo = to
		return m.last7Count, nil
	}
	return m.prev7Count, nil
}
func (m *mockRepo) AvgCaloriesPerUserInRange(from, to time.Time) (float64, error) { return 0, nil }
func (m *mockRepo) CountUsers() (int64, error)                                    { return m.usersCount, nil }
func (m *mockRepo) SumCaloriesInRange(from, to time.Time) (int64, error) {
	return m.last7CalSum, nil
}
func (m *mockRepo) SumCaloriesPerDay(userID uint, from, to time.Time) (map[string]int, error) {
	if m.calPerDay != nil {
		return m.calPerDay, m.sumCalPerDayErr
	}
	return map[string]int{}, m.sumCalPerDayErr
}
func (m *mockRepo) SumPriceInMonths(userID uint, from, to time.Time) (map[string]float64, error) {
	if m.pricePerMonth != nil {
		return m.pricePerMonth, m.sumPriceInMonthsErr
	}
	return map[string]float64{}, m.sumPriceInMonthsErr
}

// --- spy mock user repository ---

type mockUserRepo struct {
	foundUser             *user.Users
	err                   error
	getUserByIDCalledWith uint
	getUserByIDCount      int
}

func (m *mockUserRepo) GetUserByToken(token string) (*user.Users, error) { return nil, nil }
func (m *mockUserRepo) GetUserByID(id uint) (*user.Users, error) {
	m.getUserByIDCalledWith = id
	m.getUserByIDCount++
	return m.foundUser, m.err
}

func newTestSvc(repo *mockRepo) *FoodService {
	return NewFoodService(repo, &mockUserRepo{})
}

func newTestSvcFull(repo *mockRepo, userRepo *mockUserRepo) *FoodService {
	return NewFoodService(repo, userRepo)
}

// --- Test: Ownership ---

func TestOwnership(t *testing.T) {
	tests := []struct {
		name    string
		userID  uint
		wantErr error
	}{
		{"owner can access", 1, nil},
		{"wrong user gets forbidden", 2, ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &FoodEntry{UserID: 1}
			entry.ID = 10
			svc := newTestSvc(&mockRepo{entry: entry})

			_, getErr := svc.GetByID(10, tt.userID)
			if !errors.Is(getErr, tt.wantErr) {
				t.Errorf("GetByID: got %v, want %v", getErr, tt.wantErr)
			}
			_, updateErr := svc.Update(10, tt.userID, UpdateFoodEntryRequest{})
			if !errors.Is(updateErr, tt.wantErr) {
				t.Errorf("Update: got %v, want %v", updateErr, tt.wantErr)
			}
			deleteErr := svc.Delete(10, tt.userID)
			if !errors.Is(deleteErr, tt.wantErr) {
				t.Errorf("Delete: got %v, want %v", deleteErr, tt.wantErr)
			}
		})
	}
}

// --- Test: Create ---

func TestCreateValidation(t *testing.T) {
	u := &user.Users{}
	u.ID = 1

	tests := []struct {
		name     string
		foodName string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			svc := newTestSvc(repo)

			_, err := svc.Create(u, CreateFoodEntryRequest{
				FoodName:  tt.foodName,
				Calories:  intPtr(300),
				Price:     floatPtr(50.0),
				EntryDate: time.Now(),
			})

			if err == nil || err.Error() != "food_name cannot be empty" {
				t.Errorf("Create(%q): got %v, want \"food_name cannot be empty\"", tt.foodName, err)
			}
			if repo.createCalled {
				t.Error("repo.Create must not be called when validation fails")
			}
		})
	}
}

func TestCreate_Success(t *testing.T) {
	u := &user.Users{}
	u.ID = 1
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	repo := &mockRepo{}
	svc := newTestSvc(repo)

	resp, err := svc.Create(u, CreateFoodEntryRequest{
		FoodName:  "Rice",
		Calories:  intPtr(300),
		Price:     floatPtr(50.0),
		EntryDate: date,
	})

	if err != nil {
		t.Fatalf("Create: unexpected error %v", err)
	}
	if !repo.createCalled {
		t.Fatal("Create: repo.Create was not called")
	}
	if repo.createdEntry.UserID != 1 {
		t.Errorf("Create: createdEntry.UserID = %d, want 1", repo.createdEntry.UserID)
	}
	if repo.createdEntry.FoodName != "Rice" {
		t.Errorf("Create: food_name not trimmed: got %q, want \"Rice\"", repo.createdEntry.FoodName)
	}
	if repo.createdEntry.Calories != 300 {
		t.Errorf("Create: calories = %d, want 300", repo.createdEntry.Calories)
	}

	if repo.createdEntry.Price != 50.0 {
		t.Errorf("Create: price = %.1f, want 50.0", repo.createdEntry.Price)
	}
	if !repo.createdEntry.EntryDate.Equal(date) {
		t.Errorf("Create: entry_date = %v, want %v", repo.createdEntry.EntryDate, date)
	}
	if resp.ID != 1 {
		t.Errorf("Create: response ID = %d, want 1", resp.ID)
	}
}

func TestCreate_PropagatesRepoError(t *testing.T) {
	u := &user.Users{}
	u.ID = 1
	repo := &mockRepo{createErr: errors.New("db error")}
	svc := newTestSvc(repo)

	_, err := svc.Create(u, CreateFoodEntryRequest{
		FoodName:  "Rice",
		Calories:  intPtr(300),
		Price:     floatPtr(50.0),
		EntryDate: time.Now(),
	})

	if err == nil || err.Error() != "db error" {
		t.Errorf("Create: got %v, want \"db error\"", err)
	}
}

// --- Test: List ---

func TestList(t *testing.T) {
	e1 := FoodEntry{FoodName: "Rice", Calories: 300}
	e2 := FoodEntry{FoodName: "Chicken", Calories: 200}
	e1.ID = 1
	e2.ID = 2
	repo := &mockRepo{entries: []FoodEntry{e1, e2}}
	svc := newTestSvc(repo)

	result, err := svc.List(1, nil, nil)

	if err != nil {
		t.Fatalf("List: unexpected error %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("List: got %d entries, want 2", len(result))
	}
	if result[0].FoodName != "Rice" {
		t.Errorf("List: result[0].FoodName = %q, want \"Rice\"", result[0].FoodName)
	}
	if result[1].FoodName != "Chicken" {
		t.Errorf("List: result[1].FoodName = %q, want \"Chicken\"", result[1].FoodName)
	}
}

func TestList_PropagatesRepoError(t *testing.T) {
	repo := &mockRepo{findByUserIDErr: errors.New("db error")}
	svc := newTestSvc(repo)

	_, err := svc.List(1, nil, nil)

	if err == nil || err.Error() != "db error" {
		t.Errorf("List: got %v, want \"db error\"", err)
	}
}

// --- Test: Update ---

func TestUpdateValidation(t *testing.T) {
	tests := []struct {
		name     string
		foodName string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &FoodEntry{UserID: 1, FoodName: "Rice"}
			entry.ID = 10
			repo := &mockRepo{entry: entry}
			svc := newTestSvc(repo)

			_, err := svc.Update(10, 1, UpdateFoodEntryRequest{FoodName: strPtr(tt.foodName)})

			if err == nil || err.Error() != "food_name cannot be empty" {
				t.Errorf("Update(%q): got %v, want \"food_name cannot be empty\"", tt.foodName, err)
			}
			if repo.updateCalled {
				t.Error("repo.Update must not be called when validation fails")
			}
		})
	}
}

func TestUpdate_Success(t *testing.T) {
	entry := &FoodEntry{UserID: 1, FoodName: "Rice", Calories: 300, Price: 50.0}
	entry.ID = 10
	repo := &mockRepo{entry: entry}
	svc := newTestSvc(repo)

	resp, err := svc.Update(10, 1, UpdateFoodEntryRequest{
		FoodName: strPtr("  Chicken  "),
		Calories: intPtr(400),
		// Price not set — should remain 50.0
	})

	if err != nil {
		t.Fatalf("Update: unexpected error %v", err)
	}
	if !repo.updateCalled {
		t.Fatal("Update: repo.Update was not called")
	}
	if repo.updatedEntry.FoodName != "Chicken" {
		t.Errorf("Update: food_name not trimmed: got %q, want \"Chicken\"", repo.updatedEntry.FoodName)
	}
	if repo.updatedEntry.Calories != 400 {
		t.Errorf("Update: calories = %d, want 400", repo.updatedEntry.Calories)
	}
	if repo.updatedEntry.Price != 50.0 {
		t.Errorf("Update: price changed unexpectedly: got %.1f, want 50.0", repo.updatedEntry.Price)
	}
	if resp.FoodName != "Chicken" {
		t.Errorf("Update: response FoodName = %q, want \"Chicken\"", resp.FoodName)
	}
}

func TestUpdate_ForbiddenForWrongUser(t *testing.T) {
	entry := &FoodEntry{UserID: 1}
	entry.ID = 10
	repo := &mockRepo{entry: entry}
	svc := newTestSvc(repo)

	_, err := svc.Update(10, 2, UpdateFoodEntryRequest{FoodName: strPtr("Chicken")})

	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Update wrong user: got %v, want ErrForbidden", err)
	}
	if repo.updateCalled {
		t.Error("repo.Update must not be called when user is forbidden")
	}
}

func TestUpdate_PropagatesRepoError(t *testing.T) {
	entry := &FoodEntry{UserID: 1, FoodName: "Rice"}
	entry.ID = 10
	repo := &mockRepo{entry: entry, updateErr: errors.New("db error")}
	svc := newTestSvc(repo)

	_, err := svc.Update(10, 1, UpdateFoodEntryRequest{FoodName: strPtr("Chicken")})

	if err == nil || err.Error() != "db error" {
		t.Errorf("Update: got %v, want \"db error\"", err)
	}
}

// --- Test: Delete ---

func TestDelete_Success(t *testing.T) {
	entry := &FoodEntry{UserID: 1}
	entry.ID = 10
	repo := &mockRepo{entry: entry}
	svc := newTestSvc(repo)

	err := svc.Delete(10, 1)

	if err != nil {
		t.Fatalf("Delete: unexpected error %v", err)
	}
	if repo.findByIDArg != 10 {
		t.Errorf("Delete: FindByID called with id=%d, want 10", repo.findByIDArg)
	}
	if !repo.deleteCalled {
		t.Fatal("Delete: repo.Delete was not called")
	}
	if repo.deletedID != 10 {
		t.Errorf("Delete: repo.Delete called with id=%d, want 10", repo.deletedID)
	}
}

func TestDelete_ForbiddenForWrongUser(t *testing.T) {
	entry := &FoodEntry{UserID: 1}
	entry.ID = 10
	repo := &mockRepo{entry: entry}
	svc := newTestSvc(repo)

	err := svc.Delete(10, 2)

	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Delete wrong user: got %v, want ErrForbidden", err)
	}
	if repo.deleteCalled {
		t.Error("repo.Delete must not be called when user is forbidden")
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := &mockRepo{} // entry = nil → gorm.ErrRecordNotFound
	svc := newTestSvc(repo)

	err := svc.Delete(99, 1)

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Delete not found: got %v, want gorm.ErrRecordNotFound", err)
	}
}

// --- Test: DailySummary ---

func TestDailySummaryCalcsAndExceededFlag(t *testing.T) {
	u := &user.Users{DailyCalorieLimit: 2100, MonthlyPriceLimit: 1000}
	u.ID = 1
	date := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	t.Run("exceeds both limits", func(t *testing.T) {
		s, err := newTestSvc(&mockRepo{calSum: 2200, priceSum: 1100}).DailySummary(u, date)
		if err != nil {
			t.Fatal(err)
		}
		if s.TotalCalories != 2200 {
			t.Errorf("TotalCalories: got %d, want 2200", s.TotalCalories)
		}
		if s.TotalPrice != 1100 {
			t.Errorf("TotalPrice: got %.1f, want 1100.0", s.TotalPrice)
		}
		if s.CalorieLimit != 2100 {
			t.Errorf("CalorieLimit: got %d, want 2100", s.CalorieLimit)
		}
		if s.PriceLimit != 1000 {
			t.Errorf("PriceLimit: got %.1f, want 1000.0", s.PriceLimit)
		}
		if !s.CalorieExceeded {
			t.Error("CalorieExceeded should be true (2200 > 2100)")
		}
		if !s.PriceExceeded {
			t.Error("PriceExceeded should be true (1100 > 1000)")
		}
		if s.Date != "2024-06-15" {
			t.Errorf("Date: got %q, want \"2024-06-15\"", s.Date)
		}
	})

	t.Run("exactly at limit is not exceeded", func(t *testing.T) {
		s, err := newTestSvc(&mockRepo{calSum: 2100, priceSum: 1000}).DailySummary(u, date)
		if err != nil {
			t.Fatal(err)
		}
		if s.CalorieExceeded {
			t.Error("CalorieExceeded should be false at exact limit (2100 == 2100)")
		}
		if s.PriceExceeded {
			t.Error("PriceExceeded should be false at exact limit (1000 == 1000)")
		}
	})
}

func TestDailySummary_PropagatesCalorieRepoError(t *testing.T) {
	u := &user.Users{DailyCalorieLimit: 2100, MonthlyPriceLimit: 1000}
	u.ID = 1
	repo := &mockRepo{sumCalOnDayErr: errors.New("db error")}
	svc := newTestSvc(repo)

	_, err := svc.DailySummary(u, time.Now())

	if err == nil || err.Error() != "db error" {
		t.Errorf("DailySummary: got %v, want \"db error\"", err)
	}
}

func TestDailySummary_PropagatesPriceRepoError(t *testing.T) {
	u := &user.Users{DailyCalorieLimit: 2100, MonthlyPriceLimit: 1000}
	u.ID = 1
	repo := &mockRepo{sumPriceInMonthErr: errors.New("db error")}
	svc := newTestSvc(repo)

	_, err := svc.DailySummary(u, time.Now())

	if err == nil || err.Error() != "db error" {
		t.Errorf("DailySummary: got %v, want \"db error\"", err)
	}
}

// --- Test: ListDailySummaries ---

func TestListDailySummaries(t *testing.T) {
	u := &user.Users{DailyCalorieLimit: 2000, MonthlyPriceLimit: 500}
	u.ID = 1
	from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	repo := &mockRepo{
		calPerDay: map[string]int{
			"2024-06-01": 1800,
			"2024-06-02": 2100, // exceeds 2000 limit
			"2024-06-03": 0,
		},
		pricePerMonth: map[string]float64{
			"2024-06": 600, // exceeds 500 limit
		},
	}
	svc := newTestSvc(repo)

	result, err := svc.ListDailySummaries(u, from, to)

	if err != nil {
		t.Fatalf("ListDailySummaries: unexpected error %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("ListDailySummaries: got %d days, want 3", len(result))
	}

	// day 1: under calorie limit, over price limit
	if result[0].Date != "2024-06-01" {
		t.Errorf("result[0].Date = %q, want \"2024-06-01\"", result[0].Date)
	}
	if result[0].TotalCalories != 1800 {
		t.Errorf("result[0].TotalCalories = %d, want 1800", result[0].TotalCalories)
	}
	if result[0].CalorieExceeded {
		t.Error("result[0].CalorieExceeded should be false (1800 < 2000)")
	}
	if !result[0].PriceExceeded {
		t.Error("result[0].PriceExceeded should be true (600 > 500)")
	}

	// day 2: over calorie limit
	if !result[1].CalorieExceeded {
		t.Error("result[1].CalorieExceeded should be true (2100 > 2000)")
	}

	// day 3: zero calories, must still appear in result
	if result[2].TotalCalories != 0 {
		t.Errorf("result[2].TotalCalories = %d, want 0", result[2].TotalCalories)
	}
}

// --- Test: ListAll ---

func TestListAll(t *testing.T) {
	e1 := FoodEntryWithUser{FoodEntry: FoodEntry{FoodName: "Rice"}, UserName: "Alice"}
	e2 := FoodEntryWithUser{FoodEntry: FoodEntry{FoodName: "Chicken"}, UserName: "Bob"}
	e1.ID = 1
	e2.ID = 2
	repo := &mockRepo{allEntries: []FoodEntryWithUser{e1, e2}}
	svc := newTestSvc(repo)

	result, err := svc.ListAll(nil, nil)

	if err != nil {
		t.Fatalf("ListAll: unexpected error %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("ListAll: got %d entries, want 2", len(result))
	}
	if result[0].FoodName != "Rice" || result[0].UserName != "Alice" {
		t.Errorf("ListAll: result[0] = {%s %s}, want {Rice Alice}", result[0].FoodName, result[0].UserName)
	}
}

func TestListAll_PropagatesRepoError(t *testing.T) {
	repo := &mockRepo{findAllErr: errors.New("db error")}
	svc := newTestSvc(repo)

	_, err := svc.ListAll(nil, nil)

	if err == nil || err.Error() != "db error" {
		t.Errorf("ListAll: got %v, want \"db error\"", err)
	}
}

func TestListDailySummaries_PropagatesRepoError(t *testing.T) {
	u := &user.Users{DailyCalorieLimit: 2000, MonthlyPriceLimit: 500}
	u.ID = 1
	from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)
	repo := &mockRepo{sumCalPerDayErr: errors.New("db error")}
	svc := newTestSvc(repo)

	_, err := svc.ListDailySummaries(u, from, to)

	if err == nil || err.Error() != "db error" {
		t.Errorf("ListDailySummaries: got %v, want \"db error\"", err)
	}
}

// --- Test: AdminGetByID ---

func TestAdminGetByID_Found(t *testing.T) {
	entry := &FoodEntry{FoodName: "Rice", Calories: 300}
	entry.ID = 5
	entry.UserID = 1
	repo := &mockRepo{entry: entry}
	svc := newTestSvc(repo)

	resp, err := svc.AdminGetByID(5)

	if err != nil {
		t.Fatalf("AdminGetByID: unexpected error %v", err)
	}
	if resp.ID != 5 {
		t.Errorf("AdminGetByID: ID = %d, want 5", resp.ID)
	}
	if resp.FoodName != "Rice" {
		t.Errorf("AdminGetByID: FoodName = %q, want \"Rice\"", resp.FoodName)
	}
}

func TestAdminGetByID_NotFound(t *testing.T) {
	repo := &mockRepo{} // entry = nil
	svc := newTestSvc(repo)

	_, err := svc.AdminGetByID(99)

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("AdminGetByID not found: got %v, want gorm.ErrRecordNotFound", err)
	}
}

// --- Test: AdminCreate ---

func TestAdminCreate_Success(t *testing.T) {
	// FindByIDWithUser is called after Create; entry must be set for it
	entry := &FoodEntry{FoodName: "Rice", Calories: 300}
	entry.ID = 1
	entry.UserID = 2
	repo := &mockRepo{entry: entry}
	userRepo := &mockUserRepo{foundUser: &user.Users{}}
	svc := newTestSvcFull(repo, userRepo)

	resp, err := svc.AdminCreate(AdminCreateFoodEntryRequest{
		UserID:    2,
		FoodName:  "  Rice  ",
		Calories:  intPtr(300),
		Price:     floatPtr(60.0),
		EntryDate: time.Now(),
	})

	if err != nil {
		t.Fatalf("AdminCreate: unexpected error %v", err)
	}
	// verify userRepo.GetUserByID was called with the correct user id before repo.Create
	if userRepo.getUserByIDCount != 1 {
		t.Errorf("AdminCreate: GetUserByID called %d times, want 1", userRepo.getUserByIDCount)
	}
	if userRepo.getUserByIDCalledWith != 2 {
		t.Errorf("AdminCreate: GetUserByID called with id=%d, want 2", userRepo.getUserByIDCalledWith)
	}
	if !repo.createCalled {
		t.Fatal("AdminCreate: repo.Create was not called")
	}
	if repo.createdEntry.FoodName != "Rice" {
		t.Errorf("AdminCreate: food_name not trimmed: got %q, want \"Rice\"", repo.createdEntry.FoodName)
	}
	if repo.createdEntry.UserID != 2 {
		t.Errorf("AdminCreate: createdEntry.UserID = %d, want 2", repo.createdEntry.UserID)
	}
	if resp == nil {
		t.Fatal("AdminCreate: got nil response")
	}
}

func TestAdminCreate_EmptyFoodName(t *testing.T) {
	repo := &mockRepo{}
	svc := newTestSvc(repo)

	_, err := svc.AdminCreate(AdminCreateFoodEntryRequest{
		UserID:    1,
		FoodName:  "   ",
		Calories:  intPtr(300),
		Price:     floatPtr(60.0),
		EntryDate: time.Now(),
	})

	if err == nil || err.Error() != "food_name cannot be empty" {
		t.Errorf("AdminCreate empty name: got %v, want \"food_name cannot be empty\"", err)
	}
	if repo.createCalled {
		t.Error("repo.Create must not be called when validation fails")
	}
}

func TestAdminCreate_UserNotFound(t *testing.T) {
	repo := &mockRepo{}
	svc := newTestSvcFull(repo, &mockUserRepo{err: gorm.ErrRecordNotFound})

	_, err := svc.AdminCreate(AdminCreateFoodEntryRequest{
		UserID:    99,
		FoodName:  "Rice",
		Calories:  intPtr(300),
		Price:     floatPtr(60.0),
		EntryDate: time.Now(),
	})

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("AdminCreate user not found: got %v, want ErrUserNotFound", err)
	}
	if repo.createCalled {
		t.Error("repo.Create must not be called when user not found")
	}
}

// --- Test: AdminUpdate ---

func TestAdminUpdate_Success(t *testing.T) {
	entry := &FoodEntry{FoodName: "Rice", Calories: 300, Price: 50.0}
	entry.ID = 10
	entry.UserID = 1
	repo := &mockRepo{entry: entry}
	svc := newTestSvc(repo)

	resp, err := svc.AdminUpdate(10, UpdateFoodEntryRequest{
		FoodName: strPtr("  Steak  "),
		Calories: intPtr(700),
		// Price not set — should remain 50.0
	})

	if err != nil {
		t.Fatalf("AdminUpdate: unexpected error %v", err)
	}
	if !repo.updateCalled {
		t.Fatal("AdminUpdate: repo.Update was not called")
	}
	if repo.updatedEntry.FoodName != "Steak" {
		t.Errorf("AdminUpdate: food_name not trimmed: got %q, want \"Steak\"", repo.updatedEntry.FoodName)
	}
	if repo.updatedEntry.Calories != 700 {
		t.Errorf("AdminUpdate: calories = %d, want 700", repo.updatedEntry.Calories)
	}
	if repo.updatedEntry.Price != 50.0 {
		t.Errorf("AdminUpdate: price changed unexpectedly: got %.1f, want 50.0", repo.updatedEntry.Price)
	}
	if resp.FoodName != "Steak" {
		t.Errorf("AdminUpdate: response FoodName = %q, want \"Steak\"", resp.FoodName)
	}
}

func TestAdminUpdate_EmptyFoodName(t *testing.T) {
	entry := &FoodEntry{FoodName: "Rice"}
	entry.ID = 10
	repo := &mockRepo{entry: entry}
	svc := newTestSvc(repo)

	_, err := svc.AdminUpdate(10, UpdateFoodEntryRequest{FoodName: strPtr("  ")})

	if err == nil || err.Error() != "food_name cannot be empty" {
		t.Errorf("AdminUpdate empty name: got %v, want \"food_name cannot be empty\"", err)
	}
	if repo.updateCalled {
		t.Error("repo.Update must not be called when validation fails")
	}
}

func TestAdminUpdate_NotFound(t *testing.T) {
	repo := &mockRepo{} // entry = nil
	svc := newTestSvc(repo)

	_, err := svc.AdminUpdate(99, UpdateFoodEntryRequest{FoodName: strPtr("X")})

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("AdminUpdate not found: got %v, want gorm.ErrRecordNotFound", err)
	}
}

// --- Test: AdminDelete ---

func TestAdminDelete_Success(t *testing.T) {
	entry := &FoodEntry{}
	entry.ID = 10
	repo := &mockRepo{entry: entry}
	svc := newTestSvc(repo)

	err := svc.AdminDelete(10)

	if err != nil {
		t.Fatalf("AdminDelete: unexpected error %v", err)
	}
	if !repo.deleteCalled {
		t.Fatal("AdminDelete: repo.Delete was not called")
	}
	if repo.deletedID != 10 {
		t.Errorf("AdminDelete: repo.Delete called with id=%d, want 10", repo.deletedID)
	}
}

func TestAdminDelete_NotFound(t *testing.T) {
	repo := &mockRepo{} // entry = nil
	svc := newTestSvc(repo)

	err := svc.AdminDelete(99)

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("AdminDelete not found: got %v, want gorm.ErrRecordNotFound", err)
	}
}

// --- Test: GetReport ---

func TestGetReportWindowIncludesToday(t *testing.T) {
	repo := &mockRepo{last7Count: 5, prev7Count: 3, usersCount: 2, last7CalSum: 1000}
	report, err := newTestSvc(repo).GetReport()
	if err != nil {
		t.Fatal(err)
	}
	if report.EntriesLast7Days != 5 {
		t.Errorf("last7: want 5, got %d", report.EntriesLast7Days)
	}
	if report.EntriesPrevious7Days != 3 {
		t.Errorf("prev7: want 3, got %d", report.EntriesPrevious7Days)
	}
	if report.UsersCount != 2 {
		t.Errorf("users_count: want 2, got %d", report.UsersCount)
	}
	if report.AvgCaloriesPerUserLast7D != 500 {
		t.Errorf("avg_calories_per_user_last_7_days: want 500, got %v", report.AvgCaloriesPerUserLast7D)
	}
	if report.Comparison.CurrentWeek != 5 || report.Comparison.PreviousWeek != 3 || report.Comparison.Difference != 2 {
		t.Errorf("comparison: want {current:5 previous:3 difference:2}, got %+v", report.Comparison)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	expectedTo := today.AddDate(0, 0, 1)
	if !repo.capturedTo.Equal(expectedTo) {
		t.Errorf("last7 end boundary: want %v (tomorrow), got %v — today must be included", expectedTo, repo.capturedTo)
	}
}
