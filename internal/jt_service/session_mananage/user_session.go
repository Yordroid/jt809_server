package jt_session

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"jt809_server/util"
)

type UserSession struct {
	tcpServer  util.TcpServer //从连网络
	codecSlave codecBase      //从链路解码

	tcpClient   util.TcpClient //主链网络
	codecMain   codecBase      //主链路解码
	conf        UserSessionConfig
	cbEventFunc onLinkEvent

	mainIsConnected  bool //主链路是否连接
	slaveIsConnected bool //从链路是否连接

	slaveClientID int64 //当前从链路的ID

	sendMsgSN  uint32 // 发送数据SN,主从链路发送共享
	isPrintHex bool
}

func NewUserSession(conf UserSessionConfig, cbLinkEvent onLinkEvent, netMsgPacketFunc OnNetMsgPacketFunc) *UserSession {
	curSession := &UserSession{}
	curSession.initUserSession(conf, cbLinkEvent, netMsgPacketFunc)
	curSession.isPrintHex = false

	return curSession

}

func (_self *UserSession) StartMainConnect() {
	_self.tcpClient.ConnectToServer(_self.conf.RemoteAddr, _self.conf.RemotePort)
}

// BuildJTDataAndSend 构建JT消息 ,bodyData 非加密数据
func (_self *UserSession) BuildJTDataAndSend(dataType uint16, bodyData []byte, isMainLink bool) bool {
	msgBuf := _self.buildJTData(dataType, bodyData)
	if _self.isPrintHex {
		strHex := ""
		for idx := 0; idx < len(msgBuf); idx++ {
			strHex += fmt.Sprintf("%02X ", msgBuf[idx])
		}
		log.Info("ClientToJTServer:", strHex)
	}
	return _self.sendData(msgBuf, isMainLink)
}

type UserSessionConfig struct {
	BindPort   uint16
	RemoteAddr string
	RemotePort uint16
	Version    JT_VER_TYPE
	M1         uint32
	IA1        uint32
	IC1        uint32
	Key        uint32
	AccessID   uint32
	IsPrintHex bool
}

type onLinkEvent func(isMainLink, isConnect bool, sessionID int64)

func (_self *UserSession) createSplitPacket(isMainLink bool, conf UserSessionConfig, netMsgPacketFunc OnNetMsgPacketFunc) {
	switch conf.Version {
	case JT_VER_V1:
		if isMainLink {
			_self.codecMain = &splitPacketContextV1{}
		} else {
			_self.codecSlave = &splitPacketContextV1{}
		}
	case JT_VER_V2:
		if isMainLink {
			_self.codecMain = &splitPacketContextV2{}
		} else {
			_self.codecSlave = &splitPacketContextV2{}
		}

	case JT_VER_V3:
		if isMainLink {
			_self.codecMain = &splitPacketContextV3{}
		} else {
			_self.codecSlave = &splitPacketContextV3{}
		}
	default:

	}
	if isMainLink {
		_self.codecMain.initCodec(conf.IsPrintHex, conf.M1, conf.IA1, conf.IC1, conf.Key, netMsgPacketFunc)
	} else {
		_self.codecSlave.initCodec(conf.IsPrintHex, conf.M1, conf.IA1, conf.IC1, conf.Key, netMsgPacketFunc)
	}
}

func (_self *UserSession) initUserSession(conf UserSessionConfig, cbEventFunc onLinkEvent, netMsgPacketFunc OnNetMsgPacketFunc) bool {
	_self.conf = conf
	_self.cbEventFunc = cbEventFunc

	_self.tcpServer.InitServer(conf.BindPort, func(clientID int64, isConnected bool) {
		_self.mainIsConnected = isConnected
		if isConnected {
			_self.slaveClientID = clientID
			_self.createSplitPacket(false, conf, netMsgPacketFunc)
		} else {
			_self.slaveClientID = 0
		}
		if _self.cbEventFunc != nil {
			_self.cbEventFunc(false, isConnected, clientID)
		}
	}, func(clientID int64, data []byte) {
		if _self.codecSlave == nil {
			log.Error("slave link received data fail,codecSlave is nil")
			return
		}
		_self.codecSlave.putData(data)
	})
	curClientID := util.GetSnowflakeID()
	_self.tcpClient.InitClient(true, func(isConnected bool) {
		_self.mainIsConnected = isConnected
		if isConnected {
			_self.createSplitPacket(true, conf, netMsgPacketFunc)
		}
		if _self.cbEventFunc != nil {
			_self.cbEventFunc(true, isConnected, curClientID)
		}
	}, func(data []byte) {
		if _self.codecMain == nil {
			log.Error("slave link received data fail,codecMain is nil")
			return
		}
		_self.codecMain.putData(data)
	})

	return true
}

func (_self *UserSession) DeInitUserSession() bool {
	_self.tcpServer.DeInitServer()
	_self.tcpClient.DeInitClient()
	return true
}

// SendData  发送数据到上级平台,isMainLink:true 优先主链路发送,false:优先从链路发送,如果链路异常,则选择
func (_self *UserSession) sendData(msgBuf []byte, isMainLink bool) bool {
	isOK := false
	if isMainLink {
		if _self.mainIsConnected {
			isOK = _self.tcpClient.SendData(msgBuf)
		} else {
			nRet := _self.tcpServer.SendDataToClient(_self.slaveClientID, msgBuf)
			if nRet > 0 {
				isOK = true
			}
		}
	} else {
		if _self.slaveIsConnected {
			nRet := _self.tcpServer.SendDataToClient(_self.slaveClientID, msgBuf)
			if nRet > 0 {
				isOK = true
			}
		} else {
			isOK = _self.tcpClient.SendData(msgBuf)
		}
	}
	return isOK
}

// 获取发送的序列号
func (_self *UserSession) getNextMsgSN() uint32 {
	_self.sendMsgSN++
	return _self.sendMsgSN
}

func (_self *UserSession) buildJTData(dataType uint16, bodyData []byte) []byte {
	msgSN := _self.getNextMsgSN()
	if _self.codecMain != nil {
		return _self.codecMain.buildJTData(dataType, msgSN, _self.conf.AccessID, _self.conf.Key, []byte{1, 0, 0}, bodyData)
	} else if _self.codecSlave != nil {
		return _self.codecSlave.buildJTData(dataType, msgSN, _self.conf.AccessID, _self.conf.Key, []byte{1, 0, 0}, bodyData)
	}
	return []byte{}
}
