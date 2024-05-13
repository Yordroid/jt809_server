package jt_service

import (
	log "github.com/sirupsen/logrus"
	"jt809_server/internal/bu_service"
	jt_session "jt809_server/internal/jt_service/session_mananage"
	"jt809_server/models"
	"jt809_server/util"
)

//监管类

func (_self *MsgHandleContext) addTakePhotoTask(vehicleNo string, vehicleNoColor, picType, channelNo uint8, srcMsgSN uint32, srcSubDataType uint16) *vehicleTask {
	vehicleTaskObject, isExist := _self.mapVehicleTask[vehicleNo]
	if !isExist {
		vehicleTaskObject = &vehicleTask{}
		_self.mapVehicleTask[vehicleNo] = vehicleTaskObject
	}
	vehicleTaskObject.takePhotoInfo.isValid = true
	vehicleTaskObject.takePhotoInfo.vehicleNo = vehicleNo
	vehicleTaskObject.takePhotoInfo.channelNo = channelNo
	vehicleTaskObject.takePhotoInfo.vehicleNoColor = vehicleNoColor
	vehicleTaskObject.takePhotoInfo.picType = picType
	vehicleTaskObject.takePhotoInfo.srcMsgSN = srcMsgSN
	vehicleTaskObject.takePhotoInfo.srcSubDataType = srcSubDataType
	return vehicleTaskObject
}
func (_self *MsgHandleContext) deleteTakePhotoTask(vehicleNo string) {
	vehicleTaskObject := _self.getVehicleTaskByVehicleNo(vehicleNo)
	if vehicleTaskObject != nil {
		vehicleTaskObject.takePhotoInfo.isValid = false
	}
}

const (
	VS_TAKE_PHOTO_RESULT_NOT_SUPPORT      = 0 //不支持拍照
	VS_TAKE_PHOTO_RESULT_FINISH           = 1 //完成拍照
	VS_TAKE_PHOTO_RESULT_FINISH_WAIT_SEND = 2 //完成拍照,稍后发送
	VS_TAKE_PHOTO_RESULT_OFFLINE          = 3 //离线
	VS_TAKE_PHOTO_RESULT_CHANNEL          = 4 //通道
	VS_TAKE_PHOTO_RESULT_OTHER            = 5 //其它
	VS_TAKE_PHOTO_RESULT_VEHICLE_NO       = 6 //车牌号
)

// 车辆拍照请求
func (_self *MsgHandleContext) onHandleCtrlMsgTakePhotoReq(hdrObject jt_session.VsAppMsgHeader, bodyMsg []byte) bool {

	reqObject := models.VsAppTakePhotoReq{}
	readObject := util.NewBinaryRead(bodyMsg)
	reqObject.ChannelNo = readObject.ReadByte()
	reqObject.PicType = readObject.ReadByte()
	reqObject.VehicleNo = hdrObject.VehicleNo
	vehicleTaskObject := _self.addTakePhotoTask(hdrObject.VehicleNo, hdrObject.VehicleNoColor, reqObject.PicType, reqObject.ChannelNo, hdrObject.MsgSN, hdrObject.MsgID)
	log.Info("onHandleCtrlMsgTakePhotoReq", vehicleTaskObject.takePhotoInfo)
	go util.GoTaskFunc("onHandleCtrlMsgTakePhotoReq", func() {
		isOK := bu_service.MsgServiceIns().SendDevTakePhotoReq(_self.parentCtx.conf.UserKey, &reqObject)
		vehicleTaskObject.takePhotoInfo.takeGpsData = vehicleTaskObject.lastRealData
		if isOK {
			_self.sendCtrlMsgTakePhotoResp(vehicleTaskObject.takePhotoInfo, VS_TAKE_PHOTO_RESULT_FINISH_WAIT_SEND, nil)
		} else {
			_self.sendCtrlMsgTakePhotoResp(vehicleTaskObject.takePhotoInfo, VS_TAKE_PHOTO_RESULT_OTHER, nil)
			_self.parentCtx.postToService(func() {
				_self.deleteTakePhotoTask(reqObject.VehicleNo)
			})
		}

	})
	return true
}

func (_self *MsgHandleContext) sendCtrlMsgTakePhotoResp(info takePhotoInfo, result uint8, picData []byte) bool {
	dataObject := util.NewBinaryWrite()

	jtVer := _self.getJTVer()
	if jtVer == jt_session.JT_VER_V3 {
		dataObject.AppendNumber(info.srcSubDataType)
		dataObject.AppendNumber(info.srcMsgSN)
		dataObject.AppendByte(result)
		gpsBuf := _self.convertGpsData2019(&info.takeGpsData)
		dataObject.AppendByte(0)
		dataObject.AppendNumber(uint32(len(gpsBuf)))
		dataObject.AppendBytes(gpsBuf)
		dataObject.AppendFixedBytes(_self.getPlatformCode(), 11)
		dataObject.AppendNumber(info.takeGpsData.AlarmFlag)
		dataObject.AppendFixedBytes("", 11)
		dataObject.AppendNumber(uint32(0))
		dataObject.AppendFixedBytes("", 11)
		dataObject.AppendNumber(uint32(0))
	} else {
		dataObject.AppendByte(result)
		gpsBuf := _self.convertGpsData2011(&info.takeGpsData)
		dataObject.AppendBytes(gpsBuf)
	}
	dataObject.AppendByte(info.channelNo)
	dataObject.AppendNumber(uint32(len(picData)))
	dataObject.AppendByte(info.picType)
	dataObject.AppendByte(1) //1:jpg,2:gif,3:tiff,4:png
	if len(picData) > 0 {
		dataObject.AppendBytes(picData)
	}
	log.Info("sendCtrlMsgTakePhotoResp result:", result, " picLen:", len(picData), info)
	_self.buildBodyObjectByVehicleInfoAndSend(models.VS_JT_TYPE_UP_CTRL_MSG, models.VS_JT_TYPE_UP_CTRL_MSG_TAKE_PHOTO_ACK,
		dataObject.GetData(), info.vehicleNo, info.vehicleNoColor)
	return true
}
