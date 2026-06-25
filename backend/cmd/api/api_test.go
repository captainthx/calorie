//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/captainthx/calorie/backend/internal/config"
	"github.com/captainthx/calorie/backend/internal/food"
	"github.com/captainthx/calorie/backend/internal/middleware"
	routes "github.com/captainthx/calorie/backend/internal/routers"
	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	testRouter *gin.Engine
	testDB     *gorm.DB
	johnID     uint
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../../"); err != nil {
		panic(err)
	}
	initTimezone()
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	testDB = cfg.Db

	createUserRoleEnumQuery := `DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role_enum') THEN
			CREATE TYPE user_role_enum AS ENUM ('ADMIN', 'USER');
		END IF;
	END $$`
	if err := testDB.Exec(createUserRoleEnumQuery).Error; err != nil {
		panic(fmt.Sprintf("create enum: %v", err))
	}

	if err := testDB.AutoMigrate(&user.Users{}, &food.FoodEntry{}); err != nil {
		panic(fmt.Sprintf("auto migrate: %v", err))
	}
	testDB.Exec("TRUNCATE TABLE food_entries RESTART IDENTITY CASCADE")
	testDB.Model(&user.Users{}).
		Where("daily_calorie_limit = 0 AND role = ?", user.User).
		Updates(map[string]interface{}{"daily_calorie_limit": 2100, "monthly_price_limit": 1000})

	var count int64
	testDB.Model(&user.Users{}).Count(&count)
	if count == 0 {
		testDB.Create(&[]user.Users{
			{Name: "John", Role: user.User, Token: "user-token-123", DailyCalorieLimit: 2100, MonthlyPriceLimit: 1000},
			{Name: "Jane", Role: user.User, Token: "user-token-456", DailyCalorieLimit: 2100, MonthlyPriceLimit: 1000},
			{Name: "Admin", Role: user.Admin, Token: "admin-token-789"},
		})
	}
	var john, jane user.Users
	testDB.Where("token = ?", "user-token-123").First(&john)
	testDB.Where("token = ?", "user-token-456").First(&jane)
	johnID = john.ID
	seedFoodEntries(testDB, john.ID, jane.ID)

	gin.SetMode("test")
	testRouter = gin.New()
	routes.RegisterPublicRoutes(testRouter, testDB)
	userRepo := user.NewUsersRepository(testDB)
	api := testRouter.Group("/api", middleware.AuthMiddleware(userRepo))
	admin := api.Group("/admin", middleware.AdminMiddleware())
	routes.RegisterRoutes(api, testDB)
	routes.RegisterAdminRoutes(admin, testDB)

	os.Exit(m.Run())
}

func apiReq(method, path, token, body string) *httptest.ResponseRecorder {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	httpReq := httptest.NewRequest(method, path, r)
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, httpReq)
	return w
}

func decodeBody(b []byte) map[string]any {
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}

func dataField(b []byte) map[string]any {
	m := decodeBody(b)
	d, _ := m["data"].(map[string]any)
	return d
}

func checkCode(t *testing.T, name string, w *httptest.ResponseRecorder, want int) bool {
	t.Helper()
	if w.Code != want {
		t.Errorf("FAIL %s: status=%d body=%s", name, w.Code, w.Body.String())
		return false
	}
	return true
}

func assertEq(t *testing.T, name string, want, got any) {
	t.Helper()
	if want != got {
		t.Errorf("FAIL %s: expected %v, got %v", name, want, got)
	}
}

func assertFloatNear(t *testing.T, name string, want, got, delta float64) {
	t.Helper()
	if got < want-delta || got > want+delta {
		t.Errorf("FAIL %s: expected %v +/- %v, got %v", name, want, delta, got)
	}
}

type integrationState struct {
	today        string
	yesterday    string
	userEntryID  uint
	adminEntryID uint
}

func TestIntegration(t *testing.T) {
	state := &integrationState{
		today:     time.Now().Format("2006-01-02"),
		yesterday: ago(1, 0).Format("2006-01-02"),
	}

	runPublicRouteTests(t)
	runAuthChecks(t)
	runDailySummaryTests(t, state)
	runAdminReportTests(t)
	runUserFoodEntryTests(t, state)
	runAdminFoodEntryTests(t, state)
}

func runPublicRouteTests(t *testing.T) {
	t.Helper()

	t.Run("health_ok", func(t *testing.T) {
		w := apiReq("GET", "/health", "", "")
		if !checkCode(t, "health_ok", w, 200) {
			return
		}
		assertEq(t, "status", "ok", decodeBody(w.Body.Bytes())["status"])
	})
}

func runAuthChecks(t *testing.T) {
	t.Helper()

	// Auth Checks
	t.Run("no_token_401", func(t *testing.T) {
		w := apiReq("GET", "/api/food-entries", "", "")
		checkCode(t, "no_token", w, 401)
	})
	t.Run("invalid_token_401", func(t *testing.T) {
		w := apiReq("GET", "/api/food-entries", "bad-token-xyz", "")
		checkCode(t, "invalid_token", w, 401)
	})
	t.Run("user_hits_admin_403", func(t *testing.T) {
		w := apiReq("GET", "/api/admin/food-entries", "user-token-123", "")
		checkCode(t, "user_hits_admin", w, 403)
	})
}

func runDailySummaryTests(t *testing.T, state *integrationState) {
	t.Helper()

	// Daily Summary
	t.Run("daily_summary_today_john", func(t *testing.T) {
		w := apiReq("GET", "/api/daily-summary", "user-token-123", "")
		if !checkCode(t, "daily_summary_today_john", w, 200) {
			return
		}
		d := dataField(w.Body.Bytes())
		assertEq(t, "total_calories", float64(900), d["total_calories"])
		assertEq(t, "calorie_exceeded", false, d["calorie_exceeded"])
		assertEq(t, "total_price", float64(845), d["total_price"])
		assertEq(t, "price_exceeded", false, d["price_exceeded"])
		assertEq(t, "calorie_limit", float64(2100), d["calorie_limit"])
	})
	t.Run("daily_summary_yesterday_john", func(t *testing.T) {
		w := apiReq("GET", "/api/daily-summary?date="+state.yesterday, "user-token-123", "")
		if !checkCode(t, "daily_summary_yesterday_john", w, 200) {
			return
		}
		d := dataField(w.Body.Bytes())
		assertEq(t, "total_calories", float64(2200), d["total_calories"])
		assertEq(t, "calorie_exceeded", true, d["calorie_exceeded"])
		assertEq(t, "total_price", float64(845), d["total_price"])
	})
	t.Run("daily_summary_jane_price_exceeded", func(t *testing.T) {
		w := apiReq("GET", "/api/daily-summary", "user-token-456", "")
		if !checkCode(t, "daily_summary_jane_price_exceeded", w, 200) {
			return
		}
		d := dataField(w.Body.Bytes())
		assertEq(t, "total_calories", float64(600), d["total_calories"])
		assertEq(t, "calorie_exceeded", false, d["calorie_exceeded"])
		assertEq(t, "total_price", float64(1565), d["total_price"])
		assertEq(t, "price_exceeded", true, d["price_exceeded"])
	})
	t.Run("daily_summary_bad_date_400", func(t *testing.T) {
		w := apiReq("GET", "/api/daily-summary?date=not-a-date", "user-token-123", "")
		if !checkCode(t, "daily_summary_bad_date", w, 400) {
			return
		}
		assertEq(t, "success", false, decodeBody(w.Body.Bytes())["success"])
	})
	t.Run("daily_summaries_removed_404", func(t *testing.T) {
		from := ago(6, 0).Format("2006-01-02")
		w := apiReq("GET", "/api/daily-summaries?date_from="+from+"&date_to="+state.today, "user-token-123", "")
		checkCode(t, "daily_summaries_removed", w, 404)
	})
}

func runUserFoodEntryTests(t *testing.T, state *integrationState) {
	t.Helper()

	// User Food Entries
	t.Run("list_no_filter", func(t *testing.T) {
		w := apiReq("GET", "/api/food-entries", "user-token-123", "")
		if !checkCode(t, "list_no_filter", w, 200) {
			return
		}
		m := decodeBody(w.Body.Bytes())
		assertEq(t, "success", true, m["success"])
		if _, ok := m["data"].([]any); !ok {
			t.Errorf("FAIL list_no_filter: data not array: %v", m["data"])
		}
	})
	t.Run("list_date_filter", func(t *testing.T) {
		w := apiReq("GET", "/api/food-entries?date_from="+state.today+"&date_to="+state.today, "user-token-123", "")
		checkCode(t, "list_date_filter", w, 200)
	})
	t.Run("list_bad_date_400", func(t *testing.T) {
		w := apiReq("GET", "/api/food-entries?date_from=baddate", "user-token-123", "")
		checkCode(t, "list_bad_date", w, 400)
	})
	t.Run("list_date_from_after_date_to_400", func(t *testing.T) {
		w := apiReq("GET", "/api/food-entries?date_from="+state.today+"&date_to="+state.yesterday, "user-token-123", "")
		checkCode(t, "list_date_from_after_date_to", w, 400)
	})
	t.Run("create_entry", func(t *testing.T) {
		body := fmt.Sprintf(`{"food_name":"Test Meal","calories":500,"price":75.5,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("POST", "/api/food-entries", "user-token-123", body)
		if !checkCode(t, "create_entry", w, 200) {
			return
		}
		d := dataField(w.Body.Bytes())
		assertEq(t, "food_name", "Test Meal", d["food_name"])
		assertEq(t, "calories", float64(500), d["calories"])
		assertEq(t, "price", 75.5, d["price"])
		if id, ok := d["id"].(float64); ok && id > 0 {
			state.userEntryID = uint(id)
		} else {
			t.Errorf("FAIL create_entry: missing id: %s", w.Body.String())
		}
	})
	t.Run("create_calories_zero", func(t *testing.T) {
		body := fmt.Sprintf(`{"food_name":"Zero Cal Snack","calories":0,"price":10.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("POST", "/api/food-entries", "user-token-123", body)
		if !checkCode(t, "create_calories_zero", w, 200) {
			return
		}
		assertEq(t, "calories_zero", float64(0), dataField(w.Body.Bytes())["calories"])
	})
	t.Run("create_upper_bounds", func(t *testing.T) {
		body := fmt.Sprintf(`{"food_name":"Max Meal","calories":10000,"price":10000.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("POST", "/api/food-entries", "user-token-123", body)
		if !checkCode(t, "create_upper_bounds", w, 200) {
			return
		}
		d := dataField(w.Body.Bytes())
		assertEq(t, "calories", float64(10000), d["calories"])
		assertEq(t, "price", float64(10000), d["price"])
	})
	t.Run("create_empty_food_name_400", func(t *testing.T) {
		body := fmt.Sprintf(`{"food_name":"  ","calories":100,"price":10.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("POST", "/api/food-entries", "user-token-123", body)
		if !checkCode(t, "create_empty_food_name", w, 400) {
			return
		}
		assertEq(t, "success", false, decodeBody(w.Body.Bytes())["success"])
	})
	t.Run("create_calories_too_high_400", func(t *testing.T) {
		body := fmt.Sprintf(`{"food_name":"Too High Cal","calories":10001,"price":10.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("POST", "/api/food-entries", "user-token-123", body)
		checkCode(t, "create_calories_too_high", w, 400)
	})
	t.Run("create_calories_negative_400", func(t *testing.T) {
		body := fmt.Sprintf(`{"food_name":"Negative Cal","calories":-1,"price":10.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("POST", "/api/food-entries", "user-token-123", body)
		checkCode(t, "create_calories_negative", w, 400)
	})
	t.Run("create_price_too_high_400", func(t *testing.T) {
		body := fmt.Sprintf(`{"food_name":"Too Expensive","calories":100,"price":10001.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("POST", "/api/food-entries", "user-token-123", body)
		checkCode(t, "create_price_too_high", w, 400)
	})
	t.Run("create_price_negative_400", func(t *testing.T) {
		body := fmt.Sprintf(`{"food_name":"Negative Price","calories":100,"price":-1.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("POST", "/api/food-entries", "user-token-123", body)
		checkCode(t, "create_price_negative", w, 400)
	})
	t.Run("put_full_update", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID from create_entry")
		}
		body := fmt.Sprintf(`{"food_name":"Updated Meal","calories":600,"price":80.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("PUT", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-123", body)
		if !checkCode(t, "put_full_update", w, 200) {
			return
		}
		d := dataField(w.Body.Bytes())
		assertEq(t, "food_name", "Updated Meal", d["food_name"])
		assertEq(t, "calories", float64(600), d["calories"])
	})
	t.Run("put_missing_field_400", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID")
		}
		w := apiReq("PUT", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-123", `{"food_name":"Partial"}`)
		checkCode(t, "put_missing_field", w, 400)
	})
	t.Run("patch_calories_only", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID")
		}
		w := apiReq("PATCH", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-123", `{"calories":350}`)
		if !checkCode(t, "patch_calories_only", w, 200) {
			return
		}
		assertEq(t, "calories", float64(350), dataField(w.Body.Bytes())["calories"])
	})
	t.Run("patch_calories_zero", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID")
		}
		w := apiReq("PATCH", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-123", `{"calories":0}`)
		if !checkCode(t, "patch_calories_zero", w, 200) {
			return
		}
		assertEq(t, "calories_zero", float64(0), dataField(w.Body.Bytes())["calories"])
	})
	t.Run("patch_calories_too_high_400", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID")
		}
		w := apiReq("PATCH", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-123", `{"calories":10001}`)
		checkCode(t, "patch_calories_too_high", w, 400)
	})
	t.Run("patch_calories_negative_400", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID")
		}
		w := apiReq("PATCH", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-123", `{"calories":-1}`)
		checkCode(t, "patch_calories_negative", w, 400)
	})
	t.Run("put_price_too_high_400", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID")
		}
		body := fmt.Sprintf(`{"food_name":"Too Expensive Update","calories":100,"price":10001.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("PUT", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-123", body)
		checkCode(t, "put_price_too_high", w, 400)
	})
	t.Run("put_price_negative_400", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID")
		}
		body := fmt.Sprintf(`{"food_name":"Negative Price Update","calories":100,"price":-1.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("PUT", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-123", body)
		checkCode(t, "put_price_negative", w, 400)
	})
	t.Run("patch_jane_john_entry_403", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID")
		}
		w := apiReq("PATCH", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-456", `{"calories":999}`)
		if !checkCode(t, "patch_jane_john_entry", w, 403) {
			return
		}
		assertEq(t, "error", "forbidden", decodeBody(w.Body.Bytes())["error"])
	})
	t.Run("delete_wrong_user_403", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID")
		}
		w := apiReq("DELETE", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-456", "")
		if !checkCode(t, "delete_wrong_user", w, 403) {
			return
		}
		assertEq(t, "error", "forbidden", decodeBody(w.Body.Bytes())["error"])
	})
	t.Run("delete_own_entry", func(t *testing.T) {
		if state.userEntryID == 0 {
			t.Skip("no entry ID")
		}
		w := apiReq("DELETE", fmt.Sprintf("/api/food-entries/%d", state.userEntryID), "user-token-123", "")
		checkCode(t, "delete_own_entry", w, 200)
	})
}

func runAdminFoodEntryTests(t *testing.T, state *integrationState) {
	t.Helper()

	// Admin Food Entries
	t.Run("admin_list_all", func(t *testing.T) {
		w := apiReq("GET", "/api/admin/food-entries", "admin-token-789", "")
		if !checkCode(t, "admin_list_all", w, 200) {
			return
		}
		if _, ok := decodeBody(w.Body.Bytes())["data"].([]any); !ok {
			t.Errorf("FAIL admin_list_all: data not array: %s", w.Body.String())
		}
	})
	t.Run("admin_list_all_date_filter", func(t *testing.T) {
		w := apiReq("GET", "/api/admin/food-entries?date_from="+state.today+"&date_to="+state.today, "admin-token-789", "")
		checkCode(t, "admin_list_all_date_filter", w, 200)
	})
	t.Run("admin_list_all_bad_date_400", func(t *testing.T) {
		w := apiReq("GET", "/api/admin/food-entries?date_from=bad", "admin-token-789", "")
		checkCode(t, "admin_list_all_bad_date", w, 400)
	})
	t.Run("admin_create_for_john", func(t *testing.T) {
		body := fmt.Sprintf(`{"user_id":%d,"food_name":"Admin Created Meal","calories":300,"price":50.0,"entry_date":"%sT12:00:00Z"}`, johnID, state.today)
		w := apiReq("POST", "/api/admin/food-entries", "admin-token-789", body)
		if !checkCode(t, "admin_create_for_john", w, 200) {
			return
		}
		d := dataField(w.Body.Bytes())
		assertEq(t, "user_id", float64(johnID), d["user_id"])
		if id, ok := d["id"].(float64); ok && id > 0 {
			state.adminEntryID = uint(id)
		} else {
			t.Errorf("FAIL admin_create_for_john: no id: %s", w.Body.String())
		}
	})
	t.Run("admin_create_upper_bounds", func(t *testing.T) {
		body := fmt.Sprintf(`{"user_id":%d,"food_name":"Admin Max Meal","calories":10000,"price":10000.0,"entry_date":"%sT12:00:00Z"}`, johnID, state.today)
		w := apiReq("POST", "/api/admin/food-entries", "admin-token-789", body)
		if !checkCode(t, "admin_create_upper_bounds", w, 200) {
			return
		}
		d := dataField(w.Body.Bytes())
		assertEq(t, "calories", float64(10000), d["calories"])
		assertEq(t, "price", float64(10000), d["price"])
	})
	t.Run("admin_create_user_not_found_404", func(t *testing.T) {
		body := fmt.Sprintf(`{"user_id":9999,"food_name":"Ghost Meal","calories":100,"price":10.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("POST", "/api/admin/food-entries", "admin-token-789", body)
		if !checkCode(t, "admin_create_user_not_found", w, 404) {
			return
		}
		assertEq(t, "success", false, decodeBody(w.Body.Bytes())["success"])
	})
	t.Run("admin_create_price_too_high_400", func(t *testing.T) {
		body := fmt.Sprintf(`{"user_id":%d,"food_name":"Admin Too Expensive","calories":100,"price":10001.0,"entry_date":"%sT12:00:00Z"}`, johnID, state.today)
		w := apiReq("POST", "/api/admin/food-entries", "admin-token-789", body)
		checkCode(t, "admin_create_price_too_high", w, 400)
	})
	t.Run("admin_create_calories_negative_400", func(t *testing.T) {
		body := fmt.Sprintf(`{"user_id":%d,"food_name":"Admin Negative Cal","calories":-1,"price":10.0,"entry_date":"%sT12:00:00Z"}`, johnID, state.today)
		w := apiReq("POST", "/api/admin/food-entries", "admin-token-789", body)
		checkCode(t, "admin_create_calories_negative", w, 400)
	})
	t.Run("admin_create_price_negative_400", func(t *testing.T) {
		body := fmt.Sprintf(`{"user_id":%d,"food_name":"Admin Negative Price","calories":100,"price":-1.0,"entry_date":"%sT12:00:00Z"}`, johnID, state.today)
		w := apiReq("POST", "/api/admin/food-entries", "admin-token-789", body)
		checkCode(t, "admin_create_price_negative", w, 400)
	})
	t.Run("admin_get_by_id", func(t *testing.T) {
		if state.adminEntryID == 0 {
			t.Skip("no admin entry ID")
		}
		w := apiReq("GET", fmt.Sprintf("/api/admin/food-entries/%d", state.adminEntryID), "admin-token-789", "")
		if !checkCode(t, "admin_get_by_id", w, 200) {
			return
		}
		if dataField(w.Body.Bytes())["food_name"] == nil {
			t.Errorf("FAIL admin_get_by_id: food_name missing: %s", w.Body.String())
		}
	})
	t.Run("admin_put", func(t *testing.T) {
		if state.adminEntryID == 0 {
			t.Skip("no admin entry ID")
		}
		body := fmt.Sprintf(`{"food_name":"Admin Updated Meal","calories":400,"price":60.0,"entry_date":"%sT12:00:00Z"}`, state.today)
		w := apiReq("PUT", fmt.Sprintf("/api/admin/food-entries/%d", state.adminEntryID), "admin-token-789", body)
		if !checkCode(t, "admin_put", w, 200) {
			return
		}
		assertEq(t, "food_name", "Admin Updated Meal", dataField(w.Body.Bytes())["food_name"])
	})
	t.Run("admin_patch", func(t *testing.T) {
		if state.adminEntryID == 0 {
			t.Skip("no admin entry ID")
		}
		w := apiReq("PATCH", fmt.Sprintf("/api/admin/food-entries/%d", state.adminEntryID), "admin-token-789", `{"food_name":"Patched Name"}`)
		if !checkCode(t, "admin_patch", w, 200) {
			return
		}
		assertEq(t, "food_name", "Patched Name", dataField(w.Body.Bytes())["food_name"])
	})
	t.Run("admin_patch_price_too_high_400", func(t *testing.T) {
		if state.adminEntryID == 0 {
			t.Skip("no admin entry ID")
		}
		w := apiReq("PATCH", fmt.Sprintf("/api/admin/food-entries/%d", state.adminEntryID), "admin-token-789", `{"price":10001}`)
		checkCode(t, "admin_patch_price_too_high", w, 400)
	})
	t.Run("admin_patch_price_negative_400", func(t *testing.T) {
		if state.adminEntryID == 0 {
			t.Skip("no admin entry ID")
		}
		w := apiReq("PATCH", fmt.Sprintf("/api/admin/food-entries/%d", state.adminEntryID), "admin-token-789", `{"price":-1}`)
		checkCode(t, "admin_patch_price_negative", w, 400)
	})
	t.Run("admin_delete", func(t *testing.T) {
		if state.adminEntryID == 0 {
			t.Skip("no admin entry ID")
		}
		w := apiReq("DELETE", fmt.Sprintf("/api/admin/food-entries/%d", state.adminEntryID), "admin-token-789", "")
		checkCode(t, "admin_delete", w, 200)
	})
}

func runAdminReportTests(t *testing.T) {
	t.Helper()

	// Admin Reports
	t.Run("admin_get_report", func(t *testing.T) {
		w := apiReq("GET", "/api/admin/reports", "admin-token-789", "")
		if !checkCode(t, "admin_get_report", w, 200) {
			return
		}
		d := dataField(w.Body.Bytes())
		assertEq(t, "entries_last_7_days", float64(17), d["entries_last_7_days"])
		assertEq(t, "entries_previous_7_days", float64(8), d["entries_previous_7_days"])
		assertEq(t, "users_count", float64(3), d["users_count"])
		avg, ok := d["average_calories_per_user_last_7_days"].(float64)
		if !ok {
			t.Fatalf("FAIL admin_get_report: average_calories_per_user_last_7_days not float: %v", d["average_calories_per_user_last_7_days"])
		}
		assertFloatNear(t, "average_calories_per_user_last_7_days", 9550.0/3.0, avg, 0.000001)
		comparison, ok := d["entries_comparison"].(map[string]any)
		if !ok {
			t.Fatalf("FAIL admin_get_report: entries_comparison not object: %v", d["entries_comparison"])
		}
		assertEq(t, "comparison.current_week", float64(17), comparison["current_week"])
		assertEq(t, "comparison.previous_week", float64(8), comparison["previous_week"])
		assertEq(t, "comparison.difference", float64(9), comparison["difference"])
	})
}

func TestDatabasePoolConfig(t *testing.T) {
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 25 {
		t.Errorf("MaxOpenConnections = %d, want 25", stats.MaxOpenConnections)
	}
}
