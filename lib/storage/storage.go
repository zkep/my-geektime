package storage

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type Storage interface {
	Put(string, io.ReadCloser) (os.FileInfo, string, error)
	Get(string) (io.ReadCloser, os.FileInfo, string, error)
	Stat(string) (os.FileInfo, string, error)
	Delete(string) error
	GetKey(key string) string
}

var _ Storage = (*LocalStorage)(nil)

type LocalStorage struct {
	directory string
}

func NewLocalStorage(directory string) (Storage, error) {
	if stat, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(directory, os.ModePerm); err != nil {
				return nil, err
			}
		} else if !stat.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", directory)
		}
	}
	return &LocalStorage{directory: directory}, nil
}

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

func (ls *LocalStorage) Put(key string, r io.ReadCloser) (os.FileInfo, string, error) {
	dest, err := ls.mkdir(key)
	if err != nil {
		return nil, dest, err
	}
	fi, err := os.Create(dest)
	if err != nil {
		return nil, dest, err
	}
	defer func() { _ = fi.Close() }()
	_, err = io.Copy(fi, r)
	if err != nil {
		return nil, dest, err
	}
	stat, err := fi.Stat()
	if err != nil {
		return nil, dest, err
	}
	return stat, dest, nil
}

func (ls *LocalStorage) Get(key string) (io.ReadCloser, os.FileInfo, string, error) {
	dest, err := ls.mkdir(key)
	if err != nil {
		return nil, nil, dest, err
	}
	fi, err := os.Open(dest)
	if err != nil {
		return nil, nil, dest, err
	}
	stat, err := fi.Stat()
	if err != nil {
		return nil, nil, dest, err
	}
	return fi, stat, dest, nil
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

func (ls *LocalStorage) Stat(key string) (os.FileInfo, string, error) {
	dest, err := ls.mkdir(key)
	if err != nil {
		return nil, dest, err
	}
	stat, err := os.Stat(dest)
	if err != nil {
		return nil, dest, err
	}
	return stat, dest, nil
}

func (ls *LocalStorage) GetKey(key string) string {
	key = strings.TrimPrefix(key, ls.directory)
	return path.Join(ls.directory, key)
}
