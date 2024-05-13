package util

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type RunMode int

const (
	LogDebug RunMode = iota
	LogRelease
)

// 单个日志文件最大大小
const LOG_FILE_MAX_LEN = 10 * 1024 * 1024

type LogFileWriter struct {
	File       *os.File
	FileName   string
	Size       int64
	LastDay    uint8
	AppName    string
	LogFullDir string
}

// DeleteLogDirFile 删除日志文件
func (_self *LogFileWriter) DeleteLogDirFile(delDay int) {
	_dir, err := ioutil.ReadDir(_self.LogFullDir)
	if err != nil {
		log.Info("DeleteLogDirFile fail,dir read err")
		return
	}
	delTime := time.Now().AddDate(0, 0, -delDay)
	for _, _file := range _dir {
		fileName := _file.Name()
		if !strings.Contains(fileName, _self.AppName) {
			continue
		}
		strArray := strings.Split(_file.Name(), "_")
		if len(strArray) < 3 {
			log.Info("DeleteLogDirFile fail,file format err")
			return
		}
		t1, _ := ParseLocalTime("20060102", strArray[len(strArray)-2])
		if t1.Unix() < delTime.Unix() {
			DeleteFile(_self.LogFullDir + _file.Name())
			log.Info("del log file", _self.LogFullDir+_file.Name())
		}
	}
}

// InitLogger logLevel - 2:error,3:warn 4:info,5:debug
func InitLogger(appName, logFullDir string, runMode RunMode, logLevel uint32) {

	var logName string
	switch runMode {
	case LogDebug:
		log.SetOutput(os.Stdout)
	case LogRelease:
		now := time.Now()
		strTime := fmt.Sprintf("_%04d%02d%02d_%02d%02d%02d.log", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
		//配置日志
		if len(logFullDir) > 0 {
			if logFullDir[len(logFullDir)-1] != '/' {
				logFullDir += "/"
			}
		}
		fmt.Println("logPath", logFullDir)
		CreateDir(logFullDir)
		logName = logFullDir + appName + strTime
		file, err := os.OpenFile(logName, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0600)
		defer file.Close()
		if err != nil {
			log.Fatal("log  init failed")
			panic(err)
		}
		info, err := file.Stat()
		if err != nil {
			panic(err)
		}
		fileWriter := &LogFileWriter{File: file, FileName: logName, Size: info.Size(), AppName: appName, LogFullDir: logFullDir}
		log.SetOutput(fileWriter)
		log.SetLevel(log.Level(logLevel))
		//删除日志文件
		go func() {
			isNeedDel := true
			for {
				curHour := time.Now().Hour()
				if isNeedDel && curHour == 5 {
					isNeedDel = false
					fileWriter.DeleteLogDirFile(10)
				}
				if curHour != 5 {
					isNeedDel = true
				}
				time.Sleep(time.Minute * 20)
			}

		}()
	default:
		panic("run mode err")
	}
	log.SetFormatter(&log.TextFormatter{})
}

func (_self *LogFileWriter) Write(data []byte) (n int, err error) {
	if _self == nil {
		return 0, errors.New("logFileWriter is nil")
	}
	if _self.File == nil {
		return 0, errors.New("file not opened")
	}

	_self.File, _ = os.OpenFile(_self.FileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0600)
	now := time.Now()
	//每天产生一个新文件
	curDay := now.Day()
	if (_self.Size > LOG_FILE_MAX_LEN) || (0 != _self.LastDay && _self.LastDay != uint8(curDay)) {
		_self.File.Close()
		fmt.Println("log file full or new day")
		CreateFile(_self.LogFullDir)
		strTime := fmt.Sprintf("_%04d%02d%02d_%02d%02d%02d.log", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
		_self.FileName = _self.LogFullDir + _self.AppName + strTime
		_self.File, _ = os.OpenFile(_self.FileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0600)
		_self.Size = 0
	}
	_self.LastDay = uint8(curDay)
	n, e := _self.File.Write(data)
	_self.Size += int64(n)
	_self.File.Close()
	return n, e
}
