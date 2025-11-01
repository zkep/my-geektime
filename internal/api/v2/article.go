package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/geek"
	"github.com/zkep/my-geektime/internal/types/sys_dict"
)

func (p *Product) Articles(c *gin.Context) {
	var req geek.ArticlesListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	req.Size = 500
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetArticles(c, accessToken, req)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret := geek.ArticlesListResponse{Rows: make([]geek.ArticlesListRow, 0)}
	ret.Count = resp.Data.Page.Count
	for _, v := range resp.Data.List {
		if v.ID <= 0 || v.ArticleTitle == "" {
			continue
		}
		row := geek.ArticlesListRow{
			ID:               v.ID,
			ArticleTitle:     v.ArticleTitle,
			ArticleSummary:   v.ArticleSummary,
			ArticleCover:     v.ArticleCover,
			VideoCover:       v.VideoCover,
			VideoSize:        v.VideoSize,
			AudioSize:        v.AudioSize,
			AudioDownloadURL: v.AudioDownloadURL,
			AuthorName:       v.AuthorName,
			AuthorIntro:      v.AuthorIntro,
		}
		if row.VideoCover != "" && row.ArticleCover == "" {
			row.ArticleCover = row.VideoCover
		}
		row.VideoCover = service.URLProxyReplace(row.VideoCover)
		row.ArticleCover = service.URLProxyReplace(row.ArticleCover)
		ret.Rows = append(ret.Rows, row)
	}
	global.OK(c, ret)
}

func (p *Product) ArticleInfo(c *gin.Context) {
	var req geek.ArticlesInfoRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetArticleInfo(c, accessToken, req)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	resp.Data.Info.Cover.Default = service.URLProxyReplace(resp.Data.Info.Cover.Default)
	resp.Data.Info.Author.Avatar = service.URLProxyReplace(resp.Data.Info.Author.Avatar)
	resp.Data.Info.Video.Cover = service.URLProxyReplace(resp.Data.Info.Video.Cover)
	if len(resp.Data.Info.Content) > 0 {
		if contextHTML, err1 := service.HtmlURLProxyReplace(resp.Data.Info.Content); err1 == nil {
			resp.Data.Info.Content = contextHTML
		}
	}
	ret := geek.ArticleDetail{
		ArticleInfo: resp.Data.Info,
	}
	ret.Redirect = sys_dict.ProductDetailURLWithType(resp.Data.Product.Type, resp.Data.Info.Pid, resp.Data.Info.ID)
	global.OK(c, ret)
}
