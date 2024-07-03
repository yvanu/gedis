package util

import (
	"os"
	"path"
)

func checkPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

func OpenFile(filename string, dir string) (*os.File, error) {
	if !checkPermission(dir) {
		return nil, os.ErrPermission
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	return os.OpenFile(path.Join(dir, filename), os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
}
