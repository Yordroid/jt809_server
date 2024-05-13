package jt_service

import (
	log "github.com/sirupsen/logrus"
	"jt809_server/internal/bu_service"
	jt_session "jt809_server/internal/jt_service/session_mananage"
	"jt809_server/models"
	"jt809_server/util"
)

//平台间的交互消息

// 平台查岗
func (_self *MsgHandleContext) onHandlePlatformMsgPostQueryReq(hdrObject jt_session.VsAppMsgHeader, bodyMsg []byte) bool {
	jtVer := _self.getJTVer()
	readObject := util.NewBinaryRead(bodyMsg)
	reqObject := models.VsAppMsgPlatformPostQueryReq{}

	if jtVer == jt_session.JT_VER_V3 {
		reqObject.ObjectType = readObject.ReadByte() //查岗对象类型
		reqObject.ObjectID = readObject.ReadString(20)
		reqObject.AnswerTime = readObject.ReadByte()
	} else if jtVer == jt_session.JT_VER_V2 {
		reqObject.ObjectType = readObject.ReadByte() //查岗对象类型
		reqObject.ObjectID = readObject.ReadString(12)
	} else { //V1
	}
	reqObject.InfoID = readObject.ReadInt32()
	var infoLen uint32
	infoLen = readObject.ReadInt32()
	reqObject.InfoContent = readObject.ReadStringToUTF8(infoLen)
	reqInfo := userReqDataContext{}
	reqInfo.srcMsgSN = hdrObject.MsgSN
	reqInfo.srcSubDataType = hdrObject.MsgID
	taskID := util.GetSnowflakeID()
	reqObject.TaskID = models.GetTaskIDToString(taskID)
	_self.parentCtx.parent.addTaskMap(taskID, _self.parentCtx.conf.UserKey, reqInfo)
	strJson := util.ToJson(reqObject)
	log.Info("onHandlePlatformMsgPostQuery :", strJson)
	bu_service.MsgServiceIns().AsyncSendPlatformPostQueryReq(_self.parentCtx.conf.UserKey, &reqObject)
	return true
}

// SendPlatformMsgPostQueryAck 平台查岗应答
func (_self *MsgHandleContext) SendPlatformMsgPostQueryAck(reqObject *models.VsAppMsgPlatformPostQueryResp, userData userReqDataContext) {
	dataObject := util.NewBinaryWrite()
	jtVer := _self.getJTVer()
	if jtVer == jt_session.JT_VER_V3 {
		dataObject.AppendByte(reqObject.ObjectType)
		ackNameGbk := reqObject.AckName
		util.VsUtf8ToGBK(&ackNameGbk)
		dataObject.AppendFixedBytes(ackNameGbk, 16)
		dataObject.AppendFixedBytes(reqObject.AckPhone, 20)
		dataObject.AppendFixedBytes(reqObject.ObjectID, 20)
		dataObject.AppendNumber(userData.srcSubDataType)
		dataObject.AppendNumber(userData.srcMsgSN)
		gbkContent := reqObject.Content
		util.VsUtf8ToGBK(&gbkContent)
		dataObject.AppendNumber(uint32(len(gbkContent)))
		dataObject.AppendString(gbkContent)
	} else if jtVer == jt_session.JT_VER_V2 {
		dataObject.AppendByte(reqObject.ObjectType)
		dataObject.AppendFixedBytes(reqObject.ObjectID, 12)
		dataObject.AppendNumber(reqObject.InfoID)
	} else { //V1
		dataObject.AppendNumber(reqObject.InfoID)
	}
	gbkContent := reqObject.Content
	util.VsUtf8ToGBK(&gbkContent)
	dataObject.AppendNumber(uint32(len(gbkContent)))
	dataObject.AppendString(gbkContent)
	_self.buildBodyObjectAndSend(models.VS_JT_TYPE_UP_PLAFORM_MSG, models.VS_JT_TYPE_UP_PLAFORM_MSG_POST_QUERY_ACK, dataObject.GetData())
	log.Info("SendPlatformMsgPostQueryAck :", util.ToJson(reqObject))
}

// 平台报文请求
func (_self *MsgHandleContext) onHandlePlatformMsgTextInfoReq(hdrObject jt_session.VsAppMsgHeader, bodyMsg []byte) bool {
	jtVer := _self.getJTVer()
	readObject := util.NewBinaryRead(bodyMsg)
	reqObject := models.VsAppMsgPlatformTextReq{}
	if jtVer == jt_session.JT_VER_V3 {
		reqObject.ObjectType = readObject.ReadByte() //对象类型
		reqObject.ObjectID = readObject.ReadString(20)

	} else if jtVer == jt_session.JT_VER_V2 {
		reqObject.ObjectType = readObject.ReadByte() //对象类型
		reqObject.ObjectID = readObject.ReadString(12)
	} else { //V1
	}
	reqObject.InfoID = readObject.ReadInt32()
	var infoLen uint32
	infoLen = readObject.ReadInt32()
	reqObject.InfoContent = readObject.ReadStringToUTF8(infoLen)
	log.Info("onHandlePlatformMsgTextInfoReq", util.ToJson(reqObject))
	bu_service.MsgServiceIns().AsyncSendPlatformMsgTextInfoReq(_self.parentCtx.conf.UserKey, &reqObject)
	_self.SendPlatformMsgTextInfoResp(hdrObject.MsgSN, hdrObject.MsgID, reqObject.InfoID)
	return true
}

// SendPlatformMsgTextInfoResp 平台报文应答
func (_self *MsgHandleContext) SendPlatformMsgTextInfoResp(srcMsgSN uint32, srcSubDataType uint16, infoID uint32) {
	dataObject := util.NewBinaryWrite()
	jtVer := _self.getJTVer()
	if jtVer == jt_session.JT_VER_V3 {
		dataObject.AppendNumber(srcSubDataType)
		dataObject.AppendNumber(srcMsgSN)
	} else { //V1 ,V2
		dataObject.AppendNumber(infoID)
	}
	_self.buildBodyObjectAndSend(models.VS_JT_TYPE_UP_PLAFORM_MSG, models.VS_JT_TYPE_UP_PLAFORM_MSG_INFO_ACK, dataObject.GetData())
	log.Info("SendPlatformMsgTextInfoResp")
}
