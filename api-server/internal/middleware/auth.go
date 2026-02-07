package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/service"
	"github.com/task-monitor/api-server/internal/utils"
)

// JWTAuth JWT认证中间件
func JWTAuth(authService service.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "未提供认证令牌")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "认证令牌格式错误")
			c.Abort()
			return
		}

		userID, username, err := authService.ParseToken(parts[1])
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("username", username)
		c.Next()
	}
}
