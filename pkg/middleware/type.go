package middleware

// AuthUserData represents the authenticated user data extracted from JWT
type AuthUserData struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// Constants for context keys
const (
	LocalTokenKey = "user"
)
