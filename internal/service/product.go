package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/types/geek"
	"go.uber.org/zap"
)

const (
	ArticlesURL        = "https://time.geekbang.com/serv/v1/column/articles"
	ArticleInfoURL     = "https://time.geekbang.org/serv/v3/article/info"
	ProductListURL     = "https://time.geekbang.org/serv/v3/product/list"
	PvipProductListURL = "https://time.geekbang.org/serv/v4/pvip/product_list"
)

func GetArticleInfo(ctx context.Context, uid, accessToken string,
	req geek.ArticlesInfoRequest) (*geek.ArticleInfoResponse, error) {
	reqRaw, _ := json.Marshal(req)
	var resp geek.ArticleInfoResponse
	after := func(raw []byte) error {
		if err := json.Unmarshal(raw, &resp); err != nil {
			global.LOG.Error("GetArticleInfo", zap.Error(err))
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetArticleInfo", zap.Any("error", resp.Error))
			return nil
		}
		resp.Raw = raw
		go func(info geek.ArticleInfoResponse) {
			aid := fmt.Sprintf("%d", info.Data.Info.ID)
			pid := fmt.Sprintf("%d", info.Data.Info.Pid)
			if info.Data.Info.Cover.Square != "" {
				info.Data.Info.Cover.Default = info.Data.Info.Cover.Square
			}
			article := model.Article{
				Aid:   aid,
				Pid:   pid,
				Uid:   uid,
				Title: info.Data.Info.Title,
				Cover: info.Data.Info.Cover.Default,
				Raw:   raw,
			}
			if err := global.DB.
				Model(&model.Article{}).
				Where(&model.Article{Aid: aid}).
				Assign(&article).
				FirstOrCreate(&article).Error; err != nil {
				global.LOG.Error("GetArticleInfo.AutoSync", zap.Error(err))
			}
		}(resp)
		return nil
	}
	err := Request(ctx, http.MethodPost, ArticleInfoURL, bytes.NewBuffer(reqRaw), accessToken, after)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetArticles(ctx context.Context, uid, accessToken string,
	req geek.ArticlesListRequest) (*geek.ArticlesResponse, error) {
	raw, _ := json.Marshal(req)
	var resp geek.ArticlesResponse
	after := func(raw []byte) error {
		if err := json.Unmarshal(raw, &resp); err != nil {
			global.LOG.Error("GetArticles", zap.Error(err), zap.String("raw", string(raw)))
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetArticles", zap.Any("error", resp.Error))
			return nil
		}
		go func() {
			for key, value := range resp.Data.List {
				itemRaw, _ := json.Marshal(value)
				article := model.ArticleSimple{
					Aid:   fmt.Sprintf("%d", value.ID),
					Pid:   req.Cid,
					Uid:   uid,
					Title: value.ArticleTitle,
					Cover: value.ArticleCover,
					Sort:  int32(key),
					Raw:   itemRaw,
				}
				if value.VideoCover != "" {
					article.Cover = value.VideoCover
				}
				if err := global.DB.
					Model(&model.ArticleSimple{}).
					Where(&model.ArticleSimple{
						Aid: article.Aid,
					}).
					Assign(&article).
					FirstOrCreate(&article).Error; err != nil {
					global.LOG.Error("GetArticles.AutoSync", zap.Error(err))
				}
			}
		}()
		return nil
	}
	err := Request(ctx, http.MethodPost, ArticlesURL, bytes.NewBuffer(raw), accessToken, after)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetPvipProduct(ctx context.Context, uid, accessToken string,
	req geek.PvipProductRequest) (*geek.ProductResponse, error) {
	raw, _ := json.Marshal(req)
	var resp geek.ProductResponse
	after := func(raw []byte) error {
		if err := json.Unmarshal(raw, &resp); err != nil {
			global.LOG.Error("GetPvipProduct", zap.Error(err))
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetPvipProduct", zap.Any("error", resp.Error))
			return nil
		}
		go func() {
			for _, value := range resp.Data.Products {
				itemRaw, _ := json.Marshal(value)
				product := model.Product{
					Pid:        fmt.Sprintf("%d", value.ID),
					Uid:        uid,
					Title:      value.Share.Title,
					Cover:      value.Share.Cover,
					Raw:        itemRaw,
					Source:     value.Type,
					OtherType:  req.ProductType,
					OtherForm:  req.ProductForm,
					OtherGroup: req.Direction,
					OtherTag:   req.Tag,
				}
				if err := global.DB.
					Model(&model.Product{}).
					Where(&model.Product{Pid: product.Pid}).
					Assign(product).
					FirstOrCreate(&product).Error; err != nil {
					global.LOG.Error("GetPvipProduct.AutoSync", zap.Error(err))
				}
			}
		}()
		return nil
	}
	err := Request(ctx, http.MethodPost, PvipProductListURL, bytes.NewBuffer(raw), accessToken, after)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetProduct(ctx context.Context, uid, accessToken string,
	req geek.DailyProductRequest) (*geek.DailyProductResponse, error) {
	raw, _ := json.Marshal(req)
	var resp geek.DailyProductResponse
	after := func(raw []byte) error {
		if err := json.Unmarshal(raw, &resp); err != nil {
			global.LOG.Error("GetProduct", zap.Error(err))
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetProduct", zap.Any("error", resp.Error))
			return nil
		}
		go func() {
			for _, value := range resp.Data.List {
				itemRaw, _ := json.Marshal(value)
				product := model.Product{
					Pid:        fmt.Sprintf("%d", value.ID),
					Uid:        uid,
					Title:      value.Share.Title,
					Cover:      value.Share.Cover,
					Raw:        itemRaw,
					Source:     value.Type,
					OtherForm:  2,
					OtherGroup: req.Direction,
					OtherTag:   req.LabelID,
				}
				switch req.Type {
				case "d":
					product.OtherType = 6
				case "q":
					product.OtherType = 7
				}
				if err := global.DB.
					Model(&model.Product{}).
					Where(&model.Product{Pid: product.Pid}).
					Assign(product).
					FirstOrCreate(&product).Error; err != nil {
					global.LOG.Error("GetProduct.AutoSync", zap.Error(err))
				}
			}
		}()
		return nil
	}
	err := Request(ctx, http.MethodPost, ProductListURL, bytes.NewBuffer(raw), accessToken, after)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
