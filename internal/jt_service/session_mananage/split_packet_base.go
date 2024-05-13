package jt_session

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"jt809_server/util"
	"time"
)

const (
	JT_809_START_SIGN_5B  = 0x5B //开始标识
	JT_809_END_SIGN_5D    = 0x5D //结束标识
	JT_809_ESCAPE_SIGN_5A = 0x5A //转义标识
	JT_809_ESCAPE_SIGN_5E = 0x5E //转义标识
)

type JT_VER_TYPE int

const (
	JT_VER_NONE JT_VER_TYPE = 0 //未知
	JT_VER_V1               = 1 //JT809-2011
	JT_VER_V2               = 2 //JT809-2013
	JT_VER_V3               = 3 //JT809-2019
)

// 接口定义
type codecBase interface {
	initCodec(isPrintHex bool, M1, IA1, IC1, key uint32, cbNetMsgPacketFunc OnNetMsgPacketFunc) bool
	putData(data []byte) bool
	buildJTData(dataType uint16, msgSN, platformID, key uint32, verFlags, bodyData []byte) []byte
}

type splitPacketContextBase struct {
	dataBuf          *bytes.Buffer
	scanner          *util.Scanner
	M1               uint32
	IA1              uint32
	IC1              uint32
	protocolVer      JT_VER_TYPE
	cbPacketComplete onPacketComplete
	cbNetMsgDispatch OnNetMsgPacketFunc
	isPrintHex       bool
}

// 完整一包数据
type onPacketComplete func(data []byte)

func (_self *splitPacketContextBase) initCodecBase(isPrintHex bool, M1, IA1, IC1 uint32, verType JT_VER_TYPE, cbPacketComplete onPacketComplete) bool {
	_self.dataBuf = bytes.NewBuffer(nil)
	_self.scanner = util.NewScanner(_self.dataBuf)
	_self.scanner.Split(_self.splitCodec)
	_self.M1 = M1
	_self.IA1 = IA1
	_self.IC1 = IC1
	_self.cbPacketComplete = cbPacketComplete
	_self.protocolVer = verType
	_self.isPrintHex = isPrintHex
	return true
}

func (_self *splitPacketContextBase) writeDataAndParse(data []byte) bool {
	_self.dataBuf.Write(data)
	isOK := _self.scanner.Scan()
	if !isOK {
		log.Error("putData fail,scan err", _self.protocolVer)
		return false
	}
	defer _self.scanner.CloseScan()
	for _self.scanner.Next() {
		if _self.cbPacketComplete != nil {
			packetBuf := _self.scanner.Bytes()
			if _self.isPrintHex {
				strHex := ""
				for idx := 0; idx < len(packetBuf); idx++ {
					strHex += fmt.Sprintf("%02X ", packetBuf[idx])
				}
				log.Info("JTServerToClient:", strHex)
			}
			_self.cbPacketComplete(packetBuf)
		}
	}
	return true
}

// 协议解包
func (_self *splitPacketContextBase) splitCodec(data []byte) (advance int, token []byte, err error) {
	curLen := len(data)
	startIndex := 0
	isExistStartSign := false
	for idx := 0; idx < curLen; idx++ {
		curData := data[idx]
		if curData == JT_809_START_SIGN_5B {
			startIndex = idx
			isExistStartSign = true
		} else if curData == JT_809_END_SIGN_5D {
			if !isExistStartSign {
				advance = idx + 1
				err = util.ErrBadReadCount
				return
			}
			//反转义
			realLen := _self.unescape(data[startIndex : idx+1])

			isOK := _self.dataParity(data[startIndex : startIndex+int(realLen)])
			if !isOK {
				err = util.ErrBadReadCount //错误数据
			} else {
				token = data[startIndex : startIndex+int(realLen)]
			}
			advance = idx + 1
			return
		}
	}
	return 0, nil, nil
}

// 加密算法
func (_self *splitPacketContextBase) encrypt(key uint32, data []byte) {
	if _self.M1 == 0 || _self.IA1 == 0 || _self.IC1 == 0 {
		log.Error("encrypt fail m1,ia1,ic1 is zero,key:", key)
		return
	}
	var idx uint32
	if _self.protocolVer != JT_VER_V1 {
		key = _self.M1
	}
	if key == 0 {
		key = 1
	}
	dataLen := uint32(len(data))
	for idx < dataLen {
		key = _self.IA1*(key%_self.M1) + _self.IC1
		data[idx] ^= (byte)((key >> 20) & 0xFF)
		idx++
	}
}

// 数据检验，从数据头到检验码前的数据
func (_self *splitPacketContextBase) dataParity(data []byte) bool {
	curLen := len(data)
	if curLen < 5 || data[0] != JT_809_START_SIGN_5B || data[curLen-1] != JT_809_END_SIGN_5D {
		log.Error("dataParity check fail,head err ver:", _self.protocolVer, " curLen:", curLen)
		return false
	}
	calCode := util.CRC16CheckSum(data[1 : curLen-3])
	dataCode := binary.BigEndian.Uint16(data[curLen-3 : curLen-1])
	if calCode != dataCode {
		log.Error("dataParity check fail,ver:", _self.protocolVer, " calCode:", calCode, " dataCode:", dataCode, " curLen:", curLen, "dataLen:", len(data))
		return false
	}
	return true
}

// 反转义,返回真实长度
func (_self *splitPacketContextBase) unescape(data []byte) uint32 {
	var curOffset uint32
	curOffset = 1
	for idx := 1; idx < len(data)-1; idx++ {
		curData := data[idx]
		if curData == JT_809_ESCAPE_SIGN_5A && data[idx+1] == 0x01 {
			data[curOffset] = JT_809_START_SIGN_5B
			idx++
		} else if curData == JT_809_ESCAPE_SIGN_5A && data[idx+1] == 0x02 {
			data[curOffset] = JT_809_ESCAPE_SIGN_5A
			idx++
		} else if curData == JT_809_ESCAPE_SIGN_5E && data[idx+1] == 0x01 {
			data[curOffset] = JT_809_END_SIGN_5D
			idx++
		} else if curData == JT_809_ESCAPE_SIGN_5E && data[idx+1] == 0x02 {
			data[curOffset] = JT_809_ESCAPE_SIGN_5E
			idx++
		} else {
			data[curOffset] = curData
		}

		curOffset++
	}
	data[curOffset] = JT_809_END_SIGN_5D
	//加上头尾标识才是真实长度
	return curOffset + 1
}

// 转义后的数据
func (_self *splitPacketContextBase) escape(data []byte) []byte {
	newData := make([]byte, 0, len(data)+10)
	newData = append(newData, JT_809_START_SIGN_5B)
	for idx := 1; idx < len(data)-1; idx++ {
		curData := data[idx]
		if curData == JT_809_START_SIGN_5B {
			newData = append(newData, JT_809_ESCAPE_SIGN_5A)
			newData = append(newData, 0x01)
		} else if curData == JT_809_END_SIGN_5D {
			newData = append(newData, JT_809_ESCAPE_SIGN_5E)
			newData = append(newData, 0x01)
		} else if curData == JT_809_ESCAPE_SIGN_5E {
			newData = append(newData, JT_809_ESCAPE_SIGN_5E)
			newData = append(newData, 0x02)
		} else if curData == JT_809_ESCAPE_SIGN_5A {
			newData = append(newData, JT_809_ESCAPE_SIGN_5A)
			newData = append(newData, 0x02)
		} else {
			newData = append(newData, curData)
		}
	}
	newData = append(newData, JT_809_END_SIGN_5D)
	return newData
}

// 构建JT消息 ,bodyData 非加密数据
func (_self *splitPacketContextBase) buildJTData(dataType uint16, msgSN, platformID, key uint32, verFlags, bodyData []byte) []byte {
	msgLen := 0
	bodyLen := len(bodyData)
	if _self.protocolVer == JT_VER_V3 { //消息头长度(30)+头标识(2)+校验(2)
		msgLen = 34 + bodyLen
	} else if _self.protocolVer == JT_VER_V2 { //消息头长度(22)+头标识(2)+校验(2)
		msgLen = 26 + bodyLen
	} else if _self.protocolVer == JT_VER_V1 {
		msgLen = 26 + bodyLen
	} else {
		log.Error("buildJTData fail,protocol version err", _self.protocolVer)
		return []byte{}
	}

	bodyObject := util.NewBinaryWrite()
	bodyObject.AppendByte(JT_809_START_SIGN_5B)
	bodyObject.AppendNumber(uint32(msgLen))
	bodyObject.AppendNumber(msgSN)
	bodyObject.AppendNumber(dataType)
	bodyObject.AppendNumber(platformID)
	bodyObject.AppendFixedBytes(string(verFlags), 3)

	isEncrypt := false
	if key == 0 {
		bodyObject.AppendByte(0)
	} else {
		bodyObject.AppendByte(1)
		isEncrypt = true
	}
	bodyObject.AppendNumber(key)
	if _self.protocolVer == JT_VER_V3 {
		bodyObject.AppendNumber(time.Now().Unix())
	}
	if bodyLen > 0 {
		if isEncrypt {
			_self.encrypt(key, bodyData)
		}
		bodyObject.AppendBytes(bodyData)
	}
	dataParityBuf := bodyObject.GetData()
	var calCode uint16
	if len(dataParityBuf) > 5 {
		calCode = util.CRC16CheckSum(dataParityBuf[1:])
	}
	bodyObject.AppendNumber(calCode)
	bodyObject.AppendByte(JT_809_END_SIGN_5D)
	srcDataBuf := bodyObject.GetData()
	//转义
	dstDataBuf := _self.escape(srcDataBuf)
	return dstDataBuf
}
