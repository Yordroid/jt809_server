package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"jt809_server/config"
	"jt809_server/internal/bu_service"
	"jt809_server/internal/data_manage_service"
	"jt809_server/internal/jt_service"
	"jt809_server/routers"
	"jt809_server/util"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var BuildTime string

func initCommon(appName string) {
	//dataBuf := []byte{0x5B, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00, 0x0A, 0x10, 0x06, 0x01, 0x34, 0x13, 0x3B, 0x01, 0x00, 0x00, 0x01, 0x00, 0x0C, 0x65, 0x82, 0x00, 0x00, 0x00, 0x00, 0x66, 0x2F, 0x04, 0x54, 0x98, 0xD0, 0x5D}
	//dataParity(dataBuf)
	mode := ""
	flag.StringVar(&mode, "mode", "R", "调试模式") //指针
	flag.Parse()
	config.BuildTime = BuildTime
	config.LoadSystemIni(appName)
	logDir := util.GetAppPath() + "log"
	if config.BuildTime == "" || mode == "D" {
		util.SetDebugMode()
		gin.SetMode(gin.DebugMode)
		util.InitLogger(appName, logDir, util.LogDebug, 0)
	} else {
		gin.SetMode(gin.ReleaseMode)
		util.InitLogger(appName, logDir, util.LogRelease, config.GetLogLevel())
	}
	log.Info("mode:", mode, " BuildTime", BuildTime, " version:", config.CURRENT_VER_NO)
}

func main() {
	initCommon("jt809_server")
	msgNotifyFramework := util.IotMsgNotifyFramework{}
	msgNotifyFramework.InitNotifyFrame()
	data_manage_service.DataManageServiceIns().InitApi(&msgNotifyFramework)
	bu_service.MsgServiceIns().InitApi(&msgNotifyFramework)

	jt_service.JTServiceIns().InitApi(&msgNotifyFramework)
	router := routers.InitRouter()
	strPort := strconv.FormatInt(int64(config.GetHttpBindPort()), 10)
	log.Info("http bind start port:", config.GetHttpBindPort())
	go func() {
		s := &http.Server{
			Addr:           ":" + strPort,
			Handler:        router,
			ReadTimeout:    time.Second * 60 * 5,
			WriteTimeout:   time.Second * 60 * 5,
			MaxHeaderBytes: 1 << 20,
		}
		err := s.ListenAndServe()
		if err != nil {
			log.Info("http bind start port:", err.Error())
		}
	}()
	// 初始化一个os.Signal类型的channel
	// 我们必须使用缓冲通道，否则在信号发送时如果还没有准备好接收信号，就有丢失信号的风险。
	c := make(chan os.Signal, 1)
	// notify用于监听信号
	// 参数1表示接收信号的channel
	// 参数2及后面的表示要监听的信号
	// os.Interrupt 表示中断
	// os.Kill 杀死退出进程
	signal.Notify(c, os.Interrupt, os.Kill)
	// 阻塞直到接收到信息
	s := <-c
	log.Info("quit app Got signal:", s)
}
