package jt_service

import (
	log "github.com/sirupsen/logrus"
	jt_session "jt809_server/internal/jt_service/session_mananage"
	"jt809_server/models"
	"jt809_server/util"
	"time"
)

//车辆动态信息交换类

// SendExgMsgUpVehicleOnlineUpload 上传车辆注册数据,车辆上线,需要发
func (_self *MsgHandleContext) SendExgMsgUpVehicleOnlineUpload(reqObject *models.VsFrontPbMsgOnline) bool {
	if !reqObject.IsOnline {
		return false
	}
	//上线才需要
	jtVer := _self.getJTVer()
	bodyObject, vehicleInfo := _self.addSubCommonHeadInfo(models.VS_JT_TYPE_UP_EXG_MSG_REGISTER, reqObject.VehicleGuid)
	if bodyObject == nil || vehicleInfo == nil {
		return false
	}
	if jtVer == jt_session.JT_VER_V3 {
		bodyObject.AppendNumber(uint32(110))                     //后续数据长度
		bodyObject.AppendFixedBytes("1234534", 11)               //平台唯一编码
		bodyObject.AppendFixedBytes("11012033", 11)              //车载终端厂商编码
		bodyObject.AppendFixedBytes("X-Link100", 30)             //终端型号
		bodyObject.AppendFixedBytes(vehicleInfo.DeviceIMEI, 15)  //IMEI
		bodyObject.AppendFixedBytes(vehicleInfo.DeviceNo, 30)    //终端号编码
		bodyObject.AppendFixedBytesBefore(vehicleInfo.SimNo, 13) //SIM卡号
	} else if jtVer == jt_session.JT_VER_V2 {
		bodyObject.AppendNumber(uint32(61))                      //后续数据长度
		bodyObject.AppendFixedBytes("1234534", 11)               //平台唯一编码
		bodyObject.AppendFixedBytes("11012033", 11)              //车载终端厂商编码
		bodyObject.AppendFixedBytes("X-Link100", 20)             //终端型号
		bodyObject.AppendFixedBytes(vehicleInfo.DeviceIMEI, 15)  //IMEI
		bodyObject.AppendFixedBytes(vehicleInfo.DeviceNo, 7)     //终端号编码
		bodyObject.AppendFixedBytesBefore(vehicleInfo.SimNo, 12) //SIM卡号
	} else if jtVer == jt_session.JT_VER_V1 {
		bodyObject.AppendNumber(uint32(49))                      //后续数据长度
		bodyObject.AppendFixedBytes("1234534", 11)               //平台唯一编码
		bodyObject.AppendFixedBytes("11012033", 11)              //车载终端厂商编码
		bodyObject.AppendFixedBytes("X-Link100", 8)              //终端型号
		bodyObject.AppendFixedBytes(vehicleInfo.DeviceNo, 7)     //终端号编码
		bodyObject.AppendFixedBytesBefore(vehicleInfo.SimNo, 12) //SIM卡号
	}

	_self.sendDataToServer(models.VS_JT_TYPE_UP_EXG_MSG, bodyObject.GetData(), true)
	return true
}

type VsGpsExternItemID uint8

const (
	VS_EXTERN_ITEM_MILEAGE        = 0x01 //里程 uint32
	VS_EXTERN_ITEM_VEHICLE_SIGNAL = 0x25 //车辆状态 uint32
	VS_EXTERN_ITEM_NET_SIGNAL     = 0x30 //网络信号强度 uint8
	VS_EXTERN_ITEM_SATELLITENUM   = 0x31 //卫星颗数 uint8
)

func (_self *MsgHandleContext) addGpsExternItem(bodyObject *util.VsBinaryWrite, itemID VsGpsExternItemID, itemLen int, itemData any) {
	bodyObject.AppendByte(uint8(itemID))
	bodyObject.AppendByte(uint8(itemLen))
	switch v := itemData.(type) {
	case uint64, int64, uint32, int32, uint16, int16, uint8, int8:
		bodyObject.AppendNumber(itemData)
	case string:
		bodyObject.AppendString(itemData.(string))
	case []byte:
		bodyObject.AppendBytes(itemData.([]byte))
	default:
		log.Error("addGpsExternItem type not support,app will quit", v, " itemID:", itemID)
		time.Sleep(time.Second)
		panic("addGpsExternItem type not support")
		return
	}

}

func (_self *MsgHandleContext) convertGpsData2019(reqObject *models.VsPbStorageGpsData) []byte {
	//基本的定位信息
	gpsDataBuf := util.NewBinaryWrite()
	gpsDataBuf.AppendNumber(reqObject.AlarmFlag)
	gpsDataBuf.AppendNumber(reqObject.StatusFlag)
	gpsDataBuf.AppendNumber(uint32(reqObject.Latitude))
	gpsDataBuf.AppendNumber(uint32(reqObject.Longitude))
	gpsDataBuf.AppendNumber(uint16(reqObject.Altitude))
	gpsDataBuf.AppendNumber(uint16(reqObject.Speed))
	gpsDataBuf.AppendNumber(uint16(reqObject.Direction))
	gpsDataBuf.AppendBytes(util.TimeToBCD(int64(reqObject.Time)))
	//附加数据
	_self.addGpsExternItem(gpsDataBuf, VS_EXTERN_ITEM_MILEAGE, 4, reqObject.PlatformMileage)
	_self.addGpsExternItem(gpsDataBuf, VS_EXTERN_ITEM_VEHICLE_SIGNAL, 4, reqObject.VehicleExternStatus)
	_self.addGpsExternItem(gpsDataBuf, VS_EXTERN_ITEM_NET_SIGNAL, 1, uint8(reqObject.NetSignal))
	_self.addGpsExternItem(gpsDataBuf, VS_EXTERN_ITEM_SATELLITENUM, 1, uint8(reqObject.SatelliteNum))
	return gpsDataBuf.GetData()
}
func (_self *MsgHandleContext) convertGpsData2011(reqObject *models.VsPbStorageGpsData) []byte {
	//基本的定位信息
	gpsDataBuf := util.NewBinaryWrite()
	gpsDataBuf.AppendByte(0)
	tmTime := time.Unix(int64(reqObject.Time), 0)
	dateTime := []byte{0, 0, 0, 0, 0, 0, 0}
	dateTime[0] = uint8(tmTime.Day())
	dateTime[1] = uint8(tmTime.Month())
	dateTime[2] = uint8(tmTime.Year() >> 8)
	dateTime[3] = uint8(tmTime.Year() & 0xFF)
	dateTime[4] = uint8(tmTime.Hour())
	dateTime[5] = uint8(tmTime.Minute())
	dateTime[6] = uint8(tmTime.Second())
	gpsDataBuf.AppendBytes(dateTime)
	gpsDataBuf.AppendNumber(uint32(reqObject.Longitude))
	gpsDataBuf.AppendNumber(uint32(reqObject.Latitude))
	gpsDataBuf.AppendNumber(uint16(reqObject.Speed / 10))
	gpsDataBuf.AppendNumber(uint16(reqObject.Speed / 10)) //行驶记录速度
	gpsDataBuf.AppendNumber(reqObject.PlatformMileage / 1000)
	gpsDataBuf.AppendNumber(uint16(reqObject.Direction))
	gpsDataBuf.AppendNumber(uint16(reqObject.Altitude))
	gpsDataBuf.AppendNumber(reqObject.AlarmFlag)
	gpsDataBuf.AppendNumber(reqObject.StatusFlag)

	return gpsDataBuf.GetData()
}

func (_self *MsgHandleContext) updateLastGpsData(reqObject *models.VsPbStorageGpsData) {
	vehicleTaskObject, isExist := _self.mapVehicleTask[reqObject.VehicleNo]
	if !isExist {
		vehicleTaskObject = &vehicleTask{}
		_self.mapVehicleTask[reqObject.VehicleNo] = vehicleTaskObject
	}
	vehicleTaskObject.lastRealData = *reqObject
}

// SendExgMsgUpGpsDataUpload 上传定位数据
func (_self *MsgHandleContext) SendExgMsgUpGpsDataUpload(reqObject *models.VsPbStorageGpsData) bool {

	jtVer := _self.getJTVer()
	dataObject := util.NewBinaryWrite()
	if jtVer == jt_session.JT_VER_V3 {
		gpsBuf := _self.convertGpsData2019(reqObject)
		dataObject.AppendByte(0)
		dataObject.AppendNumber(uint32(len(gpsBuf)))
		dataObject.AppendBytes(gpsBuf)
		dataObject.AppendFixedBytes(_self.getPlatformCode(), 11)
		dataObject.AppendNumber(reqObject.AlarmFlag)
		dataObject.AppendFixedBytes("", 11)
		dataObject.AppendNumber(uint32(0))
		dataObject.AppendFixedBytes("", 11)
		dataObject.AppendNumber(uint32(0))
	} else {
		gpsBuf := _self.convertGpsData2011(reqObject)
		dataObject.AppendBytes(gpsBuf)
	}
	_self.updateLastGpsData(reqObject)
	return _self.buildBodyObjectByVehicleGuidAndSend(models.VS_JT_TYPE_UP_EXG_MSG, models.VS_JT_TYPE_UP_EXG_MSG_REALTIME, dataObject.GetData(), reqObject.VehicleGuid)
}

// sendExgMsgUpAutoGpsDataUpload 自动补报定位数据
func (_self *MsgHandleContext) sendExgMsgUpAutoGpsDataUpload(reqObjects []*models.VsPbStorageGpsData) bool {
	if len(reqObjects) > 5 {
		//单次不能超过5条定位
		log.Error("SendExgMsgUpAutoGpsDataUpload over 5 ")
		return false
	}
	jtVer := _self.getJTVer()
	dataObject := util.NewBinaryWrite()
	var curVehicleID uint32
	for idx := 0; idx < len(reqObjects); idx++ {
		curGpsInfo := reqObjects[idx]
		if curVehicleID != curGpsInfo.VehicleGuid && curVehicleID != 0 {
			log.Error("sendExgMsgUpAutoGpsDataUpload must same vehicle once timer ", curVehicleID, " curGpsGuid", curGpsInfo.VehicleGuid)
			return false
		}
		curVehicleID = curGpsInfo.VehicleGuid
		if jtVer == jt_session.JT_VER_V3 {
			gpsBuf := _self.convertGpsData2019(curGpsInfo)
			dataObject.AppendByte(0)
			dataObject.AppendNumber(uint32(len(gpsBuf)))
			dataObject.AppendBytes(gpsBuf)
			dataObject.AppendFixedBytes(_self.getPlatformCode(), 11)
			dataObject.AppendNumber(curGpsInfo.AlarmFlag)
			dataObject.AppendFixedBytes("", 11)
			dataObject.AppendNumber(uint32(0))
			dataObject.AppendFixedBytes("", 11)
		} else {
			dataObject.AppendBytes(_self.convertGpsData2011(curGpsInfo))
		}
	}
	return _self.buildBodyObjectByVehicleGuidAndSend(models.VS_JT_TYPE_UP_EXG_MSG, models.VS_JT_TYPE_UP_EXG_MSG_UP_EXG_MSG_HISTORY_LOCATION, dataObject.GetData(), curVehicleID)
}

// SendExgMsgUpDriverInfoUpload 主动上报驾驶员信息
func (_self *MsgHandleContext) SendExgMsgUpDriverInfoUpload(reqObject *models.VsPbStorageSwipeCardInfo) bool {
	if reqObject.GpsData == nil {
		log.Error("SendExgMsgUpDriverInfoUpload fail,gpsData is nil", reqObject.IdCard)
		return false
	}
	jtVer := _self.getJTVer()
	dataObject := util.NewBinaryWrite()
	if jtVer == jt_session.JT_VER_V3 {

		dataObject.AppendFixedBytes(reqObject.StaffName, 11)  //驾驶员名称
		dataObject.AppendFixedBytes(reqObject.IdCard, 20)     //驾驶员编号
		dataObject.AppendFixedBytes(reqObject.PqCertCode, 20) //从业资格证
		gbkIssue := reqObject.Issuing
		util.VsUtf8ToGBK(&gbkIssue)
		dataObject.AppendFixedBytes(gbkIssue, 200)           //发证机构
		dataObject.AppendNumber(uint64(reqObject.ValidDate)) //证件有效期

	} else if jtVer == jt_session.JT_VER_V2 {
		dataObject.AppendFixedBytes(reqObject.StaffName, 16)  //驾驶员名称
		dataObject.AppendFixedBytes(reqObject.IdCard, 20)     //驾驶员编号
		dataObject.AppendFixedBytes(reqObject.PqCertCode, 40) //从业资格证
		gbkIssue := reqObject.Issuing
		util.VsUtf8ToGBK(&gbkIssue)
		dataObject.AppendFixedBytes(gbkIssue, 200) //发证机构
	} else { //2011不支持
		log.Error("SendExgMsgUpDriverInfoUpload fail,jt ver is err", jtVer, reqObject.IdCard)
		return false
	}
	return _self.buildBodyObjectByVehicleGuidAndSend(models.VS_JT_TYPE_UP_EXG_MSG, models.VS_JT_TYPE_UP_EXG_MSG_REPORT_DRIVER_INFO, dataObject.GetData(), reqObject.GpsData.VehicleGuid)
}
