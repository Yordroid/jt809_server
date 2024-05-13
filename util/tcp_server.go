package util

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
)

// 最大包大小
const VS_MAX_PACKET_SIZE = 60 * 1024

// onServerReceivedDataCallback 接收数据回调
type onServerReceivedDataCallback func(clientID int64, data []byte)

// onConnectStateCallback 连接状态回调
type onServerConnectStateCallback func(clientID int64, isConnected bool)

type clientSession struct {
	conn net.Conn
}
type TcpServer struct {
	listen             net.Listener
	cbReceivedDataFunc onServerReceivedDataCallback
	cbConnectStateFunc onServerConnectStateCallback
	mapClientSession   map[int64]*clientSession
	mutexMap           sync.Mutex
	writeQueue         chan []byte
}

func (_self *TcpServer) InitServer(bindPort uint16, cbStatusFunc onServerConnectStateCallback, cbReceivedFunc onServerReceivedDataCallback) {
	_self.mapClientSession = make(map[int64]*clientSession, 1000)
	_self.cbReceivedDataFunc = cbReceivedFunc
	_self.cbConnectStateFunc = cbStatusFunc
	var err error
	_self.listen, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", bindPort))
	if err != nil {
		log.Error("Listen() failed, err: ", err.Error())
		return
	}
	log.Info("InitServer success:", bindPort)
	go func() {
		for {
			conn, errAccept := _self.listen.Accept() // 监听客户端的连接请求
			if errAccept != nil {
				log.Error("Accept() failed, err: ", errAccept.Error())
				break
			}
			newClientID := GetSnowflakeID()

			go _self.processNewClient(newClientID, conn) // 启动一个goroutine来处理客户端的连接请求
		}
	}()

}
func (_self *TcpServer) DeInitServer() {
	_self.mutexMap.Lock()
	for _, curSession := range _self.mapClientSession {
		curSession.conn.Close()
	}
	_self.mapClientSession = make(map[int64]*clientSession)
	_self.mutexMap.Unlock()
	err := _self.listen.Close()
	if err != nil {
		return
	}
}

// SendDataToClient 发送数据,支持并发发送
func (_self *TcpServer) SendDataToClient(clientID int64, data []byte) int {
	_self.mutexMap.Lock()
	curClient, isExist := _self.mapClientSession[clientID]
	if !isExist {
		_self.mutexMap.Unlock()
		log.Error("tcp server SendDataToClient fail", clientID)
		return 0
	}
	_self.mutexMap.Unlock()
	nWriteSize, err := curClient.conn.Write(data)
	if err != nil {
		log.Error("tcp server SendDataToClient fail", err.Error())
		return 0
	}
	return nWriteSize
}

// Close 主动关闭
func (_self *TcpServer) Close(clientID int64) bool {
	_self.mutexMap.Lock()
	defer _self.mutexMap.Unlock()
	curClient, isExist := _self.mapClientSession[clientID]
	if !isExist {
		log.Error("tcp server active close fail,client session no exist", clientID)
		return false
	}
	err := curClient.conn.Close()
	if err != nil {
		log.Error("tcp server active close fail")
		return false
	}
	return true
}

//内部函数

func (_self *TcpServer) processNewClient(clientID int64, conn net.Conn) {
	defer conn.Close()
	_self.mutexMap.Lock()
	curSession := &clientSession{}
	curSession.conn = conn
	_self.mapClientSession[clientID] = curSession
	_self.mutexMap.Unlock()
	buf := make([]byte, VS_MAX_PACKET_SIZE)
	if _self.cbConnectStateFunc != nil {
		_self.cbConnectStateFunc(clientID, true)
	}
	for {
		// 从网络中读
		readBytesCount, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				continue
			}
			log.Error("tcp server Failed to read", err.Error())
			if _self.cbConnectStateFunc != nil {
				_self.cbConnectStateFunc(clientID, false)
			}
			break
		}
		if _self.cbReceivedDataFunc != nil {
			_self.cbReceivedDataFunc(clientID, buf[:readBytesCount])
		}

	}
}
