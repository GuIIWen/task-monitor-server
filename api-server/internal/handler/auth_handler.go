package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/service"
	"github.com/task-monitor/api-server/internal/utils"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthServiceInterface
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请输入用户名和密码")
		return
	}
	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	utils.SuccessResponse(c, LoginResponse{Token: token, Username: req.Username})
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("userID")
	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}
	utils.SuccessResponse(c, user)
}

// ListUsers 获取用户列表
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.authService.ListUsers()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取用户列表失败")
		return
	}
	utils.SuccessResponse(c, users)
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateUser 创建用户
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请输入用户名和密码")
		return
	}
	user, err := h.authService.CreateUser(req.Username, req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(c, user)
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

// ChangePassword 修改用户密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID")
		return
	}
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请输入新密码")
		return
	}
	if err := h.authService.ChangePassword(uint(id), req.Password); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}

// DeleteUser 删除用户
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID")
		return
	}
	currentUserID, _ := c.Get("userID")
	if err := h.authService.DeleteUser(uint(id), currentUserID.(uint)); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(c, nil)
}
