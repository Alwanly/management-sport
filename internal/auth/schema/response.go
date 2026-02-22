package schema

type ResponseLogin struct {
	Token  string `json:"token"`
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

type ResponseAdminLogin struct {
	Token  string `json:"token"`
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

type ResponseRegister struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
}
