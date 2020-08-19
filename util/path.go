package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var ostype = runtime.GOOS

func NormalPath(filepath string) string {
	if ostype == "windows" {
		filepath = strings.Replace(filepath, "/", "\\", -1)
	}
	return filepath
}

func NormalPathF(path string, args ...interface{}) string {
	str := fmt.Sprintf(path, args...)
	return NormalPath(str)
}

func GetApplicationDir() string {
	tmpDir, err := os.Getwd()
	if err != nil {
		file, _ := exec.LookPath(os.Args[0])
		tfile, _ := filepath.Abs(file)
		tmpDir, _ = filepath.Split(tfile)
	}
	return tmpDir
}

func GetAbsolutePath(path string, args ...interface{}) string {
	appDir := GetApplicationDir()
	str := fmt.Sprintf(path, args...)
	str = appDir + str
	return NormalPath(str)
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
