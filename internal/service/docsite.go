package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/types/geek"
	"go.uber.org/zap"
)

// DocGenerator 文档生成器接口
type DocGenerator interface {
	// MakeDocsite 生成在线文档站点
	MakeDocsite(ctx context.Context, taskId, title, introHTML string) (string, error)

	// MakeDocsiteLocal 生成本地文档站点（带评论）
	MakeDocsiteLocal(ctx context.Context, taskId, group, title, introHTML string, commentLen int) error

	// MakeDocArchive 生成 Markdown 归档
	MakeDocArchive(ctx context.Context, taskId, title, introHTML string) (*bytes.Buffer, error)
}

// NewGoldmarkGenerator 创建 Goldmark 文档生成器
func NewGoldmarkGenerator() DocGenerator {
	return NewGoldmarkDocGenerator()
}

//go:embed docsite.html.tpl
var DocsiteHTMLTPL string

// GoldmarkDocGenerator 基于 Goldmark 的文档生成器
type GoldmarkDocGenerator struct {
	md goldmark.Markdown
}

// NewGoldmarkDocGenerator 创建 Goldmark 文档生成器
func NewGoldmarkDocGenerator() *GoldmarkDocGenerator {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // 允许原始 HTML（视频、音频等）
		),
	)

	return &GoldmarkDocGenerator{
		md: md,
	}
}

// ArticleInfo 文章信息
type ArticleInfo struct {
	Index    int
	Title    string
	Filename string
}

// MakeDocsite 使用 Goldmark 生成在线文档站点
func (g *GoldmarkDocGenerator) MakeDocsite(ctx context.Context, taskId, title, introHTML string) (string, error) {
	var tasks []model.Task
	if err := global.DB.WithContext(ctx).Model(&model.Task{}).
		Where(&model.Task{TaskPid: taskId}).
		Order("id asc").
		Find(&tasks).Error; err != nil {
		return "", fmt.Errorf("query tasks failed: %w", err)
	}

	global.LOG.Info("docsite generation started",
		zap.String("taskId", taskId),
		zap.Int("task_count", len(tasks)))

	if len(tasks) == 0 {
		return "", fmt.Errorf("no tasks found for task %s, please check if the task has completed downloading", taskId)
	}

	// 处理介绍页 HTML
	if rewrittenHTML, err := HtmlURLProxyReplace(introHTML); err == nil {
		introHTML = rewrittenHTML
	}

	// 创建目录结构（使用 Storage）
	docDir := path.Join(taskId, "docs")

	// 准备文章列表
	articles := make([]ArticleInfo, 0, len(tasks))
	articleHTMLs := make(map[string]string) // 保存每篇文章的 HTML 内容
	// articleHTMLs 的写入发生在 GPool 的并发 goroutine 中，必须加锁，
	// 否则 Go 运行时会因 concurrent map writes 直接终止进程（fatal error，无法 recover）。
	var articleHTMLsMu sync.Mutex
	batch := global.GPool.NewBatch()

	for i := range tasks {
		x, k := tasks[i], i
		batch.Queue(func(_ context.Context) (any, error) {
			var articleData geek.ArticleData
			if err := json.Unmarshal(x.Raw, &articleData); err != nil {
				return nil, err
			}

			if len(articleData.Info.Title) == 0 {
				return nil, fmt.Errorf("title is empty %s", x.TaskId)
			}

			// 使用 Cshort 如果 Content 为空
			content := articleData.Info.Content
			if len(content) == 0 && len(articleData.Info.Cshort) > 0 {
				content = articleData.Info.Cshort
			}

			// URL 代理替换
			if rewrittenContent, err := HtmlURLProxyReplace(content); err == nil {
				content = rewrittenContent
			}

			// HTML 转 Markdown
			markdown, err := HTMLConvertMarkdown(content)
			if err != nil {
				return nil, fmt.Errorf("convert html to markdown: %w", err)
			}
			var playTpl string
			coverURL := ""
			if articleData.Info.Cover.Default != "" {
				coverURL = articleData.Info.Cover.Default
			}

			if len(x.Ciphertext) > 0 || len(x.RewriteHls) > 0 {
				hlsURL := fmt.Sprintf("%s/v2/task/play.m3u8?id=%s", strings.TrimSuffix(global.CONF.Storage.Host, "/"), x.TaskId)
				if coverURL != "" {
					playTpl = `<div style="width:100%%;max-width:100%%;aspect-ratio:16/9;background:#000;border-radius:8px;overflow:hidden;margin:1.5rem 0;"><video id="video" controls preload="none" playsinline poster="%s" style="width:100%%;height:100%%;display:block;"><source src="%s" type="application/x-mpegURL"></video></div><br/> %s`
					markdown = fmt.Sprintf(playTpl, coverURL, hlsURL, markdown)
				} else {
					playTpl = `<div style="width:100%%;max-width:100%%;aspect-ratio:16/9;background:#000;border-radius:8px;overflow:hidden;margin:1.5rem 0;"><video id="video" controls preload="none" playsinline style="width:100%%;height:100%%;display:block;"><source src="%s" type="application/x-mpegURL"></video></div><br/> %s`
					markdown = fmt.Sprintf(playTpl, hlsURL, markdown)
				}
			} else if articleData.Info.Audio.DownloadURL != "" {
				if coverURL != "" {
					playTpl = `<div style="width:100%%;max-width:100%%;margin:1.5rem 0;border-radius:8px;overflow:hidden;background:#f6f7f8;"><img src="%s" style="width:100%%;max-height:300px;object-fit:cover;display:block;"><div style="padding:1rem;"><audio id="audio" controls preload="none" style="width:100%%;display:block;"><source id="mp3" src="%s"></audio></div></div><br/> %s`
					markdown = fmt.Sprintf(playTpl, coverURL, articleData.Info.Audio.DownloadURL, markdown)
				} else {
					playTpl = `<div style="width:100%%;max-width:100%%;margin:1.5rem 0;padding:1rem;background:#f6f7f8;border-radius:8px;"><audio id="audio" controls preload="none" style="width:100%%;display:block;"><source id="mp3" src="%s"></audio></div><br/> %s`
					markdown = fmt.Sprintf(playTpl, articleData.Info.Audio.DownloadURL, markdown)
				}
			}
			// 获取评论
			var aid int64
			if _, err := fmt.Sscanf(x.OtherId, "%d", &aid); err != nil {
				global.LOG.Warn("parse other id failed",
					zap.String("other_id", x.OtherId),
					zap.Error(err))
			}
			commentHTML := g.getCommentsHTML(ctx, aid, 0)
			if commentHTML != "" {
				markdown += fmt.Sprintf("\n\n<div><strong>全部留言（%d）</strong></div>", strings.Count(commentHTML, "<li>"))
				markdown += fmt.Sprintf("<ul>\n%s\n</ul>", commentHTML)
			}

			// Markdown 转 HTML
			var htmlBuf bytes.Buffer
			if err := g.md.Convert([]byte(markdown), &htmlBuf); err != nil {
				return nil, fmt.Errorf("convert markdown to html: %w", err)
			}

			// 生成文件名
			baseName := VerifyFileName(articleData.Info.Title)
			fileName := fmt.Sprintf("%d_%s.html", k+1, baseName)
			filePath := path.Join(docDir, fileName)

			// 渲染完整 HTML 页面
			tmpl, err := template.New("docsite").Funcs(template.FuncMap{
				"sub": func(a, b int) int { return a - b },
				"div": func(a, b int) int {
					if b == 0 {
						return 0
					}
					return a / b
				},
			}).Parse(DocsiteHTMLTPL)
			if err != nil {
				return nil, fmt.Errorf("parse template: %w", err)
			}

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, map[string]interface{}{
				"SiteTitle":     VerifyFileName(title),
				"Title":         articleData.Info.Title,
				"Content":       template.HTML(htmlBuf.String()), // #nosec G203 - HTML content is from trusted source (converted from markdown)
				"MediaPlayer":   template.HTML(playTpl),          // #nosec G203 - Template is generated internally
				"Articles":      []ArticleInfo{},
				"CurrentIndex":  k + 1,
				"TotalArticles": len(tasks),
				"ArticleIndex":  k + 1,
				"IsIndex":       false,
				"PrevURL":       "",
				"NextURL":       "",
			})
			if err != nil {
				return nil, fmt.Errorf("execute template: %w", err)
			}

			// 使用 Storage 写入文件
			if _, err := global.Storage.Put(filePath, io.NopCloser(bytes.NewBuffer(buf.Bytes()))); err != nil {
				return nil, fmt.Errorf("write file via storage: %w", err)
			}

			global.LOG.Info("generated article html",
				zap.String("file", fileName),
				zap.String("title", articleData.Info.Title))

			// 保存 HTML 内容到内存
			articleHTMLsMu.Lock()
			articleHTMLs[fileName] = htmlBuf.String()
			articleHTMLsMu.Unlock()

			return &ArticleInfo{
				Index:    k + 1,
				Title:    articleData.Info.Title,
				Filename: fileName,
			}, nil
		})
	}

	// 等待所有任务完成
	ws, err := batch.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("batch wait: %w", err)
	}

	// 收集文章信息并记录失败的任务
	failedCount := 0
	for _, w := range ws {
		if w.Err != nil {
			failedCount++
			global.LOG.Warn("article generation failed",
				zap.Error(w.Err))
			continue
		}
		if val, ok := w.Value.(*ArticleInfo); ok {
			articles = append(articles, *val)
		}
	}

	global.LOG.Info("article generation summary",
		zap.String("taskId", taskId),
		zap.Int("total_tasks", len(tasks)),
		zap.Int("success", len(articles)),
		zap.Int("failed", failedCount))

	// 按索引排序
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Index < articles[j].Index
	})

	// 检查是否有文章
	if len(articles) == 0 {
		return "", fmt.Errorf("no articles generated for task %s", taskId)
	}

	// 重新生成所有文章，添加导航链接
	for i, article := range articles {
		filePath := path.Join(docDir, article.Filename)

		// 从内存获取 HTML 内容
		htmlContent, ok := articleHTMLs[article.Filename]
		if !ok {
			global.LOG.Warn("article HTML not found in memory",
				zap.String("file", article.Filename))
			continue
		}

		// 提取 body 内容（简单方式：直接使用原内容）
		prevURL := ""
		if i > 0 {
			prevURL = articles[i-1].Filename
		}
		nextURL := ""
		if i < len(articles)-1 {
			nextURL = articles[i+1].Filename
		}

		// 重新渲染带导航的页面
		tmpl, err := template.New("docsite").Funcs(template.FuncMap{
			"sub": func(a, b int) int { return a - b },
			"div": func(a, b int) int {
				if b == 0 {
					return 0
				}
				return a / b
			},
		}).Parse(DocsiteHTMLTPL)
		if err != nil {
			continue
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]interface{}{
			"SiteTitle":     VerifyFileName(title),
			"Title":         article.Title,
			"Content":       template.HTML(htmlContent), // #nosec G203 - HTML content is from trusted source (converted from markdown)
			"Articles":      articles,
			"CurrentIndex":  article.Index,
			"TotalArticles": len(articles),
			"ArticleIndex":  article.Index,
			"IsIndex":       false,
			"PrevURL":       prevURL,
			"NextURL":       nextURL,
		})
		if err != nil {
			continue
		}

		// 使用 Storage 写入文件
		if _, err := global.Storage.Put(filePath, io.NopCloser(bytes.NewBuffer(buf.Bytes()))); err != nil {
			global.LOG.Warn("write article file failed",
				zap.String("file", filePath),
				zap.Error(err))
		}
	}

	// 生成首页（介绍页）
	indexPath := path.Join(docDir, "index.html")
	tmpl, err := template.New("docsite-index").Funcs(template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}).Parse(DocsiteHTMLTPL)
	if err != nil {
		return "", fmt.Errorf("parse index template: %w", err)
	}

	var indexBuf bytes.Buffer
	err = tmpl.Execute(&indexBuf, map[string]interface{}{
		"SiteTitle":     VerifyFileName(title),
		"Title":         title,
		"Content":       template.HTML(introHTML), // #nosec G203 - HTML content is from trusted source (course introduction)
		"Articles":      articles,
		"CurrentIndex":  0,
		"TotalArticles": len(articles),
		"ArticleIndex":  0,
		"IsIndex":       true,
		"PrevURL":       "",
		"NextURL":       articles[0].Filename,
	})
	if err != nil {
		return "", fmt.Errorf("execute index template: %w", err)
	}

	// 使用 Storage 写入首页
	if _, err := global.Storage.Put(indexPath, io.NopCloser(bytes.NewBuffer(indexBuf.Bytes()))); err != nil {
		return "", fmt.Errorf("write index file via storage: %w", err)
	}

	global.LOG.Info("index page generated",
		zap.String("file", indexPath),
		zap.Int("articles", len(articles)))

	global.LOG.Info("docsite generation completed",
		zap.String("taskId", taskId),
		zap.Int("articles", len(articles)))

	return global.Storage.GetKey(docDir, false), nil
}

// MakeDocsiteLocal 生成本地文档站点（带评论）
func (g *GoldmarkDocGenerator) MakeDocsiteLocal(ctx context.Context, taskId, group, title, introHTML string, commentLen int) error {
	var tasks []model.Task
	if err := global.DB.Model(&model.Task{}).
		Where(&model.Task{TaskPid: taskId}).
		Order("id asc").
		Find(&tasks).Error; err != nil {
		return fmt.Errorf("query tasks failed: %w", err)
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no tasks found for task %s", taskId)
	}

	docDir := path.Join(group, VerifyFileName(title))
	var articles []ArticleInfo
	articleHTMLs := make(map[string]string) // 内存中保存 HTML 内容
	// 同 MakeDocsite：map 的写入发生在 GPool 并发 goroutine 中，必须加锁。
	var articleHTMLsMu sync.Mutex

	// 第一步：生成所有文章的 HTML（带评论）
	batch := global.GPool.NewBatch()
	for i := range tasks {
		x, k := tasks[i], i
		batch.Queue(func(_ context.Context) (any, error) {
			var articleData geek.ArticleData
			if err := json.Unmarshal(x.Raw, &articleData); err != nil {
				return nil, err
			}

			if len(articleData.Info.Title) == 0 {
				return nil, fmt.Errorf("title is empty %s", x.TaskId)
			}

			// 使用 Cshort 如果 Content 为空
			content := articleData.Info.Content
			if len(content) == 0 && len(articleData.Info.Cshort) > 0 {
				content = articleData.Info.Cshort
			}

			// URL 代理替换
			if rewrittenContent, err := HtmlURLProxyReplace(content); err == nil {
				content = rewrittenContent
			}

			// HTML 转 Markdown
			markdown, err := HTMLConvertMarkdown(content)
			if err != nil {
				return nil, fmt.Errorf("convert html to markdown: %w", err)
			}

			coverURL := ""
			if articleData.Info.Cover.Default != "" {
				coverURL = articleData.Info.Cover.Default
			}

			if len(x.Ciphertext) > 0 || len(x.RewriteHls) > 0 {
				hlsURL := fmt.Sprintf("%s/v2/task/play.m3u8?id=%s", strings.TrimSuffix(global.CONF.Storage.Host, "/"), x.TaskId)
				if coverURL != "" {
					playTpl := `<div style="width:100%%;max-width:100%%;aspect-ratio:16/9;background:#000;border-radius:8px;overflow:hidden;margin:1.5rem 0;"><video id="video" controls preload="none" playsinline poster="%s" style="width:100%%;height:100%%;display:block;"><source src="%s" type="application/x-mpegURL"></video></div><br/> %s`
					markdown = fmt.Sprintf(playTpl, coverURL, hlsURL, markdown)
				} else {
					playTpl := `<div style="width:100%%;max-width:100%%;aspect-ratio:16/9;background:#000;border-radius:8px;overflow:hidden;margin:1.5rem 0;"><video id="video" controls preload="none" playsinline style="width:100%%;height:100%%;display:block;"><source src="%s" type="application/x-mpegURL"></video></div><br/> %s`
					markdown = fmt.Sprintf(playTpl, hlsURL, markdown)
				}
			} else if articleData.Info.Audio.DownloadURL != "" {
				if coverURL != "" {
					playTpl := `<div style="width:100%%;max-width:100%%;margin:1.5rem 0;border-radius:8px;overflow:hidden;background:#f6f7f8;"><img src="%s" style="width:100%%;max-height:300px;object-fit:cover;display:block;"><div style="padding:1rem;"><audio id="audio" controls preload="none" style="width:100%%;display:block;"><source id="mp3" src="%s"></audio></div></div><br/> %s`
					markdown = fmt.Sprintf(playTpl, coverURL, articleData.Info.Audio.DownloadURL, markdown)
				} else {
					playTpl := `<div style="width:100%%;max-width:100%%;margin:1.5rem 0;padding:1rem;background:#f6f7f8;border-radius:8px;"><audio id="audio" controls preload="none" style="width:100%%;display:block;"><source id="mp3" src="%s"></audio></div><br/> %s`
					markdown = fmt.Sprintf(playTpl, articleData.Info.Audio.DownloadURL, markdown)
				}
			}

			// 获取评论（本地版本带评论）
			var aid int64
			if _, err := fmt.Sscanf(x.OtherId, "%d", &aid); err != nil {
				global.LOG.Warn("parse other id failed",
					zap.String("other_id", x.OtherId),
					zap.Error(err))
			}
			commentHTML := g.getCommentsHTML(ctx, aid, commentLen)
			if commentHTML != "" {
				label := "全部留言"
				if commentLen > 0 {
					label = "精选留言"
				}
				commentCount := strings.Count(commentHTML, "<li>")
				markdown += fmt.Sprintf("\n\n<div><strong>%s（%d）</strong></div>", label, commentCount)
				markdown += fmt.Sprintf("<ul>\n%s\n</ul>", commentHTML)
			}

			// Markdown 转 HTML
			var htmlBuf bytes.Buffer
			if err := g.md.Convert([]byte(markdown), &htmlBuf); err != nil {
				return nil, fmt.Errorf("convert markdown to html: %w", err)
			}

			// 生成文件名
			baseName := VerifyFileName(articleData.Info.Title)
			fileName := fmt.Sprintf("%d_%s.html", k+1, baseName)

			// 临时渲染（不包含导航，后续会重新渲染）
			// 保存 HTML 内容到内存
			articleHTMLsMu.Lock()
			articleHTMLs[fileName] = htmlBuf.String()
			articleHTMLsMu.Unlock()

			global.LOG.Info("generated article html with comments",
				zap.String("file", fileName),
				zap.String("title", articleData.Info.Title),
				zap.Int("comments", strings.Count(commentHTML, "<li>")))

			return &ArticleInfo{
				Index:    k + 1,
				Title:    articleData.Info.Title,
				Filename: fileName,
			}, nil
		})
	}

	// 等待所有任务完成
	ws, err := batch.Wait(ctx)
	if err != nil {
		return fmt.Errorf("batch wait: %w", err)
	}

	// 收集文章信息
	for _, w := range ws {
		if val, ok := w.Value.(*ArticleInfo); ok {
			articles = append(articles, *val)
		}
	}

	// 按索引排序
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Index < articles[j].Index
	})

	// 检查是否有文章
	if len(articles) == 0 {
		return fmt.Errorf("no articles generated for task %s", taskId)
	}

	// 第二步：重新生成所有文章，添加导航链接
	for i, article := range articles {
		filePath := path.Join(docDir, article.Filename)

		// 从内存获取 HTML 内容
		htmlContent, ok := articleHTMLs[article.Filename]
		if !ok {
			global.LOG.Warn("article HTML not found in memory",
				zap.String("file", article.Filename))
			continue
		}

		// 提取 body 内容（简单方式：直接使用原内容）
		prevURL := ""
		if i > 0 {
			prevURL = articles[i-1].Filename
		}
		nextURL := ""
		if i < len(articles)-1 {
			nextURL = articles[i+1].Filename
		}

		// 重新渲染带导航的页面
		tmpl, err := template.New("docsite").Funcs(template.FuncMap{
			"sub": func(a, b int) int { return a - b },
			"div": func(a, b int) int {
				if b == 0 {
					return 0
				}
				return a / b
			},
		}).Parse(DocsiteHTMLTPL)
		if err != nil {
			continue
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]interface{}{
			"SiteTitle":     VerifyFileName(title),
			"Title":         article.Title,
			"Content":       template.HTML(htmlContent), // #nosec G203 - HTML content is from trusted source (converted from markdown)
			"Articles":      articles,
			"CurrentIndex":  article.Index,
			"TotalArticles": len(articles),
			"ArticleIndex":  article.Index,
			"IsIndex":       false,
			"PrevURL":       prevURL,
			"NextURL":       nextURL,
		})
		if err != nil {
			continue
		}

		// 使用 Storage 写入文件
		if _, err := global.Storage.Put(filePath, io.NopCloser(bytes.NewBuffer(buf.Bytes()))); err != nil {
			global.LOG.Warn("write article file failed",
				zap.String("file", filePath),
				zap.Error(err))
		}
	}

	// 第三步：生成首页（介绍页）
	indexPath := path.Join(docDir, "index.html")
	tmpl, err := template.New("docsite-index").Funcs(template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}).Parse(DocsiteHTMLTPL)
	if err != nil {
		return fmt.Errorf("parse index template: %w", err)
	}

	var indexBuf bytes.Buffer
	err = tmpl.Execute(&indexBuf, map[string]interface{}{
		"SiteTitle":     VerifyFileName(title),
		"Title":         title,
		"Content":       template.HTML(introHTML), // #nosec G203 - HTML content is from trusted source (course introduction)
		"Articles":      articles,
		"CurrentIndex":  0,
		"TotalArticles": len(articles),
		"ArticleIndex":  0,
		"IsIndex":       true,
		"PrevURL":       "",
		"NextURL":       articles[0].Filename,
	})
	if err != nil {
		return fmt.Errorf("execute index template: %w", err)
	}

	// 使用 Storage 写入首页
	if _, err := global.Storage.Put(indexPath, io.NopCloser(bytes.NewBuffer(indexBuf.Bytes()))); err != nil {
		return fmt.Errorf("write index file via storage: %w", err)
	}

	global.LOG.Info("local docsite generation completed",
		zap.String("taskId", taskId),
		zap.String("group", group),
		zap.String("dir", docDir),
		zap.Int("articles", len(articles)))

	return nil
}

// MakeDocArchive 生成 Markdown 归档
func (g *GoldmarkDocGenerator) MakeDocArchive(ctx context.Context, taskId, _ string, introHTML string) (*bytes.Buffer, error) {
	var tasks []model.Task
	if err := global.DB.WithContext(ctx).Model(&model.Task{}).
		Where(&model.Task{TaskPid: taskId}).
		Order("id asc").
		Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("query tasks failed: %w", err)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks found for task %s", taskId)
	}

	// 创建 tar.gz 归档
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// 添加介绍文件
	introMarkdown, err := HTMLConvertMarkdown(introHTML)
	if err != nil {
		return nil, fmt.Errorf("convert intro to markdown: %w", err)
	}
	if err := tw.WriteHeader(&tar.Header{
		Name: "README.md",
		Mode: 0644,
		Size: int64(len(introMarkdown)),
	}); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(introMarkdown)); err != nil {
		return nil, err
	}

	// 添加每篇文章
	for i, task := range tasks {
		var articleData geek.ArticleData
		if err := json.Unmarshal(task.Raw, &articleData); err != nil {
			continue
		}

		content := articleData.Info.Content
		if len(content) == 0 && len(articleData.Info.Cshort) > 0 {
			content = articleData.Info.Cshort
		}

		markdown, err := HTMLConvertMarkdown(content)
		if err != nil {
			continue
		}

		fileName := fmt.Sprintf("%d_%s.md", i+1, VerifyFileName(articleData.Info.Title))
		if err := tw.WriteHeader(&tar.Header{
			Name: fileName,
			Mode: 0644,
			Size: int64(len(markdown)),
		}); err != nil {
			return nil, err
		}
		if _, err := tw.Write([]byte(markdown)); err != nil {
			return nil, err
		}
	}

	// 关闭 writers
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	global.LOG.Info("doc archive generated",
		zap.String("taskId", taskId),
		zap.Int("articles", len(tasks)),
		zap.Int("size", buf.Len()))

	return &buf, nil
}

// getCommentsHTML 获取评论 HTML
func (g *GoldmarkDocGenerator) getCommentsHTML(_ context.Context, aid int64, commentLen int) string {
	hasNext := true
	perPage := 20
	if commentLen > 0 {
		perPage = commentLen
	}
	page := 1
	commentHTML := ""
	count := 0

	commentHtmlFormat := `<li><img src="%s" width="30px" referrerpolicy="no-referrer"><span>%s</span> 👍（%d） 💬（%d）<p>%s</p>%s</li><br/>`
	commentSimpleHtmlFormat := `<li><span>%s</span> 👍（%d） 💬（%d）<p>%s</p>%s</li><br/>`

	for hasNext {
		var comments []*model.ArticleComment
		tx := global.DB.Model(&model.ArticleComment{})
		var commentCount int64
		if err := tx.Where("aid = ?", aid).
			Count(&commentCount).Offset((page - 1) * perPage).
			Limit(perPage + 1).Find(&comments).Error; err != nil {
			break
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
			row.UserHeader = URLProxyReplace(row.UserHeader)

			format := commentSimpleHtmlFormat
			if row.UserHeader != "" {
				format = commentHtmlFormat
				commentHTML += fmt.Sprintf(format, row.UserHeader,
					row.UserName, row.LikeCount, row.DiscussionCount, row.CommentContent,
					time.Unix(row.CommentCtime, 0).Format(time.DateOnly))
			} else {
				commentHTML += fmt.Sprintf(format,
					row.UserName, row.LikeCount, row.DiscussionCount, row.CommentContent,
					time.Unix(row.CommentCtime, 0).Format(time.DateOnly))
			}
		}
	}

	return commentHTML
}

type TaskMessage struct {
	Object string `json:"object"`
}
