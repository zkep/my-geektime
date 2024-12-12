package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/zkep/mygeektime/lib/color"
	"github.com/zkep/mygeektime/lib/zhttp"
)

type BrowserFlags struct {
	Driver  string `name:"driver" description:"driver to use " default:"chrome"`
	Version string `name:"version" description:"browser version of dependency"`
}

var (
	// https://mirrors.huaweicloud.com/chromedriver/
	chromeHubURL = "https://storage.googleapis.com/chrome-for-testing-public/"

	chromeOS = map[string]string{
		"linux64":      "linux64/chromedriver-linux64.zip",
		"darwin-arm64": "mac-arm64/chromedriver-mac-arm64.zip",
		"darwin-x64":   "mac-x64/chromedriver-mac-x64.zip",
		"win64":        "win64/chromedriver-win64.zip",
		"win32":        "win32/chromedriver-win32.zip",
	}

	chromeVersionHelp = []string{
		color.Red("Browser version is required ."),
		color.Cyan("For example: 131.0.6778.109"),
		color.Red("You can execute in the address bar of Google Chrome browser:"),
		color.Cyan("chrome://version"),
	}

	repo = "repo"
)

func browser(f *BrowserFlags) error {
	if f.Version == "" {
		return errors.New(strings.Join(chromeVersionHelp, "\n"))
	}
	systemOS := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("OS: %s\n", color.Cyan(systemOS))
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*15)
	defer cancel()
	switch f.Driver {
	case "chrome":
		versionName, ok := chromeOS[systemOS]
		if !ok {
			return fmt.Errorf(color.Red("Unsupported drive type %s"), systemOS)
		}
		dirName := path.Join(repo, path.Dir(versionName))
		if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
			return err
		}
		fileName := path.Join(dirName, path.Base(versionName))
		downloadURL := fmt.Sprintf("%s%s/%s", chromeHubURL, f.Version, versionName)
		if err := zhttp.R.
			Client(&http.Client{Timeout: time.Minute * 15}).
			Before(func(r *http.Request) {
				r.Header.Set("User-Agent", zhttp.RandomUserAgent())
			}).
			After(func(r *http.Response) error {
				if !zhttp.IsHTTPSuccessStatus(r.StatusCode) {
					return fmt.Errorf("http status code %d from %s", r.StatusCode, downloadURL)
				}
				fi, err := os.Create(fileName)
				if err != nil {
					return err
				}
				defer func() { _ = fi.Close() }()
				go func(uri string, size int64) {
					ticker := time.NewTicker(time.Second)
					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							ticker.Reset(time.Second)
							stat, _ := os.Stat(fileName)
							fmt.Printf("%s, progress: %d / %d", uri, stat.Size(), size)
						}
					}
				}(downloadURL, r.ContentLength)
				_, err = io.Copy(fi, r.Body)
				return err
			}).
			DoWithRetry(ctx, http.MethodGet, downloadURL, nil); err != nil {
			return err
		}
		switch path.Ext(fileName) {
		case ".zip":
			fmt.Printf("Unzipping %q", fileName)
			if err := exec.Command("unzip", "-o", fileName, "-d", dirName).Run(); err != nil {
				return fmt.Errorf("Error unzipping %s: %w", fileName, err)
			}
			rename := strings.ReplaceAll(fileName, ".zip", "/chromedriver")
			if err := os.Rename(rename, "chromedriver"); err != nil {
				return err
			}
			_ = os.RemoveAll(repo)
			fmt.Println("download finished: chromedriver")
		}

	default:
		return fmt.Errorf("unsupported driver: %s", f.Driver)
	}
	return nil
}
