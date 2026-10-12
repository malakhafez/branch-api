package branch

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes بتسجّل مسارات الـ module على group بيجيلها من main.go.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/branches", h.create)
	rg.GET("/branches", h.list)
	rg.GET("/branches/:id", h.get)
	rg.PATCH("/branches/:id", h.update)
}

func (h *Handler) create(c *gin.Context) {
	var req CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "request body is not valid JSON")
		return
	}
	b, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ToResponse(b))
}

func (h *Handler) list(c *gin.Context) {
	branches, err := h.svc.List(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	out := make([]BranchResponse, 0, len(branches))
	for _, b := range branches {
		out = append(out, ToResponse(b))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_ID", "branch id is not a valid UUID")
		return
	}
	b, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToResponse(b))
}

func (h *Handler) update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_ID", "branch id is not a valid UUID")
		return
	}
	var req UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "request body is not valid JSON")
		return
	}
	b, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToResponse(b))
}

// handleError بتحوّل أخطاء الـ module لكود HTTP وكود خطأ.
// أي خطأ مش معرّف هنا بيرجع 500 برسالة عامة (من غير تفاصيل الداتا بيز).
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(c, http.StatusNotFound, "BRANCH_NOT_FOUND", err.Error())
	case errors.Is(err, ErrNameTaken):
		writeError(c, http.StatusConflict, "BRANCH_NAME_TAKEN", err.Error())
	case errors.Is(err, ErrNameRequired):
		writeError(c, http.StatusBadRequest, "BRANCH_NAME_REQUIRED", err.Error())
	case errors.Is(err, ErrNameTooLong):
		writeError(c, http.StatusBadRequest, "BRANCH_NAME_TOO_LONG", err.Error())
	case errors.Is(err, ErrInvalidType):
		writeError(c, http.StatusBadRequest, "BRANCH_INVALID_TYPE", err.Error())
	case errors.Is(err, ErrInvalidTime):
		writeError(c, http.StatusBadRequest, "BRANCH_INVALID_TIME", err.Error())
	case errors.Is(err, ErrLateRateRequired):
		writeError(c, http.StatusBadRequest, "BRANCH_LATE_RATE_REQUIRED", err.Error())
	case errors.Is(err, ErrInvalidLateRate):
		writeError(c, http.StatusBadRequest, "BRANCH_INVALID_LATE_RATE", err.Error())
	case errors.Is(err, ErrLateRateTooLarge):
		writeError(c, http.StatusBadRequest, "BRANCH_LATE_RATE_TOO_LARGE", err.Error())
	case errors.Is(err, ErrLateRatePrecision):
		writeError(c, http.StatusBadRequest, "BRANCH_LATE_RATE_PRECISION", err.Error())
	default:
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

// writeError بتكتب الخطأ بالشكل المتفق عليه في الفريق.
func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{"code": code, "message": message},
	})
}
