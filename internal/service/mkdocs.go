package service

import (
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

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/internal/types/task"
	"go.uber.org/zap"
)

//go:embed mkdocs.tpl
var MkdocsYML string

type Mkdocs struct {
	SiteName string
	Navs     []Nav
}

type Nav struct {
	Name  string
	Items []string
}

func MakeDocsite(ctx context.Context, taskId, title, introHTML string) (string, error) {
	converter := md.NewConverter("", true, nil)
	var ls []model.Task
	if err := global.DB.Model(&model.Task{}).
		Where(&model.Task{TaskPid: taskId}).
		Order("id asc").
		Find(&ls).Error; err != nil {
		return "", err
	}
	indexMarkdown, err1 := converter.ConvertString(introHTML)
	if err1 != nil {
		return "", err1
	}
	title = html.EscapeString(VerifyFileName(title))
	docs := Mkdocs{
		SiteName: title,
		Navs:     make([]Nav, 0, len(ls)),
	}
	for _, x := range ls {
		var articleData geek.ArticleData
		if err := json.Unmarshal(x.Raw, &articleData); err != nil {
			return "", err
		}
		if len(articleData.Info.Cshort) > len(articleData.Info.Content) {
			articleData.Info.Content = articleData.Info.Cshort
		}
		if markdown, err2 := converter.ConvertString(articleData.Info.Content); err2 != nil {
			return "", err2
		} else if len(markdown) > 0 {
			baseName := html.EscapeString(VerifyFileName(articleData.Info.Title))
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

			fileName := baseName + ".md"
			fpath := path.Join(taskId, "docs", fileName)
			if _, err := global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(markdown)))); err != nil {
				return "", err
			}
			docs.Navs = append(docs.Navs, Nav{Items: []string{fileName}})
		}
	}
	fpath := path.Join(taskId, "docs", "index.md")
	if _, err := global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(indexMarkdown)))); err != nil {
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
	mkdocsPath := path.Join(taskId, "mkdocs.yml")
	if _, err = global.Storage.Put(mkdocsPath, io.NopCloser(&buf)); err != nil {
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

func MakeDocsiteLocal(taskId, group, title, introHTML string) error {
	converter := md.NewConverter("", true, nil)
	var ls []model.Task
	if err := global.DB.Model(&model.Task{}).
		Where(&model.Task{TaskPid: taskId}).
		Order("id asc").
		Find(&ls).Error; err != nil {
		return err
	}
	indexMarkdown, err1 := converter.ConvertString(introHTML)
	if err1 != nil {
		return err1
	}
	title = html.EscapeString(VerifyFileName(title))
	docs := Mkdocs{
		SiteName: title,
		Navs:     make([]Nav, 0, len(ls)),
	}
	for _, x := range ls {
		var articleData geek.ArticleData
		if err := json.Unmarshal(x.Raw, &articleData); err != nil {
			return err
		}
		if len(articleData.Info.Cshort) > len(articleData.Info.Content) {
			articleData.Info.Content = articleData.Info.Cshort
		}
		if markdown, err2 := converter.ConvertString(articleData.Info.Content); err2 != nil {
			return err2
		} else if len(markdown) > 0 {
			baseName := html.EscapeString(VerifyFileName(articleData.Info.Title))
			fileName := baseName + ".md"
			fpath := path.Join(group, title, "docs", fileName)
			if _, err := global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(markdown)))); err != nil {
				return err
			}
			docs.Navs = append(docs.Navs, Nav{Items: []string{fileName}})
		}
	}
	fpath := path.Join(group, title, "docs", "index.md")
	if _, err := global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(indexMarkdown)))); err != nil {
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
	mkdocsPath := path.Join(group, title, "mkdocs.yml")
	if _, err = global.Storage.Put(mkdocsPath, io.NopCloser(&buf)); err != nil {
		return err
	}

	return nil
}
