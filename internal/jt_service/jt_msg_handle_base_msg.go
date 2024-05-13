package jt_service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"jt809_server/internal/data_manage_service"
	jt_session "jt809_server/internal/jt_service/session_mananage"
	"jt809_server/models"
	"jt809_server/util"
)

//静态数据交互

type jtVehicleInfo struct {
	strVehicleInfo string
}

func (_self *jtVehicleInfo) addVehicleKeyInfo(key string, value any) {
	if len(_self.strVehicleInfo) > 0 {
		_self.strVehicleInfo += fmt.Sprintf(";%s:=%v", key, value)
	} else {
		_self.strVehicleInfo += fmt.Sprintf("%s:=%v", key, value)
	}

}
func (_self *jtVehicleInfo) getString() string {
	util.VsUtf8ToGBK(&_self.strVehicleInfo)
	return _self.strVehicleInfo
}

// 构建车辆静态数据
func buildStaticVehicleInfo(jtVer jt_session.JT_VER_TYPE, info *models.VsVehicleBaseInfo) string {
	buildVehicleInfo := jtVehicleInfo{}
	buildVehicleInfo.addVehicleKeyInfo("VIN", info.VehicleNo)                          //车牌号
	buildVehicleInfo.addVehicleKeyInfo("VEHICLE_COLOR", info.VehicleNoColor)           //车牌颜色
	buildVehicleInfo.addVehicleKeyInfo("VEHICLE_TYPE", info.VehicleType)               //车辆类型
	buildVehicleInfo.addVehicleKeyInfo("TRANS_TYPE", info.TransType)                   //运输行业编码
	buildVehicleInfo.addVehicleKeyInfo("VEHICLE_NATIONALITY", info.VehicleNationality) //车籍地
	if jtVer == jt_session.JT_VER_V3 {
		buildVehicleInfo.addVehicleKeyInfo("BUSINESSSCOPECODE", info.BusinessScopeCode) //经营范围代码
	}
	buildVehicleInfo.addVehicleKeyInfo("OWERS_ID", info.OwnerID)     //业户 ID
	buildVehicleInfo.addVehicleKeyInfo("OWERS_NAME", info.OwnerName) //业户名称
	buildVehicleInfo.addVehicleKeyInfo("OWERS_TEL", info.OwnerTel)   //业户联系电话
	return buildVehicleInfo.getString()
}

// 车辆拍照请求
func (_self *MsgHandleContext) onHandleBaseMsgVehicleStaticInfoReq(hdrObject jt_session.VsAppMsgHeader, bodyMsg []byte) bool {
	srcDataType := hdrObject.MsgID
	srcMsgSN := hdrObject.MsgSN
	vehicleBaseInfo := data_manage_service.DataManageServiceIns().GetVehicleInfoByVehicleNo(hdrObject.VehicleNo)
	if vehicleBaseInfo == nil {
		log.Error("onHandleBaseMsgVehicleStaticInfoReq fail,not found,", hdrObject.VehicleNo)
		return false
	}
	jtVer := _self.getJTVer()
	strVehicleInfo := buildStaticVehicleInfo(jtVer, vehicleBaseInfo)
	dataObject := util.NewBinaryWrite()

	if jtVer == jt_session.JT_VER_V3 {
		dataObject.AppendNumber(srcDataType)
		dataObject.AppendNumber(srcMsgSN)
	}
	dataObject.AppendString(strVehicleInfo)
	log.Info("onHandleBaseMsgVehicleStaticInfoReq: ", strVehicleInfo)
	_self.buildBodyObjectByVehicleInfoAndSend(models.VS_JT_TYPE_UP_BASE_MSG, models.VS_JT_TYPE_UP_BASE_MSG_VEHICLE_ADDED_ACK,
		dataObject.GetData(), hdrObject.VehicleNo, hdrObject.VehicleNoColor)
	return true
}
