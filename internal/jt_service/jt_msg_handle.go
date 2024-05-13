package jt_service

import (
	log "github.com/sirupsen/logrus"
	"jt809_server/internal/data_manage_service"
	jt_session "jt809_server/internal/jt_service/session_mananage"
	"jt809_server/models"
	"jt809_server/util"
	"time"
)

type takePhotoInfo struct {
	isValid        bool                      //是否存在请求
	channelNo      uint8                     //通道号
	picType        uint8                     //图片分辨率
	vehicleNo      string                    //车牌号
	vehicleNoColor uint8                     //车牌号颜色
	srcSubDataType uint16                    //源请求子业务ID
	srcMsgSN       uint32                    //源请求子业务序列号
	takeGpsData    models.VsPbStorageGpsData //抓拍时的定位数据
}

// 车辆任务
type vehicleTask struct {
	takePhotoInfo takePhotoInfo
	lastRealData  models.VsPbStorageGpsData
}

type eventVehicleMapInfo struct {
	vehicleNo string
	startTime int64 //开始时间,用来做超时处理
}

type MsgHandleContext struct {
	parentCtx           *jtUserInfoContext
	lastAliveTime       int64 //最后数据交互时间
	lastMinCheckS       int64
	loginVerifyCode     uint32                          //登录应答较验码
	mapVehicleTask      map[string]*vehicleTask         //key:车牌号
	mapEventIDVehicleNo map[uint64]*eventVehicleMapInfo //key:事件ID,value:车牌号
}

func NewMsgHandleContext(parent *jtUserInfoContext) *MsgHandleContext {
	msgHandleCtx := &MsgHandleContext{}
	msgHandleCtx.parentCtx = parent
	msgHandleCtx.lastMinCheckS = time.Now().Unix()
	msgHandleCtx.mapVehicleTask = make(map[string]*vehicleTask, 5000)
	msgHandleCtx.mapEventIDVehicleNo = make(map[uint64]*eventVehicleMapInfo, 5000)
	return msgHandleCtx
}

func (_self *MsgHandleContext) updateLastAliveTime() {
	_self.lastAliveTime = time.Now().Unix()
}

func (_self *MsgHandleContext) onTimerTaskSecond() {
	isOK := util.HasTimerTask(&_self.lastMinCheckS, 60*1000)
	if isOK {
		_self.SendLinkManageUpKeepAliveReq()
		_self.handleTimerTaskEventVehicleInfo()
	}
}

func (_self *MsgHandleContext) handleTimerTaskEventVehicleInfo() {
	curTime := time.Now().Unix()
	var delKeys []uint64
	for eventID, vehicleInfo := range _self.mapEventIDVehicleNo {
		if curTime-vehicleInfo.startTime > 300 {
			delKeys = append(delKeys, eventID)
		}
	}
	for idx := 0; idx < len(delKeys); idx++ {
		delete(_self.mapEventIDVehicleNo, delKeys[idx])
	}
}

// 获取车辆任务根据事件ID
func (_self *MsgHandleContext) getVehicleTaskInfoByEventID(eventID uint64) *vehicleTask {
	vehicleNo := _self.getVehicleNoByEventID(eventID)
	if vehicleNo == "" {
		return nil
	}
	return _self.getVehicleTaskByVehicleNo(vehicleNo)
}

// 获取车牌号根据事件ID
func (_self *MsgHandleContext) getVehicleNoByEventID(eventID uint64) string {
	vehicleInfo, isExist := _self.mapEventIDVehicleNo[eventID]
	if isExist {
		//获取完就清除
		vehicleNo := vehicleInfo.vehicleNo
		delete(_self.mapEventIDVehicleNo, eventID)
		return vehicleNo
	}
	return ""
}

// 获取车辆的任务信息
func (_self *MsgHandleContext) getVehicleTaskByVehicleNo(vehicleNo string) *vehicleTask {
	vehicleTaskObject, isExist := _self.mapVehicleTask[vehicleNo]
	if isExist {
		return vehicleTaskObject
	}
	return nil
}

func (_self *MsgHandleContext) sendDataToServer(dataType models.VsJTDataType, bodyData []byte, isMainLink bool) bool {
	if _self.parentCtx == nil {
		log.Error("MsgHandleContext-sendDataToServer fail,parentCtx is nil")
		return false
	}
	return _self.parentCtx.SendDataToServer(dataType, bodyData, isMainLink)
}

func (_self *MsgHandleContext) getJTVer() jt_session.JT_VER_TYPE {
	if _self.parentCtx == nil {
		log.Error("MsgHandleContext-getJTVer fail,parentCtx is nil")
		return jt_session.JT_VER_NONE
	}
	return jt_session.JT_VER_TYPE(_self.parentCtx.conf.JTVer)
}

// 获取平台编码
func (_self *MsgHandleContext) getPlatformCode() string {
	if _self.parentCtx == nil {
		log.Error("MsgHandleContext-getPlatformCode fail,parentCtx is nil")
		return ""
	}
	return _self.parentCtx.conf.JTPlatformCode
}

func (_self *MsgHandleContext) addSubCommonHeadInfo(subDataType models.VsJTDataType, vehicleID uint32) (*util.VsBinaryWrite, *models.VsVehicleBaseInfo) {
	bodyObject := util.NewBinaryWrite()
	vehicleInfo := data_manage_service.DataManageServiceIns().GetVehicleInfoByID(vehicleID)
	if vehicleInfo == nil {
		log.Error("addSubCommonHeadInfo fail,not found vehicleID:", vehicleID, " subDataType:", subDataType)
		return nil, nil
	}
	gbkVehicleNo := vehicleInfo.VehicleNo
	util.VsUtf8ToGBK(&gbkVehicleNo)
	bodyObject.AppendFixedBytes(gbkVehicleNo, models.VS_JT_VEHCLE_NO_LEN)
	bodyObject.AppendByte(vehicleInfo.VehicleNoColor)
	bodyObject.AppendNumber(uint16(subDataType)) //子业务类型
	return bodyObject, vehicleInfo
}

// 构建消息体并发送到服务器
func (_self *MsgHandleContext) buildBodyObjectAndSend(dataType, subDataType models.VsJTDataType, dataBuf []byte) {
	bodyObject := util.NewBinaryWrite()
	bodyObject.AppendNumber(uint16(subDataType))
	bodyObject.AppendNumber(uint32(len(dataBuf)))
	bodyObject.AppendBytes(dataBuf)
	_self.sendDataToServer(dataType, bodyObject.GetData(), true)
}

// 构建消息体并发送到服务器 存在车牌号的消息体
func (_self *MsgHandleContext) buildBodyObjectByVehicleInfoAndSend(dataType, subDataType models.VsJTDataType, dataBuf []byte, vehicleNo string, vehicleNoColor uint8) {
	bodyObject := util.NewBinaryWrite()
	gbkVehicleNo := vehicleNo
	util.VsUtf8ToGBK(&gbkVehicleNo)
	bodyObject.AppendFixedBytes(gbkVehicleNo, models.VS_JT_VEHCLE_NO_LEN)
	bodyObject.AppendByte(vehicleNoColor)
	bodyObject.AppendNumber(uint16(subDataType))
	bodyObject.AppendNumber(uint32(len(dataBuf)))
	bodyObject.AppendBytes(dataBuf)
	_self.sendDataToServer(dataType, bodyObject.GetData(), true)
}

// 构建消息体并发送到服务器 根据车辆ID
func (_self *MsgHandleContext) buildBodyObjectByVehicleGuidAndSend(dataType, subDataType models.VsJTDataType, dataBuf []byte, vehicleID uint32) bool {
	vehicleInfo := data_manage_service.DataManageServiceIns().GetVehicleInfoByID(vehicleID)
	if vehicleInfo == nil {
		log.Error("buildBodyObjectByVehicleGuid fail,not found vehicleID:", vehicleID, " subDataType:", subDataType)
		return false
	}
	bodyObject := util.NewBinaryWrite()

	gbkVehicleNo := vehicleInfo.VehicleNo
	util.VsUtf8ToGBK(&gbkVehicleNo)
	bodyObject.AppendFixedBytes(gbkVehicleNo, models.VS_JT_VEHCLE_NO_LEN)
	bodyObject.AppendByte(vehicleInfo.VehicleNoColor)
	bodyObject.AppendNumber(uint16(subDataType)) //子业务类型
	bodyObject.AppendNumber(uint32(len(dataBuf)))
	bodyObject.AppendBytes(dataBuf)
	_self.sendDataToServer(dataType, bodyObject.GetData(), true)
	return true
}
