package util

import (
	"bytes"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
)

type VsBinaryWrite struct {
	buf *bytes.Buffer
}

func NewBinaryWrite() *VsBinaryWrite {
	curObject := &VsBinaryWrite{}
	curObject.init()
	return curObject
}

func (_self *VsBinaryWrite) init() {
	_self.buf = new(bytes.Buffer)
	//初始化1k
	_self.buf.Grow(1024)
}

func copyString(dstBytes []byte, dstByteMaxLen int, srcString string) {
	if len(srcString) > dstByteMaxLen {
		copy(dstBytes, []byte(srcString)[:dstByteMaxLen])
	} else {
		copy(dstBytes, []byte(srcString)[:len(srcString)])
	}
}

// AppendNumber 追加整型数据
func (_self *VsBinaryWrite) AppendNumber(data any) bool {
	if _self.buf == nil {
		log.Error("VsBinaryWrite AppendData fail,buf is nil")
		return false
	}
	err := binary.Write(_self.buf, binary.BigEndian, data)
	if err != nil {
		log.Error("VsBinaryWrite AppendData fail,write fail", err.Error(), " data:", data)
		return false
	}
	return true
}

// AppendByte 追加单字节数据
func (_self *VsBinaryWrite) AppendByte(data byte) bool {
	if _self.buf == nil {
		log.Error("VsBinaryWrite AppendByte fail,buf is nil")
		return false
	}
	_self.buf.WriteByte(data)
	return true
}

// AppendBytes 追加字节数据
func (_self *VsBinaryWrite) AppendBytes(dataBuf []byte) bool {
	if _self.buf == nil {
		log.Error("VsBinaryWrite AppendBytes fail,buf is nil")
		return false
	}
	if len(dataBuf) > 0 {
		_self.buf.Write(dataBuf)
	}
	return true
}

// AppendFixedBytes 追加固定长度的字节数据
func (_self *VsBinaryWrite) AppendFixedBytes(data string, dataLen int) bool {
	if _self.buf == nil {
		log.Error("VsBinaryWrite AppendBytes fail,buf is nil")
		return false
	}
	if dataLen == 0 {
		log.Error("VsBinaryWrite AppendBytes fail,data len == 0")
		return false
	}
	curData := make([]byte, dataLen)
	copyString(curData, dataLen, data)
	_self.buf.Write(curData)
	return true
}

// AppendFixedBytesBefore 追加固定长度的字节数据,不足的在前补0
func (_self *VsBinaryWrite) AppendFixedBytesBefore(data string, dataLen int) bool {
	if _self.buf == nil {
		log.Error("VsBinaryWrite AppendBytes fail,buf is nil")
		return false
	}
	if dataLen == 0 {
		log.Error("VsBinaryWrite AppendBytes fail,data len == 0")
		return false
	}
	curData := make([]byte, dataLen)
	if len(data) > dataLen {
		copy(curData, []byte(data)[:dataLen])
	} else {
		remainLen := dataLen - len(data)
		copy(curData[remainLen:dataLen], data)
	}
	_self.buf.Write(curData)
	return true
}

// AppendString 追加字符数据
func (_self *VsBinaryWrite) AppendString(data string) bool {
	if _self.buf == nil {
		log.Error("VsBinaryWrite AppendString fail,buf is nil")
		return false
	}
	_self.buf.Write([]byte(data))
	return true
}

// GetData 获取数据
func (_self *VsBinaryWrite) GetData() []byte {
	if _self.buf == nil {
		log.Error("VsBinaryWrite GetData fail,buf is nil")
		return nil
	}
	return _self.buf.Bytes()
}
