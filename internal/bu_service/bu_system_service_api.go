package bu_service

import (
	log "github.com/sirupsen/logrus"
	"jt809_server/config"
	"jt809_server/models"
	"jt809_server/util"
	"sync"
)

//业务系统管理,用来登录业务系统,获取基础数据,MQ消息接入管理,第三方业务系统只需要改对这个

// VsBuServiceApi 业务系统服务对外接口,单例

type VsBuSystemServiceApi struct {
	ctx buSystemBase
}

var api *VsBuSystemServiceApi
var apiOnce sync.Once

// MsgServiceIns 获取消息接入实例
func MsgServiceIns() *VsBuSystemServiceApi {
	apiOnce.Do(func() {
		api = &VsBuSystemServiceApi{}
	})
	return api
}

// vsBuType 业务系统类型,可以通过配置文件来更改
type vsBuType int

const (
	VS_BU_TYPE_SELF_V4  vsBuType = 0 //对接自己V4平台
	VS_BU_TYPE_OTHER_V1 vsBuType = 1 //其它平台
)

func (_self *VsBuSystemServiceApi) InitApi(notifyFrameWork *util.IotMsgNotifyFramework) {
	buType := config.GetBuType()
	log.Info("init bu type:", buType)
	if VS_BU_TYPE_OTHER_V1 == vsBuType(buType) {

	} else { //默认 VS_BU_TYPE_SELF_V4
		selfBuCtx := &buSystemServiceContext{}
		selfBuCtx.StartService("buSystemService", 50000, notifyFrameWork, func() {
			log.Info("buSystemService start finish")
			selfBuCtx.initService()
		})
		_self.ctx = selfBuCtx
	}

}

// AddBuUserInfo 添加用户信息
func (_self *VsBuSystemServiceApi) AddBuUserInfo(userName, userPwdMd5 string, userKey int64) {
	if _self.ctx == nil {
		log.Error("AddBuUserInfo fail,not init bu ctx")
		return
	}
	_self.ctx.addBuUserInfo(userName, userPwdMd5, userKey)
}

// DeleteBuUserInfo 删除用户信息
func (_self *VsBuSystemServiceApi) DeleteBuUserInfo(userKey int64) {
	if _self.ctx == nil {
		log.Error("DeleteBuUserInfo fail,not init bu ctx")
		return
	}
	_self.ctx.deleteBuUserInfo(userKey)
}

//所有http接口,都需要开一个协程来调用,如果需要结果的,由调用方创建,因为可能要根据结果做数据同步

// SendDevTakePhotoReq 请求抓拍
func (_self *VsBuSystemServiceApi) SendDevTakePhotoReq(userKey int64, reqObject *models.VsAppTakePhotoReq) bool {
	if _self.ctx == nil {
		log.Error("SendDevTakePhotoReq fail,not init bu ctx")
		return false
	}
	return _self.ctx.sendDevTakePhotoReq(userKey, reqObject)
}

// AsyncSendPlatformPostQueryReq 上级平台查岗
func (_self *VsBuSystemServiceApi) AsyncSendPlatformPostQueryReq(userKey int64, reqObject *models.VsAppMsgPlatformPostQueryReq) {
	if _self.ctx == nil {
		log.Error("AsyncSendPlatformPostQueryReq fail,not init bu ctx")
		return
	}
	go func() {
		_self.ctx.sendPlatformPostQueryReq(userKey, reqObject)
	}()
}

// AsyncSendPlatformMsgTextInfoReq 上级平台报文下发
func (_self *VsBuSystemServiceApi) AsyncSendPlatformMsgTextInfoReq(userKey int64, reqObject *models.VsAppMsgPlatformTextReq) {
	if _self.ctx == nil {
		log.Error("AsyncSendPlatformMsgTextInfoReq fail,not init bu ctx")
		return
	}
	go func() {
		_self.ctx.sendPlatformMsgTextInfoReq(userKey, reqObject)
	}()

}

// AsyncSendPlatformWarnUrgeTodoReq 报警督办请求
func (_self *VsBuSystemServiceApi) AsyncSendPlatformWarnUrgeTodoReq(userKey int64, reqObject *models.VsAppMsgWarnUrgeTodoReq) {
	if _self.ctx == nil {
		log.Error("AsyncSendPlatformWarnUrgeTodoReq fail,not init bu ctx")
		return
	}
	go func() {
		_self.ctx.sendPlatformWarnUrgeTodoReq(userKey, reqObject)
	}()

}
