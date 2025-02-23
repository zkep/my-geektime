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

type User struct {
	// uid
	Uid string `json:"uid,omitempty"`
	// user_name
	UserName string `json:"user_name,omitempty"`
	// nick_name
	NickName string `json:"nick_name,omitempty"`
	// avatar
	Avatar string `json:"avatar,omitempty"`
	// geek auth
	GeekAuth bool `json:"geek_auth,omitempty"`
	// role_id
	RoleId int32 `json:"role_id,omitempty"`
}
