package service

import (
	"context"
	"sort"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/types/sys_dict"
	"github.com/zkep/my-geektime/libs/utils"
	"gorm.io/gorm"
)

type Dict struct{}

func (d *Dict) QueryWithKey(ctx context.Context, key string) (l *model.SysDict, err error) {
	err = global.DB.WithContext(ctx).Model(&model.SysDict{}).Where("key = ?", key).Find(&l).Error
	return
}

func (d *Dict) QueryWithPKey(ctx context.Context, pkey string) (ls []*model.SysDict, err error) {
	err = global.DB.WithContext(ctx).Model(&model.SysDict{}).Where("pkey = ?", pkey).Find(&ls).Error
	return
}

func (d *Dict) Queries(ctx context.Context, ids ...int64) (map[int64]*model.SysDict, error) {
	m := make(map[int64]*model.SysDict, len(ids))
	if len(ids) == 0 {
		return m, nil
	}
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		if _, ok := m[id]; !ok {
			args = append(args, id)
			m[id] = nil
		}
	}
	m = map[int64]*model.SysDict{}
	ls := make([]*model.SysDict, 0, len(ids))
	if err := global.DB.WithContext(ctx).
		Model(&model.SysDict{}).
		Where("id IN ?", args).Find(&ls).Error; err != nil {
		return nil, err
	}
	for _, l := range ls {
		m[l.Model.Id] = l
	}
	return m, nil
}

func (d *Dict) ALL(
	ctx context.Context,
	scopes ...func(*gorm.DB) *gorm.DB,
) ([]*model.SysDict, error) {
	var ls []*model.SysDict
	tx := global.DB.WithContext(ctx).Model(&model.SysDict{})
	if len(scopes) > 0 {
		tx = tx.Scopes(scopes...)
	}
	tx = tx.Where("deleted = ?", 0)
	tx = tx.Order("id ASC")
	tx = tx.Order("sort DESC")
	if err := tx.Find(&ls).Error; err != nil {
		return nil, err
	}
	return ls, nil
}

func (d *Dict) GetTreeRecursive(
	ls []*sys_dict.DictTree,
	parentKey string,
) []*sys_dict.DictTree {
	res := make([]*sys_dict.DictTree, 0, len(ls))
	for _, v := range ls {
		if v.Pkey == parentKey {
			v.Children = d.GetTreeRecursive(ls, v.Key)
			res = append(res, v)
		}
	}
	return res
}

func (d *Dict) GetBreadCrumb(ls []*model.SysDict, key string) []string {
	res := make(map[int64]*model.SysDict, len(ls))
	ids := make([]int64, 0, len(ls))
	currKey := key
	sort.Slice(ls, func(i, j int) bool {
		return ls[i].Model.Id > ls[j].Model.Id
	})
	for _, v := range ls {
		if currKey == v.Base.Key {
			ids = append(ids, v.Model.Id)
			currKey = v.Base.Pkey
		}
		res[v.Model.Id] = v
	}
	labels := make([]string, 0, len(ids))
	sort.Slice(ids, func(i, j int) bool {
		return i > j
	})
	for _, v := range ids {
		if item, ok := res[v]; ok {
			labels = append(labels, item.Base.Name)
		}
	}
	return labels
}

func (d *Dict) CollectCategoryInitialize(ctx context.Context) error {
	collectCategories := []model.SysDictBase{
		{
			Key:     "collectCategory",
			Pkey:    "",
			Name:    "收藏分类",
			Content: []byte("{}"),
		},
		{
			Key:     utils.HalfUUID(),
			Pkey:    "collectCategory",
			Name:    "全部",
			Content: []byte(`{"value":""}`),
		},
		{
			Key:     utils.HalfUUID(),
			Pkey:    "collectCategory",
			Name:    "后端/架构",
			Content: []byte(`{"value":"3"}`),
		},
		{
			Key:     utils.HalfUUID(),
			Pkey:    "collectCategory",
			Name:    "前端/移动",
			Content: []byte(`{"value":"5"}`),
		},
		{
			Key:     utils.HalfUUID(),
			Pkey:    "collectCategory",
			Name:    "计算机基础",
			Content: []byte(`{"value":"9"}`),
		},
		{
			Key:     utils.HalfUUID(),
			Pkey:    "collectCategory",
			Name:    "AI/大数据",
			Content: []byte(`{"value":"8"}`),
		},
		{
			Key:     utils.HalfUUID(),
			Pkey:    "collectCategory",
			Name:    "运维/测试",
			Content: []byte(`{"value":"6"}`),
		},
		{
			Key:     utils.HalfUUID(),
			Pkey:    "collectCategory",
			Name:    "产品/运营",
			Content: []byte(`{"value":"7"}`),
		},
		{
			Key:     utils.HalfUUID(),
			Pkey:    "collectCategory",
			Name:    "管理/成长",
			Content: []byte(`{"value":"4"}`),
		},
	}
	return global.DB.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			for _, category := range collectCategories {
				info := model.SysDict{Base: &category}
				if err := tx.Model(&model.SysDict{}).
					Where(&model.SysDict{
						Base: &model.SysDictBase{
							Pkey: category.Pkey,
							Key:  category.Key,
						},
					}).
					FirstOrCreate(&info).Error; err != nil {
					return err
				}
			}
			return nil
		})
}
