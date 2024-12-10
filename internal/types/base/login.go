package base

type LoginRequest struct {
	UserName string `json:"username,omitempty" validate:"min=5,max=50"`
	Password string `json:"password,omitempty" validate:"min=5,max=50"`
}
