package base

type RegisterRequest struct {
	Email    string `json:"email,omitempty" binding:"required,email"`
	Code     string `json:"code,omitempty" binding:"required,min=5,max=50"`
	Password string `json:"password,omitempty" binding:"required,min=5,max=50"`
	Type     string `json:"type,omitempty" binding:"required,min=2,max=50"`
}
