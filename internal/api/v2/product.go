package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/geek"
	"github.com/zkep/my-geektime/internal/types/sys_dict"
	"github.com/zkep/my-geektime/internal/types/task"
	"github.com/zkep/my-geektime/libs/utils"
	"gorm.io/gorm"
)

type Product struct{}

func NewProduct() *Product {
	return &Product{}
}

func (p *Product) Download(c *gin.Context) {
	var req geek.DowloadRequest
	if err := c.BindJSON(&req); err != nil {
		global.FAIL(c, "fail.msg", err)
		return
	}
	identity := c.GetString(global.Identity)
	accessToken := c.GetString(global.AccessToken)
	if accessToken == "" {
		global.FAIL(c, "product.no_cookie")
		return
	}
	// check geektime cookie
	var auth geek.AuthResponse
	if err := service.Authority(accessToken, service.SaveCookie(accessToken, identity, &auth)); err != nil {
		if errors.Is(err, service.ErrorGeekAccountNotLogin) {
			global.JSON(c, 10002, nil, "product.no_cookie", "")
		} else {
			global.FAIL(c, "fail.msg", err.Error())
		}
		return
	}
	articlesMap := make(map[int64]*model.Article, 10)
	ids := make([]int64, 0, 1)
	switch x := req.Ids.(type) {
	case string:
		for _, v := range strings.Split(x, ",") {
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				global.FAIL(c, "fail.msg", err)
				return
			}
			ids = append(ids, i)
		}
	case float64:
		ids = append(ids, int64(x))
	default:
		if req.Pid <= 0 {
			global.FAIL(c, "product.no_exists_pid")
			return
		}
		resp, err := service.GetArticles(c, accessToken,
			geek.ArticlesListRequest{
				Cid:   fmt.Sprintf("%d", req.Pid),
				Order: "earliest",
				Prev:  1,
				Size:  500,
			})
		if err != nil {
			global.FailWithError(c, err)
			return
		}
		if len(resp.Data.List) == 0 {
			global.FAIL(c, "product.api_busy")
			return
		}
		for _, v := range resp.Data.List {
			if v.ID <= 0 || v.ArticleTitle == "" {
				continue
			}
			ids = append(ids, v.ID)
			itemRaw, _ := json.Marshal(v)
			info := &model.Article{
				Aid:   fmt.Sprintf("%d", v.ID),
				Pid:   fmt.Sprintf("%d", req.Pid),
				Title: v.ArticleTitle,
				Cover: v.ArticleCover,
				Raw:   itemRaw,
			}
			if v.VideoCover != "" && info.Cover == "" {
				info.Cover = v.VideoCover
			}
			articlesMap[v.ID] = info
		}
	}
	var product model.Product
	if err := global.DB.Model(&model.Product{}).
		Where(&model.Product{Pid: fmt.Sprintf("%d", req.Pid)}).Find(&product).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if product.Pid == "" {
		ret, err := service.GetColumnInfo(c, accessToken,
			geek.ColumnRequest{ProductID: req.Pid, WithRecommendArticle: true})
		if err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		product.Title = ret.Data.Title
		product.Cover = ret.Data.Cover.Square
		product.Raw, _ = json.Marshal(ret.Data)
	}
	if len(articlesMap) == 0 {
		var articles []*model.Article
		if err := global.DB.Model(&model.Article{}).
			Where("aid IN ?", ids).Find(&articles).Error; err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		for _, v := range articles {
			articlesMap[v.Id] = v
		}
	}
	jobId := utils.HalfUUID()
	job := &model.Task{
		TaskId:     jobId,
		TaskName:   product.Title,
		TaskType:   service.TASK_TYPE_PRODUCT,
		OtherId:    fmt.Sprintf("%d", req.Pid),
		Cover:      product.Cover,
		Raw:        product.Raw,
		OtherType:  product.OtherType,
		OtherForm:  product.OtherForm,
		OtherGroup: product.OtherGroup,
		OtherTag:   product.OtherTag,
	}
	tasks := make([]*model.Task, 0, len(ids))
	for _, id := range ids {
		var (
			raw      []byte
			otherId  string
			taskName string
			cover    string
		)
		if article, ok := articlesMap[id]; !ok {
			info, err := service.GetArticleInfo(c, accessToken, geek.ArticlesInfoRequest{Id: id})
			if err != nil {
				global.FAIL(c, "fail.msg", err.Error())
				return
			}
			var m geek.ArticleInfoRaw
			if err = json.Unmarshal(info.Raw, &m); err != nil {
				global.FAIL(c, "fail.msg", err.Error())
				return
			}
			raw = m.Data
			otherId = fmt.Sprintf("%d", info.Data.Info.ID)
			taskName = info.Data.Info.Title
			cover = info.Data.Info.Cover.Default
		} else {
			raw = article.Raw
			otherId = article.Aid
			taskName = article.Title
			cover = article.Cover
		}
		item := model.Task{
			TaskPid:    jobId,
			TaskId:     utils.HalfUUID(),
			OtherId:    otherId,
			TaskName:   taskName,
			TaskType:   service.TASK_TYPE_ARTICLE,
			Cover:      cover,
			Raw:        raw,
			OtherType:  product.OtherType,
			OtherForm:  product.OtherForm,
			OtherGroup: product.OtherGroup,
			OtherTag:   product.OtherTag,
		}
		if global.CONF.Site.Download {
			item.Bstatus = service.TASK_STATUS_PENDING
		}
		tasks = append(tasks, &item)
	}
	count := len(tasks)
	statistics := task.TaskStatistics{
		Count: count,
		Items: map[int]int{
			service.TASK_STATUS_PENDING:  count,
			service.TASK_STATUS_RUNNING:  0,
			service.TASK_STATUS_FINISHED: 0,
			service.TASK_STATUS_ERROR:    0,
		},
	}
	job.Statistics, _ = json.Marshal(statistics)
	if global.CONF.Site.Download {
		job.Bstatus = service.TASK_STATUS_PENDING
	}
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(job).Error; err != nil {
			return err
		}
		for _, x := range tasks {
			if err := tx.Create(x).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	resp := geek.DowloadResponse{JobID: jobId}
	global.OK(c, resp)
}

func (p *Product) ProductList(c *gin.Context) {
	var req geek.DailyProductRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	req.Size = req.PerPage
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetProduct(c, accessToken, req)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret := geek.ProductListResponse{Rows: make([]geek.ProductListRow, 0)}
	ret.Score = resp.Data.Page.Score
	ret.Count = resp.Data.Page.Count
	if resp.Data.Page.Count == 0 {
		ret.HasNext = resp.Data.Page.More
	}
	for _, v := range resp.Data.List {
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
			IsColumn:      v.IsColumn,
			IsCore:        v.IsCore,
			IsDailylesson: v.IsDailylesson,
			IsUniversity:  v.IsUniversity,
			IsOpencourse:  v.IsOpencourse,
			IsQconp:       v.IsQconp,
			IsMentor:      v.IsMentor,
			IsSale:        v.IsSale,
			Sale:          v.Price.Sale,
			SaleType:      v.Price.SaleType,
			Share:         v.Share,
			Author:        v.Author,
			Cover:         v.Cover,
			Article:       v.Article,
		}
		row.Cover.Square = service.URLProxyReplace(row.Cover.Square)
		row.Author.Avatar = service.URLProxyReplace(row.Author.Avatar)
		if len(row.IntroHTML) > 0 {
			if introHTML, err1 := service.HtmlURLProxyReplace(row.IntroHTML); err1 == nil {
				row.IntroHTML = introHTML
			}
		}
		row.Redirect = sys_dict.ProductURLWithType(v.Type, v.ID)
		ret.Rows = append(ret.Rows, row)
	}
	global.OK(c, ret)
}

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
	accessToken := c.GetString(global.AccessToken)
	ret := geek.ProductListResponse{Rows: make([]geek.ProductListRow, 0)}
	if len(req.Keyword) > 0 {
		searchReq := geek.SearchRequest{
			Keyword:  req.Keyword,
			Category: "product",
			Platform: "pc",
			Prev:     req.Prev,
			Size:     req.Size + 1,
		}
		searchRet, err := service.GeekTimeSearch(c, accessToken, searchReq)
		if err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		if len(searchRet.Data.List) > req.Size {
			ret.HasNext = true
			searchRet.Data.List = searchRet.Data.List[:req.Size]
		}
		for _, v := range searchRet.Data.List {
			if v.ItemType != "product" {
				continue
			}
			item := v.Product
			row := geek.ProductListRow{
				ID:       item.ID,
				Title:    item.Title,
				Subtitle: item.Subtitle,
				IsVideo:  item.Type == "c3",
				IsAudio:  item.Type == "c1",
			}
			row.Article.Count = item.TotalLesson
			row.Author.Name = item.AuthorName
			row.Author.Info = item.AuthorIntro
			row.Cover.Square = item.Cover
			row.Cover.Square = service.URLProxyReplace(row.Cover.Square)
			if len(row.IntroHTML) > 0 {
				if introHTML, err1 := service.HtmlURLProxyReplace(row.IntroHTML); err1 == nil {
					row.IntroHTML = introHTML
				}
			}
			ret.Rows = append(ret.Rows, row)
		}

		global.OK(c, ret)
		return
	}
	resp, err := service.GetPvipProduct(c, accessToken, req)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret.Count = resp.Data.Page.Total
	if resp.Data.Page.Total == 0 {
		ret.HasNext = resp.Data.Page.More
	}
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
			IsColumn:      v.IsColumn,
			IsCore:        v.IsCore,
			IsDailylesson: v.IsDailylesson,
			IsUniversity:  v.IsUniversity,
			IsOpencourse:  v.IsOpencourse,
			IsQconp:       v.IsQconp,
			IsMentor:      v.IsMentor,
			IsSale:        v.IsSale,
			Sale:          v.Price.Sale,
			SaleType:      v.Price.SaleType,
			Share:         v.Share,
			Author:        v.Author,
			Cover:         v.Cover,
			Article:       v.Article,
		}
		row.Cover.Square = service.URLProxyReplace(row.Cover.Square)
		row.Author.Avatar = service.URLProxyReplace(row.Author.Avatar)
		if len(row.IntroHTML) > 0 {
			if introHTML, err1 := service.HtmlURLProxyReplace(row.IntroHTML); err1 == nil {
				row.IntroHTML = introHTML
			}
		}
		row.Redirect = sys_dict.ProductURLWithType(v.Type, v.ID)
		ret.Rows = append(ret.Rows, row)
	}
	global.OK(c, ret)
}
