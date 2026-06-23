package food

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/captainthx/calorie/backend/pkg/response"
)

type AdminHandler struct {
	svc *FoodService
}

func NewAdminHandler(svc *FoodService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) ListAll(c *gin.Context) {
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
	entries, err := h.svc.ListAll(dateFrom, dateTo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, entries)
}

func (h *AdminHandler) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	entry, err := h.svc.AdminGetByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, entry)
}

func (h *AdminHandler) Create(c *gin.Context) {
	var req AdminCreateFoodEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.svc.AdminCreate(req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, res)
}

func (h *AdminHandler) Update(c *gin.Context) {
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
	res, err := h.svc.AdminUpdate(id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, res)
}

func (h *AdminHandler) FullUpdate(c *gin.Context) {
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
	res, err := h.svc.AdminUpdate(id, patchReq)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, res)
}

func (h *AdminHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.AdminDelete(id); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Entry deleted successfully"})
}

func (h *AdminHandler) Report(c *gin.Context) {
	report, err := h.svc.GetReport()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, report)
}
