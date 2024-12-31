package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/types/user"
)

type User struct{}

func NewUser() *User {
	return &User{}
}

func (u *User) List(c *gin.Context) {
	var req user.UserListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if req.PerPage <= 0 || (req.PerPage > 200) {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	roleId := c.GetFloat64(global.Role)
	if roleId != user.AdminRoleId {
		global.FAIL(c, "fail.msg", "no auth")
		return
	}
	ret := user.UserListResponse{
		Rows: make([]user.User, 0, 10),
	}
	var ls []*model.User
	tx := global.DB.Model(&model.User{})
	if req.Status > 0 {
		tx = tx.Where("status = ?", req.Status)
	}
	tx = tx.Where("role_id = ?", user.MemeberRoleId)
	tx = tx.Order("id DESC")
	if err := tx.Count(&ret.Count).
		Offset((req.Page - 1) * req.PerPage).
		Limit(req.PerPage).
		Find(&ls).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	for _, l := range ls {
		ret.Rows = append(ret.Rows, user.User{
			Uid:         l.Uid,
			UserName:    l.UserName,
			NickName:    l.NickName,
			Avatar:      l.Avatar,
			Status:      l.Status,
			AccessToken: l.AccessToken,
			RoleId:      l.RoleId,
			CreatedAt:   l.CreatedAt,
			UpdatedAt:   l.UpdatedAt,
		})
	}
	global.OK(c, ret)
}

func (u *User) Status(c *gin.Context) {
	var req user.UserStatusRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if err := global.DB.Model(&model.User{}).
		Where(&model.User{Uid: req.Uid}).
		Updates(&model.User{Status: req.Status}).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	global.OK(c, nil)
}
