package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

type I18nRender struct {
	I18n I18n
	*gin.Context
}

func NewI18nRender(i18n I18n) *I18nRender {
	return &I18nRender{I18n: i18n}
}

func (r *I18nRender) OK(c *gin.Context, obj any) {
	r.JSON(c, 0, obj, "OK!", "OK!")
}

func (r *I18nRender) OkWithMsg(c *gin.Context, obj any, msg, other string, params ...any) {
	r.JSON(c, 0, obj, msg, other, params...)
}

func (r *I18nRender) FAIL(c *gin.Context, msg string, params ...any) {
	r.JSON(c, 100, struct{}{}, msg, "FAIL!!!", params...)
}

func (r *I18nRender) FailWithMsg(c *gin.Context, msg, other string, params ...any) {
	r.JSON(c, 100, struct{}{}, msg, other, params...)
}

func (r *I18nRender) FailWithError(c *gin.Context, err error) {
	r.JSON(c, 100, struct{}{}, "", err.Error())
}

func (r *I18nRender) JSON(c *gin.Context, code int, obj any, msg, other string, params ...any) {
	//if obj == nil {
	//	obj = struct{}{}
	//}
	msg = r.I18n.HttpValue(c.Request, msg, other, params...)
	c.Render(http.StatusOK, render.JSON{Data: gin.H{"status": code, "msg": msg, "data": obj}})
}
