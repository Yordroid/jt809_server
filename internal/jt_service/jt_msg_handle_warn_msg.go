package jt_service

//车辆报警信息类

import (
	log "github.com/sirupsen/logrus"
	"jt809_server/internal/bu_service"
	"jt809_server/internal/data_manage_service"
	jt_session "jt809_server/internal/jt_service/session_mananage"
	"jt809_server/models"
	"jt809_server/util"
	"time"
)

// SendWarnMsgEventDataUpload 上报报警信息
func (_self *MsgHandleContext) SendWarnMsgEventDataUpload(reqObject *models.VsPbStorageEventInfo) bool {
	jtVer := _self.getJTVer()
	jtAlarmType, alarmString := models.ToJT809AlarmCode(reqObject.AlarmType, jtVer)
	if jtAlarmType == models.VS_JT_ALARM_CODE_NONE {
		return false
	}
	dataObject := util.NewBinaryWrite()
	if jtVer == jt_session.JT_VER_V3 {
		dataObject.AppendFixedBytes(_self.getPlatformCode(), 11)

		dataObject.AppendNumber(uint16(jtAlarmType))
		dataObject.AppendNumber(uint64(reqObject.Time))
		dataObject.AppendNumber(uint64(reqObject.Time))
		if reqObject.ExternInfo != nil {
			dataObject.AppendNumber(uint64(reqObject.ExternInfo.EndTime))
		}
		dataObject.AppendFixedBytes(reqObject.VehicleNo, models.VS_JT_VEHCLE_NO_LEN)
		var lineGuid uint32
		vehicleInfo := data_manage_service.DataManageServiceIns().GetVehicleInfoByID(reqObject.VehicleGuid)
		if vehicleInfo != nil {
			dataObject.AppendByte(vehicleInfo.VehicleNoColor)
			lineGuid = vehicleInfo.LineGuid
		} else {
			dataObject.AppendByte(0)
		}
		dataObject.AppendFixedBytes("", 11)
		dataObject.AppendNumber(lineGuid)
		targetAlarmString := alarmString
		util.VsUtf8ToGBK(&targetAlarmString)
		dataObject.AppendNumber(uint32(len(targetAlarmString)))
		if len(targetAlarmString) > 0 {
			dataObject.AppendString(targetAlarmString)
		}
		_self.buildBodyObjectAndSend(models.VS_JT_TYPE_UP_WARN_MSG,
			models.VS_JT_TYPE_UP_WARN_MSG_ADPT_INFO, dataObject.GetData())
	} else {
		dataObject.AppendByte(1) //报警信息来源
		if jtAlarmType > 7 {     //老协议不存在超过7的类型
			return false
		}
		dataObject.AppendNumber(uint16(jtAlarmType))
		dataObject.AppendNumber(reqObject.Time)
		dataObject.AppendNumber(uint32(time.Now().Unix())) //信息ID
		targetAlarmString := alarmString
		util.VsUtf8ToGBK(&targetAlarmString)
		dataObject.AppendNumber(uint32(len(targetAlarmString)))
		if len(targetAlarmString) > 0 {
			dataObject.AppendString(targetAlarmString)
		}
		_self.buildBodyObjectByVehicleGuidAndSend(models.VS_JT_TYPE_UP_WARN_MSG,
			models.VS_JT_TYPE_UP_WARN_MSG_ADPT_INFO, dataObject.GetData(), reqObject.VehicleGuid)
	}

	return true
}

// 报警督办请求
func (_self *MsgHandleContext) onHandleWarnMsgUrgeTodoReq(hdrObject jt_session.VsAppMsgHeader, bodyMsg []byte) bool {
	reqObject := models.VsAppMsgWarnUrgeTodoReq{}
	jtVer := _self.getJTVer()
	readObject := util.NewBinaryRead(bodyMsg)
	reqInfo := userReqDataContext{}
	if jtVer == jt_session.JT_VER_V3 {
		reqObject.PlatformID = readObject.ReadString(11)
		reqObject.AlarmTime = uint32(readObject.ReadInt64())
		reqInfo.srcSubDataType = readObject.ReadInt16()
		reqInfo.srcMsgSN = readObject.ReadInt32()
		//以下两个新协议不存在
		reqObject.SuperVisionID = uint32(time.Now().Unix())
		reqObject.AlarmSrc = 9 //其它
		reqObject.ReqMsgSN = hdrObject.MsgSN
	} else { //V1,V2
		reqObject.AlarmSrc = readObject.ReadByte()
		isOK, appType := models.ToAppAlarmCode(models.VS_JT_ALARM_CODE(readObject.ReadInt16()), jtVer)
		if isOK {
			reqObject.AlarmType = uint16(appType)
		}
		reqObject.AlarmTime = uint32(readObject.ReadInt64())
		reqObject.SuperVisionID = readObject.ReadInt32()
		reqObject.PlatformID = _self.getPlatformCode()
		reqInfo.vehicleNo = hdrObject.VehicleNo
		reqInfo.vehicleNoColor = hdrObject.VehicleNoColor
	}
	reqObject.SupervisionEndTime = uint32(readObject.ReadInt64())
	reqObject.SuperVisionLevel = readObject.ReadByte()
	reqObject.SuperVisor = readObject.ReadStringToUTF8(16)
	reqObject.SuperVisorTel = readObject.ReadString(20)
	reqObject.SuperVisorEmail = readObject.ReadString(32)
	taskID := util.GetSnowflakeID()
	reqObject.TaskID = models.GetTaskIDToString(taskID)
	_self.parentCtx.parent.addTaskMap(taskID, _self.parentCtx.conf.UserKey, reqInfo)
	log.Info("onHandleWarnMsgUrgeTodoReq ", util.ToJson(reqObject))
	bu_service.MsgServiceIns().AsyncSendPlatformWarnUrgeTodoReq(_self.parentCtx.conf.UserKey, &reqObject)
	return true
}

func (_self *MsgHandleContext) SendWarnMsgUrgeTodoResp(reqObject *models.VsAppMsgWarnUrgeTodoResp, userData userReqDataContext) bool {
	jtVer := _self.getJTVer()

	dataObject := util.NewBinaryWrite()
	if jtVer == jt_session.JT_VER_V3 {
		dataObject.AppendNumber(reqObject.ReqMsgSN)
		dataObject.AppendNumber(userData.srcSubDataType)
		dataObject.AppendNumber(userData.srcMsgSN)
		dataObject.AppendByte(reqObject.Result)
		_self.buildBodyObjectAndSend(models.VS_JT_TYPE_UP_WARN_MSG,
			models.VS_JT_TYPE_UP_WARN_MSG_URGE_TODO_ACK, dataObject.GetData())
	} else {
		dataObject.AppendNumber(reqObject.SuperVisionID)
		dataObject.AppendByte(reqObject.Result)
		_self.buildBodyObjectByVehicleInfoAndSend(models.VS_JT_TYPE_UP_WARN_MSG,
			models.VS_JT_TYPE_UP_WARN_MSG_URGE_TODO_ACK, dataObject.GetData(), userData.vehicleNo, userData.vehicleNoColor)
	}
	log.Info("SendWarnMsgUrgeTodoResp ", util.ToJson(reqObject))
	return true
}
