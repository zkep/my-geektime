package base

import "encoding/json"

const (
	LoginWithName  = "name"
	LoginWithEmail = "email"
)

type LoginRequest struct {
	Type string          `json:"type,omitempty" binding:"required,min=3,max=10"`
	Data json.RawMessage `json:"data,omitempty" binding:"required,min=10"`
}

type LoginWithEmailRequest struct {
	Email    string `json:"email,omitempty"  binding:"required,email"`
	Password string `json:"password,omitempty" binding:"required,min=5,max=50"`
}

type LoginWithNameRequest struct {
	Account  string `json:"account,omitempty"  binding:"required,min=5,max=50"`
	Password string `json:"password,omitempty" binding:"required,min=5,max=50"`
}

type RedirectRequest struct {
	Token string `form:"token,omitempty" binding:"required,min=5"`
}

type SendEmailRequest struct {
	Email string `form:"email,omitempty"  binding:"required,email"`
}

type RefreshCookieRequest struct {
	Cookie string `json:"cookie,omitempty" binding:"required,min=100,max=5000"`
}
