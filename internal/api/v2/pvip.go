package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/geek"
)

func (p *Product) PvipProductList(c *gin.Context) {
	var req geek.PvipProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	if req.Tag > 0 {
		req.TagIds = []int{req.Tag}
	}
	req.Size = req.PerPage
	req.Prev = req.Page
	identity := c.GetString(global.Identity)
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetPvipProduct(c, identity, accessToken, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	ret := geek.ProductListResponse{Rows: make([]geek.ProductListRow, 0)}
	ret.Count = resp.Data.Page.Total
	if resp.Data.Page.Total == 0 {
		ret.HasNext = resp.Data.Page.More
	}
	for _, v := range resp.Data.Products {
		ret.Rows = append(ret.Rows, geek.ProductListRow{
			ID:            v.ID,
			Title:         v.Title,
			Subtitle:      v.Subtitle,
			Intro:         v.Intro,
			IntroHTML:     v.IntroHTML,
			Ucode:         v.Ucode,
			IsFinish:      v.IsFinish,
			IsVideo:       v.IsVideo,
			IsAudio:       v.IsAudio,
			IsDailylesson: v.IsDailylesson,
			IsUniversity:  v.IsUniversity,
			IsOpencourse:  v.IsOpencourse,
			IsQconp:       v.IsQconp,
			Share:         v.Share,
			Author:        v.Author,
			Cover:         v.Cover,
			Article:       v.Article,
		})
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK", "data": ret})
}
