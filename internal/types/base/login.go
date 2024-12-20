package base

type LoginRequest struct {
	Email    string `form:"email,omitempty"  binding:"email"`
	Password string `json:"password,omitempty" binding:"required,min=5,max=50"`
	Type     string `json:"type,omitempty" binding:"required,min=3,max=10"`
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
