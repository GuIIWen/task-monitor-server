package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/task-monitor/api-server/internal/model"
)

// MockAuthService is a mock implementation of AuthServiceInterface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ParseToken(tokenString string) (uint, string, error) {
	args := m.Called(tokenString)
	return args.Get(0).(uint), args.String(1), args.Error(2)
}

func (m *MockAuthService) GetUserByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) ListUsers() ([]model.User, error) {
	args := m.Called()
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockAuthService) CreateUser(username, password string) (*model.User, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) ChangePassword(userID uint, newPassword string) error {
	args := m.Called(userID, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) DeleteUser(userID uint, currentUserID uint) error {
	args := m.Called(userID, currentUserID)
	return args.Error(0)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("Login", "admin", "admin123").Return("jwt-token-123", nil)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "jwt-token-123", data["token"])
	assert.Equal(t, "admin", data["username"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("Login", "admin", "wrong").Return("", errors.New("用户名或密码错误"))

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc)

	body, _ := json.Marshal(map[string]string{"username": "admin"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_ListUsers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("ListUsers").Return([]model.User{
		{ID: 1, Username: "admin"},
		{ID: 2, Username: "user2"},
	}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/users", nil)

	h.ListUsers(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 2)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_CreateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("CreateUser", "newuser", "pass123").Return(&model.User{
		ID: 2, Username: "newuser",
	}, nil)

	body, _ := json.Marshal(map[string]string{"username": "newuser", "password": "pass123"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.CreateUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_CreateUser_Duplicate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("CreateUser", "admin", "pass123").Return(nil, errors.New("用户名已存在"))

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "pass123"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.CreateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("ChangePassword", uint(1), "newpass123").Return(nil)

	body, _ := json.Marshal(map[string]string{"password": "newpass123"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/v1/users/1/password", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.ChangePassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_DeleteUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("DeleteUser", uint(2), uint(1)).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/v1/users/2", nil)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Set("userID", uint(1))

	h.DeleteUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_DeleteUser_Self(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc)

	mockSvc.On("DeleteUser", uint(1), uint(1)).Return(errors.New("不能删除当前登录用户"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/v1/users/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("userID", uint(1))

	h.DeleteUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertExpectations(t)
}
