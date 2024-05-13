package jt_session

import (
	"jt809_server/util"
)

const (
	JT_MSG_HDR_2019_LEN = 30 //消息头长度
)

type splitPacketContextV3 struct {
	splitPacketContextBase
}

type codecHeaderV3 struct {
	msgLen             uint32 //数据长度
	msgSN              uint32 //消息序列号
	msgID              uint16 //业务数据类型
	platformAccessCode uint32 //平台接入码
	versionFlag        []byte //版本号
	isEncrypt          bool   //是否加密
	encryptKey         uint32 //加密KEY
	dateTime           int64  //utc时间
}

func (_self *splitPacketContextV3) initCodec(isPrintHex bool, M1, IA1, IC1, key uint32, cbNetMsgPacketFunc OnNetMsgPacketFunc) bool {
	isOK := _self.initCodecBase(isPrintHex, M1, IA1, IC1, JT_VER_V3, func(data []byte) {
		isOK, hdrObject, bodyData := _self.getMsgBaseObjectFromData(data)
		if isOK {
			if hdrObject.isEncrypt {
				_self.encrypt(key, bodyData)
			}
			appHdr := VsAppMsgHeader{}
			appHdr.MsgSN = hdrObject.msgSN
			appHdr.EncryptKey = hdrObject.encryptKey
			appHdr.DateTime = hdrObject.dateTime
			appHdr.MsgID = hdrObject.msgID
			appHdr.JTVer = _self.protocolVer
			if cbNetMsgPacketFunc != nil {
				cbNetMsgPacketFunc(appHdr, bodyData)
			}
		}
	})
	return isOK
}
func (_self *splitPacketContextV3) putData(data []byte) bool {
	return _self.writeDataAndParse(data)
}

// 获取消息头和消息体
func (_self *splitPacketContextV3) getMsgBaseObjectFromData(data []byte) (bool, codecHeaderV3, []byte) {
	hdrObject := codecHeaderV3{}
	dataLen := len(data)
	if dataLen < JT_MSG_HDR_2019_LEN+1 {
		return false, hdrObject, nil
	}
	hdrData := data[1 : JT_MSG_HDR_2019_LEN+1]
	readObject := util.NewBinaryRead(hdrData)
	hdrObject.msgLen = readObject.ReadInt32()
	hdrObject.msgSN = readObject.ReadInt32()
	hdrObject.msgID = readObject.ReadInt16()
	hdrObject.platformAccessCode = readObject.ReadInt32()
	hdrObject.versionFlag = readObject.ReadBytes(3)
	bVer := readObject.ReadBytes(1)
	if len(bVer) == 1 && bVer[0] == 1 {
		hdrObject.isEncrypt = true
	}
	hdrObject.encryptKey = readObject.ReadInt32()
	hdrObject.dateTime = readObject.ReadInt64()
	var bodyData []byte
	if JT_MSG_HDR_2019_LEN+4 < dataLen {
		bodyData = data[JT_MSG_HDR_2019_LEN+1 : dataLen-3]
	}
	return true, hdrObject, bodyData
}

// 消息头和消息体 转成数据
func (_self *splitPacketContextV3) getDataFromMsgObject(header VsAppMsgHeader, bodyData []byte) (bool, []byte) {
	dataBufs := make([]byte, 0, JT_MSG_HDR_2019_LEN+len(bodyData)+3)
	dataBufs = append(dataBufs, JT_809_START_SIGN_5B)
	return true, dataBufs
}
