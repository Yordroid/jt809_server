package util

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
)

type vsBinaryRead struct {
	offset uint32
	srcBuf []byte
	srcLen uint32
}

func NewBinaryRead(srcBuf []byte) *vsBinaryRead {
	curObject := &vsBinaryRead{}
	curObject.init()
	curObject.srcBuf = srcBuf
	curObject.srcLen = uint32(len(srcBuf))
	return curObject
}

func (_self *vsBinaryRead) init() {
	_self.offset = 0
}

func (_self *vsBinaryRead) ReadInt64() int64 {
	if _self.offset+8 > _self.srcLen {
		log.Error("ReadInt64 fail,over size,offset:", _self.offset, " srcLen:", _self.srcLen)
		return 0
	}
	value := binary.BigEndian.Uint64(_self.srcBuf[_self.offset : _self.offset+8])
	_self.offset += 8
	return int64(value)
}

func (_self *vsBinaryRead) ReadInt32() uint32 {
	if _self.offset+4 > _self.srcLen {
		log.Error("ReadInt32 fail,over size,offset:", _self.offset, " srcLen:", _self.srcLen)
		return 0
	}
	value := binary.BigEndian.Uint32(_self.srcBuf[_self.offset : _self.offset+4])
	_self.offset += 4
	return value
}
func (_self *vsBinaryRead) ReadInt16() uint16 {
	if _self.offset+2 > _self.srcLen {
		log.Error("ReadUInt16 fail,over size,offset:", _self.offset, " srcLen:", _self.srcLen)
		return 0
	}
	value := binary.BigEndian.Uint16(_self.srcBuf[_self.offset : _self.offset+2])
	_self.offset += 2
	return value
}

func (_self *vsBinaryRead) ReadByte() byte {
	if _self.offset+1 > _self.srcLen {
		log.Error("ReadByte fail,over size,offset:", _self.offset, " srcLen:", _self.srcLen)
		return 0
	}
	value := _self.srcBuf[_self.offset : _self.offset+1][0]
	_self.offset++
	return value
}

func (_self *vsBinaryRead) ReadBytes(readLen uint32) []byte {
	if _self.offset+readLen > _self.srcLen {
		log.Error("ReadBytes fail,over size,offset:", _self.offset, " srcLen:", _self.srcLen)
		return nil
	}
	value := _self.srcBuf[_self.offset : _self.offset+readLen]
	_self.offset += readLen
	return value
}

// ReadString 会把0的去掉
func (_self *vsBinaryRead) ReadString(readLen uint32) string {
	if _self.offset+readLen > _self.srcLen {
		log.Error("ReadString fail,over size,offset:", _self.offset, " srcLen:", _self.srcLen)
		return ""
	}
	var idx uint32
	var actualLen uint32
	for idx = 0; idx < readLen; idx++ {
		if 0 == _self.srcBuf[int(_self.offset+idx)] {
			break
		}
		actualLen++
	}
	value := _self.srcBuf[_self.offset : _self.offset+actualLen]
	_self.offset += readLen
	return string(value)
}

// ReadStringToUTF8 GBK 转 UTF8
func (_self *vsBinaryRead) ReadStringToUTF8(readLen uint32) string {
	value := _self.ReadString(readLen)
	err := VsGBKToUtf8(&value)
	if err != nil {
		log.Error("ReadStringToUTF8 fail", err.Error())
		return ""
	}
	return value
}
