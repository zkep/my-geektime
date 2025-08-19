package service

import (
	"archive/tar"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"os/exec"
	"path"
	"sort"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/types/geek"
	"github.com/zkep/my-geektime/internal/types/task"
	"go.uber.org/zap"
)

//go:embed mkdocs.tpl
var MkdocsYML string

type Mkdocs struct {
	SiteName string
	Navs     []Nav
}

type Nav struct {
	Index int
	Name  string
	Items []string
}

var (
	commentHtmlFormat       = `<li><img src="%s" width="30px"><span>%s</span> 👍（%d） 💬（%d）<p>%s</p>%s</li><br/>`
	commentSimpleHtmlFormat = `<li><span>%s</span> 👍（%d） 💬（%d）<p>%s</p>%s</li><br/>`
)

func MakeDocsite(ctx context.Context, taskId, title, introHTML string) (string, error) {
	var ls []model.Task
	if err := global.DB.Model(&model.Task{}).
		Where(&model.Task{TaskPid: taskId}).
		Order("id asc").
		Find(&ls).Error; err != nil {
		return "", err
	}
	if rewrittenHTML, err := HtmlURLProxyReplace(introHTML); err == nil {
		introHTML = rewrittenHTML
	}
	indexMarkdown, err1 := htmltomarkdown.ConvertString(introHTML)
	if err1 != nil {
		return "", err1
	}
	docs := Mkdocs{
		SiteName: VerifyFileName(title),
		Navs:     make([]Nav, 0, len(ls)),
	}
	intro := fmt.Sprintf("%s.md", title)
	docs.Navs = append(docs.Navs, Nav{Items: []string{intro}})
	batch := global.GPool.NewBatch()
	for i := range ls {
		x, k := ls[i], i
		batch.Queue(func(_ context.Context) (any, error) {
			var articleData geek.ArticleData
			if err := json.Unmarshal(x.Raw, &articleData); err != nil {
				return nil, err
			}
			if len(articleData.Info.Title) == 0 {
				return nil, fmt.Errorf("title is empty %s", x.TaskId)
			}
			if len(articleData.Info.Cshort) > len(articleData.Info.Content) {
				articleData.Info.Content = articleData.Info.Cshort
			}
			if rewrittenContent, err2 := HtmlURLProxyReplace(articleData.Info.Content); err2 == nil {
				articleData.Info.Content = rewrittenContent
			}
			markdown, err2 := htmltomarkdown.ConvertString(articleData.Info.Content)
			if err2 != nil {
				return nil, err2
			}
			baseName := VerifyFileName(articleData.Info.Title)
			var itemMessage task.TaskMessage
			if len(x.Message) > 0 {
				_ = json.Unmarshal(x.Message, &itemMessage)
				if len(itemMessage.Object) > 0 {
					object := global.Storage.GetUrl(itemMessage.Object)
					playTpl := `<video id="video" controls="" preload="none"><source id="mp4" src="%s"></video><br/> %s`
					if articleData.Info.Audio.URL != "" {
						playTpl = `<audio id="audio" controls="" preload="none"><source id="mp3" src="%s"></audio><br/> %s`
					}
					markdown = fmt.Sprintf(playTpl, object, markdown)
				}
			}
			// article comments
			hasNext := true
			perPage := 20
			page := 1
			commentCount := int64(0)
			commentHtml := ""
			for hasNext {
				var comments []*model.ArticleComment
				tx := global.DB.Model(&model.ArticleComment{})
				if err := tx.Where("aid = ?", x.OtherId).
					Count(&commentCount).Offset((page - 1) * perPage).
					Limit(perPage + 1).Find(&comments).Error; err != nil {
					return nil, err
				}
				page++
				if len(comments) > perPage {
					comments = comments[:perPage]
				} else {
					hasNext = false
				}
				for _, comment := range comments {
					var row geek.ArticleComment
					if err := json.Unmarshal(comment.Raw, &row); err != nil {
						continue
					}
					row.UserHeader = URLProxyReplace(row.UserHeader)
					commentHtml += fmt.Sprintf(commentHtmlFormat, row.UserHeader,
						row.UserName, row.LikeCount, row.DiscussionCount, row.CommentContent,
						time.Unix(row.CommentCtime, 0).Format(time.DateOnly))
				}
			}

			if commentCount > 0 {
				markdown += fmt.Sprintf("\n<div><strong>全部留言（%d）</strong></div>", commentCount)
				markdown += fmt.Sprintf("<ul>\n%s\n</ul>", commentHtml)
			}
			fileName := baseName + ".md"
			fpath := path.Join(taskId, "docs", fileName)
			if _, err := global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(markdown)))); err != nil {
				return nil, err
			}
			return &Nav{Index: k + 1, Items: []string{fileName}}, nil
		})
	}
	ws, err := batch.Wait(ctx)
	if err != nil {
		return "", err
	}
	for _, w := range ws {
		if val, ok := w.Value.(*Nav); ok {
			docs.Navs = append(docs.Navs, *val)
		}
	}
	sort.Slice(docs.Navs, func(i, j int) bool {
		return docs.Navs[i].Index < docs.Navs[j].Index
	})
	fpath := path.Join(taskId, "docs", intro)
	if _, err = global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(indexMarkdown)))); err != nil {
		return "", err
	}
	indexPath := path.Join(taskId, "docs", "index.md")
	if _, err = global.Storage.Put(indexPath, io.NopCloser(bytes.NewBuffer([]byte(title)))); err != nil {
		return "", err
	}
	tpl, err := template.New("template").Parse(MkdocsYML)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err = tpl.Execute(&buf, docs); err != nil {
		return "", err
	}
	buff := bytes.NewBufferString(html.UnescapeString(buf.String()))
	mkdocsPath := path.Join(taskId, "mkdocs.yml")
	if _, err = global.Storage.Put(mkdocsPath, io.NopCloser(buff)); err != nil {
		return "", err
	}
	realDir := global.Storage.GetKey(taskId, true)
	cmd := exec.CommandContext(ctx, "mkdocs", "build")
	cmd.Dir = realDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		global.LOG.Error("docsite", zap.Error(err), zap.String("output", string(output)))
		return "", err
	}
	docDir := path.Join(taskId, "site")
	docURL := global.Storage.GetKey(docDir, false)
	return docURL, nil
}

func MakeDocsiteLocal(ctx context.Context, taskId, group, title, introHTML string, commentLen int) error {
	var ls []model.Task
	if err := global.DB.Model(&model.Task{}).
		Where(&model.Task{TaskPid: taskId}).
		Order("id asc").
		Find(&ls).Error; err != nil {
		return err
	}
	indexMarkdown, err1 := htmltomarkdown.ConvertString(introHTML)
	if err1 != nil {
		return err1
	}
	docs := Mkdocs{
		SiteName: VerifyFileName(title),
		Navs:     make([]Nav, 0, len(ls)),
	}
	intro := fmt.Sprintf("%s.md", title)
	docs.Navs = append(docs.Navs, Nav{Items: []string{intro}})
	batch := global.GPool.NewBatch()
	for i := range ls {
		x, k := ls[i], i
		batch.Queue(func(_ context.Context) (any, error) {
			var articleData geek.ArticleData
			if err := json.Unmarshal(x.Raw, &articleData); err != nil {
				return nil, err
			}
			if len(articleData.Info.Title) == 0 {
				return nil, fmt.Errorf("title is empty %s", x.TaskId)
			}
			if len(articleData.Info.Cshort) > len(articleData.Info.Content) {
				articleData.Info.Content = articleData.Info.Cshort
			}
			markdown, err2 := htmltomarkdown.ConvertString(articleData.Info.Content)
			if err2 != nil {
				return nil, err2
			}
			baseName := VerifyFileName(articleData.Info.Title)
			// article comments
			hasNext := true
			perPage := 20
			if commentLen > 0 {
				perPage = commentLen
			}
			page := 1
			commentCount := int64(0)
			commentHtml := ""
			count := 0
			for hasNext {
				var comments []*model.ArticleComment
				tx := global.DB.Model(&model.ArticleComment{})
				if err := tx.Where("aid = ?", x.OtherId).
					Count(&commentCount).Offset((page - 1) * perPage).
					Limit(perPage + 1).Find(&comments).Error; err != nil {
					return nil, err
				}
				page++
				if len(comments) > perPage {
					comments = comments[:perPage]
				} else {
					hasNext = false
				}
				for _, comment := range comments {
					if commentLen > 0 && count >= commentLen {
						hasNext = false
						break
					}
					count++
					var row geek.ArticleComment
					if err := json.Unmarshal(comment.Raw, &row); err != nil {
						continue
					}
					commentHtml += fmt.Sprintf(commentSimpleHtmlFormat,
						row.UserName, row.LikeCount, row.DiscussionCount, row.CommentContent,
						time.Unix(row.CommentCtime, 0).Format(time.DateOnly))
				}
			}

			if commentCount > 0 {
				label := "全部留言"
				if commentLen > 0 {
					label = "精选留言"
					commentCount = int64(count)
				}
				markdown += fmt.Sprintf("\n<div><strong>%s（%d）</strong></div>", label, commentCount)
				markdown += fmt.Sprintf("<ul>\n%s\n</ul>", commentHtml)
			}
			fileName := baseName + ".md"
			fpath := path.Join(group, title, "docs", fileName)
			if _, err := global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(markdown)))); err != nil {
				return nil, err
			}
			return &Nav{Index: k + 1, Items: []string{fileName}}, nil
		})
	}
	ws, err1 := batch.Wait(ctx)
	if err1 != nil {
		return err1
	}
	for _, w := range ws {
		if val, ok := w.Value.(*Nav); ok {
			docs.Navs = append(docs.Navs, *val)
		}
	}
	sort.Slice(docs.Navs, func(i, j int) bool {
		return docs.Navs[i].Index < docs.Navs[j].Index
	})

	fpath := path.Join(group, title, "docs", intro)
	if _, err := global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(indexMarkdown)))); err != nil {
		return err
	}
	indexPath := path.Join(group, title, "docs", "index.md")
	if _, err := global.Storage.Put(indexPath, io.NopCloser(bytes.NewBuffer([]byte(title)))); err != nil {
		return err
	}
	tpl, err := template.New("template").Parse(MkdocsYML)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err = tpl.Execute(&buf, docs); err != nil {
		return err
	}
	buff := bytes.NewBufferString(html.UnescapeString(buf.String()))
	mkdocsPath := path.Join(group, title, "mkdocs.yml")
	if _, err = global.Storage.Put(mkdocsPath, io.NopCloser(buff)); err != nil {
		return err
	}
	return nil
}

func MakeDocArchive(_ context.Context, taskId, title, introHTML string) (*bytes.Buffer, error) {

	buf := new(bytes.Buffer)
	archiveWriter := tar.NewWriter(buf)

	defer func() { _ = archiveWriter.Close() }()

	var ls []model.Task
	if err := global.DB.Model(&model.Task{}).
		Where(&model.Task{TaskPid: taskId}).
		Order("id asc").
		Find(&ls).Error; err != nil {
		return nil, err
	}
	indexMarkdown, err1 := htmltomarkdown.ConvertString(introHTML)
	if err1 != nil {
		return nil, err1
	}
	indexMarkdown += "\n\n\n\n\n\n"

	docs := Mkdocs{
		SiteName: VerifyFileName(title),
		Navs:     make([]Nav, 0, len(ls)),
	}

	docs.Navs = append(docs.Navs, Nav{Items: []string{"index.md"}})
	for _, x := range ls {
		var articleData geek.ArticleData
		if err := json.Unmarshal(x.Raw, &articleData); err != nil {
			return nil, err
		}
		if len(articleData.Info.Cshort) > len(articleData.Info.Content) {
			articleData.Info.Content = articleData.Info.Cshort
		}
		if markdown, err2 := htmltomarkdown.ConvertString(articleData.Info.Content); err2 != nil {
			return nil, err2
		} else if len(markdown) > 0 {
			baseName := VerifyFileName(articleData.Info.Title)
			fileName := fmt.Sprintf("%s.md", baseName)
			archiveFileName := fmt.Sprintf("docs/%s", fileName)
			docs.Navs = append(docs.Navs, Nav{Items: []string{fileName}})
			hdr := &tar.Header{
				Name: archiveFileName,
				Mode: 0600,
				Size: int64(len(markdown)),
			}
			if err := archiveWriter.WriteHeader(hdr); err != nil {
				return nil, err
			}
			if _, err := archiveWriter.Write([]byte(markdown)); err != nil {
				return nil, err
			}
			indexMarkdown += fmt.Sprintf("\n * [%s](./%s) \n", baseName, fileName)
		}
	}

	hdr := &tar.Header{
		Name: "docs/index.md",
		Mode: 0600,
		Size: int64(len(indexMarkdown)),
	}
	if err := archiveWriter.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := archiveWriter.Write([]byte(indexMarkdown)); err != nil {
		return nil, err
	}
	tpl, err := template.New("template").Parse(MkdocsYML)
	if err != nil {
		return nil, err
	}
	var buff bytes.Buffer
	if err = tpl.Execute(&buff, docs); err != nil {
		return nil, err
	}
	buff2 := bytes.NewBufferString(html.UnescapeString(buff.String()))

	mkdocsHdr := &tar.Header{
		Name: "mkdocs.yml",
		Mode: 0600,
		Size: int64(len(buff2.Bytes())),
	}
	if err = archiveWriter.WriteHeader(mkdocsHdr); err != nil {
		return nil, err
	}
	if _, err = archiveWriter.Write(buff2.Bytes()); err != nil {
		return nil, err
	}
	return buf, nil
}
