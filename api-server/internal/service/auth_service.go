package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/task-monitor/api-server/internal/model"
	"github.com/task-monitor/api-server/internal/repository"
)

// AuthService 认证服务
type AuthService struct {
	userRepo      repository.UserRepositoryInterface
	jwtSecret     string
	expireMinutes int
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repository.UserRepositoryInterface, jwtSecret string, expireMinutes int) *AuthService {
	if expireMinutes <= 0 {
		expireMinutes = 1440 // 24h
	}
	return &AuthService{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		expireMinutes: expireMinutes,
	}
}

// Login 用户登录，返回JWT token
func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return "", errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("用户名或密码错误")
	}
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Duration(s.expireMinutes) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ParseToken 解析JWT token
func (s *AuthService) ParseToken(tokenString string) (uint, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", errors.New("invalid claims")
	}

	userIDRaw, ok := claims["user_id"]
	if !ok {
		return 0, "", errors.New("invalid claims")
	}

	var userID uint
	switch v := userIDRaw.(type) {
	case float64:
		if v < 0 {
			return 0, "", errors.New("invalid claims")
		}
		userID = uint(v)
	case json.Number:
		parsed, err := strconv.ParseUint(v.String(), 10, 64)
		if err != nil {
			return 0, "", errors.New("invalid claims")
		}
		userID = uint(parsed)
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, "", errors.New("invalid claims")
		}
		userID = uint(parsed)
	default:
		return 0, "", errors.New("invalid claims")
	}

	usernameRaw, ok := claims["username"]
	if !ok {
		return 0, "", errors.New("invalid claims")
	}
	username, ok := usernameRaw.(string)
	if !ok || username == "" {
		return 0, "", errors.New("invalid claims")
	}

	return userID, username, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

// ListUsers 获取所有用户
func (s *AuthService) ListUsers() ([]model.User, error) {
	return s.userRepo.FindAll()
}

// CreateUser 创建用户
func (s *AuthService) CreateUser(username, password string) (*model.User, error) {
	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return nil, errors.New("用户名已存在")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{Username: username, Password: string(hashedPassword)}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uint, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}

// DeleteUser 删除用户（不能删除自己）
func (s *AuthService) DeleteUser(userID uint, currentUserID uint) error {
	if userID == currentUserID {
		return errors.New("不能删除当前登录用户")
	}
	if _, err := s.userRepo.FindByID(userID); err != nil {
		return errors.New("用户不存在")
	}
	return s.userRepo.Delete(userID)
}
