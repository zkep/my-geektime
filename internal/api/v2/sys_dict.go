package v2

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/sys_dict"
)

type Dict struct {
	dict *service.Dict
}

func NewDict() *Dict {
	return &Dict{
		dict: &service.Dict{},
	}
}

func (s *Dict) Create(c *gin.Context) {
	var r sys_dict.Request
	if err := c.ShouldBind(&r); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	base := model.SysDictBase{
		Key:     r.Key,
		Pkey:    r.Pkey,
		Name:    r.Name,
		Summary: r.Summary,
		Content: r.Content,
		Sort:    r.Sort,
	}
	info := model.SysDict{Base: &base}
	if err := global.DB.WithContext(c).
		Model(&model.SysDict{}).
		Where(&model.SysDict{
			Base: &model.SysDictBase{
				Pkey: r.Pkey,
				Key:  r.Key,
			},
		}).
		FirstOrCreate(&info).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret := sys_dict.Response{
		Id:      info.Model.Id,
		Request: r,
		Created: info.Model.Created,
		Updated: info.Model.Updated,
	}
	global.OK(c, ret)
}

func (s *Dict) Update(c *gin.Context) {
	var r sys_dict.UpdateDictRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	base := model.SysDictBase{
		Key:     r.Key,
		Pkey:    r.Pkey,
		Name:    r.Name,
		Summary: r.Summary,
		Content: r.Content,
		Sort:    r.Sort,
	}
	info := model.SysDict{Base: &base, Model: &model.Model{Id: r.Id}}
	if err := global.DB.
		Model(&info).
		Updates(&info).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret := sys_dict.Response{
		Id:      info.Model.Id,
		Request: r.Request,
		Created: info.Model.Created,
		Updated: info.Model.Updated,
	}
	global.OK(c, ret)
}

func (s *Dict) Delete(c *gin.Context) {
	var r sys_dict.Query
	if err := c.ShouldBindJSON(&r); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	info := model.SysDict{Model: &model.Model{Id: r.Id, Deleted: time.Now().Unix()}}
	if err := global.DB.
		Model(&info).
		Updates(&info).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	global.OK(c, nil)
}

func (s *Dict) List(c *gin.Context) {
	var req sys_dict.ListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if req.PerPage <= 0 || (req.PerPage > 100) {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	var ls []*model.SysDict
	tx := global.DB.Model(&model.SysDict{})
	if req.Key != "" {
		tx = tx.Where("`key` = ?", req.Key)
	}
	if req.Pkey != "" {
		tx = tx.Where("pkey = ?", req.Pkey)
	}
	if req.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+req.Name+"%")
	}

	tx = tx.Where("deleted = ?", 0)
	tx = tx.Order("id ASC")
	tx = tx.Order("sort DESC")

	if len(req.Pkey) <= 0 {
		ret := sys_dict.ListResponse{
			Rows: make([]sys_dict.Response, 0, 10),
		}
		if err := tx.Count(&ret.Count).
			Where("pkey = ?", req.Pkey).
			Offset((req.Page - 1) * req.PerPage).
			Limit(req.PerPage).
			Find(&ls).Error; err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		for _, l := range ls {
			var childCount int64
			if err := global.DB.Model(&model.SysDict{}).
				Where("pkey = ?", l.Base.Key).Count(&childCount).Error; err != nil {
				global.FAIL(c, "fail.msg", err.Error())
				return
			}
			row := sys_dict.Response{
				Request: sys_dict.Request{
					Name:    l.Base.Name,
					Key:     l.Base.Key,
					Pkey:    l.Base.Pkey,
					Summary: l.Base.Summary,
					Content: l.Base.Content,
					Sort:    l.Base.Sort,
				},
				Id:      l.Model.Id,
				Created: l.Model.Created,
				Updated: l.Model.Updated,
				Defer:   childCount > 0,
			}
			ret.Rows = append(ret.Rows, row)
		}
		global.OK(c, ret)
		return
	}
	if err := tx.Find(&ls).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret := sys_dict.ListItem{
		Children: make([]*sys_dict.Response, 0, req.PerPage),
	}
	var info model.SysDict
	if err := global.DB.Model(&model.SysDict{}).
		Where("`key` = ?", req.Pkey).First(&info).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret.Response = sys_dict.Response{
		Request: sys_dict.Request{
			Name:    info.Base.Name,
			Key:     info.Base.Key,
			Pkey:    info.Base.Pkey,
			Summary: info.Base.Summary,
			Sort:    info.Base.Sort,
			Content: info.Base.Content,
		},
		Id:      info.Model.Id,
		Created: info.Model.Created,
		Updated: info.Model.Updated,
	}
	for _, l := range ls {
		var childCount int64
		if err := global.DB.Model(&model.SysDict{}).
			Where("pkey = ?", l.Base.Key).Count(&childCount).Error; err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		row := sys_dict.Response{
			Request: sys_dict.Request{
				Name:    l.Base.Name,
				Key:     l.Base.Key,
				Pkey:    l.Base.Pkey,
				Summary: l.Base.Summary,
				Content: l.Base.Content,
				Sort:    l.Base.Sort,
			},
			Id:      l.Model.Id,
			Created: l.Model.Created,
			Updated: l.Model.Updated,
			Defer:   childCount > 0,
		}
		ret.Children = append(ret.Children, &row)
	}
	global.OK(c, ret)
}

func (s *Dict) Tree(c *gin.Context) {
	var r sys_dict.QueryWithKey
	if err := c.ShouldBindQuery(&r); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ls, err := s.dict.ALL(c)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if len(ls) == 0 && r.Key == "collectCategory" {
		if err = s.dict.CollectCategoryInitialize(c); err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		ls, err = s.dict.ALL(c)
		if err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
	}
	defaultKey := "options"
	ret := make(map[string][]*sys_dict.DictTree, len(ls))
	keys := make([]string, 0, 1)
	if len(r.Key) > 0 {
		keys = strings.Split(r.Key, ",")
	} else {
		keys = append(keys, defaultKey)
	}

	tree := make([]*sys_dict.DictTree, 0, len(ls))
	parentKeys := make(map[string]struct{}, len(keys))
	for _, v := range ls {
		item := sys_dict.DictTree{
			Pkey:     v.Base.Pkey,
			Key:      v.Base.Key,
			Label:    v.Base.Name,
			Value:    v.Base.Key,
			ID:       v.Model.Id,
			Children: nil,
		}
		switch r.Option {
		case "value":
			var m sys_dict.DictValue
			_ = json.Unmarshal(v.Base.Content, &m)
			item.Value = m.Value
		}
		for _, key := range keys {
			if strings.Contains(key, v.Base.Key) {
				parentKeys[key] = struct{}{}
			}
		}
		tree = append(tree, &item)
	}
	for _, key := range keys {
		var parentKey string
		if _, ok := parentKeys[key]; ok {
			parentKey = key
		}
		opts := s.dict.GetTreeRecursive(tree, parentKey)
		ret[key] = opts
	}
	global.OK(c, ret)
}
