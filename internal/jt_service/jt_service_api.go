package jt_service

import (
	log "github.com/sirupsen/logrus"
	"jt809_server/models"
	"jt809_server/util"
	"sync"
)

// VsJTServiceApi 转发到JT809上级平台,单例

type VsJTServiceApi struct {
	ctx *JTServiceContext
}

var api *VsJTServiceApi
var apiOnce sync.Once

// JTServiceIns 获取服务实例
func JTServiceIns() *VsJTServiceApi {
	apiOnce.Do(func() {
		api = &VsJTServiceApi{}
	})
	return api
}

func (_self *VsJTServiceApi) InitApi(notifyFrameWork *util.IotMsgNotifyFramework) {
	_self.ctx = &JTServiceContext{}
	_self.ctx.StartService("vsJTServiceContext", 50000, notifyFrameWork, func() {
		log.Info("vsJTServiceContext start finish")
		_self.ctx.initService()
	})
}

// SendPlatformPostQueryResp 响应平台查岗应答
func (_self *VsJTServiceApi) SendPlatformPostQueryResp(reqObject *models.VsAppMsgPlatformPostQueryResp) {
	_self.ctx.onHandleApiMsgPostQueryResp(reqObject)
}

// SendWarnUrgeTodoResp 响应平台查岗应答
func (_self *VsJTServiceApi) SendWarnUrgeTodoResp(reqObject *models.VsAppMsgWarnUrgeTodoResp) {
	_self.ctx.onHandleApiMsgWarnUrgeTodoResp(reqObject)
}
