package config

import (
	"fmt"
	"jt809_server/util"
	"strconv"
	"strings"
	"sync"
)

const CURRENT_VER_NO string = "JT_809_SERVER_V4.0.0.240307"

var BuildTime = ""

// GetVersionNo 获取版本号
func GetVersionNo() string {
	return CURRENT_VER_NO + BuildTime
}

var systemConfig *SystemConfig
var confOnce sync.Once

// SystemConf 获取配置单实例
func SystemConf() *SystemConfig {
	confOnce.Do(func() {
		systemConfig = &SystemConfig{}
	})
	return systemConfig
}

type JTUserConfigInfo struct {
	BindPort       uint16 //监听端口
	PublicAddr     string //公网地址
	RemoteAddr     string //远程地址
	RemotePort     uint16 //远程端口
	JTVer          uint8  //部标版本 0:无,1:2011,2:2013,3:2019
	M1             uint32
	IA1            uint32
	IC1            uint32
	Key            uint32 //用户KEY
	AccessID       uint32 //接入码
	JTUserID       uint32 //部标平台的用户ID
	JTUserPwd      string //部标平台的密码
	JTPlatformCode string //平台唯一编码
	BuUserName     string //业务系统的用户名
	BuUserPwdMD5   string //业务系统的用户密码
	UserKey        int64  //用户唯一KEY adminGuid+userGuid
	IsPrintHex     int    //是否打印接收和发送二进制
}

// SystemConfig 全局配置
type SystemConfig struct {
	logLevel      uint32 //logLevel - 2:error,3:warn 4:info,5:debug 默认4
	isDurable     uint32 //消息是否持久化
	version       string
	appDataPath   string //应用数据路径
	brokenInfo    string //rabbitMQ broken 消息
	httpBindPort  uint32 //http 服务
	webUrlBase    string //web URL
	dataUrlBase   string //data URL
	buType        int    //业务系统类型,0:自己的V4平台,1:其它待定
	userKeys      string //用户key,多个用,号隔开,目前由管理员id(高32)+用户ID组成(低32)
	jtUserConfigs []JTUserConfigInfo
}

// LoadSystemIni 加载ini配置
func LoadSystemIni(configName string) bool {
	iniParser := util.IniParser{}
	configPath := util.GetAppConfigFilePath(configName + "_config.ini")
	if err := iniParser.Load(configPath); err != nil {
		fmt.Printf("try load config file[%s] error[%s]\n", configPath, err.Error())
		return false
	}
	curConfig := SystemConf()

	curConfig.logLevel = iniParser.GetUint32("baseConfig", "logLevel", 0)
	if curConfig.logLevel == 0 {
		curConfig.logLevel = 4
		iniParser.SetUInt("baseConfig", "logLevel", curConfig.logLevel)
	}

	curConfig.isDurable = iniParser.GetUint32("baseConfig", "isDurable", 0)
	if curConfig.isDurable == 0 {
		curConfig.isDurable = 0
		iniParser.SetUInt("baseConfig", "isDurable", curConfig.isDurable)
	}

	curConfig.appDataPath = iniParser.GetString("baseConfig", "appDataPath", "")
	if curConfig.appDataPath == "" {
		curConfig.appDataPath = util.GetDefaultAppDir()
		iniParser.SetString("baseConfig", "appDataPath", curConfig.appDataPath)
	}
	curConfig.brokenInfo = iniParser.GetString("baseConfig", "brokenInfo", "")
	if curConfig.brokenInfo == "" {
		curConfig.brokenInfo = "amqp://guest:guest@localhost:5672"
		iniParser.SetString("baseConfig", "brokenInfo", curConfig.brokenInfo)
	}
	curConfig.httpBindPort = iniParser.GetUint32("baseConfig", "httpBindPort", 0)
	if curConfig.httpBindPort == 0 {
		curConfig.httpBindPort = 7222
		iniParser.SetUInt("baseConfig", "httpBindPort", curConfig.httpBindPort)
	}
	curConfig.webUrlBase = iniParser.GetString("baseConfig", "webUrlBase", "")
	if curConfig.webUrlBase == "" {
		curConfig.webUrlBase = "http://127.0.0.1:7215"
		iniParser.SetString("baseConfig", "webUrlBase", curConfig.webUrlBase)
	}
	curConfig.dataUrlBase = iniParser.GetString("baseConfig", "dataUrlBase", "")
	if curConfig.dataUrlBase == "" {
		curConfig.dataUrlBase = "http://127.0.0.1:7213"
		iniParser.SetString("baseConfig", "dataUrlBase", curConfig.dataUrlBase)
	}
	curConfig.buType = int(iniParser.GetInt32("baseConfig", "buType", -1))
	if curConfig.buType == -1 {
		curConfig.buType = 0
		iniParser.SetInt("baseConfig", "buType", int32(curConfig.buType))
	}

	curConfig.userKeys = iniParser.GetString("baseConfig", "userKeys", "")
	if curConfig.userKeys == "" {
		curConfig.userKeys = ""
		iniParser.SetString("baseConfig", "userKeys", curConfig.userKeys)
	}
	aUserKey := strings.Split(curConfig.userKeys, ",")
	for idx := 0; idx < len(aUserKey); idx++ {
		curUserKey := aUserKey[idx]
		if curUserKey == "" {
			continue
		}
		curJTUser := JTUserConfigInfo{}
		curJTUser.BuUserName = iniParser.GetString(curUserKey, "buUserName", "")
		if curJTUser.BuUserName == "" {
			curJTUser.BuUserName = ""
			iniParser.SetString(curUserKey, "buUserName", curJTUser.BuUserName)
		}
		curJTUser.BuUserPwdMD5 = iniParser.GetString(curUserKey, "buUserPwd", "")
		if curJTUser.BuUserPwdMD5 == "" {
			curJTUser.BuUserPwdMD5 = ""
			iniParser.SetString(curUserKey, "buUserPwd", curJTUser.BuUserPwdMD5)
		}
		curJTUser.BindPort = uint16(iniParser.GetUint32(curUserKey, "bindPort", 0))
		if curJTUser.BindPort == 0 {
			curJTUser.BindPort = 7223
			iniParser.SetUInt(curUserKey, "bindPort", uint32(curJTUser.BindPort))
		}
		curJTUser.PublicAddr = iniParser.GetString(curUserKey, "publicAddr", "")
		if curJTUser.PublicAddr == "" {
			curJTUser.PublicAddr = "127.0.0.1"
			iniParser.SetString(curUserKey, "publicAddr", curJTUser.PublicAddr)
		}
		curJTUser.RemoteAddr = iniParser.GetString(curUserKey, "remoteAddr", "")
		if curJTUser.RemoteAddr == "" {
			curJTUser.RemoteAddr = "127.0.0.1"
			iniParser.SetString(curUserKey, "remoteAddr", curJTUser.RemoteAddr)
		}
		curJTUser.RemotePort = uint16(iniParser.GetUint32(curUserKey, "remotePort", 0))
		if curJTUser.RemotePort == 0 {
			curJTUser.RemotePort = 809
			iniParser.SetUInt(curUserKey, "remotePort", uint32(curJTUser.RemotePort))
		}
		curJTUser.JTVer = uint8(iniParser.GetUint32(curUserKey, "jtVer", 0))
		if curJTUser.JTVer == 0 {
			curJTUser.JTVer = 3
			iniParser.SetUInt(curUserKey, "jtVer", uint32(curJTUser.JTVer))
		}
		curJTUser.M1 = iniParser.GetUint32(curUserKey, "M1", 0)
		if curJTUser.M1 == 0 {
			curJTUser.M1 = 0
			iniParser.SetUInt(curUserKey, "M1", curJTUser.M1)
		}
		curJTUser.IA1 = iniParser.GetUint32(curUserKey, "IA1", 0)
		if curJTUser.IA1 == 0 {
			curJTUser.IA1 = 0
			iniParser.SetUInt(curUserKey, "IA1", curJTUser.IA1)
		}
		curJTUser.IC1 = iniParser.GetUint32(curUserKey, "IC1", 0)
		if curJTUser.IC1 == 0 {
			curJTUser.IC1 = 0
			iniParser.SetUInt(curUserKey, "IC1", curJTUser.IC1)
		}
		curJTUser.Key = iniParser.GetUint32(curUserKey, "key", 0)
		if curJTUser.Key == 0 {
			curJTUser.Key = 0
			iniParser.SetUInt(curUserKey, "key", curJTUser.Key)
		}
		curJTUser.AccessID = iniParser.GetUint32(curUserKey, "accessID", 0)
		if curJTUser.AccessID == 0 {
			curJTUser.AccessID = 0
			iniParser.SetUInt(curUserKey, "accessID", curJTUser.AccessID)
		}
		curJTUser.JTUserID = iniParser.GetUint32(curUserKey, "jtUserID", 0)
		if curJTUser.JTUserID == 0 {
			curJTUser.JTUserID = 0
			iniParser.SetUInt(curUserKey, "jtUserID", curJTUser.JTUserID)
		}
		curJTUser.JTUserPwd = iniParser.GetString(curUserKey, "jtUserPwd", "")
		if curJTUser.JTUserPwd == "" {
			curJTUser.JTUserPwd = ""
			iniParser.SetString(curUserKey, "jtUserPwd", curJTUser.JTUserPwd)
		}
		curJTUser.JTPlatformCode = iniParser.GetString(curUserKey, "jtPlatformCode", "")
		if curJTUser.JTPlatformCode == "" {
			curJTUser.JTPlatformCode = ""
			iniParser.SetString(curUserKey, "jtPlatformCode", curJTUser.JTPlatformCode)
		}
		curJTUser.IsPrintHex = int(iniParser.GetInt32("baseConfig", "isPrintHex", -1))
		if curJTUser.IsPrintHex == -1 {
			curJTUser.IsPrintHex = 0
			iniParser.SetInt("baseConfig", "isPrintHex", 0)
		}

		curJTUser.UserKey, _ = strconv.ParseInt(curUserKey, 10, 64)
		curConfig.jtUserConfigs = append(curConfig.jtUserConfigs, curJTUser)

	}
	//写版本
	iniParser.SetString("baseConfig", "version", GetVersionNo())
	curConfig.version = GetVersionNo()
	iniParser.Save()
	return true
}

func GetAppDataPath() string {
	return SystemConf().appDataPath
}

func GetMQBroken() string {
	return SystemConf().brokenInfo
}

func IsDurable() bool {
	if SystemConf().isDurable == 0 {
		return false
	}
	return true
}

func GetLogLevel() uint32 {
	return SystemConf().logLevel
}

func GetHttpBindPort() uint32 {
	return SystemConf().httpBindPort
}
func GetWebUrlBase() string {
	return SystemConf().webUrlBase
}
func GetDataUrlBase() string {
	return SystemConf().dataUrlBase
}

func GetJTUserConfig() []JTUserConfigInfo {
	return SystemConf().jtUserConfigs
}
func GetBuType() int {
	return SystemConf().buType
}
