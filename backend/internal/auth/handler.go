package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"cau-used-goods-app/backend/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type wechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type devLoginRequest struct {
	OpenID string `json:"openid"`
	Role   string `json:"role"`
}

type reactivateRequest struct {
	ReactivationToken string `json:"reactivationToken" binding:"required"`
	Confirm           bool   `json:"confirm"`
}

func (h *Handler) DevLogin(c *gin.Context) {
	var req devLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数格式不正确")
		return
	}

	result, err := h.service.DevLogin(c.Request.Context(), DevLoginInput{
		OpenID: req.OpenID,
		Role:   req.Role,
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *Handler) WechatLogin(c *gin.Context) {
	var req wechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "微信登录 code 不能为空")
		return
	}

	result, err := h.service.WechatLogin(c.Request.Context(), req.Code)
	if err != nil {
		if strings.Contains(err.Error(), "code") {
			response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *Handler) Reactivate(c *gin.Context) {
	var req reactivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "请求参数格式不正确")
		return
	}
	result, err := h.service.Reactivate(c.Request.Context(), req.ReactivationToken, req.Confirm)
	if err != nil {
		response.Error(c, http.StatusConflict, response.CodeConflict, err.Error())
		return
	}
	response.Success(c, result)
}
