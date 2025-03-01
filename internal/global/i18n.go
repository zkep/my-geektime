package global

import "github.com/zkep/my-geektime/lib/rest"

var (
	I18N rest.I18n

	Render *rest.I18nRender

	OK            = Render.OK
	OkWithMsg     = Render.OkWithMsg
	FAIL          = Render.FAIL
	FailWithMsg   = Render.FailWithMsg
	FailWithError = Render.FailWithError
	JSON          = Render.JSON
)
