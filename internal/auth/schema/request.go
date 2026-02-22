package schema

type RequestLogin struct {
	Username string `json:"username" validate:"required,min=3,max=255"`
	Password string `json:"password" validate:"required,min=6"`
}

type RequestAdminLogin struct {
	Username string `json:"username" validate:"required,min=3,max=255"`
	Password string `json:"password" validate:"required,min=6"`
}

type RequestRegister struct {
	Username string `json:"username" validate:"required,min=3,max=255"`
	Password string `json:"password" validate:"required,min=6"`
}
