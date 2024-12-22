package rest

import (
	"container/list"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	LangDefault = "zh-CN"
	AcceptLang  = "Accept-Language"
)

type I18n interface {
	HttpValue(req *http.Request, key, defaultVal string, params ...any) string
	LangValue(lang, key, defaultVal string, params ...any) string
}

var _ I18n = (*defaultI18n)(nil)

var (
	locker sync.RWMutex
)

type defaultI18n struct {
	i18nVals map[string]string
}

func InitI18nWithDir(i18nDir string) (I18n, error) {
	i18n := &defaultI18n{i18nVals: make(map[string]string)}
	err := loadI18nWithDir(i18n.i18nVals, i18nDir)
	if err != nil {
		return nil, err
	}
	return i18n, nil
}

func InitI18nWithFsFile(files ...fs.File) (I18n, error) {
	i18n := &defaultI18n{i18nVals: make(map[string]string)}
	locker.Lock()
	defer locker.Unlock()
	for _, fi := range files {
		defer func() { _ = fi.Close() }()
		stat, err := fi.Stat()
		if err != nil {
			return nil, err
		}
		fname := filepath.Base(stat.Name())
		lang := strings.TrimSuffix(fname, filepath.Ext(fname))
		if err = i18nParse(i18n.i18nVals, lang, fi); err != nil {
			return nil, err
		}
	}
	return i18n, nil
}

func HttpLanguage(req *http.Request) string {
	lang := req.Header.Get(AcceptLang)
	if strings.Index(lang, ",") > 0 {
		lang = lang[:strings.Index(lang, ",")]
	}
	if len(lang) == 0 {
		lang = LangDefault
	}
	return lang
}

func langKey(lang, key string) string {
	return fmt.Sprintf("%s.%s", lang, strings.ToLower(key))
}

func (i *defaultI18n) HttpValue(req *http.Request, key, defaultVal string, params ...any) string {
	return i.LangValue(HttpLanguage(req), key, defaultVal, params...)
}

func (i *defaultI18n) LangValue(lang, key, defaultVal string, params ...any) string {
	lkey := langKey(lang, key)
	val, ok := i.i18nVals[lkey]
	if !ok {
		return defaultVal
	}
	if len(params) > 0 {
		return fmt.Sprintf(val, params...)
	}
	return val
}

type Ele struct {
	key   string
	value any
}

func loadI18nWithDir(i18nVals map[string]string, i18nDir string) error {
	return filepath.WalkDir(i18nDir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, er := d.Info()
		if er != nil {
			return er
		}
		file := filepath.Join(i18nDir, info.Name())
		return loadI18nWithFile(i18nVals, file)
	})
}

func loadI18nWithFile(i18nVals map[string]string, files ...string) error {
	locker.Lock()
	defer locker.Unlock()
	for _, file := range files {
		fi, err := os.Open(file)
		if err != nil {
			return err
		}
		defer func() { _ = fi.Close() }()
		fname := filepath.Base(file)
		lang := strings.TrimSuffix(fname, filepath.Ext(fname))
		if err = i18nParse(i18nVals, lang, fi); err != nil {
			return err
		}
	}
	return nil
}

func i18nParse(i18nVals map[string]string, lang string, fi io.Reader) error {
	var values map[string]any
	if err := yaml.NewDecoder(fi).Decode(&values); err != nil {
		return err
	}
	l := list.New()
	l.PushBack(Ele{key: lang, value: values})
	for e := l.Front(); e != nil; e = e.Next() {
		ele := e.Value.(Ele)
		switch ele.value.(type) {
		case string:
			i18nVals[ele.key] = ele.value.(string)
		case map[string]any:
			for k, v := range ele.value.(map[string]any) {
				l.PushBack(Ele{key: langKey(ele.key, k), value: v})
			}
		}
	}
	return nil
}
