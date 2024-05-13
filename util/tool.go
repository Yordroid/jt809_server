package util

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// DecimalToBCD 将十进制数转换为BCD码
func DecimalToBCD(decimal byte) byte {
	return (decimal / 10 << 4) + (decimal % 10)
}

func TimeToBCD(utcTime int64) []byte {
	bcdTime := []byte{0, 0, 0, 0, 0, 0}
	tmTime := time.Unix(utcTime, 0).Local()
	bcdTime[0] = DecimalToBCD(byte(tmTime.Year() % 100))
	bcdTime[1] = DecimalToBCD(byte(tmTime.Month()))
	bcdTime[2] = DecimalToBCD(byte(tmTime.Day()))
	bcdTime[3] = DecimalToBCD(byte(tmTime.Hour()))
	bcdTime[4] = DecimalToBCD(byte(tmTime.Minute()))
	bcdTime[5] = DecimalToBCD(byte(tmTime.Second()))
	return bcdTime
}

func ParseLocalTime(layout, strTime string) (time.Time, error) {
	return time.ParseInLocation(layout, strTime, time.Local)
}

// GetRealAddr 获取真实IP
func GetRealAddr(request *http.Request) string {
	if request == nil {
		return ""
	}
	realIP := request.Header.Get("X-Forwarded-For")
	if realIP == "" {
		realIP = request.Header.Get("X-Real-Ip")
		if realIP == "" {
			realIP = request.RemoteAddr
		}
	}
	return realIP
}

const InvalidGUID = 0xFFFFFFFF

type VsGoTaskFuncManage struct {
	mapTaskID     sync.Map
	maxUseTime    int64
	maxTaskNum    int32
	curTaskNum    int32
	lastCheckTime int64
}

var taskFuncManage *VsGoTaskFuncManage
var taskFuncManageOnce sync.Once

func TaskFuncIns() *VsGoTaskFuncManage {
	taskFuncManageOnce.Do(func() {
		taskFuncManage = &VsGoTaskFuncManage{}
		taskFuncManage.initTask()
	})
	return taskFuncManage
}

func HasTimerTask(lastTickS *int64, timeOutMs int64) bool {
	curTime := time.Now()
	if curTime.Sub(time.Unix(*lastTickS, 0)).Milliseconds() > timeOutMs {
		*lastTickS = curTime.Unix()
		return true
	}
	return false
}

const VS_GO_TASK_CHECK_PRINT_TIME = 600 * 1000 //10分钟打印一次

func (_self *VsGoTaskFuncManage) initTask() {
	go func() {
		for {
			_self.mapTaskID.Range(func(key, value any) bool {
				taskInfoCtx := value.(VsTaskContext)
				taskID := key.(int64)
				curTime := time.Now().Unix()
				if taskInfoCtx.startTime < curTime {
					if curTime-taskInfoCtx.startTime > 60 {
						log.Info("task func no quit over 30s,taskName:", taskInfoCtx.taskName, " taskID:", taskID)
					}
				}
				return true
			})
			time.Sleep(time.Minute * 5)
			hasTask := HasTimerTask(&_self.lastCheckTime, VS_GO_TASK_CHECK_PRINT_TIME)
			if hasTask {
				log.Info("go task timer check,maxUseTime:", _self.maxUseTime, " maxTaskNum:", _self.maxTaskNum, " curTaskNum:", _self.curTaskNum)
			}

		}

	}()
}

type VsTaskContext struct {
	startTime int64
	taskName  string
}

func (_self *VsGoTaskFuncManage) addTaskID(taskName string, taskID int64) {
	curTaskInfo := VsTaskContext{}
	curTaskInfo.taskName = taskName
	curTaskInfo.startTime = time.Now().Unix()
	_self.mapTaskID.Store(taskID, curTaskInfo)
	atomic.AddInt32(&_self.curTaskNum, 1)
	if _self.curTaskNum > _self.maxTaskNum {
		_self.maxTaskNum = _self.curTaskNum
	}
}
func (_self *VsGoTaskFuncManage) delTaskID(taskID int64) {
	taskInfoI, isOK := _self.mapTaskID.LoadAndDelete(taskID)
	if isOK {
		curTime := time.Now().Unix()
		taskInfoCtx := taskInfoI.(VsTaskContext)
		if taskInfoCtx.startTime < curTime {
			if curTime-taskInfoCtx.startTime > 30 {
				log.Info("task func exec over 30s,taskName:", taskInfoCtx.taskName)
			}
			if curTime-taskInfoCtx.startTime > _self.maxUseTime {
				_self.maxUseTime = curTime - taskInfoCtx.startTime
				log.Info("task func update max use time,taskName:", taskInfoCtx.taskName, "useTime:", _self.maxUseTime)
			}
		}
	}
	atomic.AddInt32(&_self.curTaskNum, -1)
}
func GoTaskFunc(taskName string, taskFunc func()) {
	curID := GetSnowflakeID()
	TaskFuncIns().addTaskID(taskName, curID)
	defer TaskFuncIns().delTaskID(curID)
	taskFunc()

}

// GetRoutineID 获取当前协程ID
func GetRoutineID() uint64 {
	b := make([]byte, 64)
	runtime.Stack(b, false)
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func ExceptionRecover() {
	if r := recover(); r != nil {
		log.Info("ExceptionRecover err,%v", r)
	}
}

// ParseJson json to struct
func ParseJson(jsonData []byte, jsonObject interface{}) bool {
	if string(jsonData) == "" {
		return false
	}
	if err := json.Unmarshal(jsonData, jsonObject); err != nil {
		log.Info("ParseJson err,", err.Error(), "content:", string(jsonData))
		return false
	}
	return true
}

func ToJson(jsonObject interface{}) string {
	var (
		byteString []byte
		err        error
	)
	if byteString, err = json.Marshal(jsonObject); err != nil {
		log.Info("ToJson err,", err.Error())
		return ""
	}
	return string(byteString)
}

// 解析url参数
func ParseUrlParam(urlParam string) map[string]string {
	paramKeyValue := map[string]string{}
	if urlParam != "" {
		params := strings.Split(urlParam, "&")
		if len(params) > 0 {
			for _, curParam := range params {
				retParam := strings.Split(curParam, "=")
				if len(retParam) == 2 {
					paramKeyValue[retParam[0]] = retParam[1]
				}
			}
		} else {
			retParam := strings.Split(urlParam, "=")
			if len(retParam) == 2 {
				paramKeyValue[retParam[0]] = retParam[1]
			}
		}
	}
	return paramKeyValue
}

// zip压缩
func ZipBytes(data []byte) []byte {

	var in bytes.Buffer
	z := zlib.NewWriter(&in)
	z.Write(data)
	z.Close()
	return in.Bytes()
}

// zip解压
func UZipBytes(data []byte) []byte {
	var out bytes.Buffer
	var in bytes.Buffer
	in.Write(data)
	r, _ := zlib.NewReader(&in)
	r.Close()
	io.Copy(&out, r)
	return out.Bytes()
}

// gzip压缩
func GZipBytes(data []byte) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		panic(err)
	}
	if err := gz.Flush(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	return b.Bytes()
}

// 获取正在运行的函数名
func GetFuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

// PrintException 严重异常
func PrintException(param ...any) {
	log.Error(param, "printException stack:", string(debug.Stack()))
	panic(param)
}

type None struct{}

var none None

type SetType interface {
	uint32 | string
}

// Set 结构,用于去重
type Set[SetType comparable] struct {
	item    map[SetType]None
	capSize int
	isSafe  bool
	rw      sync.Mutex
}

func (_self *Set[SetType]) InitSize(size int, isSafe bool) {
	if nil == _self.item {
		_self.item = make(map[SetType]None, size)
		_self.capSize = size
		_self.isSafe = isSafe
	}
}

func (_self *Set[SetType]) Add(key SetType) {
	if nil == _self.item {
		_self.item = make(map[SetType]None)
	}
	if _self.isSafe {
		_self.rw.Lock()
		defer _self.rw.Unlock()
	}
	_self.item[key] = none

}
func (_self *Set[SetType]) Del(key SetType) {
	if _self.isSafe {
		_self.rw.Lock()
		defer _self.rw.Unlock()
	}
	delete(_self.item, key)
}
func (_self *Set[SetType]) AddIsRepeated(key SetType) bool {
	if _self.isSafe {
		_self.rw.Lock()
		defer _self.rw.Unlock()
	}
	if nil == _self.item {
		_self.item = make(map[SetType]None)
	}
	_, isExist := _self.item[key]
	_self.item[key] = none
	return isExist
}

// Combine 合并
func (_self *Set[SetType]) Combine(key1 []SetType, key2 []SetType) {
	if _self.isSafe {
		_self.rw.Lock()
		defer _self.rw.Unlock()
	}
	for idx := 0; idx < len(key1); idx++ {
		_self.item[key1[idx]] = none
	}
	for idx := 0; idx < len(key2); idx++ {
		_self.item[key2[idx]] = none
	}
}

func (_self *Set[SetType]) ForEach(keyFunc func(SetType)) {
	if _self.isSafe {
		_self.rw.Lock()
		defer _self.rw.Unlock()
	}
	for key, _ := range _self.item {
		if keyFunc != nil {
			keyFunc(key)
		}
	}
}

func (_self *Set[SetType]) Clear() {
	if _self.isSafe {
		_self.rw.Lock()
		defer _self.rw.Unlock()
	}
	if _self.capSize == 0 {
		_self.item = make(map[SetType]None)
	} else {
		_self.item = make(map[SetType]None, _self.capSize)
	}

}

func (_self *Set[SetType]) GetAllKey() []SetType {
	if _self.isSafe {
		_self.rw.Lock()
		defer _self.rw.Unlock()
	}
	var resultArray []SetType
	for key, _ := range _self.item {
		resultArray = append(resultArray, key)
	}
	return resultArray
}
func (_self *Set[SetType]) GetSize() int {
	if _self.isSafe {
		_self.rw.Lock()
		defer _self.rw.Unlock()
	}
	return len(_self.item)
}
func (_self *Set[SetType]) HasKey(key SetType) bool {
	if _self.isSafe {
		_self.rw.Lock()
		defer _self.rw.Unlock()
	}
	_, isExist := _self.item[key]
	return isExist
}

type SnowFlake struct {
	node *snowflake.Node
}

var instance *SnowFlake
var once sync.Once

func GetIns() *SnowFlake {
	once.Do(func() {
		instance = &SnowFlake{}
	})
	return instance
}

func (_self *SnowFlake) Init(nodeCode int64) {
	node, err := snowflake.NewNode(nodeCode)
	if err != nil {
		log.Error("init snowflake id fail", err.Error())
		return
	}
	_self.node = node
}

func GetSnowflakeID() int64 {
	if instance == nil {
		GetIns().Init(10)
	}
	if instance.node == nil {
		log.Info("snowflake instance.node is null,auto set default")
		GetIns().Init(10)
	}
	return int64(instance.node.Generate())
}

var isDebugMode bool

func IsDebugMode() bool {
	return isDebugMode
}

func SetDebugMode() {
	isDebugMode = true
}

func GetAppConfigDir() string {
	confDir := GetAppPath() + "config"
	CreateDir(confDir)
	return confDir
}

func GetAppConfigFilePath(fileName string) string {
	configPath := GetAppConfigDir() + "/" + fileName
	if !IsExist(configPath) {
		CreateFile(configPath)
	}
	return configPath
}

func GetAppDataDir() string {
	confDir := GetAppPath() + "data"
	CreateDir(confDir)
	return confDir
}

// GetOsHostName 获取主机名称
func GetOsHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		log.Error("GetOsHostName fail", err.Error())
		return ""
	}
	return hostName
}

// CRC16-CCITT

func CRC16CheckSum(data []byte) uint16 {
	iLenIn := len(data)
	var wTemp uint16
	var wCRC uint16 = 0xFFFF
	var pCRCOut uint16 = 0
	for i := 0; i < iLenIn; i++ {
		for j := 0; j < 8; j++ {
			wTemp = (uint16(data[i]<<j) & 0x80) ^ ((wCRC & 0x8000) >> 8)
			wCRC <<= 1
			if wTemp != 0 {
				wCRC ^= 0x1021
			}
		}
	}
	pCRCOut = wCRC
	return pCRCOut
}
