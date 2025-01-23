package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/lib/zhttp"
	"go.uber.org/zap"
)

type PlayMeta struct {
	Spec         []byte
	LocalSpec    []byte
	KeyPath      string
	Ciphertext   string
	CipherMethod string
	Parts        []Part
}

type PlayMetaRequest struct {
	DowloadURL string
	Dir        string
	Filename   string
	TaskId     string
	Ciphertext []byte
	Spec       []byte
}

func playSpec(ctx context.Context, req *PlayMetaRequest) error {
	before := func(r *http.Request) {
		r.Header.Set("Accept", "application/json, text/plain, */*")
		r.Header.Set("User-Agent", zhttp.RandomUserAgent())
		r.Header.Set("Referer", r.URL.String())
		r.Header.Set("Origin", "https://time.geekbang.org")
	}
	after := func(r *http.Response) error {
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
	}
	err := zhttp.NewRequest().Before(before).
		After(after).DoWithRetry(ctx, http.MethodGet, req.DowloadURL, nil)
	if err != nil {
		return err
	}
	return nil
}

func playCiphertext(ctx context.Context, req *PlayMetaRequest, uri string) error {
	before := func(r *http.Request) {
		r.Header.Set("Accept", "application/json, text/plain, */*")
		r.Header.Set("User-Agent", zhttp.RandomUserAgent())
		r.Header.Set("Referer", r.URL.String())
		r.Header.Set("Origin", "https://time.geekbang.org")
	}
	after := func(r *http.Response) error {
		if zhttp.IsHTTPSuccessStatus(r.StatusCode) {
			raw, err := io.ReadAll(r.Body)
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				return fmt.Errorf("cipher key is empty %v", r.Request.URL)
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
	}
	err := zhttp.NewRequest().Before(before).After(after).DoWithRetry(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func RewritePlay(ctx context.Context, req PlayMetaRequest) (*PlayMeta, error) {
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute*3)
	defer retryCancel()

	if len(req.Spec) == 0 {
		if err := playSpec(retryCtx, &req); err != nil {
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
		switch {
		case strings.HasPrefix(l, "#EXT-X-KEY:METHOD=AES-128"):
			meta.CipherMethod = "AES-128"
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
				if err := playCiphertext(retryCtx, &req, sps[1]); err != nil {
					return nil, err
				}
			}
		case strings.HasSuffix(l, ".ts"):
			playHost := req.DowloadURL[:strings.LastIndex(req.DowloadURL, "/")+1]
			if !strings.HasPrefix(l, "https://") {
				destName := path.Join(req.Dir, req.Filename, l)
				meta.Parts = append(meta.Parts, Part{playHost + l, destName, false})
				rl = playHost + l
			} else {
				tsPath := strings.TrimPrefix(l, playHost)
				if strings.HasPrefix(tsPath, playHost) {
					rl = tsPath
				}
				l = l[strings.LastIndex(l, "/")+1:]
				destName := path.Join(req.Dir, req.Filename, tsPath)
				meta.Parts = append(meta.Parts, Part{rl, destName, false})
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

func Video(ctx context.Context, dir, fileName string, req *PlayMeta) (string, error) {
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute*10)
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
	before := func(r *http.Request) {
		r.Header.Set("Accept", "application/json, text/plain, */*")
		r.Header.Set("User-Agent", zhttp.RandomUserAgent())
	}

	after := func(destName string) func(r *http.Response) error {
		return func(r *http.Response) error {
			if zhttp.IsHTTPSuccessStatus(r.StatusCode) {
				if partStat, err1 := global.Storage.Put(destName, r.Body); err1 != nil {
					return err1
				} else if partStat.Size() <= 0 {
					return fmt.Errorf("[%s] is empty", destName)
				}
				global.LOG.Info("video part end", zap.String("part", destName))
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
		}
	}

	for _, t := range req.Parts {
		partURL, destName := t.Src, t.Dest
		if partStat, _ := global.Storage.Stat(destName); partStat != nil && partStat.Size() > 0 {
			continue
		}
		err = zhttp.NewRequest().Before(before).
			After(after(destName)).DoWithRetry(retryCtx, http.MethodGet, partURL, nil)
		if err != nil {
			return "", err
		}
	}

	ffmpeg_command := []string{
		"-allowed_extensions",
		"ALL",
		"-protocol_whitelist",
		"concat,file,http,https,tcp,tls,crypto",
		"-i",
		path.Join(destDir, "index.m3u8"),
	}
	if len(req.KeyPath) > 0 {
		ffmpeg_command = append(ffmpeg_command, "-hls_key_info_file", path.Join(destDir, "key.key"))
	}
	ffmpeg_command = append(ffmpeg_command, "-c", "copy", concatPath)
	global.LOG.Info("video", zap.String("concatPath", concatPath))
	output, err := exec.CommandContext(retryCtx, "ffmpeg", ffmpeg_command...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s,%s", err.Error(), string(output))
	}
	global.LOG.Info("video", zap.String("removeAll", destDir))
	_ = os.RemoveAll(destDir)
	return concatPath, nil
}
