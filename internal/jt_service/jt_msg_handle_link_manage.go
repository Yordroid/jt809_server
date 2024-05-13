package jt_service

import (
	log "github.com/sirupsen/logrus"
	jt_session "jt809_server/internal/jt_service/session_mananage"
	"jt809_server/models"
	"jt809_server/util"
)

//链路管理类

// SendLinkManageUpLoginReq 主链路登录请求
func (_self *MsgHandleContext) SendLinkManageUpLoginReq(reqObject *models.VsAppMsgLoginInfoReq) {
	bodyObject := util.NewBinaryWrite()
	bodyObject.AppendNumber(reqObject.UserID)
	bodyObject.AppendFixedBytes(reqObject.Password, 8)
	bodyObject.AppendNumber(reqObject.AccessID)
	bodyObject.AppendFixedBytes(reqObject.DownLinkIP, 32)
	bodyObject.AppendNumber(reqObject.DownLinkPort)
	_self.sendDataToServer(models.VS_JT_TYPE_UP_LOGIN_REQ, bodyObject.GetData(), true)
	log.Info("SendLinkManageUpLoginReq verifyCode:", util.ToJson(reqObject))
}

// onHandleLinkManageUpLoginResp 登录应答
func (_self *MsgHandleContext) onHandleLinkManageUpLoginResp(hdrObject jt_session.VsAppMsgHeader, bodyMsg []byte) bool {
	respObject := models.VsAppMsgLoginResp{}
	readObject := util.NewBinaryRead(bodyMsg)
	respObject.Result = readObject.ReadByte()
	respObject.VerifyCode = readObject.ReadInt32()
	_self.loginVerifyCode = respObject.VerifyCode
	log.Info("onHandleLoginResp result:", respObject.Result, " verifyCode:", respObject.VerifyCode)
	return true
}

// SendLinkManageUpKeepAliveReq 主链路连接保持请求
// 下级平台登录成功后,如与上级平台之间有应用业务数据包往来
// 则不需要发送主链路连接保持请求数据包;否则,下级平台应每分钟向上级平台发送一个主链路连接保持请求数据包,以保持主链路连接
func (_self *MsgHandleContext) SendLinkManageUpKeepAliveReq() {
	_self.sendDataToServer(models.VS_JT_TYPE_UP_LINK_KEEP_ALIVE_REQ, nil, true)
	log.Info("SendUpKeepAliveReq")
}

// onHandleLinkManageUpKeepLiveResp 主链路保持应答
func (_self *MsgHandleContext) onHandleLinkManageUpKeepLiveResp(hdrObject jt_session.VsAppMsgHeader, bodyMsg []byte) bool {
	log.Info("onHandleLinkManageUpKeepLiveResp")
	return true
}

// onHandleLinkManageDownLoginReq 从链路登录请求
func (_self *MsgHandleContext) onHandleLinkManageDownLoginReq(hdrObject jt_session.VsAppMsgHeader, bodyMsg []byte) bool {
	readObject := util.NewBinaryRead(bodyMsg)
	verifyCode := readObject.ReadInt32()
	bodyObject := util.NewBinaryWrite()
	if _self.loginVerifyCode > 0 && verifyCode == _self.loginVerifyCode {
		bodyObject.AppendNumber(uint8(0)) //0:成功,1:verifyCode不匹配,2:资源紧张,FF:其它
	} else {
		bodyObject.AppendNumber(uint8(1))
	}
	_self.sendDataToServer(models.VS_JT_TYPE_DOWN_LOGIN_RESP, bodyObject.GetData(), false)
	log.Info("onHandleLinkManageDownLoginReq verifyCode:", verifyCode, " loginVerify:", _self.loginVerifyCode)
	return true
}

// onHandleLinkManageDownKeepLiveReq 从链路保持请求
func (_self *MsgHandleContext) onHandleLinkManageDownKeepLiveReq(hdrObject jt_session.VsAppMsgHeader, bodyMsg []byte) bool {
	_self.sendDataToServer(models.VS_JT_TYPE_DOWN_LINK_KEEP_ALIVE_RESP, nil, false)
	log.Info("onHandleLinkManageDownKeepLiveReq")
	return true
}
