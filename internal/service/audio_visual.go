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
	fileName := VerifyFileName(data.Info.Title)
	dir := path.Join(x.TaskPid, VerifyFileName(data.Product.Title))
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
	} else if data.Info.Audio.DownloadURL != "" {
		downloadURL = data.Info.Audio.DownloadURL
	}

	if len(downloadURL) > 0 {
		rewritePlayReq := PlayMetaRequest{
			DowloadURL: downloadURL,
			Dir:        dir,
			Filename:   fileName,
			TaskId:     x.TaskId,
			Spec:       x.RewriteHls,
		}
		if len(x.Ciphertext) > 0 {
			cipher, err1 := base64.StdEncoding.DecodeString(x.Ciphertext)
			if err1 != nil {
				global.LOG.Error("download rewritePlay", zap.Error(err1), zap.String("taskId", x.TaskId))
				return err1
			}
			rewritePlayReq.Ciphertext = cipher
		}
		meta, err1 := RewritePlay(ctx, rewritePlayReq)
		if err1 != nil {
			global.LOG.Error("download rewritePlay", zap.Error(err1), zap.String("taskId", x.TaskId))
			return err1
		}
		x.RewriteHls = meta.Spec
		x.Ciphertext = meta.Ciphertext
		if global.CONF.Site.Download {
			if data.Info.IsVideo {
				source, err = Video(ctx, dir, fileName, meta)
				if err != nil {
					global.LOG.Error("download video", zap.Error(err), zap.String("taskId", x.TaskId))
					return err
				}
			} else {
				source, err = Audio(ctx, x, downloadURL, dir, fileName)
				if err != nil {
					global.LOG.Error("download audio", zap.Error(err), zap.String("taskId", x.TaskId))
					return err
				}
			}
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
	}

	global.LOG.Info("download end", zap.String("taskId", x.TaskId),
		zap.String("url", downloadURL), zap.Duration("cost", time.Since(t0)),
	)
	return nil
}

func Video(ctx context.Context, dir, fileName string, req *PlayMeta) (string, error) {
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute*30)
	defer retryCancel()

	if len(req.Parts) == 0 {
		return "", fmt.Errorf("[%s] The parameter is incorrect", fileName)
	}

	if len(req.KeyPath) > 0 {
		cipher, err1 := base64.StdEncoding.DecodeString(req.Ciphertext)
		if err1 != nil {
			global.LOG.Error("video cipher", zap.Error(err1), zap.String("fileName", fileName))
			return "", err1
		}
		if err := os.MkdirAll(path.Dir(req.KeyPath), os.ModePerm); err != nil {
			return "", err
		}
		if err := os.WriteFile(req.KeyPath, cipher, os.ModePerm); err != nil {
			return "", err
		}
		global.LOG.Info("video cipher key", zap.String("KeyPath", req.KeyPath))
	}

	m3u8Path := path.Join(dir, fileName, "index.m3u8")
	stat, err := global.Storage.Put(m3u8Path, io.NopCloser(bytes.NewBuffer(req.LocalSpec)))
	if err != nil {
		return "", err
	} else if stat.Size() <= 0 {
		return "", fmt.Errorf("[%s] size is zero", m3u8Path)
	}
	destKey := global.Storage.GetKey(m3u8Path, true)
	destDir := path.Dir(destKey)
	concatPath := path.Join(path.Dir(destDir), fmt.Sprintf("%s.mp4", fileName))
	if s, _ := os.Stat(concatPath); s != nil && s.Size() > 0 {
		_ = os.RemoveAll(destDir)
		return concatPath, nil
	}

	batch := global.GPool.NewBatch()
	for _, t := range req.Parts {
		partURL, destName := t.Src, t.Dest
		if partStat, _ := global.Storage.Stat(destName); partStat != nil && partStat.Size() > 0 {
			continue
		}
		global.LOG.Info("video part start", zap.String("part", path.Base(destName)))
		batch.Queue(func(pctx context.Context) (any, error) {
			err = zhttp.R.
				Before(func(r *http.Request) {
					r.Header.Set("Accept", "application/json, text/plain, */*")
					r.Header.Set("User-Agent", zhttp.RandomUserAgent())
				}).
				After(func(r *http.Response) error {
					if zhttp.IsHTTPSuccessStatus(r.StatusCode) {
						if partStat, err1 := global.Storage.Put(destName, r.Body); err1 != nil {
							return err1
						} else if partStat.Size() <= 0 {
							return fmt.Errorf("[%s] is empty", destName)
						}
						global.LOG.Info("video part end", zap.String("part", path.Base(destName)))
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
				DoWithRetry(pctx, http.MethodGet, partURL, nil)
			if err != nil {
				return "", err
			}
			return nil, nil
		})
	}

	if _, err = batch.Wait(ctx); err != nil {
		return "", err
	}

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
	_ = os.RemoveAll(destDir)
	global.LOG.Info("video", zap.String("removeAll", destDir))
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

type PlayMeta struct {
	Spec       []byte
	LocalSpec  []byte
	KeyPath    string
	Ciphertext string
	Parts      []Part
}

type PlayMetaRequest struct {
	DowloadURL string
	Dir        string
	Filename   string
	TaskId     string
	Ciphertext []byte
	Spec       []byte
}

func RewritePlay(ctx context.Context, req PlayMetaRequest) (*PlayMeta, error) {
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute*30)
	defer retryCancel()
	if len(req.Spec) == 0 {
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
					req.Spec = raw
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
			DoWithRetry(retryCtx, http.MethodGet, req.DowloadURL, nil)
		if err != nil {
			return nil, err
		}
	}
	if len(req.Spec) == 0 {
		return nil, fmt.Errorf("[%s] hls file is zero", req.DowloadURL)
	}
	meta := PlayMeta{
		Spec:      make([]byte, 0, len(req.Spec)),
		LocalSpec: make([]byte, 0, len(req.Spec)),
		Parts:     make([]Part, 0, 10),
	}
	bio := bufio.NewReader(bytes.NewReader(req.Spec))
	for {
		line, _, err1 := bio.ReadLine()
		if err1 != nil {
			break
		}
		l, rl := string(line), string(line)
		if strings.HasPrefix(l, "#EXT-X-KEY:") {
			sps := strings.Split(l, `"`)
			token, _, er := global.JWT.TokenGenerator(func(claims jwt.MapClaims) {
				claims["task_id"] = req.TaskId
			})
			if er != nil {
				return nil, er
			}
			destName := path.Join(req.Dir, req.Filename, "key.key")
			meta.KeyPath = global.Storage.GetKey(destName, true)
			l = fmt.Sprintf(`%s"file:///%s"`, sps[0], meta.KeyPath)
			rl = fmt.Sprintf(`%s"{host}/v2/task/kms?Ciphertext=%s"`, sps[0], token)
			if len(req.Ciphertext) == 0 {
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
							req.Ciphertext = raw
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
					return nil, err
				}
			}
		} else if strings.HasSuffix(l, ".ts") {
			playHost := req.DowloadURL[:strings.LastIndex(req.DowloadURL, "/")+1]
			if !strings.HasPrefix(l, "https://") {
				meta.Parts = append(meta.Parts, Part{playHost + l, path.Join(req.Dir, l), false})
				l = playHost + l
				rl = l
			} else {
				tsPath := strings.TrimPrefix(l, playHost)
				if strings.HasPrefix(tsPath, playHost) {
					tsPath = strings.TrimPrefix(tsPath, playHost)
					l = strings.TrimPrefix(l, playHost)
				}
				destName := path.Join(req.Dir, req.Filename, tsPath)
				meta.Parts = append(meta.Parts, Part{l, destName, false})
				rl = l
			}
		}
		meta.Spec = append(meta.Spec, []byte(rl+"\n")...)
		meta.LocalSpec = append(meta.LocalSpec, []byte(l+"\n")...)
	}
	if len(req.Ciphertext) > 0 {
		meta.Ciphertext = base64.StdEncoding.EncodeToString(req.Ciphertext)
	}
	return &meta, nil
}
