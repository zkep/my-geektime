package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/zkep/my-geektime/internal/config"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/initialize"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/geek"
	"github.com/zkep/my-geektime/internal/types/sys_dict"
	"github.com/zkep/my-geektime/internal/types/task"
	"github.com/zkep/my-geektime/internal/types/user"
	"github.com/zkep/my-geektime/libs/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type DataFlags struct {
	Config     string   `name:"config" description:"Path to config file"`
	Types      []string `name:"types" description:"1: 体系课，4:公开课，5:线下大会，6:社区课，d:每日一课，q:大厂案例"`
	Cookies    string   `name:"cookies" description:"geektime cookies string"`
	Download   bool     `name:"download" description:"download geektime source" default:"false"`
	Preview    bool     `name:"preview" description:"preview geektime source" default:"false"`
	PreviewNum int      `name:"preview_num" description:"preview geektime source number" default:"1"`
}

func (app *App) Data(f *DataFlags) error {
	var (
		cfg         config.Config
		accessToken string
		configRaw   []byte
		err         error
	)
	if len(f.Types) == 0 {
		return fmt.Errorf("no the types %+v", f.Types)
	}
	if f.Config == "" {
		configRaw, err = app.assets.ReadFile("config.yml")
	} else {
		configRaw, err = os.ReadFile(f.Config)
	}
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(configRaw, &cfg); err != nil {
		return err
	}
	global.CONF = &cfg
	global.CONF.Site.Download = f.Download
	global.ASSETS = app.assets
	if err = initialize.Gorm(app.ctx); err != nil {
		return err
	}
	if err = initialize.Logger(app.ctx); err != nil {
		return err
	}
	if err = initialize.Storage(app.ctx); err != nil {
		return err
	}
	if len(f.Cookies) > 0 {
		accessToken = f.Cookies
		if cookies := os.Getenv("cookies"); len(cookies) > 0 {
			accessToken = cookies
		}
	} else {
		var u model.User
		if err = global.DB.
			Where(&model.User{RoleId: user.AdminRoleId}).
			First(&u).Error; err != nil {
			return err
		}
		accessToken = u.AccessToken
	}
	if accessToken == "" {
		return errors.New("no access token")
	}
	after := func(r *http.Response) error {
		var auth geek.AuthResponse
		authData, err1 := service.GetGeekUser(r, &auth)
		if err1 != nil {
			global.LOG.Error("GetGeekUser", zap.Error(err1))
			return err1
		}
		if authData.UID <= 0 {
			return fmt.Errorf("no user")
		}
		return nil
	}
	if err = service.Authority(accessToken, after); err != nil {
		return err
	}
	tagRaw, err := app.assets.ReadFile("web/pages/tags.json")
	if err != nil {
		return err
	}
	var tagData sys_dict.TagData
	if err = json.Unmarshal(tagRaw, &tagData); err != nil {
		return err
	}
	var allTask []*model.Task
	if err = global.DB.Model(&model.Task{}).
		Select([]string{"id", "other_id", "other_type", "other_tag", "other_form", "other_group"}).
		Where("task_pid=?", "").
		Find(&allTask).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	for _, x := range allTask {
		dataTasksMap[x.OtherId] = &DataTask{
			TaskId:     x.TaskId,
			OtherType:  x.OtherType,
			OtherTag:   x.OtherTag,
			OtherForm:  x.OtherForm,
			OtherGroup: x.OtherGroup,
		}
	}
	fmt.Printf("database exists product [%d]\n\n", len(dataTasksMap))
	for _, typ := range f.Types {
		otherType, ok := sys_dict.ProductTypes[typ]
		if !ok {
			fmt.Printf("not found product typ [%s]", typ)
			continue
		}
		if f.Preview {
			switch typ {
			case "d":
				if err = app.iteratorsDailyLesson(f.Preview, f.PreviewNum, otherType,
					emptyOption, emptyOption, accessToken, typ); err != nil {
					return err
				}
			case "q":
				if err = app.iteratorsRealCase(f.Preview, otherType, accessToken, typ); err != nil {
					return err
				}
			default:
				if err = app.iteratorsProduct(f.Preview, f.PreviewNum,
					otherType, emptyOption, emptyOption, emptyOption, accessToken); err != nil {
					return err
				}
			}
		} else {
			switch typ {
			case "d":
				for _, otherGroup := range tagData.Data {
					for _, otherTag := range otherGroup.Options {
						if err = app.iteratorsDailyLesson(f.Preview, f.PreviewNum, otherType,
							otherTag, otherGroup.Option, accessToken, typ); err != nil {
							return err
						}
					}
					if err = app.iteratorsDailyLesson(f.Preview, f.PreviewNum, otherType,
						emptyOption, otherGroup.Option, accessToken, typ); err != nil {
						return err
					}
				}
				if err = app.iteratorsDailyLesson(f.Preview, f.PreviewNum, otherType,
					emptyOption, emptyOption, accessToken, typ); err != nil {
					return err
				}
			case "q":
				if err = app.iteratorsRealCase(f.Preview, otherType, accessToken, typ); err != nil {
					return err
				}
			default:
				for _, otherForm := range sys_dict.ProductForms {
					for _, otherGroup := range tagData.Data {
						for _, otherTag := range otherGroup.Options {
							if err = app.iteratorsProduct(f.Preview, f.PreviewNum, otherType, otherForm,
								otherTag, otherGroup.Option, accessToken); err != nil {
								return err
							}
						}
						if err = app.iteratorsProduct(f.Preview, f.PreviewNum, otherType, otherForm, emptyOption,
							otherGroup.Option, accessToken, otherGroup.Value); err != nil {
							return err
						}
					}
					if err = app.iteratorsProduct(f.Preview, f.PreviewNum,
						otherType, otherForm, emptyOption, emptyOption, accessToken); err != nil {
						return err
					}
				}
			}
		}
	}

	fmt.Printf("product len [%d]\n\n", len(dataTasksMap))
	for k, v := range dataTasksMap {
		otherType, exists := sys_dict.OriginTypes[v.OtherType]
		if !exists {
			fmt.Printf("add articles otherType not exists, productId:%s, otherType:%d\n", k, v.OtherType)
			continue
		}
		if otherType == "d" {
			err = app.addArticlesWithID(accessToken, k, v)
			if err != nil {
				fmt.Printf("addArticlesWithID failed, productId:%s, err:%v\n", k, err)
				continue
			}
		} else {
			err = app.addArticlesWithProductID(accessToken, k, v)
			if err != nil {
				fmt.Printf("addArticlesWithProductID failed, productId:%s, err:%v\n", k, err)
				continue
			}
		}
	}
	return nil
}

var (
	emptyOption  = sys_dict.Option{}
	dataTasksMap = make(map[string]*DataTask, 500)
)

func (app *App) iteratorsProduct(preview bool, previewNum int, otherType, otherForm, otherTag,
	otherGroup sys_dict.Option, accessToken string, tags ...int32) error {

	prev, psize, hasNext, total := 1, 20, true, 0
	fmt.Printf(
		"iteratorsProduct start [%s/%s/%s/%s] \n",
		otherType.Label, otherForm.Label, otherGroup.Label, otherTag.Label,
	)
	if otherTag.Value > 0 {
		tags = append(tags, otherTag.Value)
	}
	req := geek.PvipProductRequest{
		TagIds:       tags,
		ProductType:  otherType.Value,
		ProductForm:  otherForm.Value,
		Sort:         8,
		Size:         psize,
		Prev:         prev,
		WithArticles: true,
	}
	for hasNext {
		req.Prev = prev
		resp, err := service.GetPvipProduct(app.ctx, accessToken, req)
		if err != nil {
			return err
		}
		if preview {
			hasNext = false
			if len(resp.Data.Products) > 0 {
				if previewNum <= 0 {
					previewNum = 1
				}
				resp.Data.Products = resp.Data.Products[:previewNum]
			}
		}
		total += len(resp.Data.Products)
		fmt.Printf(
			"iteratorsProduct [%s/%s/%s/%s] total: %d , pageTotal: %d, prev:%d \n",
			otherType.Label, otherForm.Label, otherGroup.Label,
			otherTag.Label, total, resp.Data.Page.Total, prev,
		)
		currLen := len(resp.Data.Products)
		if total >= resp.Data.Page.Total || currLen <= 0 {
			fmt.Printf(
				"iteratorsProduct end [%s/%s/%s/%s] total: %d , pageTotal: %d, prev:%d, currLen:%d \n",
				otherType.Label, otherForm.Label, otherGroup.Label,
				otherTag.Label, total, resp.Data.Page.Total, prev, currLen,
			)
			hasNext = false
		}
		prev++
		for _, product := range resp.Data.Products {
			if product.ID <= 0 {
				continue
			}
			otherId := fmt.Sprintf("%d", product.ID)
			if _, exists := dataTasksMap[otherId]; exists {
				continue
			}
			jobId := utils.HalfUUID()
			itemRaw, _ := json.Marshal(product)
			job := &model.Task{
				TaskId:     jobId,
				TaskName:   product.Title,
				TaskType:   service.TASK_TYPE_PRODUCT,
				OtherId:    otherId,
				Cover:      product.Cover.Square,
				Raw:        itemRaw,
				OtherType:  otherType.Value,
				OtherForm:  otherForm.Value,
				OtherGroup: otherGroup.Value,
				OtherTag:   otherTag.Value,
				Status:     service.TASK_STATUS_PENDING,
			}
			if err = global.DB.Model(&model.Task{}).
				Where(&model.Task{OtherId: job.OtherId}).
				Assign(job).FirstOrCreate(job).Error; err != nil {
				fmt.Println(err)
				continue
			}
			dataTasksMap[otherId] = &DataTask{
				TaskId:     job.TaskId,
				ArticleID:  int64(product.Article.ID),
				OtherType:  job.OtherType,
				OtherTag:   job.OtherTag,
				OtherForm:  job.OtherForm,
				OtherGroup: job.OtherGroup,
			}
		}
	}
	return nil
}

func (app *App) iteratorsDailyLesson(preview bool, previewNum int, otherType, otherTag,
	otherGroup sys_dict.Option, accessToken string, typ string) error {

	prev, psize, hasNext, total := 0, 20, true, 0
	fmt.Printf(
		"iteratorsDailyLesson start [%s/%s/%s] \n",
		otherType.Label, otherGroup.Label, otherTag.Label,
	)
	req := geek.DailyProductRequest{
		Type:      typ,
		Orderby:   "new",
		LabelID:   otherTag.Value,
		Direction: otherTag.Value,
		Size:      psize,
		Prev:      prev,
	}
	for hasNext {
		req.Prev = prev
		resp, err := service.GetProduct(app.ctx, accessToken, req)
		if err != nil {
			return err
		}
		if preview {
			hasNext = false
			if len(resp.Data.List) > 0 {
				if previewNum <= 0 {
					previewNum = 1
				}
				resp.Data.List = resp.Data.List[:previewNum]
			}
		}
		total += len(resp.Data.List)
		fmt.Printf(
			"iteratorsDailyLesson [%s/%s/%s] total: %d , pageTotal: %d, prev:%d \n",
			otherType.Label, otherGroup.Label,
			otherTag.Label, total, resp.Data.Page.Count, prev,
		)
		currLen := len(resp.Data.List)
		if total >= resp.Data.Page.Count || currLen <= 0 {
			fmt.Printf(
				"iteratorsDailyLesson end [%s/%s/%s] total: %d , pageTotal: %d, prev:%d, currLen:%d \n",
				otherType.Label, otherGroup.Label,
				otherTag.Label, total, resp.Data.Page.Count, prev, currLen,
			)
			hasNext = false
		}
		prev++
		for _, product := range resp.Data.List {
			if product.ID <= 0 {
				continue
			}
			otherId := fmt.Sprintf("%d", product.ID)
			jobId := utils.HalfUUID()
			itemRaw, _ := json.Marshal(product)
			job := &model.Task{
				TaskId:     jobId,
				TaskName:   product.Title,
				TaskType:   service.TASK_TYPE_PRODUCT,
				OtherId:    otherId,
				Cover:      product.Cover.Square,
				Raw:        itemRaw,
				OtherType:  otherType.Value,
				OtherGroup: otherGroup.Value,
				OtherTag:   otherTag.Value,
				Status:     service.TASK_STATUS_PENDING,
			}
			if err = global.DB.Model(&model.Task{}).
				Where(&model.Task{OtherId: job.OtherId}).
				Assign(job).FirstOrCreate(job).Error; err != nil {
				fmt.Println(err)
				continue
			}
			dataTasksMap[otherId] = &DataTask{
				TaskId:     job.TaskId,
				ArticleID:  int64(product.Article.ID),
				OtherType:  job.OtherType,
				OtherTag:   job.OtherTag,
				OtherForm:  job.OtherForm,
				OtherGroup: job.OtherGroup,
			}
		}
	}
	return nil
}

func (app *App) iteratorsRealCase(preview bool, otherType sys_dict.Option, accessToken string, typ string) error {
	prev, psize, hasNext, total := 0, 20, true, 0
	fmt.Printf("iteratorsRealCase start [%s] \n", otherType.Label)
	req := geek.DailyProductRequest{
		Type:    typ,
		Orderby: "new",
		Size:    psize,
		Prev:    prev,
	}
	for hasNext {
		req.Prev = prev
		resp, err := service.GetProduct(app.ctx, accessToken, req)
		if err != nil {
			return err
		}
		if preview {
			hasNext = false
		}
		total += len(resp.Data.List)
		fmt.Printf(
			"iteratorsRealCase [%s] total: %d , pageTotal: %d, prev:%d \n",
			otherType.Label, total, resp.Data.Page.Count, prev,
		)
		currLen := len(resp.Data.List)
		if total >= resp.Data.Page.Count || currLen <= 0 {
			fmt.Printf(
				"iteratorsRealCase end [%s] total: %d , pageTotal: %d, prev:%d, currLen:%d \n",
				otherType.Label, total, resp.Data.Page.Count, prev, currLen,
			)
			hasNext = false
		}
		prev = resp.Data.Page.Score
		itemMap := make(map[int]*geek.ProductItem, len(resp.Data.List))
		for _, product := range resp.Data.List {
			if product.ID <= 0 {
				continue
			}
			itemMap[product.ID] = &product
		}

		for _, topic := range resp.Data.Topics {
			tasks := make([]*model.Task, 0, len(topic.Pids)+1)
			jobId := utils.HalfUUID()
			for k, v := range topic.Pids {
				p, ok := itemMap[v]
				if !ok {
					continue
				}
				if k == 0 {
					itemRaw, _ := json.Marshal(p)
					job := &model.Task{
						TaskId:    jobId,
						TaskName:  topic.Title,
						TaskType:  service.TASK_TYPE_PRODUCT,
						OtherId:   fmt.Sprintf("%d", topic.ID),
						Cover:     topic.Cover,
						Raw:       itemRaw,
						OtherType: otherType.Value,
						Status:    service.TASK_STATUS_PENDING,
					}
					tasks = append(tasks, job)
				}
				article, err := service.GetArticleInfo(app.ctx,
					accessToken, geek.ArticlesInfoRequest{Id: int64(p.Article.ID)})
				if err != nil {
					fmt.Println(err)
					continue
				}
				if article.Data.Info.ID <= 0 {
					fmt.Printf(
						"iteratorsRealCase [%s]  articleInfo is empty, articleID: %d, ID: %d\n",
						otherType.Label, p.Article.ID, p.ID,
					)
					continue
				}
				var m geek.ArticleInfoRaw
				if err = json.Unmarshal(article.Raw, &m); err != nil {
					fmt.Println(err)
					continue
				}
				taskName := article.Data.Info.Title
				cover := article.Data.Info.Cover.Default
				item := &model.Task{
					TaskPid:   jobId,
					TaskId:    utils.HalfUUID(),
					OtherId:   fmt.Sprintf("%d", p.Article.ID),
					TaskName:  taskName,
					TaskType:  service.TASK_TYPE_ARTICLE,
					Cover:     cover,
					Raw:       m.Data,
					OtherType: otherType.Value,
					Status:    service.TASK_STATUS_PENDING,
				}
				tasks = append(tasks, item)
			}
			err = global.DB.Transaction(func(tx *gorm.DB) error {
				for _, x := range tasks {
					if err = tx.Model(&model.Task{}).
						Where(&model.Task{OtherId: x.OtherId}).
						Assign(x).FirstOrCreate(x).Error; err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
			for _, x := range tasks {
				dataTasksMap[x.OtherId] = &DataTask{
					TaskId:     x.TaskId,
					OtherType:  x.OtherType,
					OtherTag:   x.OtherTag,
					OtherForm:  x.OtherForm,
					OtherGroup: x.OtherGroup,
				}
			}
		}
	}
	return nil
}

func (app *App) addArticlesWithProductID(accessToken, productID string, info *DataTask) error {
	req := geek.ArticlesListRequest{
		Cid:   productID,
		Order: "earliest",
		Prev:  0,
		Size:  500,
	}
	articles, err1 := service.GetArticles(app.ctx, accessToken, req)
	if err1 != nil {
		fmt.Println(err1)
		return err1
	}
	tasks := make([]*model.Task, 0, len(articles.Data.List))
	for _, data := range articles.Data.List {
		if data.ID <= 0 || data.ArticleTitle == "" {
			fmt.Printf("articleID not found: %d, title: %s\n", data.ID, data.ArticleTitle)
			continue
		}
		article, err := service.GetArticleInfo(app.ctx, accessToken, geek.ArticlesInfoRequest{Id: data.ID})
		if err != nil {
			fmt.Println(err)
			continue
		}
		var m geek.ArticleInfoRaw
		if err = json.Unmarshal(article.Raw, &m); err != nil {
			fmt.Println(err)
			continue
		}
		raw := m.Data
		taskName := article.Data.Info.Title
		cover := article.Data.Info.Cover.Default
		item := model.Task{
			TaskPid:    info.TaskId,
			TaskId:     utils.HalfUUID(),
			OtherId:    fmt.Sprintf("%d", article.Data.Info.ID),
			TaskName:   taskName,
			TaskType:   service.TASK_TYPE_ARTICLE,
			Cover:      cover,
			Raw:        raw,
			OtherType:  info.OtherType,
			OtherForm:  info.OtherForm,
			OtherGroup: info.OtherGroup,
			OtherTag:   info.OtherTag,
			Status:     service.TASK_STATUS_PENDING,
		}
		tasks = append(tasks, &item)
	}
	statistics := task.TaskStatistics{
		Count: len(tasks),
		Items: map[int]int{
			service.TASK_STATUS_PENDING:  len(tasks),
			service.TASK_STATUS_RUNNING:  0,
			service.TASK_STATUS_FINISHED: 0,
			service.TASK_STATUS_ERROR:    0,
		},
	}
	statisticsRaw, _ := json.Marshal(statistics)
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Task{}).Where(&model.Task{OtherId: productID}).
			UpdateColumns(map[string]any{"statistics": statisticsRaw}).Error; err != nil {
			return err
		}
		for _, x := range tasks {
			if err := tx.Model(&model.Task{}).Where(&model.Task{OtherId: x.OtherId}).
				Assign(x).FirstOrCreate(x).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (app *App) addArticlesWithID(accessToken, productID string, info *DataTask) error {
	article, err := service.GetArticleInfo(app.ctx, accessToken, geek.ArticlesInfoRequest{Id: info.ArticleID})
	if err != nil {
		fmt.Println(err)
		return err
	}
	var m geek.ArticleInfoRaw
	if err = json.Unmarshal(article.Raw, &m); err != nil {
		fmt.Println(err)
		return err
	}
	raw := m.Data
	taskName := article.Data.Info.Title
	cover := article.Data.Info.Cover.Default
	tasks := make([]*model.Task, 0, 2)
	item := model.Task{
		TaskPid:    info.TaskId,
		TaskId:     utils.HalfUUID(),
		OtherId:    fmt.Sprintf("%d", info.ArticleID),
		TaskName:   taskName,
		TaskType:   service.TASK_TYPE_ARTICLE,
		Cover:      cover,
		Raw:        raw,
		OtherType:  info.OtherType,
		OtherForm:  info.OtherForm,
		OtherGroup: info.OtherGroup,
		OtherTag:   info.OtherTag,
		Status:     service.TASK_STATUS_PENDING,
	}
	tasks = append(tasks, &item)
	statistics := task.TaskStatistics{
		Count: len(tasks),
		Items: map[int]int{
			service.TASK_STATUS_PENDING:  len(tasks),
			service.TASK_STATUS_RUNNING:  0,
			service.TASK_STATUS_FINISHED: 0,
			service.TASK_STATUS_ERROR:    0,
		},
	}
	statisticsRaw, _ := json.Marshal(statistics)
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		if err = tx.Model(&model.Task{}).Where(&model.Task{OtherId: productID, TaskType: service.TASK_TYPE_PRODUCT}).
			UpdateColumns(map[string]any{"statistics": statisticsRaw}).Error; err != nil {
			return err
		}
		for _, x := range tasks {
			if err = tx.Model(&model.Task{}).Where(&model.Task{OtherId: x.OtherId, TaskType: service.TASK_TYPE_ARTICLE}).
				Assign(x).FirstOrCreate(x).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

type DataTask struct {
	TaskId     string
	ArticleID  int64
	OtherType  int32
	OtherTag   int32
	OtherForm  int32
	OtherGroup int32
}
