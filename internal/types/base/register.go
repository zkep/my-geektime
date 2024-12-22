package base

import "encoding/json"

type RegisterRequest struct {
	Type string          `json:"type,omitempty" binding:"required,min=3,max=10"`
	Data json.RawMessage `json:"data,omitempty" binding:"required,min=10"`
}

type RegisterWithEmailRequest struct {
	Email    string `json:"email,omitempty"  binding:"required,email"`
	Code     string `json:"code,omitempty" binding:"required,min=6,max=9"`
	Password string `json:"password,omitempty" binding:"required,min=5,max=50"`
}

type RegisterWithNameRequest struct {
	Account  string `json:"account,omitempty"  binding:"required,min=4,max=20"`
	Password string `json:"password,omitempty" binding:"required,min=6,max=50"`
}
