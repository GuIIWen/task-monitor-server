package service

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/task-monitor/api-server/internal/model"
)

// MockUserRepository is a mock implementation of UserRepositoryInterface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindAll() ([]model.User, error) {
	args := m.Called()
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func newHashedPassword(plain string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	return string(h)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	hashed := newHashedPassword("admin123")
	mockRepo.On("FindByUsername", "admin").Return(&model.User{
		ID: 1, Username: "admin", Password: hashed,
	}, nil)

	token, err := svc.Login("admin", "admin123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	hashed := newHashedPassword("admin123")
	mockRepo.On("FindByUsername", "admin").Return(&model.User{
		ID: 1, Username: "admin", Password: hashed,
	}, nil)

	token, err := svc.Login("admin", "wrongpass")
	assert.Error(t, err)
	assert.Equal(t, "用户名或密码错误", err.Error())
	assert.Empty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	mockRepo.On("FindByUsername", "nobody").Return(nil, errors.New("not found"))

	token, err := svc.Login("nobody", "pass")
	assert.Error(t, err)
	assert.Equal(t, "用户名或密码错误", err.Error())
	assert.Empty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ParseToken_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	hashed := newHashedPassword("admin123")
	mockRepo.On("FindByUsername", "admin").Return(&model.User{
		ID: 1, Username: "admin", Password: hashed,
	}, nil)

	token, err := svc.Login("admin", "admin123")
	assert.NoError(t, err)

	userID, username, err := svc.ParseToken(token)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), userID)
	assert.Equal(t, "admin", username)
}

func TestAuthService_ParseToken_Invalid(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	_, _, err := svc.ParseToken("invalid-token")
	assert.Error(t, err)
	assert.Equal(t, "invalid token", err.Error())
}

func TestAuthService_ParseToken_WrongSecret(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc1 := NewAuthService(mockRepo, "secret-1", 24)
	svc2 := NewAuthService(mockRepo, "secret-2", 24)

	hashed := newHashedPassword("pass")
	mockRepo.On("FindByUsername", "user1").Return(&model.User{
		ID: 1, Username: "user1", Password: hashed,
	}, nil)

	token, _ := svc1.Login("user1", "pass")
	_, _, err := svc2.ParseToken(token)
	assert.Error(t, err)
}

func TestAuthService_ParseToken_InvalidClaims_NoPanic(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	claims := jwt.MapClaims{
		"user_id":  "not-a-number",
		"username": 123,
		"exp":      time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	assert.NotPanics(t, func() {
		_, _, parseErr := svc.ParseToken(signed)
		assert.Error(t, parseErr)
		assert.Equal(t, "invalid claims", parseErr.Error())
	})
}

func TestAuthService_CreateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	mockRepo.On("FindByUsername", "newuser").Return(nil, errors.New("not found"))
	mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil)

	user, err := svc.CreateUser("newuser", "password123")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newuser", user.Username)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_CreateUser_Duplicate(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	mockRepo.On("FindByUsername", "admin").Return(&model.User{
		ID: 1, Username: "admin",
	}, nil)

	user, err := svc.CreateUser("admin", "password123")
	assert.Error(t, err)
	assert.Equal(t, "用户名已存在", err.Error())
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	hashed := newHashedPassword("oldpass")
	mockRepo.On("FindByID", uint(1)).Return(&model.User{
		ID: 1, Username: "admin", Password: hashed,
	}, nil)
	mockRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	err := svc.ChangePassword(1, "newpass123")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ChangePassword_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	mockRepo.On("FindByID", uint(99)).Return(nil, errors.New("not found"))

	err := svc.ChangePassword(99, "newpass")
	assert.Error(t, err)
	assert.Equal(t, "用户不存在", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestAuthService_DeleteUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	mockRepo.On("FindByID", uint(2)).Return(&model.User{
		ID: 2, Username: "user2",
	}, nil)
	mockRepo.On("Delete", uint(2)).Return(nil)

	err := svc.DeleteUser(2, 1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_DeleteUser_Self(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	err := svc.DeleteUser(1, 1)
	assert.Error(t, err)
	assert.Equal(t, "不能删除当前登录用户", err.Error())
}

func TestAuthService_DeleteUser_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	mockRepo.On("FindByID", uint(99)).Return(nil, errors.New("not found"))

	err := svc.DeleteUser(99, 1)
	assert.Error(t, err)
	assert.Equal(t, "用户不存在", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ListUsers(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewAuthService(mockRepo, "test-secret", 24)

	mockRepo.On("FindAll").Return([]model.User{
		{ID: 1, Username: "admin"},
		{ID: 2, Username: "user2"},
	}, nil)

	users, err := svc.ListUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	mockRepo.AssertExpectations(t)
}
