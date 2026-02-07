package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/task-monitor/api-server/internal/model"
)

// MockAuthService is a mock for AuthServiceInterface
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

func TestJWTAuth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)

	mockSvc.On("ParseToken", "valid-token").Return(uint(1), "admin", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/nodes", nil)
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	handler := JWTAuth(mockSvc)
	handler(c)

	assert.False(t, c.IsAborted())
	userID, _ := c.Get("userID")
	assert.Equal(t, uint(1), userID)
	username, _ := c.Get("username")
	assert.Equal(t, "admin", username)
	mockSvc.AssertExpectations(t)
}

func TestJWTAuth_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/nodes", nil)

	handler := JWTAuth(mockSvc)
	handler(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/nodes", nil)
	c.Request.Header.Set("Authorization", "InvalidFormat")

	handler := JWTAuth(mockSvc)
	handler(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)

	mockSvc.On("ParseToken", "expired-token").Return(uint(0), "", errors.New("invalid token"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/nodes", nil)
	c.Request.Header.Set("Authorization", "Bearer expired-token")

	handler := JWTAuth(mockSvc)
	handler(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertExpectations(t)
}
