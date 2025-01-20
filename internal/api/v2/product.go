package v2

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/internal/types/task"
	"github.com/zkep/mygeektime/lib/utils"
	"gorm.io/gorm"
)

type Product struct{}

func NewProduct() *Product {
	return &Product{}
}

func (p *Product) Articles(c *gin.Context) {
	var req geek.ArticlesListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	req.Size = req.PerPage
	req.Prev = req.Page
	identity := c.GetString(global.Identity)
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetArticles(c, identity, accessToken, req)
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
	identity := c.GetString(global.Identity)
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetArticleInfo(c, identity, accessToken, req)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	converter := md.NewConverter("", true, nil)
	if len(resp.Data.Info.Content) > 0 {
		if markdown, err := converter.ConvertString(resp.Data.Info.Content); err == nil {
			resp.Data.Info.Content = markdown
		}
	}
	global.OK(c, resp.Data.Info)
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
		resp, err := service.GetArticles(c,
			identity, accessToken,
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
			ids = append(ids, v.ID)
			itemRaw, _ := json.Marshal(v)
			articlesMap[v.ID] = &model.Article{
				Aid:   fmt.Sprintf("%d", v.ID),
				Pid:   fmt.Sprintf("%d", req.Pid),
				Title: v.ArticleTitle,
				Cover: v.ArticleCover,
				Raw:   itemRaw,
			}
		}
	}
	var product model.Product
	if err := global.DB.Model(&model.Product{}).
		Where(&model.Product{Pid: fmt.Sprintf("%d", req.Pid)}).Find(&product).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
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
		Uid:        identity,
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
			info, err := service.GetArticleInfo(c, identity, accessToken, geek.ArticlesInfoRequest{Id: id})
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
			Uid:        identity,
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
	identity := c.GetString(global.Identity)
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetProduct(c, identity, accessToken, req)
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
	converter := md.NewConverter("", true, nil)
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
