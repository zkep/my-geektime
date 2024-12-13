package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/zkep/mygeektime/internal/types/task"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/lib/zhttp"
	"go.uber.org/zap"
)

var (
	SpecialCharacters = map[string]string{
		"|": "_",
		"?": "_",
		"？": "_",
		"：": "_",
		"/": "_",
		"*": "_",
		" ": "",
	}
)

func VerifyFileName(name string) string {
	for src, dest := range SpecialCharacters {
		name = strings.ReplaceAll(name, src, dest)
	}
	return name
}

type Part struct {
	Src  string
	Dest string
}

func Download(ctx context.Context, x *model.Task) error {
	t0 := time.Now()
	var articleInfo geek.ArticleInfoResponse
	if err := json.Unmarshal(x.Raw, &articleInfo); err != nil {
		return err
	}
	data := articleInfo.Data
	if data.Info.IsVideo {
		global.LOG.Info("download video start",
			zap.String("taskId", x.TaskId),
			zap.String("otherId", x.OtherId))
		aid, err := strconv.ParseInt(x.OtherId, 10, 64)
		if err != nil {
			return err
		}
		article, err := GetArticleInfo(ctx, geek.ArticlesInfoRequest{Id: aid})
		if err != nil {
			return err
		}
		if len(article.Data.Info.Video.HlsMedias) == 0 {
			return fmt.Errorf("article info not found %d", aid)
		}
		data = article.Data

		sort.Slice(data.Info.Video.HlsMedias, func(i, j int) bool {
			return data.Info.Video.HlsMedias[i].Size > data.Info.Video.HlsMedias[j].Size
		})

		hlsURL := data.Info.Video.HlsMedias[0].URL
		fileName := VerifyFileName(data.Info.Title)
		dir := path.Join(VerifyFileName(data.Product.Title), fileName)
		source, err := Video(ctx, hlsURL, dir, fileName)
		if err != nil {
			global.LOG.Error("download video", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
		message := task.TaskMessage{
			Object: global.Storage.GetKey(source, false),
		}
		x.Message, _ = json.Marshal(message)
		global.LOG.Info("download video end",
			zap.String("taskId", x.TaskId),
			zap.String("url", hlsURL),
			zap.Duration("cost", time.Since(t0)),
		)
	} else if data.Info.Audio.DownloadURL != "" {
		global.LOG.Info("download audio start",
			zap.String("taskId", x.TaskId),
			zap.String("url", data.Info.Audio.DownloadURL))
		source, err := Audio(ctx, data.Info.Audio.DownloadURL,
			VerifyFileName(data.Product.Title), VerifyFileName(data.Info.Title))
		if err != nil {
			global.LOG.Error("download audio", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
		message := task.TaskMessage{
			Object: global.Storage.GetKey(source, false),
		}
		x.Message, _ = json.Marshal(message)
		global.LOG.Info("download audio end",
			zap.String("taskId", x.TaskId),
			zap.String("url", data.Info.Audio.DownloadURL),
			zap.Duration("cost", time.Since(t0)),
		)
	}
	return nil
}

func Video(ctx context.Context, hlsURL, dir, fileName string) (string, error) {
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute*30)
	defer retryCancel()

	var hlsRaw []byte
	err := zhttp.R.
		Before(func(r *http.Request) {
			r.Header.Set("Accept", "application/json, text/plain, */*")
			r.Header.Set("User-Agent", zhttp.RandomUserAgent())
			r.Header.Set("Referer", r.URL.String())
			r.Header.Set("Origin", "https://time.geekbang.org")
		}).
		After(func(r *http.Response) error {
			if zhttp.IsHTTPSuccessStatus(r.StatusCode) {
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					return err
				}
				hlsRaw = raw
				return nil
			}
			if zhttp.IsHTTPStatusSleep(r.StatusCode) {
				time.Sleep(time.Second * 10)
			}
			if zhttp.IsHTTPStatusRetryable(r.StatusCode) {
				return fmt.Errorf("http status: %s", r.Status)
			}
			return zhttp.BreakRetryError(fmt.Errorf("break http status: %s", r.Status))
		}).
		DoWithRetry(retryCtx, http.MethodGet, hlsURL, nil)
	if err != nil {
		return "", err
	}
	if len(hlsRaw) == 0 {
		return "", fmt.Errorf("[%s] hls file is zero", hlsURL)
	}
	var buff bytes.Buffer
	bio := bufio.NewReader(bytes.NewReader(hlsRaw))
	ts := make([]Part, 0, 10)
	for {
		line, _, err1 := bio.ReadLine()
		if err1 != nil {
			break
		}
		l := string(line)
		srcURL, destName := "", ""
		if strings.HasPrefix(l, "#EXT-X-KEY:") {
			sps := strings.Split(l, `"`)
			srcURL = sps[1]
			destName = path.Join(dir, "key.key")
			l = fmt.Sprintf(`%s"file:///%s"`, sps[0], global.Storage.GetKey(destName, true))
		} else if strings.HasSuffix(l, ".ts") {
			srcURL = hlsURL[:strings.LastIndex(hlsURL, "/")+1] + l
			destName = path.Join(dir, l)
		}
		buff.WriteString(l + "\n")
		if len(srcURL) == 0 || len(destName) == 0 {
			continue
		}
		ts = append(ts, Part{srcURL, destName})
	}

	if len(ts) == 0 {
		return "", fmt.Errorf("[%s] part is zero", hlsURL)
	}
	for _, t := range ts {
		dowloadURL, destName := t.Src, t.Dest
		err = zhttp.R.
			Before(func(r *http.Request) {
				r.Header.Set("Accept", "application/json, text/plain, */*")
				r.Header.Set("User-Agent", zhttp.RandomUserAgent())
			}).
			After(func(r *http.Response) error {
				if zhttp.IsHTTPSuccessStatus(r.StatusCode) {
					if stat, err := global.Storage.Put(destName, r.Body); err != nil {
						return err
					} else if stat.Size() <= 0 {
						return fmt.Errorf("[%s] is empty", destName)
					}
					return nil
				}
				if zhttp.IsHTTPStatusSleep(r.StatusCode) {
					time.Sleep(time.Second * 10)
				}
				if zhttp.IsHTTPStatusRetryable(r.StatusCode) {
					return fmt.Errorf("http status: %s", r.Status)
				}
				return zhttp.BreakRetryError(fmt.Errorf("break http status: %s", r.Status))
			}).
			DoWithRetry(retryCtx, http.MethodGet, dowloadURL, nil)
		if err != nil {
			return "", err
		}
	}
	m3u8Path := path.Join(dir, "index.m3u8")
	stat, err := global.Storage.Put(m3u8Path, io.NopCloser(&buff))
	if err != nil {
		return "", err
	} else if stat.Size() <= 0 {
		return "", fmt.Errorf("[%s] size is zero", m3u8Path)
	}
	destKey := global.Storage.GetKey(m3u8Path, true)
	destDir := path.Dir(destKey)
	concatPath := path.Join(path.Dir(destDir), fmt.Sprintf("%s.mp4", fileName))
	ffmpeg_command := []string{
		"-allowed_extensions",
		"ALL",
		"-protocol_whitelist",
		"concat,file,http,https,tcp,tls,crypto",
		"-i",
		path.Join(destDir, "index.m3u8"),
		"-hls_key_info_file",
		path.Join(destDir, "key.key"),
		"-c",
		"copy",
		concatPath,
	}
	global.LOG.Info("video", zap.String("concatPath", concatPath))
	output, err := exec.CommandContext(retryCtx, "ffmpeg", ffmpeg_command...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s,%s", err.Error(), string(output))
	}
	if s, _ := os.Stat(concatPath); s != nil && s.Size() > 0 {
		_ = os.RemoveAll(destDir)
	}
	return concatPath, nil
}

func Audio(ctx context.Context, dowloadURL, dir, fileName string) (string, error) {
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute*5)
	defer retryCancel()
	destName := path.Join(dir, fmt.Sprintf("%s.mp3", fileName))
	err := zhttp.R.
		Before(func(r *http.Request) {
			r.Header.Set("Accept", "application/json, text/plain, */*")
			r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
			r.Header.Set("User-Agent", zhttp.RandomUserAgent())
			r.Header.Set("Referer", r.URL.String())
			r.Header.Set("Origin", "https://time.geekbang.org")
		}).
		After(func(r *http.Response) error {
			if zhttp.IsHTTPSuccessStatus(r.StatusCode) {
				stat, err := global.Storage.Put(destName, r.Body)
				if err != nil {
					return err
				} else if stat.Size() <= 0 {
					return fmt.Errorf("audio empty: [%s]", destName)
				}
				destName = global.Storage.GetKey(destName, true)
				return nil
			}
			if zhttp.IsHTTPStatusSleep(r.StatusCode) {
				time.Sleep(time.Second * 10)
			}
			if zhttp.IsHTTPStatusRetryable(r.StatusCode) {
				return fmt.Errorf("http status: %s", r.Status)
			}
			return zhttp.BreakRetryError(fmt.Errorf("http status: %s", r.Status))
		}).
		DoWithRetry(retryCtx, http.MethodGet, dowloadURL, nil)
	if err != nil {
		return "", err
	}
	return destName, nil
}
