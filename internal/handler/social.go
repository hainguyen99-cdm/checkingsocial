package handler

import (
	"checkingsocial/internal/model"
	"checkingsocial/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SocialHandler xử lý các request liên quan đến mạng xã hội.
type SocialHandler struct {
	service service.Checker
}

// NewSocialHandler tạo một instance mới của SocialHandler.
func NewSocialHandler(s service.Checker) *SocialHandler {
	return &SocialHandler{service: s}
}

// RegisterRoutes đăng ký các route cho social handler.
func (h *SocialHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// Route mới cho social action
		api.POST("/social-action", h.SocialAction)
	}
}

// SocialAction xử lý request thực hiện hành động trên mạng xã hội.
// @Summary Thực hiện một hành động trên mạng xã hội
// @Description Nhận một hành động và trả về true nếu xử lý thành công.
// @Tags Social
// @Accept json
// @Produce json
// @Param request body model.SocialActionRequest true "Yêu cầu hành động"
// @Success 200 {boolean} boolean "Trả về true hoặc false"
// @Failure 400 {object} map[string]string "Lỗi validation"
// @Failure 500 {object} map[string]string "Lỗi server"
// @Router /social-action [post]
func (h *SocialHandler) SocialAction(c *gin.Context) {
	var req model.SocialActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.CheckSocialAction(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
