package task

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/lib/zhttp"
	"go.uber.org/zap"
)

var (
	keyLock = "_key_lock_"
	lock    = &sync.Map{}
)

func TaskHandler(t time.Time) error {
	global.LOG.Debug("task handler Start", zap.Time("time", t))
	_, loaded := lock.LoadOrStore(keyLock, t)
	if loaded {
		global.LOG.Debug("task handler running", zap.Time("time", t))
		return nil
	}
	defer lock.Delete(keyLock)

	hasMore, page, psize := true, 1, 5
	ctx := context.Background()
	for hasMore {
		var ls []*model.Task
		t1 := time.Now().AddDate(0, 0, -1).Unix()
		if err := global.DB.Model(&model.Task{}).
			Where("created_at >= ?", t1).
			Where("status = ?", service.TASK_STATUS_PENDING).
			Order("id ASC").
			Offset((page - 1) * psize).
			Limit(psize + 1).
			Find(&ls).Error; err != nil {
			global.LOG.Error("task handler find", zap.Error(err))
			return err
		}
		if len(ls) <= psize {
			hasMore = false
		} else {
			ls = ls[:psize]
		}
		page++
		for idx := range ls {
			x := ls[idx]
			err := worker(ctx, x)
			if err != nil {
				global.LOG.Error("task handler worker", zap.Error(err), zap.String("taskId", x.TaskId))
			}
		}
	}

	global.LOG.Debug("task handler End", zap.Time("time", time.Now()))
	return nil
}

func worker(ctx context.Context, x *model.Task) error {
	switch x.TaskType {
	case service.TASK_TYPE_PRODUCT:
		global.LOG.Debug("task worker", zap.String("type", x.TaskType))
		var count int64
		if err := global.DB.Model(&model.Task{}).
			Where("task_pid = ?", x.TaskId).
			Where("status <= ?", service.TASK_STATUS_RUNNING).
			Count(&count).Error; err != nil {
			global.LOG.Error("task handler Count",
				zap.Error(err),
				zap.String("taskId", x.TaskId),
			)
			return err
		}
		if count > 0 {
			global.LOG.Info("task worker sub task",
				zap.Int64("pending", count),
				zap.String("taskId", x.TaskId),
			)
			return nil
		}
		// all subtask do with
		m := map[string]any{
			"status":     service.TASK_STATUS_FINISHED,
			"updated_at": time.Now().Unix(),
		}
		if err := global.DB.Model(&model.Task{Id: x.Id}).UpdateColumns(m).Error; err != nil {
			global.LOG.Error("task worker UpdateColumns",
				zap.Error(err),
				zap.String("taskId", x.TaskId),
			)
			return err
		}
	case service.TASK_TYPE_ARTICLE:
		m := map[string]any{
			"status":     service.TASK_STATUS_RUNNING,
			"updated_at": time.Now().Unix(),
		}
		if err := global.DB.Model(&model.Task{Id: x.Id}).UpdateColumns(m).Error; err != nil {
			global.LOG.Error("task worker UpdateColumns",
				zap.Error(err),
				zap.String("taskId", x.TaskId),
			)
			return err
		}
		err := download(ctx, x)
		if err != nil {
			global.LOG.Error("task worker download",
				zap.Error(err), zap.String("taskId", x.TaskId))
		}
		message := bytes.NewBuffer(nil)
		status := service.TASK_STATUS_FINISHED
		if err != nil {
			status = service.TASK_STATUS_ERROR
			message.WriteString(err.Error())
		}
		m = map[string]any{
			"status":     status,
			"updated_at": time.Now().Unix(),
			"message":    message.Bytes(),
		}
		err = global.DB.Model(&model.Task{Id: x.Id}).UpdateColumns(m).Error
		if err != nil {
			global.LOG.Error("task worker UpdateColumns",
				zap.Error(err), zap.String("taskId", x.TaskId),
			)
			return err
		}
	}
	return nil
}

func download(ctx context.Context, x *model.Task) error {
	var data geek.ArticleData
	if err := json.Unmarshal(x.Raw, &data); err != nil {
		return err
	}
	if data.Info.IsVideo {
		if err := video(ctx, data); err != nil {
			global.LOG.Error("download video", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
	} else if data.Info.Audio.DownloadURL != "" {
		if err := audio(ctx, data); err != nil {
			global.LOG.Error("download audio", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
	}
	return nil
}

func video(ctx context.Context, data geek.ArticleData) error {
	sort.Slice(data.Info.Video.HlsMedias, func(i, j int) bool {
		return data.Info.Video.HlsMedias[i].Size > data.Info.Video.HlsMedias[j].Size
	})
	hlsURL := data.Info.Video.HlsMedias[0].URL
	fileName := getVerifyFileName(data.Info.Title)
	vodDir := path.Join(getVerifyFileName(data.Product.Title), fileName)
	return concat(ctx, hlsURL, vodDir, fileName)
}

func audio(ctx context.Context, data geek.ArticleData) error {
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute*30)
	defer retryCancel()
	dowloadURL := data.Info.Audio.DownloadURL
	fileName := getVerifyFileName(data.Info.Title)
	dstName := path.Join(getVerifyFileName(data.Product.Title), fmt.Sprintf("%s.mp3", fileName))
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
				fmt.Println(dstName, data.Info.Audio.DownloadURL)
				if _, err := global.Storage.Put(dstName, r.Body); err != nil {
					return err
				}
				if stat, _ := global.Storage.Stat(dstName); stat != nil && stat.Size() <= 0 {
					return fmt.Errorf("audio empty: %s", dstName)
				}
				return nil
			}
			if r.StatusCode == http.StatusTooManyRequests ||
				r.StatusCode == http.StatusUnavailableForLegalReasons {
				time.Sleep(time.Second * 10)
			}
			if zhttp.IsHTTPStatusRetryable(r.StatusCode) {
				return fmt.Errorf("http status: %s", r.Status)
			}
			return zhttp.BreakRetryError(fmt.Errorf("http status: %s", r.Status))
		}).
		DoWithRetry(retryCtx, http.MethodGet, dowloadURL, nil)
	if err != nil {
		global.LOG.Error("audio dowload error",
			zap.Error(err),
			zap.String("dowloadURL", dowloadURL),
		)
		return err
	}
	return nil
}

func getVerifyFileName(name string) string {
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "|", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "？", "_")
	name = strings.ReplaceAll(name, "：", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "*", "_")
	return name
}

func concat(ctx context.Context, hlsURL, vodDir, fileName string) error {
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute*30)
	defer retryCancel()
	var hlsRaw []byte
	err := zhttp.R.
		Client(global.HttpClient).
		Before(func(r *http.Request) {
			r.Header.Set("Accept", "application/json, text/plain, */*")
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
			r.Header.Set("Cookie", global.GeekCookies)
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
			if zhttp.IsHTTPStatusRetryable(r.StatusCode) {
				return fmt.Errorf("http status: %s", r.Status)
			}
			return zhttp.BreakRetryError(fmt.Errorf("http status: %s", r.Status))
		}).
		DoWithRetry(retryCtx, http.MethodGet, hlsURL, nil)
	if err != nil {
		global.LOG.Error("concat Error", zap.Error(err))
		return err
	}
	var buff bytes.Buffer
	bio := bufio.NewReader(bytes.NewReader(hlsRaw))
	for {
		line, _, err1 := bio.ReadLine()
		if err1 != nil {
			if !errors.Is(err1, io.EOF) {
				global.LOG.Error("concat ReadLine", zap.Error(err1))
			}
			break
		}
		l := string(line)
		dowloadURL, dstName := "", ""
		if strings.HasPrefix(l, "#EXT-X-KEY:") {
			sps := strings.Split(l, `"`)
			dowloadURL = sps[1]
			dstName = path.Join(vodDir, "key.key")
			l = fmt.Sprintf(`%s"file:///%s/key.key"`, sps[0], global.Storage.GetKey(vodDir))
		} else if strings.HasSuffix(l, ".ts") {
			dowloadURL = hlsURL[:strings.LastIndex(hlsURL, "/")+1] + l
			dstName = path.Join(vodDir, l)
		}
		buff.WriteString(l + "\n")

		if len(dowloadURL) == 0 || len(dstName) == 0 {
			continue
		}

		if stat, _ := global.Storage.Stat(dstName); stat != nil && stat.Size() > 0 {
			continue
		}
		err = zhttp.R.
			Before(func(r *http.Request) {
				r.Header.Set("Accept", "application/json, text/plain, */*")
				r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
				r.Header.Set("User-Agent", zhttp.RandomUserAgent())
				r.Header.Set("Referer", r.URL.String())
				r.Header.Set("Origin", "https://time.geekbang.org")
			}).
			After(func(r *http.Response) error {
				if zhttp.IsHTTPSuccessStatus(r.StatusCode) {
					if _, err = global.Storage.Put(dstName, r.Body); err != nil {
						return err
					}
					if stat, _ := global.Storage.Stat(dstName); stat != nil && stat.Size() <= 0 {
						return fmt.Errorf("concat empty: %s", dstName)
					}
					return nil
				}
				if r.StatusCode == http.StatusTooManyRequests ||
					r.StatusCode == http.StatusUnavailableForLegalReasons {
					time.Sleep(time.Second * 10)
				}
				if zhttp.IsHTTPStatusRetryable(r.StatusCode) {
					return fmt.Errorf("http status: %s", r.Status)
				}
				return zhttp.BreakRetryError(fmt.Errorf("http status: %s", r.Status))
			}).
			DoWithRetry(retryCtx, http.MethodGet, dowloadURL, nil)
		if err != nil {
			global.LOG.Error("concat dowload error",
				zap.Error(err),
				zap.String("dowloadURL", dowloadURL),
			)
			return err
		}
		global.LOG.Info("concat process", zap.String("dowloadURL", dowloadURL))
	}

	m3u8Path := path.Join(vodDir, "index.m3u8")
	destKey, err := global.Storage.Put(m3u8Path, io.NopCloser(&buff))
	if err != nil {
		global.LOG.Error("concat WriteFile Error", zap.Error(err))
		return err
	}
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
	global.LOG.Info("concat", zap.String("cmd", "ffmpeg "+strings.Join(ffmpeg_command, " ")))
	output, err := exec.CommandContext(retryCtx, "ffmpeg", ffmpeg_command...).CombinedOutput()
	if err != nil {
		global.LOG.Error("concat cmd", zap.Error(err), zap.String("output", string(output)))
		return fmt.Errorf("%s: %s", err.Error(), string(output))
	}
	if s, _ := os.Stat(concatPath); s != nil && s.Size() > 0 {
		_ = os.RemoveAll(destDir)
	}
	return nil
}
