package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/internal/types/task"
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
	Src   string
	Dest  string
	IsKey bool
}

func Download(ctx context.Context, x *model.Task) error {
	t0 := time.Now()
	var data geek.ArticleData
	if err := json.Unmarshal(x.Raw, &data); err != nil {
		return err
	}
	var (
		source      string
		downloadURL string
		err         error
	)
	if data.Info.IsVideo {
		if len(data.Info.Video.HlsMedias) == 0 && len(data.Info.VideoPreview.Medias) > 0 {
			data.Info.Video.HlsMedias = data.Info.VideoPreview.Medias
		}
		if len(data.Info.Video.HlsMedias) == 0 && len(data.Info.VideoPreview.Medias) == 0 {
			return fmt.Errorf("article info not found %s", x.OtherId)
		}
		sort.Slice(data.Info.Video.HlsMedias, func(i, j int) bool {
			return data.Info.Video.HlsMedias[i].Size > data.Info.Video.HlsMedias[j].Size
		})
		downloadURL = data.Info.Video.HlsMedias[0].URL
		if global.CONF.Site.Download {
			fileName := VerifyFileName(data.Info.Title)
			dir := path.Join(x.TaskPid, VerifyFileName(data.Product.Title), fileName)
			source, err = Video(ctx, x, downloadURL, dir, fileName)
			if err != nil {
				global.LOG.Error("download video", zap.Error(err), zap.String("taskId", x.TaskId))
				return err
			}
		} else if len(x.RewriteHls) == 0 {
			ciphertext, rewriteHls, err1 := RewritePlay(ctx, downloadURL, x.TaskId)
			if err1 != nil {
				global.LOG.Error("download rewritePlay", zap.Error(err1), zap.String("taskId", x.TaskId))
				return err1
			}
			x.RewriteHls = rewriteHls
			x.Ciphertext = ciphertext
		}
	} else if data.Info.Audio.DownloadURL != "" {
		downloadURL = data.Info.Audio.DownloadURL
		if global.CONF.Site.Download {
			dir := path.Join(x.TaskPid, VerifyFileName(data.Product.Title))
			source, err = Audio(ctx, x, downloadURL, dir, VerifyFileName(data.Info.Title))
			if err != nil {
				global.LOG.Error("download audio", zap.Error(err), zap.String("taskId", x.TaskId))
				return err
			}
		}
		if len(x.RewriteHls) == 0 {
			ciphertext, rewriteHls, err1 := RewritePlay(ctx, data.Info.Audio.URL, x.TaskId)
			if err1 != nil {
				global.LOG.Error("download rewritePlay", zap.Error(err1), zap.String("taskId", x.TaskId))
				return err1
			}
			x.RewriteHls = rewriteHls
			x.Ciphertext = ciphertext
		}
	}

	if global.CONF.Site.Download {
		message := task.TaskMessage{}
		if len(source) > 0 {
			message.Object = global.Storage.GetKey(source, false)
		} else {
			message.Text = "not found download url"
		}
		x.Message, _ = json.Marshal(message)
		if len(data.Info.Content) > 0 {
			converter := md.NewConverter("", true, nil)
			if markdown, err := converter.ConvertString(data.Info.Content); err == nil {
				realFile := global.Storage.GetKey(source, true)
				realFile = strings.TrimSuffix(realFile, ".mp4")
				realFile = strings.TrimSuffix(realFile, ".mp3")
				mdPath := fmt.Sprintf("%s.md", realFile)
				_ = os.WriteFile(mdPath, []byte(markdown), os.ModePerm)
			}
		}
	}

	global.LOG.Info("download end", zap.String("taskId", x.TaskId),
		zap.String("url", downloadURL), zap.Duration("cost", time.Since(t0)),
	)
	return nil
}

func Video(ctx context.Context, x *model.Task, hlsURL, dir, fileName string) (string, error) {
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute*30)
	defer retryCancel()

	hlsRaw := x.RewriteHls
	if len(hlsRaw) == 0 {
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
					return fmt.Errorf("http status: %s, %s", r.Status, r.Request.URL.String())
				}
				return zhttp.BreakRetryError(fmt.Errorf(
					"break http status: %s,%s", r.Status, r.Request.URL.String()))
			}).
			DoWithRetry(retryCtx, http.MethodGet, hlsURL, nil)
		if err != nil {
			return "", err
		}
		if len(hlsRaw) == 0 {
			return "", fmt.Errorf("[%s] hls file is zero", hlsURL)
		}
	}

	var (
		buff        bytes.Buffer
		rewriteBuff bytes.Buffer
		ciphertext  []byte
	)
	bio := bufio.NewReader(bytes.NewReader(hlsRaw))
	ts := make([]Part, 0, 10)
	for {
		line, _, err1 := bio.ReadLine()
		if err1 != nil {
			break
		}
		l, rl := string(line), string(line)
		srcURL, destName, isKey := "", "", false
		if strings.HasPrefix(l, "#EXT-X-KEY:") {
			isKey = true
			sps := strings.Split(l, `"`)
			srcURL = sps[1]
			destName = path.Join(dir, "key.key")
			realDestPath := global.Storage.GetKey(destName, true)
			l = fmt.Sprintf(`%s"file:///%s"`, sps[0], realDestPath)
			token, _, er := global.JWT.TokenGenerator(func(claims jwt.MapClaims) {
				claims["task_id"] = x.TaskId
			})
			if er != nil {
				return "", er
			}
			rl = fmt.Sprintf(`%s"{host}/v2/task/kms?Ciphertext=%s"`, sps[0], token)
			if len(x.Ciphertext) > 0 {
				cipher, err := base64.StdEncoding.DecodeString(x.Ciphertext)
				if err != nil {
					return "", err
				}
				if err = os.WriteFile(realDestPath, cipher, os.ModePerm); err != nil {
					return "", err
				}
				srcURL, destName = "", ""
			}
		} else if strings.HasSuffix(l, ".ts") {
			playHost := hlsURL[:strings.LastIndex(hlsURL, "/")+1]
			if !strings.HasPrefix(l, "https://") {
				srcURL = playHost + l
				destName = path.Join(dir, l)
			} else {
				srcURL = l
				destName = path.Join(dir, strings.TrimPrefix(l, playHost))
			}
			rl = srcURL
		}
		rewriteBuff.WriteString(rl + "\n")
		buff.WriteString(l + "\n")
		if len(srcURL) == 0 || len(destName) == 0 {
			continue
		}
		ts = append(ts, Part{srcURL, destName, isKey})
	}

	if len(ts) == 0 {
		return "", fmt.Errorf("[%s] part is zero", hlsURL)
	}
	for _, t := range ts {
		dowloadURL, destName, isKey := t.Src, t.Dest, t.IsKey
		err := zhttp.R.
			Before(func(r *http.Request) {
				r.Header.Set("Accept", "application/json, text/plain, */*")
				r.Header.Set("User-Agent", zhttp.RandomUserAgent())
			}).
			After(func(r *http.Response) error {
				if zhttp.IsHTTPSuccessStatus(r.StatusCode) {
					if isKey {
						raw, err := io.ReadAll(r.Body)
						if err != nil {
							return err
						}
						ciphertext = raw
						r.Body = io.NopCloser(bytes.NewBuffer(raw))
					}
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
					return fmt.Errorf("http status: %s, %s", r.Status, r.Request.URL.String())
				}
				return zhttp.BreakRetryError(fmt.Errorf(
					"break http status: %s,%s", r.Status, r.Request.URL.String()))
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
	x.Ciphertext = base64.StdEncoding.EncodeToString(ciphertext)
	x.RewriteHls = rewriteBuff.Bytes()
	return concatPath, nil
}

func Audio(ctx context.Context, _ *model.Task, dowloadURL, dir, fileName string) (string, error) {
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

func RewritePlay(ctx context.Context, hlsURL, taskId string) (string, []byte, error) {
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
		return "", nil, err
	}
	if len(hlsRaw) == 0 {
		return "", nil, fmt.Errorf("[%s] hls file is zero", hlsURL)
	}
	var (
		buff       bytes.Buffer
		ciphertext []byte
	)
	bio := bufio.NewReader(bytes.NewReader(hlsRaw))
	for {
		line, _, err1 := bio.ReadLine()
		if err1 != nil {
			break
		}
		l := string(line)
		if strings.HasPrefix(l, "#EXT-X-KEY:") {
			sps := strings.Split(l, `"`)
			token, _, er := global.JWT.TokenGenerator(func(claims jwt.MapClaims) {
				claims["task_id"] = taskId
			})
			if er != nil {
				return "", nil, er
			}
			l = fmt.Sprintf(`%s"{host}/v2/task/kms?Ciphertext=%s"`, sps[0], token)
			err = zhttp.R.
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
						ciphertext = raw
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
				DoWithRetry(retryCtx, http.MethodGet, sps[1], nil)
			if err != nil {
				return "", nil, err
			}
		} else if strings.HasSuffix(l, ".ts") {
			l = hlsURL[:strings.LastIndex(hlsURL, "/")+1] + l
		}
		buff.WriteString(l + "\n")
	}
	cipher := ""
	if len(ciphertext) > 0 {
		cipher = base64.StdEncoding.EncodeToString(ciphertext)
	}
	return cipher, buff.Bytes(), nil
}
