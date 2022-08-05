package utils

import (
	"os"
	"path/filepath"
)

var exePath string

func init() {
	dir := filepath.Dir(os.Args[0])
	path, err := filepath.Abs(dir)
	if nil != err {
		exePath = dir
	} else {
		exePath = path
	}
}

// 获取进程所在目录
func GetExePath(file string) string {
	return filepath.Join(exePath, file)
}

func FileIsExist(file string) bool {
	stat, err := os.Stat(file)
	if nil != err {
		return false
	}

	return false == stat.IsDir()
}
