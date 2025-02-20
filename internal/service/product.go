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
	ArticlesURL                 = "https://time.geekbang.com/serv/v1/column/articles"
	ArticleInfoURL              = "https://time.geekbang.org/serv/v3/article/info"
	ProductListURL              = "https://time.geekbang.org/serv/v3/product/list"
	PvipProductListURL          = "https://time.geekbang.org/serv/v4/pvip/product_list"
	ArticleCommentURL           = "https://time.geekbang.org/serv/v4/comment/list"
	ArticleCommentDiscussionURL = "https://time.geekbang.org/serv/discussion/v1/root_list"
	SearchURL                   = "https://time.geekbang.org/serv/v3/search"
	ColumnInfoURL               = "https://time.geekbang.org/serv/v3/column/info"
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
		go func(ret geek.ArticleInfoResponse) {
			aid := fmt.Sprintf("%d", ret.Data.Info.ID)
			pid := fmt.Sprintf("%d", ret.Data.Info.Pid)
			if ret.Data.Info.Cover.Square != "" {
				ret.Data.Info.Cover.Default = ret.Data.Info.Cover.Square
			}
			info := model.Article{
				Aid:   aid,
				Pid:   pid,
				Uid:   uid,
				Title: ret.Data.Info.Title,
				Cover: ret.Data.Info.Cover.Default,
				Raw:   raw,
			}
			if err := global.DB.
				Model(&model.Article{}).
				Where(&model.Article{Aid: aid}).
				Assign(&info).
				FirstOrCreate(&info).Error; err != nil {
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
				info := model.ArticleSimple{
					Aid:   fmt.Sprintf("%d", value.ID),
					Pid:   req.Cid,
					Uid:   uid,
					Title: value.ArticleTitle,
					Cover: value.ArticleCover,
					Sort:  int32(key),
					Raw:   itemRaw,
				}
				if value.VideoCover != "" {
					info.Cover = value.VideoCover
				}
				if err := global.DB.
					Model(&model.ArticleSimple{}).
					Where(&model.ArticleSimple{Aid: info.Aid}).
					Assign(&info).
					FirstOrCreate(&info).Error; err != nil {
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
				info := model.Product{
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
					Where(&model.Product{Pid: info.Pid}).
					Assign(&info).
					FirstOrCreate(&info).Error; err != nil {
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
				info := model.Product{
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
					info.OtherType = 6
				case "q":
					info.OtherType = 7
				}
				if err := global.DB.
					Model(&model.Product{}).
					Where(&model.Product{Pid: info.Pid}).
					Assign(info).
					FirstOrCreate(&info).Error; err != nil {
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

func GetArticleComment(ctx context.Context, _, accessToken string,
	req geek.ArticleCommentListRequest) (*geek.ArticleCommentList, error) {
	raw, _ := json.Marshal(req)
	var resp geek.ArticleCommentList
	after := func(raw []byte) error {
		if err := json.Unmarshal(raw, &resp); err != nil {
			global.LOG.Error("GetArticleComment", zap.Error(err), zap.String("raw", string(raw)))
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetArticleComment", zap.Any("error", resp.Error))
			return nil
		}
		go func() {
			for _, value := range resp.Data.List {
				itemRaw, _ := json.Marshal(value)
				info := model.ArticleComment{
					Aid:             req.Aid,
					Cid:             value.ID,
					DiscussionCount: value.DiscussionCount,
					LikeCount:       value.LikeCount,
					CommentCtime:    value.CommentCtime,
					Raw:             itemRaw,
				}
				if err := global.DB.
					Model(&model.ArticleComment{}).
					Where(&model.ArticleComment{Aid: info.Aid, Cid: info.Cid}).
					Assign(&info).
					FirstOrCreate(&info).Error; err != nil {
					global.LOG.Error("GetArticleComment.AutoSync", zap.Error(err))
				}
			}
		}()
		return nil
	}
	err := Request(ctx, http.MethodPost, ArticleCommentURL, bytes.NewBuffer(raw), accessToken, after)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetArticleCommentDiscussion(ctx context.Context, _, accessToken string,
	req geek.DiscussionListRequest) (*geek.DiscussionOriginListResponse, error) {
	raw, _ := json.Marshal(req)
	var resp geek.DiscussionOriginListResponse
	after := func(raw []byte) error {
		if err := json.Unmarshal(raw, &resp); err != nil {
			global.LOG.Error("GetArticleCommentDiscussion", zap.Error(err), zap.String("raw", string(raw)))
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetArticleCommentDiscussion", zap.Any("error", resp.Error))
			return nil
		}
		go func() {
			for _, value := range resp.Data.List {
				itemRaw, _ := json.Marshal(value)
				info := model.ArticleCommentDiscussion{
					Cid:         req.TargetID,
					Did:         value.Discussion.ID,
					LikesNumber: value.Discussion.LikesNumber,
					Ctime:       value.Discussion.Ctime,
					Raw:         itemRaw,
				}
				if err := global.DB.
					Model(&model.ArticleCommentDiscussion{}).
					Where(&model.ArticleCommentDiscussion{Cid: info.Cid, Did: info.Did}).
					Assign(&info).
					FirstOrCreate(&info).Error; err != nil {
					global.LOG.Error("GetArticleCommentDiscussion.AutoSync", zap.Error(err))
				}
			}
		}()
		return nil
	}
	err := Request(ctx, http.MethodPost, ArticleCommentDiscussionURL, bytes.NewBuffer(raw), accessToken, after)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func ArticleAllComment(ctx context.Context, _, accessToken string, id int64) error {
	req := geek.ArticleCommentListRequest{Aid: id}
	hasMore := true
	after := func(raw []byte) error {
		var resp geek.ArticleCommentList
		if err := json.Unmarshal(raw, &resp); err != nil {
			global.LOG.Error("ArticleAllComment", zap.Error(err), zap.String("raw", string(raw)))
			return err
		}
		hasMore = resp.Data.Page.More
		if resp.Code != 0 {
			global.LOG.Warn("ArticleAllComment", zap.Any("error", resp.Error))
			return nil
		}
		for _, value := range resp.Data.List {
			itemRaw, _ := json.Marshal(value)
			info := model.ArticleComment{
				Aid:             req.Aid,
				Cid:             value.ID,
				DiscussionCount: value.DiscussionCount,
				LikeCount:       value.LikeCount,
				CommentCtime:    value.CommentCtime,
				Raw:             itemRaw,
			}
			if err := global.DB.
				Model(&model.ArticleComment{}).
				Where(&model.ArticleComment{Aid: info.Aid, Cid: info.Cid}).
				Assign(&info).
				FirstOrCreate(&info).Error; err != nil {
				global.LOG.Error("ArticleAllComment.AutoSync", zap.Error(err))
			}
			if info.DiscussionCount > 0 {
				discussionReq := geek.DiscussionListRequest{
					UseLikesOrder: true,
					TargetID:      info.Cid,
					TargetType:    1,
					PageType:      1,
					Size:          50,
				}
				hasNext := true
				discussionAfter := func(raw []byte) error {
					var discussionResp geek.DiscussionOriginListResponse
					if err := json.Unmarshal(raw, &discussionResp); err != nil {
						global.LOG.Error("ArticleAllComment", zap.Error(err), zap.String("raw", string(raw)))
						return err
					}
					hasNext = discussionResp.Data.Page.More
					if discussionResp.Code != 0 {
						global.LOG.Warn("ArticleAllComment", zap.Any("error", discussionResp.Error))
						return nil
					}
					for _, x := range discussionResp.Data.List {
						valueRaw, _ := json.Marshal(x)
						dinfo := model.ArticleCommentDiscussion{
							Cid:         discussionReq.TargetID,
							Did:         x.Discussion.ID,
							LikesNumber: x.Discussion.LikesNumber,
							Ctime:       x.Discussion.Ctime,
							Raw:         valueRaw,
						}
						if err := global.DB.
							Model(&model.ArticleCommentDiscussion{}).
							Where(&model.ArticleCommentDiscussion{Cid: dinfo.Cid, Did: dinfo.Did}).
							Assign(&dinfo).
							FirstOrCreate(&dinfo).Error; err != nil {
							global.LOG.Error("ArticleAllComment.AutoSync", zap.Error(err))
						}
					}
					return nil
				}
				for hasNext {
					discussionReq.Prev++
					discussionRaw, _ := json.Marshal(discussionReq)
					err := Request(ctx, http.MethodPost,
						ArticleCommentDiscussionURL, bytes.NewBuffer(discussionRaw), accessToken, discussionAfter)
					if err != nil {
						return err
					}
				}

			}
		}
		return nil
	}
	for hasMore {
		req.Prev++
		raw, _ := json.Marshal(req)
		err := Request(ctx, http.MethodPost, ArticleCommentURL, bytes.NewBuffer(raw), accessToken, after)
		if err != nil {
			return err
		}
	}
	return nil
}

func GeekTimeSearch(ctx context.Context, accessToken string, req geek.SearchRequest) (*geek.SearchResponse, error) {
	raw, _ := json.Marshal(req)
	var resp geek.SearchResponse
	after := func(raw []byte) error {
		if err := json.Unmarshal(raw, &resp); err != nil {
			global.LOG.Error("GeekTimeSearch", zap.Error(err), zap.String("raw", string(raw)))
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GeekTimeSearch", zap.Any("error", resp.Error))
			return nil
		}
		return nil
	}
	err := Request(ctx, http.MethodPost, SearchURL, bytes.NewBuffer(raw), accessToken, after)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetColumnInfo(ctx context.Context, uid, accessToken string,
	req geek.ColumnRequest) (*geek.ColumnResponse, error) {
	reqRaw, _ := json.Marshal(req)
	var resp geek.ColumnResponse
	after := func(raw []byte) error {
		if err := json.Unmarshal(raw, &resp); err != nil {
			global.LOG.Error("GetArticleInfo", zap.Error(err))
			return err
		}
		if resp.Code != 0 {
			global.LOG.Warn("GetArticleInfo", zap.Any("error", resp.Error))
			return nil
		}
		value := resp.Data
		itemRaw, _ := json.Marshal(value)
		info := model.Product{
			Pid:    fmt.Sprintf("%d", value.ID),
			Uid:    uid,
			Title:  value.Share.Title,
			Cover:  value.Share.Cover,
			Raw:    itemRaw,
			Source: value.Type,
		}
		if err := global.DB.
			Model(&model.Product{}).
			Where(&model.Product{Pid: info.Pid}).
			Assign(&info).
			FirstOrCreate(&info).Error; err != nil {
			global.LOG.Error("GetColumnInfo.AutoSync", zap.Error(err))
		}
		return nil
	}
	err := Request(ctx, http.MethodPost, ColumnInfoURL, bytes.NewBuffer(reqRaw), accessToken, after)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
