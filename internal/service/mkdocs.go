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
	docs := Mkdocs{
		SiteName: VerifyFileName(title),
		Navs:     make([]Nav, 0, len(ls)),
	}
	intro := fmt.Sprintf("%s.md", title)
	docs.Navs = append(docs.Navs, Nav{Items: []string{intro}})
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

			fileName := baseName + ".md"
			fpath := path.Join(taskId, "docs", fileName)
			if _, err := global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(markdown)))); err != nil {
				return "", err
			}
			docs.Navs = append(docs.Navs, Nav{Items: []string{fileName}})
		}
	}
	fpath := path.Join(taskId, "docs", intro)
	if _, err := global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(indexMarkdown)))); err != nil {
		return "", err
	}
	indexPath := path.Join(taskId, "docs", "index.md")
	if _, err := global.Storage.Put(indexPath, io.NopCloser(bytes.NewBuffer([]byte(title)))); err != nil {
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
	docs := Mkdocs{
		SiteName: VerifyFileName(title),
		Navs:     make([]Nav, 0, len(ls)),
	}
	intro := fmt.Sprintf("%s.md", title)
	docs.Navs = append(docs.Navs, Nav{Items: []string{intro}})
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
			baseName := VerifyFileName(articleData.Info.Title)
			fileName := baseName + ".md"
			fpath := path.Join(group, title, "docs", fileName)
			if _, err := global.Storage.Put(fpath, io.NopCloser(bytes.NewBuffer([]byte(markdown)))); err != nil {
				return err
			}
			docs.Navs = append(docs.Navs, Nav{Items: []string{fileName}})
		}
	}
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

	converter := md.NewConverter("", true, nil)
	var ls []model.Task
	if err := global.DB.Model(&model.Task{}).
		Where(&model.Task{TaskPid: taskId}).
		Order("id asc").
		Find(&ls).Error; err != nil {
		return nil, err
	}
	indexMarkdown, err1 := converter.ConvertString(introHTML)
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
		if markdown, err2 := converter.ConvertString(articleData.Info.Content); err2 != nil {
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
