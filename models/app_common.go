package models

const (
	VS_APP_SERVICE_MSG_ID_GPS_DATA                 = 0x10000 //定位数据 VsPbStorageGpsData
	VS_APP_SERVICE_MSG_ID_ALARM_DATA               = 0x10001 //报警数据 VsPbStorageEventInfo
	VS_APP_SERVICE_MSG_ID_DEV_ONOFF_LINE           = 0x10002 //车辆上下线 VsFrontPbMsgOnline
	VS_APP_SERVICE_MSG_ID_DEV_ATTACH_INFO_UPLOAD   = 0x10003 //附件信息上传 VsFrontAttachInfoMsg
	VS_APP_SERVICE_MSG_ID_DEV_ATTACH_INFO_COMPLETE = 0x10004 //附件信息完成上传 VsFrontAttachUploadCompleteMsg
	VS_APP_SERVICE_MSG_ID_SWIPE_CARD               = 0x10005 //刷卡信息 VsPbStorageSwipeCardInfo
	VS_APP_SERVICE_MSG_ID_MULTI_MEDIA_INFO         = 0x10006 //多媒体信息上传 VsPbStorageMultiMediaEventInfo

	//请求应答类

)

// VsVehicleBaseInfo 车辆基本信息
type VsVehicleBaseInfo struct {
	VehicleNoColor     uint8  //车牌号颜色
	VehicleNo          string //车牌号
	DeviceIMEI         string //终端的通讯模块IMEI
	DeviceNo           string //终端号
	SimNo              string //SIM卡号
	VehicleID          uint32 //车辆ID
	LineGuid           uint32 //线路ID
	VehicleType        int    //车辆类型
	TransType          string //运输行业编码
	VehicleNationality string //车籍地
	BusinessScopeCode  string //经营范围代码
	OwnerID            string //业户ID
	OwnerName          string //业户名称
	OwnerTel           string //业户联系人
}

type VsAppTakePhotoReq struct {
	ChannelNo uint8  //通道
	PicType   uint8  //图片尺寸类型
	VehicleNo string //车牌号
}
