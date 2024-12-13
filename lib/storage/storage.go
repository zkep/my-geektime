package storage

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type Storage interface {
	Name() string
	Put(string, io.ReadCloser) (os.FileInfo, error)
	Get(string) (io.ReadCloser, os.FileInfo, error)
	Stat(string) (os.FileInfo, error)
	Delete(string) error
	GetKey(key string, isReal bool) string
	GetUrl(key string) string
}

var _ Storage = (*LocalStorage)(nil)

type LocalStorage struct {
	host      string
	bucket    string
	directory string
	prefix    string
}

func NewLocalStorage(host, bucket, directory string) (Storage, error) {
	s := LocalStorage{
		host:      strings.TrimSuffix(host, "/"),
		bucket:    strings.TrimSuffix(bucket, "/"),
		directory: strings.TrimSuffix(directory, "/"),
		prefix:    "local://",
	}
	if stat, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(directory, os.ModePerm); err != nil {
				return nil, err
			}
		} else if !stat.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", directory)
		}
	}
	return &s, nil
}

func (s *LocalStorage) Name() string { return "local" }

func (ls *LocalStorage) mkdir(key string) (string, error) {
	key = strings.TrimPrefix(key, ls.directory)
	objectPath := path.Join(ls.directory, key)
	objectDir := path.Dir(objectPath)
	_, err := os.Stat(objectDir)
	if os.IsNotExist(err) {
		return objectPath, os.MkdirAll(objectDir, os.ModePerm)
	}
	return objectPath, err
}

func (ls *LocalStorage) Put(key string, r io.ReadCloser) (os.FileInfo, error) {
	dest, err := ls.mkdir(key)
	if err != nil {
		return nil, err
	}
	fi, err := os.Create(dest)
	if err != nil {
		return nil, err
	}
	defer func() { _ = fi.Close() }()
	_, err = io.Copy(fi, r)
	if err != nil {
		return nil, err
	}
	stat, err := fi.Stat()
	if err != nil {
		return nil, err
	}
	return stat, nil
}

func (ls *LocalStorage) Get(key string) (io.ReadCloser, os.FileInfo, error) {
	dest, err := ls.mkdir(key)
	if err != nil {
		return nil, nil, err
	}
	fi, err := os.Open(dest)
	if err != nil {
		return nil, nil, err
	}
	stat, err := fi.Stat()
	if err != nil {
		return nil, nil, err
	}
	return fi, stat, nil
}

func (ls *LocalStorage) Delete(key string) error {
	dest, err := ls.mkdir(key)
	if err != nil {
		return err
	}
	_, err = os.Stat(dest)
	if os.IsExist(err) {
		return os.Remove(dest)
	}
	return err
}

func (ls *LocalStorage) Stat(key string) (os.FileInfo, error) {
	dest, err := ls.mkdir(key)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(dest)
	if err != nil {
		return nil, err
	}
	return stat, nil
}

func (ls *LocalStorage) GetKey(key string, isReal bool) string {
	if strings.HasPrefix(key, ls.prefix) {
		key = strings.TrimPrefix(key, ls.prefix)
		key = strings.TrimPrefix(key, ls.bucket)
	} else {
		key = strings.TrimPrefix(key, ls.directory)
	}
	key = strings.TrimPrefix(key, "/")
	if isReal {
		return path.Join(ls.directory, key)
	}
	return fmt.Sprintf("%s%s/%s", ls.prefix, ls.bucket, key)
}

func (ls *LocalStorage) GetUrl(key string) string {
	if strings.HasPrefix(key, ls.prefix) {
		return fmt.Sprintf("%s/%s", ls.host, strings.TrimPrefix(key, ls.prefix))
	}
	key = strings.TrimPrefix(key, ls.directory)
	key = strings.TrimPrefix(key, "/")
	return fmt.Sprintf("%s/%s/%s", ls.host, ls.bucket, key)
}
