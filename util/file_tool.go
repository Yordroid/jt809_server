package util

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DeleteFile 删除文件
func DeleteFile(fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		log.Error(fmt.Sprintf("delete file err,filename:[%s],errInfo[%s]", fileName, err.Error()))
	}
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

// GetParentDirectory 获取上级目录，level =1为上一级，2为上上级
func GetParentDirectory(dirctory string, level int32) string {
	if 0 == level {
		return dirctory
	}
	level--
	var parentPath string
	if strings.Contains(dirctory, "\\") {
		parentPath = substr(dirctory, 0, strings.LastIndex(dirctory, "\\"))
	} else {
		parentPath = substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
	}
	return GetParentDirectory(parentPath, level)
}

// CreateFile 调用os.MkdirAll递归创建文件夹
func CreateFile(filePath string) error {
	if !IsExist(filePath) {
		parentDir := GetParentDirectory(filePath, 1)
		err := os.MkdirAll(parentDir, os.ModePerm)
		if err != nil {
			return err
		}
		var (
			f *os.File
		)
		if f, err = os.Create(filePath); err != nil {
			return err
		}
		defer f.Close()
		return err
	}
	return nil
}

func CreateDir(dir string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// IsExist 判断所给路径文件/文件夹是否存在(返回true是存在)
func IsExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

// GetAppPath 获取当前应用目录
func GetAppPath() string {

	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return ""
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return ""
	}
	path = filepath.ToSlash(path)
	i := strings.LastIndex(path, "/")
	//if i < 0 {
	//	i = strings.LastIndex(path, "\\")
	//}
	if i < 0 {
		return ""
	}
	return path[:i+1]
}
