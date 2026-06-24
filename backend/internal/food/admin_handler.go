package food

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/captainthx/calorie/backend/pkg/response"
)

type AdminHandler struct {
	svc *FoodService
}

func NewAdminHandler(svc *FoodService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// ListAll godoc
// @Summary List all food entries
// @Description List all food entries across users for admin.
// @Tags Admin Food Entries
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} response.SuccessBody{data=[]AdminFoodEntryResponse}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/admin/food-entries [get]
func (h *AdminHandler) ListAll(c *gin.Context) {
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
	entries, err := h.svc.ListAll(dateFrom, dateTo)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	response.Success(c, entries)
}

// GetByID godoc
// @Summary Get food entry by ID
// @Description Get a food entry by ID for admin.
// @Tags Admin Food Entries
// @Security BearerAuth
// @Produce json
// @Param id path int true "Food entry ID"
// @Success 200 {object} response.SuccessBody{data=AdminFoodEntryResponse}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/admin/food-entries/{id} [get]
func (h *AdminHandler) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	entry, err := h.svc.AdminGetByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, entry)
}

// Create godoc
// @Summary Create food entry for user
// @Description Create a food entry for any user as admin.
// @Tags Admin Food Entries
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AdminCreateFoodEntryRequest true "Admin create payload"
// @Success 200 {object} response.SuccessBody{data=AdminFoodEntryResponse}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/admin/food-entries [post]
func (h *AdminHandler) Create(c *gin.Context) {
	var req AdminCreateFoodEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.FoodName = strings.TrimSpace(req.FoodName)
	if req.FoodName == "" {
		response.BadRequest(c, "food_name cannot be empty")
		return
	}
	res, err := h.svc.AdminCreate(req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, res)
}

// Update godoc
// @Summary Update food entry for admin
// @Description Partially update any food entry as admin.
// @Tags Admin Food Entries
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Food entry ID"
// @Param request body UpdateFoodEntryRequest true "Partial update payload"
// @Success 200 {object} response.SuccessBody{data=AdminFoodEntryResponse}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/admin/food-entries/{id} [patch]
func (h *AdminHandler) Update(c *gin.Context) {
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
	res, err := h.svc.AdminUpdate(id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, res)
}

// FullUpdate godoc
// @Summary Replace food entry for admin
// @Description Fully replace any food entry as admin.
// @Tags Admin Food Entries
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Food entry ID"
// @Param request body PutFoodEntryRequest true "Full update payload"
// @Success 200 {object} response.SuccessBody{data=AdminFoodEntryResponse}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/admin/food-entries/{id} [put]
func (h *AdminHandler) FullUpdate(c *gin.Context) {
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
	res, err := h.svc.AdminUpdate(id, patchReq)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, res)
}

// Delete godoc
// @Summary Delete food entry for admin
// @Description Delete any food entry as admin.
// @Tags Admin Food Entries
// @Security BearerAuth
// @Produce json
// @Param id path int true "Food entry ID"
// @Success 200 {object} response.SuccessBody{data=response.MessageData}
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/admin/food-entries/{id} [delete]
func (h *AdminHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.AdminDelete(id); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, response.MessageData{Message: "Entry deleted successfully"})
}

// Report godoc
// @Summary Get admin report
// @Description Get aggregate admin report across recent food entries.
// @Tags Admin Reports
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.SuccessBody{data=ReportResponse}
// @Failure 401 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /api/admin/reports [get]
func (h *AdminHandler) Report(c *gin.Context) {
	report, err := h.svc.GetReport()
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	response.Success(c, report)
}
