package jt_session

type VsAppMsgHeader struct {
	MsgSN          uint32 //消息序列号
	MsgID          uint16 //业务数据类型
	EncryptKey     uint32 //加密KEY
	DateTime       int64  //utc时间
	JTVer          JT_VER_TYPE
	VehicleNo      string //车牌号 子业务类型需要
	VehicleNoColor uint8  //车牌号颜色
}

// OnNetMsgPacketFunc 消息处理,从网络来的数据
type OnNetMsgPacketFunc func(appHdr VsAppMsgHeader, bodyData []byte) bool
