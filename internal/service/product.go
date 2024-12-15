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
	LearnProductURL    = "https://time.geekbang.org/serv/v3/learn/product"
	ArticlesURL        = "https://time.geekbang.com/serv/v1/column/articles"
	ArticleInfoURL     = "https://time.geekbang.org/serv/v3/article/info"
	PvipProductListURL = "https://time.geekbang.org/serv/v4/pvip/product_list"
)

func GetArticleInfo(ctx context.Context, req geek.ArticlesInfoRequest) (*geek.ArticleInfoResponse, error) {
	reqRaw, _ := json.Marshal(req)
	var resp geek.ArticleInfoResponse
	err := Request(ctx, http.MethodPost, ArticleInfoURL, bytes.NewBuffer(reqRaw), func(raw []byte) error {
		// auto sync to db
		if !global.CONF.Geektime.AutoSync {
			return nil
		}
		if err := json.Unmarshal(raw, &resp); err != nil {
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetArticleInfo", zap.String("raw", string(raw)))
			return nil
		}
		go func(info geek.ArticleInfoResponse) {
			aid := fmt.Sprintf("%d", info.Data.Info.Id)
			pid := fmt.Sprintf("%d", info.Data.Info.Pid)
			article := model.Article{
				Aid:   aid,
				Pid:   pid,
				Title: info.Data.Info.Title,
				Cover: info.Data.Info.Cover.Default,
				Raw:   raw,
			}
			if err := global.DB.
				Model(&model.Article{}).
				Where("aid=?", aid).
				Assign(article).
				FirstOrCreate(&article).Error; err != nil {
				global.LOG.Error("GetArticleInfo.AutoSync",
					zap.Error(err), zap.String("raw", string(raw)))
			}
		}(resp)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetArticles(ctx context.Context, req geek.ArticlesListRequest) (*geek.ArticlesResponse, error) {
	raw, _ := json.Marshal(req)
	var resp geek.ArticlesResponse
	err := Request(ctx, http.MethodPost, ArticlesURL, bytes.NewBuffer(raw), func(raw []byte) error {
		// auto sync to db
		if !global.CONF.Geektime.AutoSync {
			return nil
		}
		if err := json.Unmarshal(raw, &resp); err != nil {
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetArticles", zap.String("raw", string(raw)))
			return nil
		}
		go func() {
			for key, value := range resp.Data.List {
				itemRaw, _ := json.Marshal(value)
				article := model.ArticleSimple{
					Aid:   fmt.Sprintf("%d", value.ID),
					Pid:   req.Cid,
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
					Where("aid=?", article.Aid).
					Assign(article).
					FirstOrCreate(&article).Error; err != nil {
					global.LOG.Error("GetArticles.AutoSync",
						zap.Error(err), zap.String("raw", string(raw)))
				}
			}
		}()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetLearnProduct(ctx context.Context, req geek.ProductListRequest) (*geek.ProductResponse, error) {
	raw, _ := json.Marshal(req)
	var resp geek.ProductResponse
	err := Request(ctx, http.MethodPost, LearnProductURL, bytes.NewBuffer(raw), func(raw []byte) error {
		// auto sync to db
		if !global.CONF.Geektime.AutoSync {
			return nil
		}
		if err := json.Unmarshal(raw, &resp); err != nil {
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetLearnProduct", zap.String("raw", string(raw)))
			return nil
		}
		go func() {
			for _, value := range resp.Data.Products {
				itemRaw, _ := json.Marshal(value)
				product := model.Product{
					Pid:    fmt.Sprintf("%d", value.ID),
					Title:  value.Share.Title,
					Cover:  value.Share.Cover,
					Raw:    itemRaw,
					Source: value.Type,
				}
				if err := global.DB.
					Model(&model.Product{}).
					Where("pid=?", product.Pid).
					Assign(product).
					FirstOrCreate(&product).Error; err != nil {
					global.LOG.Error("GetLearnProduct.AutoSync",
						zap.Error(err), zap.String("raw", string(raw)))
				}
			}
		}()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetPvipProduct(ctx context.Context, req geek.PvipProductRequest) (*geek.ProductResponse, error) {
	raw, _ := json.Marshal(req)
	var resp geek.ProductResponse
	err := Request(ctx, http.MethodPost, PvipProductListURL, bytes.NewBuffer(raw), func(raw []byte) error {
		// auto sync to db
		if !global.CONF.Geektime.AutoSync {
			return nil
		}
		if err := json.Unmarshal(raw, &resp); err != nil {
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetPvipProduct", zap.String("raw", string(raw)))
			return nil
		}
		go func() {
			for _, value := range resp.Data.Products {
				itemRaw, _ := json.Marshal(value)
				product := model.Product{
					Pid:    fmt.Sprintf("%d", value.ID),
					Title:  value.Share.Title,
					Cover:  value.Share.Cover,
					Raw:    itemRaw,
					Source: value.Type,
				}
				if err := global.DB.
					Model(&model.Product{}).
					Where("pid=?", product.Pid).
					Assign(product).
					FirstOrCreate(&product).Error; err != nil {
					global.LOG.Error("GetPvipProduct.AutoSync",
						zap.Error(err), zap.String("raw", string(raw)))
				}
			}
		}()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
