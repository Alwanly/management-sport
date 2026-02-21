package middleware

import (
	"net/http"
	"strings"

	"github.com/Alwanly/management-sport/pkg/authentication"
	"github.com/gin-gonic/gin"
)

type IAuthMiddleware interface {
	JwtAuth() gin.HandlerFunc
	BasicAuth() gin.HandlerFunc
}

type AuthMiddleware struct {
	Jwt   authentication.IJwtService
	Basic authentication.IBasicAuthService
}

// AuthUserData defined in pkg/middleware/type.go

func NewAuthMiddleware(jwt authentication.IJwtService, basic authentication.IBasicAuthService) *AuthMiddleware {
	return &AuthMiddleware{
		Jwt:   jwt,
		Basic: basic,
	}
}

func (a *AuthMiddleware) JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if !strings.Contains(token, "Bearer") {
			responseUnauthorized(c, "Bearer", "Invalid token")
			return
		}

		token = strings.Replace(token, "Bearer ", "", 1)
		if token == "" {
			responseUnauthorized(c, "Bearer", "Invalid token")
			return
		}

		auth, err := a.Jwt.ParseToken(token)
		if err != nil {
			responseUnauthorized(c, "Bearer", "Invalid token")
			return
		}

		c.Set(LocalTokenKey, decodeAuthToken(*auth))
		c.Next()
	}
}

func (a *AuthMiddleware) AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First run JWT auth
		a.JwtAuth()(c)

		// If JWT auth failed, it would have aborted already
		if c.IsAborted() {
			return
		}

		// Get user data from context
		authUserValue, exists := c.Get(LocalTokenKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Access denied: admin role required"})
			return
		}

		authUser, ok := authUserValue.(*AuthUserData)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Access denied: invalid user data"})
			return
		}

		// Check if user has admin role
		if authUser.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Access denied: admin role required"})
			return
		}

		c.Next()
	}
}

func (a *AuthMiddleware) BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.Contains(auth, "Basic") {
			responseUnauthorized(c, "Basic", "Invalid auth")
			return
		}

		username, password := a.Basic.DecodeFromHeader(auth)
		if !a.Basic.Validate(username, password) {
			responseUnauthorized(c, "Basic", "Invalid auth")
			return
		}
		c.Next()
	}
}

func responseUnauthorized(c *gin.Context, _ string, message ...string) {
	c.Header("WWW-Authenticate", "Basic realm=Restricted")
	response := gin.H{"message": message[0]}
	if len(message) > 1 {
		response["statusCode"] = message[1]
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, response)
}

func decodeAuthToken(auth authentication.JWTClaims) *AuthUserData {
	userData := &AuthUserData{
		UserID: auth["userId"].(string),
	}
	if role, ok := auth["role"].(string); ok {
		userData.Role = role
	}
	return userData
}
