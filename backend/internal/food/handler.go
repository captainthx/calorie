package food

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// Create godoc
// @Summary Create food entry
// @Description Create a food entry for the authenticated user.
// @Tags Food Entries
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateFoodEntryRequest true "Food entry payload"
// @Success 200 {object} response.SuccessBody{data=FoodEntryResponse}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/food-entries [post]
func (h *Handler) Create(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	var req CreateFoodEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.FoodName = strings.TrimSpace(req.FoodName)
	if req.FoodName == "" {
		response.BadRequest(c, "food_name cannot be empty")
		return
	}
	res, err := h.svc.Create(u, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	response.Success(c, res)
}

// List godoc
// @Summary List food entries
// @Description List food entries for the authenticated user, optionally filtered by date range.
// @Tags Food Entries
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} response.SuccessBody{data=[]FoodEntryResponse}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/food-entries [get]
func (h *Handler) List(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	dateFrom, err := parseOptionalDate(c.Query("date_from"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	dateTo, err := parseOptionalDate(c.Query("date_to"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if dateFrom != nil && dateTo != nil && dateFrom.After(*dateTo) {
		response.BadRequest(c, "date_from must not be after date_to")
		return
	}
	entries, err := h.svc.List(u.ID, dateFrom, dateTo)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	response.Success(c, entries)
}

// Update godoc
// @Summary Update food entry
// @Description Partially update a food entry owned by the authenticated user.
// @Tags Food Entries
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Food entry ID"
// @Param request body UpdateFoodEntryRequest true "Partial update payload"
// @Success 200 {object} response.SuccessBody{data=FoodEntryResponse}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/food-entries/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req UpdateFoodEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.FoodName != nil && strings.TrimSpace(*req.FoodName) == "" {
		response.BadRequest(c, "food_name cannot be empty")
		return
	}
	res, err := h.svc.Update(id, u.ID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, res)
}

// FullUpdate godoc
// @Summary Replace food entry
// @Description Fully replace a food entry owned by the authenticated user.
// @Tags Food Entries
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Food entry ID"
// @Param request body PutFoodEntryRequest true "Full update payload"
// @Success 200 {object} response.SuccessBody{data=FoodEntryResponse}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/food-entries/{id} [put]
func (h *Handler) FullUpdate(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req PutFoodEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.FoodName = strings.TrimSpace(req.FoodName)
	if req.FoodName == "" {
		response.BadRequest(c, "food_name cannot be empty")
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

// Delete godoc
// @Summary Delete food entry
// @Description Delete a food entry owned by the authenticated user.
// @Tags Food Entries
// @Security BearerAuth
// @Produce json
// @Param id path int true "Food entry ID"
// @Success 200 {object} response.SuccessBody{data=response.MessageData}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/food-entries/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.Delete(id, u.ID); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, response.MessageData{Message: "Entry deleted successfully"})
}

// DailySummary godoc
// @Summary Get daily summary
// @Description Get the authenticated user's daily calorie summary and monthly price total.
// @Tags Summaries
// @Security BearerAuth
// @Produce json
// @Param date query string false "Date (YYYY-MM-DD)"
// @Success 200 {object} response.SuccessBody{data=DailySummaryResponse}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/daily-summary [get]
func (h *Handler) DailySummary(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	var date time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			response.BadRequest(c, "invalid date format, use YYYY-MM-DD")
			return
		}
	} else {
		date = time.Now()
	}
	summary, err := h.svc.DailySummary(u, date)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	response.Success(c, summary)
}

// DailySummaryRange godoc
// @Summary Get daily summary for date range
// @Description Get per-day calorie summary for the authenticated user over a date range.
// @Tags Summaries
// @Security BearerAuth
// @Produce json
// @Param date_from query string true "Start date (YYYY-MM-DD)"
// @Param date_to query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.SuccessBody{data=[]DailySummaryRangeItem}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/daily-summary-range [get]
func (h *Handler) DailySummaryRange(c *gin.Context) {
	u := c.MustGet("user").(*user.Users)
	from, err := parseOptionalDate(c.Query("date_from"))
	if err != nil || from == nil {
		response.BadRequest(c, "date_from is required and must be YYYY-MM-DD")
		return
	}
	to, err := parseOptionalDate(c.Query("date_to"))
	if err != nil || to == nil {
		response.BadRequest(c, "date_to is required and must be YYYY-MM-DD")
		return
	}
	if from.After(*to) {
		response.BadRequest(c, "date_from must not be after date_to")
		return
	}
	result, err := h.svc.DailySummaryRange(u, *from, *to)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	response.Success(c, result)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrForbidden):
		response.Forbidden(c, "forbidden")
	case errors.Is(err, ErrUserNotFound):
		response.NotFound(c, "user not found")
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.NotFound(c, "entry not found")
	default:
		response.InternalServerError(c, err)
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
