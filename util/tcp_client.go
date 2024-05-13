package util

import (
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"time"
)

// 最大包大小
const MAX_PACKET_SIZE = 60 * 1024

// 接收数据回调
type onRecvDataCallback func(data []byte)

// 连接状态回调
type onConnectStateCallback func(isConnected bool)
type TcpClient struct {
	conn                net.Conn
	isConnected         bool
	cbRecvDataFunc      onRecvDataCallback
	cbConnectStateFunc  onConnectStateCallback
	isNeedAutoReconnect bool
	isInit              bool
	isQuitTimer         bool
	serverAddr          string
	serverPort          uint16
	writeQueue          chan []byte
}

// InitClient 初始化,一开始调用
func (_self *TcpClient) InitClient(isNeedAutoReconnect bool, onConnectStateCb onConnectStateCallback, onRecvFunc onRecvDataCallback) {
	if _self.isInit {
		log.Info("client already init")
		return
	}
	_self.isQuitTimer = false
	_self.isInit = true
	_self.isConnected = false
	_self.cbRecvDataFunc = onRecvFunc
	_self.isNeedAutoReconnect = isNeedAutoReconnect
	_self.cbConnectStateFunc = onConnectStateCb
	_self.writeQueue = make(chan []byte, 1*1024*1024)
	go func() {
		timerOut := 0
		for !_self.isQuitTimer {
			timerOut = timerOut + 1
			if timerOut > 5 {
				timerOut = 0
				if !_self.isNeedAutoReconnect {
					continue
				}
				if !_self.isConnected {
					_self.ConnectToServer(_self.serverAddr, _self.serverPort)
				}
			}

			time.Sleep(time.Second)
		}

	}()
	go func() {
		for {
			data, isOK := <-_self.writeQueue
			if !isOK {
				log.Info("tcp client write err")
				break
			}
			_, err := _self.conn.Write(data)
			if err != nil {
				log.Error("socket send fail")
				return
			}
		}

	}()
}
func (_self *TcpClient) DeInitClient() {
	if !_self.isInit {
		log.Error("DeInitClient no init")
		return
	}
	_self.isQuitTimer = true
	_self.Close()
	close(_self.writeQueue)
	_self.writeQueue = nil
	_self.isInit = false

}

// ConnectToServer 连接到服务器
func (_self *TcpClient) ConnectToServer(serverAddr string, serverPort uint16) {
	if !_self.isInit {
		log.Info("client no init,forget call InitClient?")
		return
	}
	if _self.isConnected {
		log.Info("connected addr:", serverAddr, " port:", serverPort)
		return
	}
	_self.serverAddr = serverAddr
	_self.serverPort = serverPort
	log.Info("start connect server,remote addr:", serverAddr, " port:", serverPort)
	var err error
	remoteAddr := serverAddr + ":" + strconv.Itoa(int(serverPort))
	_self.conn, err = net.DialTimeout("tcp", remoteAddr, time.Second*5)
	if err != nil {
		log.Info("connect failed, err :", err.Error())
		_self.isConnected = false
		_self.cbConnectStateFunc(false)
		return
	}
	_self.isConnected = true
	if nil != _self.cbConnectStateFunc {
		_self.cbConnectStateFunc(true)
	}

	go _self.handleConnect()
}

// SendData 发送数据,支持并发发送
func (_self *TcpClient) SendData(data []byte) bool {
	if !_self.isConnected || !_self.isInit {
		log.Info("sendData fail,no connect server")
		return false
	}
	_self.writeQueue <- data
	return true
}

// Close 主动关闭
func (_self *TcpClient) Close() {
	if !_self.isConnected || _self.isInit {
		log.Info("sendData fail,no connect server")
		return
	}
	err := _self.conn.Close()
	if err != nil {
		log.Error("tcp client close fail", err.Error())
		return
	}
	_self.isConnected = false
}

func (_self *TcpClient) handleConnect() {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Error("handleConnect close socket fail", err.Error())
		}
	}(_self.conn)
	var buf [MAX_PACKET_SIZE]byte
	for {
		n, err := _self.conn.Read(buf[0:])
		if err != nil {
			log.Info("tcp read err:", err)
			break
			//if err == io.EOF {
			//	continue
			//} else {
			//
			//}
		}
		if nil != _self.cbRecvDataFunc {
			_self.cbRecvDataFunc(buf[0:n])
		}
	}
	_self.isConnected = false
	if nil != _self.cbConnectStateFunc {
		_self.cbConnectStateFunc(false)
	}
}
