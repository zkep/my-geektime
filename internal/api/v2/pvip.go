package v2

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/geek"
)

func (p *Product) PvipProductList(c *gin.Context) {
	var req geek.PvipProductRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if req.Tag > 0 {
		req.TagIds = []int32{req.Tag}
	}
	req.Size = req.PerPage
	req.Prev = req.Page
	identity := c.GetString(global.Identity)
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetPvipProduct(c, identity, accessToken, req)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret := geek.ProductListResponse{Rows: make([]geek.ProductListRow, 0)}
	ret.Count = resp.Data.Page.Total
	if resp.Data.Page.Total == 0 {
		ret.HasNext = resp.Data.Page.More
	}
	converter := md.NewConverter("", true, nil)
	for _, v := range resp.Data.Products {
		row := geek.ProductListRow{
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
			IsSale:        v.IsSale,
			Sale:          v.Price.Sale,
			SaleType:      v.Price.SaleType,
			Share:         v.Share,
			Author:        v.Author,
			Cover:         v.Cover,
			Article:       v.Article,
		}
		if len(row.IntroHTML) > 0 {
			if markdown, err := converter.ConvertString(row.IntroHTML); err == nil {
				row.IntroHTML = markdown
			}
		}
		ret.Rows = append(ret.Rows, row)
	}
	global.OK(c, ret)
}
