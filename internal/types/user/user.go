package user

const (
	UserStatusActive   = 0x01
	UserStatusInactive = 0x02
)

const (
	AdminRoleId   = 0x01
	MemeberRoleId = 0x02
)

type UserListRequest struct {
	Uid     string `json:"uid"  form:"uid"`
	Status  int32  `json:"status" form:"status"`
	Page    int    `json:"page" form:"page"`
	PerPage int    `json:"perPage"  form:"perPage"`
}

type UserListResponse struct {
	Count int64  `json:"count"`
	Rows  []User `json:"rows"`
}

type User struct {
	// uid
	Uid string `json:"uid,omitempty"`
	// user_name
	UserName string `json:"user_name,omitempty"`
	// nike_name
	NikeName string `json:"nike_name,omitempty"`
	// avatar
	Avatar string `json:"avatar,omitempty"`
	// access_token
	AccessToken string `json:"access_token,omitempty"`
	// status
	Status int32 `json:"status,omitempty"`
	// phone
	Phone string `json:"phone,omitempty"`
	// role_id
	RoleId int32 `json:"role_id,omitempty"`
	// created_at
	CreatedAt int64 `json:"created_at,omitempty"`
	// updated_at
	UpdatedAt int64 `json:"updated_at,omitempty"`
}

type UserStatusRequest struct {
	// uid
	Uid string `json:"uid,omitempty"`
	// status
	Status int32 `json:"status,omitempty"`
}
