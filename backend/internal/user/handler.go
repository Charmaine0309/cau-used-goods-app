package user

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"cau-used-goods-app/backend/internal/middleware"
	"cau-used-goods-app/backend/pkg/response"
)

type Handler struct {
	service *Service
}

type updateProfileRequest struct {
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatarUrl"`
	Phone     *string `json:"phone"`
}

type submitStudentVerificationRequest struct {
	StudentID string `json:"studentId" binding:"required"`
	RealName  string `json:"realName" binding:"required"`
	College   string `json:"college" binding:"required"`
}

type reviewStudentVerificationRequest struct {
	AuthStatus  string `json:"authStatus" binding:"required"`
	Description string `json:"description"`
}

type updateAccountStatusRequest struct {
	AccountStatus string `json:"accountStatus" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
	RelatedType   string `json:"relatedType"`
	RelatedID     uint64 `json:"relatedId"`
}

type updateRoleRequest struct {
	Role   string `json:"role" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

type cancelAccountRequest struct {
	Confirm bool `json:"confirm"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Me(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.Me(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, err.Error())
		return
	}
	response.Success(c, user)
}

func (h *Handler) PublicProfile(c *gin.Context) {
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	item, err := h.service.PublicProfile(c.Request.Context(), userID)
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) Restriction(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}
	item, err := h.service.Restriction(c.Request.Context(), userID)
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) CancelAccount(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}
	var req cancelAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid request body")
		return
	}
	if err := h.service.CancelAccount(c.Request.Context(), userID, req.Confirm); err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, gin.H{"accountStatus": "CANCELED"})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}
	if err := h.service.EnsureAccountNormal(c.Request.Context(), userID); err != nil {
		writeUserError(c, err)
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid request body")
		return
	}

	if req.AvatarURL != nil {
		avatarURL, err := saveRemoteAvatar(*req.AvatarURL)
		if err != nil {
			response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
			return
		}
		req.AvatarURL = &avatarURL
	}

	user, err := h.service.UpdateProfile(c.Request.Context(), userID, UpdateProfileInput{
		Nickname:  req.Nickname,
		AvatarURL: req.AvatarURL,
		Phone:     req.Phone,
	})
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, user)
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}
	if err := h.service.EnsureAccountNormal(c.Request.Context(), userID); err != nil {
		writeUserError(c, err)
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "avatar file is required")
		return
	}
	if file.Size > 2*1024*1024 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "avatar file must be <= 2MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "avatar file must be jpg, jpeg, png or webp")
		return
	}

	name, err := randomFileName(ext)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "generate avatar filename failed")
		return
	}

	dir := filepath.Join("uploads", "avatar")
	if err := os.MkdirAll(dir, 0755); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "create avatar directory failed")
		return
	}

	dst := filepath.Join(dir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "save avatar file failed")
		return
	}

	avatarURL := "/" + filepath.ToSlash(dst)
	user, err := h.service.UpdateProfile(c.Request.Context(), userID, UpdateProfileInput{
		AvatarURL: &avatarURL,
	})
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, gin.H{
		"avatarUrl": avatarURL,
		"user":      user,
	})
}

func (h *Handler) SubmitStudentVerification(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	var req submitStudentVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid request body")
		return
	}

	verification, err := h.service.SubmitStudentVerification(c.Request.Context(), userID, SubmitStudentVerificationInput{
		StudentID: req.StudentID,
		RealName:  req.RealName,
		College:   req.College,
	})
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, verification)
}

func randomFileName(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf) + ext, nil
}

func saveRemoteAvatar(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("avatarUrl is empty")
	}
	if strings.HasPrefix(rawURL, "/uploads/avatar/") {
		return rawURL, nil
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", fmt.Errorf("avatarUrl must be http(s) URL or /uploads/avatar path")
	}

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("download avatar failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download avatar failed with status %d", resp.StatusCode)
	}
	if resp.ContentLength > 2*1024*1024 {
		return "", fmt.Errorf("avatar file must be <= 2MB")
	}

	ext := avatarExt(resp.Header.Get("Content-Type"), filepath.Ext(parsed.Path))
	if ext == "" {
		return "", fmt.Errorf("avatar file must be jpg, jpeg, png or webp")
	}

	name, err := randomFileName(ext)
	if err != nil {
		return "", fmt.Errorf("generate avatar filename failed")
	}

	dir := filepath.Join("uploads", "avatar")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create avatar directory failed")
	}

	dst := filepath.Join(dir, name)
	out, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create avatar file failed")
	}
	defer out.Close()

	limited := io.LimitReader(resp.Body, 2*1024*1024+1)
	written, err := io.Copy(out, limited)
	if err != nil {
		return "", fmt.Errorf("save avatar file failed")
	}
	if written > 2*1024*1024 {
		_ = os.Remove(dst)
		return "", fmt.Errorf("avatar file must be <= 2MB")
	}

	return "/" + filepath.ToSlash(dst), nil
}

func avatarExt(contentType string, pathExt string) string {
	switch strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	}

	switch strings.ToLower(pathExt) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return strings.ToLower(pathExt)
	default:
		return ""
	}
}

func (h *Handler) StudentVerification(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	verification, err := h.service.StudentVerification(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, err.Error())
		return
	}
	response.Success(c, verification)
}

func (h *Handler) ListStudentVerifications(c *gin.Context) {
	status := c.DefaultQuery("authStatus", "PENDING")
	items, err := h.service.ListStudentVerifications(c.Request.Context(), status)
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *Handler) ListAdminUsers(c *gin.Context) {
	page, pageSize := parsePage(c)
	role, _ := middleware.CurrentRole(c)
	result, err := h.service.ListAdminUsers(c.Request.Context(), AdminUserQuery{
		Keyword:       c.Query("keyword"),
		AuthStatus:    c.Query("authStatus"),
		AccountStatus: c.Query("accountStatus"),
		Role:          c.Query("role"),
		Page:          page,
		PageSize:      pageSize,
	}, role == roleSuperAdmin)
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) AdminUserDetail(c *gin.Context) {
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	role, _ := middleware.CurrentRole(c)
	result, err := h.service.AdminUserDetail(c.Request.Context(), userID, role == roleSuperAdmin, role == roleSuperAdmin)
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) UpdateAccountStatus(c *gin.Context) {
	adminID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	var req updateAccountStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid request body")
		return
	}
	user, err := h.service.UpdateAccountStatus(c.Request.Context(), adminID, UpdateAccountStatusInput{
		UserID:        userID,
		AccountStatus: req.AccountStatus,
		Reason:        req.Reason,
		IPAddress:     c.ClientIP(),
		RelatedType:   req.RelatedType,
		RelatedID:     req.RelatedID,
	})
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, user)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	adminID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid request body")
		return
	}
	user, err := h.service.UpdateRole(c.Request.Context(), adminID, UpdateRoleInput{
		UserID:    userID,
		Role:      req.Role,
		Reason:    req.Reason,
		IPAddress: c.ClientIP(),
	})
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, user)
}

func (h *Handler) ListUserProducts(c *gin.Context) {
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	result, err := h.service.ListUserProducts(c.Request.Context(), userID, page, pageSize, c.Query("dataScope"))
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListUserOrders(c *gin.Context) {
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	result, err := h.service.ListUserOrders(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListUserReports(c *gin.Context) {
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	result, err := h.service.ListUserReports(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListUserAppeals(c *gin.Context) {
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	result, err := h.service.ListUserAppeals(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListUserReviews(c *gin.Context) {
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	result, err := h.service.ListUserReviews(c.Request.Context(), userID, page, pageSize, c.Query("dataScope"))
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListUserLogs(c *gin.Context) {
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	role, _ := middleware.CurrentRole(c)
	result, err := h.service.ListUserLogs(c.Request.Context(), userID, page, pageSize, role == roleSuperAdmin)
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ReviewStudentVerification(c *gin.Context) {
	adminID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid user id")
		return
	}

	var req reviewStudentVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid request body")
		return
	}

	verification, err := h.service.ReviewStudentVerification(c.Request.Context(), adminID, ReviewStudentVerificationInput{
		UserID:      userID,
		AuthStatus:  req.AuthStatus,
		Description: req.Description,
	})
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, verification)
}

func writeUserError(c *gin.Context, err error) {
	message := err.Error()
	switch message {
	case "当前账号状态不可操作", "目标用户账号状态不可审核", "管理员不能审核自己的认证",
		"管理员不能修改自己的账号状态", "管理员不能修改自己的角色", "需要超级管理员权限",
		"不能修改超级管理员账号状态", "普通管理员只能修改普通用户账号状态",
		"不能通过接口修改超级管理员角色", "只有普通用户可以主动注销":
		response.Error(c, http.StatusForbidden, response.CodeForbidden, message)
	case "用户不存在":
		response.Error(c, http.StatusNotFound, response.CodeNotFound, message)
	case "学号已被使用",
		"学生认证正在审核中，请勿重复提交",
		"学生认证已通过，不能重复提交",
		"当前认证状态不可提交",
		"认证状态已变化，请刷新后重试",
		"当前账号状态不可变更",
		"目标用户账号状态不可操作",
		"不允许的账号状态流转",
		"撤销永久封禁必须关联已通过的申诉",
		"当前账号状态不可注销",
		"存在进行中订单，暂不能注销",
		"当前账号状态不可重新激活",
		"只有正常账号可以修改角色",
		"目标用户已经是该角色":
		response.Error(c, http.StatusConflict, response.CodeConflict, message)
	default:
		if isInternalUserError(message) {
			response.Error(c, http.StatusInternalServerError, response.CodeInternal, "服务器内部错误")
			return
		}
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, message)
	}
}

func isInternalUserError(message string) bool {
	prefixes := []string{
		"begin ", "commit ", "find ", "list ", "count ", "scan ", "iterate ",
		"query ", "check related record", "lock ", "update ", "create ",
		"off shelf ", "cancel account:", "reactivate account:", "submit student",
		"review student", "get affected rows",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	return false
}

func parseUserIDParam(c *gin.Context) (uint64, bool) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid user id")
		return 0, false
	}
	return userID, true
}

func parsePage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	return page, pageSize
}
