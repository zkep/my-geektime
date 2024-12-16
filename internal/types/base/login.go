package base

type LoginRequest struct {
	Account  string `json:"account,omitempty" validate:"min=5,max=50"`
	Password string `json:"password,omitempty" validate:"min=5,max=50"`
	Type     string `json:"type,omitempty" validate:"min=5,max=50"`
}

type RedirectRequest struct {
	Token string `form:"token,omitempty" validate:"min=5"`
}
