package food

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/captainthx/calorie/backend/pkg/response"
)

type Handler struct {
	svc *FoodService
}

func NewHandler(svc *FoodService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	var req CreateFoodEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.svc.Create(u, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, res)
}

func (h *Handler) List(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	dateFrom, err := parseOptionalDate(c.Query("date_from"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	dateTo, err := parseOptionalDate(c.Query("date_to"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if dateFrom != nil && dateTo != nil && dateFrom.After(*dateTo) {
		response.Error(c, http.StatusBadRequest, "date_from must not be after date_to")
		return
	}
	entries, err := h.svc.List(u.ID, dateFrom, dateTo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, entries)
}

func (h *Handler) Update(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	id, err := parseID(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req UpdateFoodEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.svc.Update(id, u.ID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, res)
}

func (h *Handler) FullUpdate(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	id, err := parseID(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req PutFoodEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	patchReq := UpdateFoodEntryRequest{
		FoodName:  &req.FoodName,
		Calories:  req.Calories,
		Price:     req.Price,
		EntryDate: &req.EntryDate,
	}
	res, err := h.svc.Update(id, u.ID, patchReq)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, res)
}

func (h *Handler) Delete(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	id, err := parseID(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(id, u.ID); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Entry deleted successfully"})
}

func (h *Handler) DailySummary(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	var date time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
	} else {
		date = time.Now()
	}
	summary, err := h.svc.DailySummary(u, date)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, summary)
}

func (h *Handler) DailySummaryRange(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	dateFrom, err := parseOptionalDate(c.Query("date_from"))
	if err != nil || dateFrom == nil {
		response.Error(c, http.StatusBadRequest, "date_from is required (YYYY-MM-DD)")
		return
	}
	dateTo, err := parseOptionalDate(c.Query("date_to"))
	if err != nil || dateTo == nil {
		response.Error(c, http.StatusBadRequest, "date_to is required (YYYY-MM-DD)")
		return
	}
	if dateFrom.After(*dateTo) {
		response.Error(c, http.StatusBadRequest, "date_from must not be after date_to")
		return
	}
	summaries, err := h.svc.ListDailySummaries(u, *dateFrom, *dateTo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, summaries)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden")
	case errors.Is(err, ErrUserNotFound):
		response.Error(c, http.StatusNotFound, "user not found")
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(c, http.StatusNotFound, "entry not found")
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id), err
}

func parseOptionalDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q, use YYYY-MM-DD", s)
	}
	return &t, nil
}
