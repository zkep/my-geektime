package v2

import (
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gin-gonic/gin"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/geek"
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
	resp.Data.Info.Author.Avatar = service.URLProxyReplace(resp.Data.Info.Author.Avatar)
	resp.Data.Info.Video.Cover = service.URLProxyReplace(resp.Data.Info.Video.Cover)
	if len(resp.Data.Info.Content) > 0 {
		if contextHTML, err1 := service.HtmlURLProxyReplace(resp.Data.Info.Content); err1 == nil {
			resp.Data.Info.Content = contextHTML
		}
		if markdown, err1 := htmltomarkdown.ConvertString(resp.Data.Info.Content); err1 == nil {
			resp.Data.Info.Content = markdown
		}
	}
	global.OK(c, resp.Data.Info)
}
