package v2

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/sys_dict"
	"gorm.io/gorm"
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
	err := global.DB.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if r.Pkey == "" {
			info.Base.Rkey = r.Key
		} else {
			var parent model.SysDict
			if err := tx.Model(&model.SysDict{}).
				Where("`key` = ?", r.Pkey).
				First(&parent).Error; err != nil {
				return err
			}
			info.Base.Rkey = parent.Base.Rkey
		}
		err := tx.Model(&model.SysDict{}).
			Where(&model.SysDict{
				Base: &model.SysDictBase{
					Pkey: info.Base.Pkey,
					Key:  info.Base.Key,
				},
			}).
			FirstOrCreate(&info).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
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
	info := model.SysDict{Model: &model.Model{Id: r.Id}}
	if err := global.DB.Model(&info).First(&info).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	info.Base.Name = r.Name
	info.Base.Summary = r.Summary
	info.Base.Sort = r.Sort
	info.Base.Content = r.Content
	if err := global.DB.Model(&info).Updates(&info).Error; err != nil {
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
	if req.Name+req.Key+req.Pkey == "" {
		tx = tx.Where("pkey = ?", req.Pkey)
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
	var r sys_dict.QueryTree
	if err := c.ShouldBind(&r); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	keys := strings.Split(r.Key, ",")
	if len(r.Key) == 0 {
		keys = []string{""}
		r.FiledName = "options"
	}
	filedNames := make([]string, 0, len(keys))
	if len(r.FiledName) > 0 {
		filedNames = strings.Split(r.FiledName, ",")
	}
	if len(filedNames) == 0 {
		filedNames = keys
	}
	if len(filedNames) != len(keys) {
		global.FAIL(c, "fail")
		return
	}
	noChilds := make([]bool, len(filedNames))
	noChildArr := strings.Split(r.NoChild, ",")
	if len(noChildArr) == len(filedNames) {
		for i := 0; i < len(filedNames); i++ {
			noChilds[i], _ = strconv.ParseBool(noChildArr[i])
		}
	}
	scores := make([]func(db *gorm.DB) *gorm.DB, 0, len(keys))
	if len(r.Key) > 0 {
		scope := func(db *gorm.DB) *gorm.DB {
			return db.Where("rkey IN (?)", keys)
		}
		scores = append(scores, scope)
	}
	ls, err := s.dict.ALL(c, scores...)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if len(ls) == 0 {
		raw, err1 := global.ASSETS.ReadFile("web/pages/tags.json")
		if err1 != nil {
			global.FAIL(c, "fail.msg", err1.Error())
			return
		}
		var tagData sys_dict.TagData
		if err = json.Unmarshal(raw, &tagData); err != nil {
			global.FAIL(c, "fail.msg", err1.Error())
			return
		}
		for _, key := range keys {
			switch key {
			case sys_dict.CollectCategoryKey:
				if err = service.CollectCategoryInitialize(c, tagData); err != nil {
					global.FAIL(c, "fail.msg", err.Error())
					return
				}
			case sys_dict.GeektimeCategoryKey:
				if err = service.GeektimeCategory(c, tagData); err != nil {
					global.FAIL(c, "fail.msg", err.Error())
					return
				}
			}
		}
		ls, err = s.dict.ALL(c, scores...)
		if err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
	}
	tree := make([]*sys_dict.DictTree, 0, len(ls))
	for _, v := range ls {
		item := sys_dict.DictTree{
			Data:     v.Base.Content,
			Pkey:     v.Base.Pkey,
			Key:      v.Base.Key,
			Label:    v.Base.Name,
			Value:    v.Base.Key,
			ID:       v.Model.Id,
			Children: nil,
		}
		tree = append(tree, &item)
	}

	ret := make(map[string][]*sys_dict.DictTree, len(ls))
	for k, v := range keys {
		opts := s.dict.GetTreeRecursive(tree, v, noChilds[k])
		ret[filedNames[k]] = opts
	}
	global.OK(c, ret)
}
