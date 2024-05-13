package util

import (
	"context"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

// 服务间通信的起始ID
const VS_SERVICE_MSG_NOTIFY_ID_START = 0x100000
const IOT_DEFAULT_TIMEOUT = 10 //10S
type MsgObject any

// 处理消息函数
type HandleMsgFunc func(reqObject MsgObject) (bool, MsgObject)
type AsyncHandleMsgFunc func(isSuccess bool, respObject MsgObject)
type PostTaskFunc func()
type SyncTaskFunc func() (bool, interface{})

// 消息框架
type IotMsgNotifyFramework struct {
	//服务对象
	mapMutex sync.Mutex
	//key服务唯一ID
	mapServiceObject map[int64]GoServiceI
	//key:消息类型，value服务id
	mapMsgTypeRegService map[uint32]map[int64]struct{}
}

func (_self *IotMsgNotifyFramework) InitNotifyFrame() {
	_self.mapServiceObject = make(map[int64]GoServiceI)
	_self.mapMsgTypeRegService = make(map[uint32]map[int64]struct{})
}

func (_self *IotMsgNotifyFramework) addService(serviceBase GoServiceI) {
	_self.mapMutex.Lock()
	defer _self.mapMutex.Unlock()
	serviceID := serviceBase.GetServiceID()
	if _, isExist := _self.mapServiceObject[serviceID]; !isExist {
		_self.mapServiceObject[serviceID] = serviceBase
	}
}

func (_self *IotMsgNotifyFramework) removeService(serviceBase GoServiceI) {
	_self.mapMutex.Lock()
	defer _self.mapMutex.Unlock()
	delete(_self.mapServiceObject, serviceBase.GetServiceID())
}

func (_self *IotMsgNotifyFramework) bindRegMsgService(msgType uint32, serviceBase GoServiceI) {
	_self.mapMutex.Lock()
	defer _self.mapMutex.Unlock()
	uniqueID := serviceBase.GetServiceID()
	mapService, isExistServiceID := _self.mapMsgTypeRegService[msgType]
	if !isExistServiceID {
		curServiceIDMap := make(map[int64]struct{})
		curServiceIDMap[uniqueID] = struct{}{}
		_self.mapMsgTypeRegService[msgType] = curServiceIDMap
	} else {
		mapService[uniqueID] = struct{}{}
	}
}

func (_self *IotMsgNotifyFramework) unbindRegMsgService(msgType uint32, serviceBase GoServiceI) {
	_self.mapMutex.Lock()
	defer _self.mapMutex.Unlock()
	mapService, isExistServiceID := _self.mapMsgTypeRegService[msgType]
	if isExistServiceID {
		delete(mapService, serviceBase.GetServiceID())
	}
}

// 服务通信结构
type serviceReqPacket struct {
	msgType   uint32
	msgObject MsgObject
	//async
	respFunc   AsyncHandleMsgFunc
	reqTimeout uint32
	//sync
	respCh chan serviceRespPacket
}

type serviceRespPacket struct {
	isOK      bool
	msgObject MsgObject
}

type serviceSyncTaskPacket struct {
	funcTask SyncTaskFunc
	respCh   chan serviceRespPacket
}

// 服务定义
type GoServiceI interface {
	StartService(serviceName string, msgPacketMaxNum uint32, notifyFramework *IotMsgNotifyFramework, funcStarted func()) bool
	StopService() bool
	SyncRequest(msgType uint32, reqObject MsgObject, timeout uint32) (bool, MsgObject)
	AsyncRequest(msgType uint32, reqObject MsgObject, respFunc AsyncHandleMsgFunc, timeout uint32)
	PostTask(taskFunc PostTaskFunc) bool
	SyncTask(taskFunc SyncTaskFunc, timeout uint32) (bool, interface{})
	PublishMsg(msgType uint32, reqObject MsgObject)
	GetServiceID() int64
	sendMsgObject(msgType uint32, reqObject MsgObject)
	SetTimer(timerID uint32, timeMs int64, timerFunc OnTimerFunc)
	StopTimer(timerID uint32)
	GetServiceName() string
}
type OnTimerFunc func(uint32)
type timerInfo struct {
	durationMs int64
	lastTimeMs int64
	timerFunc  OnTimerFunc
}
type GoService struct {
	mapMsgFunc           map[uint32]HandleMsgFunc
	isInitFinish         bool
	funcTaskCh           chan PostTaskFunc
	syncTaskCh           chan serviceSyncTaskPacket
	svrPacketCh          chan serviceReqPacket
	quitCh               chan bool
	serviceName          string
	serviceId            int64
	routineID            uint64
	notifyFramework      *IotMsgNotifyFramework
	isSyncWait           bool
	mapTimerTask         map[uint32]*timerInfo
	funcStarted          func()
	isWaitStartedFinish  bool
	maxPacketNum         uint32
	isNeedPrintPacketLog bool
	lostPacketNum        uint32
}

func (_self *GoService) GetServiceID() int64 {
	return _self.serviceId
}

func (_self *GoService) GetServiceName() string {
	return _self.serviceName
}

func (_self *GoService) SetServiceStartedFinish() {
	if !_self.isWaitStartedFinish {
		log.Info("SetServiceStartedFinish:", _self.serviceName)
		_self.isWaitStartedFinish = true
	}
}

func (_self *GoService) StartService(serviceName string, msgPacketMaxNum uint32, notifyFramework *IotMsgNotifyFramework, funcStarted func()) bool {
	if nil == notifyFramework || _self.isInitFinish {
		log.Info("start service fail, notify is nil or already init", serviceName)
		return false
	}

	_self.notifyFramework = notifyFramework
	_self.isInitFinish = false
	_self.serviceName = serviceName
	_self.funcStarted = funcStarted
	if msgPacketMaxNum == 0 {
		msgPacketMaxNum = 1000
	}
	_self.serviceId = GetSnowflakeID()
	_self.quitCh = make(chan bool)
	_self.funcTaskCh = make(chan PostTaskFunc, msgPacketMaxNum)
	_self.syncTaskCh = make(chan serviceSyncTaskPacket, msgPacketMaxNum)
	_self.mapMsgFunc = make(map[uint32]HandleMsgFunc)
	_self.svrPacketCh = make(chan serviceReqPacket, msgPacketMaxNum)
	_self.maxPacketNum = msgPacketMaxNum
	_self.mapTimerTask = make(map[uint32]*timerInfo)

	go func() {
		_self.isInitFinish = true
		_self.routineID = GetRoutineID()
		ticker := time.NewTicker(time.Millisecond * 10)
		lastCheckTime := time.Now().Unix()
		lastCheckServiceAliveTime := time.Now().Unix()
		defer ticker.Stop()
		isQuit := false
		if _self.funcStarted != nil {
			log.Info("goroutine start id:", GetRoutineID(), " serviceName:", _self.serviceName)
			_self.funcStarted()
		}
		for {
			select {
			case <-ticker.C:
				//1 min print queue size
				curTime := time.Now().Unix()
				if curTime-lastCheckTime > 60 {
					lastCheckTime = curTime
					if len(_self.svrPacketCh) > int(_self.maxPacketNum/2) {
						log.Info("queue warn ----service name:", _self.serviceName, " useQueueSize: ", len(_self.svrPacketCh), " curMaxQueue:", _self.maxPacketNum)
					}
				}
				for timerID, curTimerInfo := range _self.mapTimerTask {
					curTick := time.Now().UnixMilli()
					if curTick < curTimerInfo.lastTimeMs {
						curTimerInfo.lastTimeMs = curTick
					}
					diff := curTick - curTimerInfo.lastTimeMs
					if diff > curTimerInfo.durationMs {
						curTimerInfo.timerFunc(timerID)
						curTimerInfo.lastTimeMs = curTick
					}
				}
				hasTask := HasTimerTask(&lastCheckServiceAliveTime, 1800*1000)
				if hasTask {
					log.Info("service alive ,name:", _self.getServiceBaseInfo(), " totalServiceNum:", len(_self.notifyFramework.mapServiceObject))
				}
				break
			case curSync := <-_self.syncTaskCh:
				isOK, respObject := curSync.funcTask()
				curRespPacket := serviceRespPacket{}
				curRespPacket.isOK = isOK
				curRespPacket.msgObject = respObject
				curSync.respCh <- curRespPacket
				break
			case curFunc := <-_self.funcTaskCh:
				curFunc()
				break
			case <-_self.quitCh:
				isQuit = true
				break
			case curPacket := <-_self.svrPacketCh:
				var curTicker *time.Ticker
				if curPacket.reqTimeout > 0 {
					curTicker = time.NewTicker(time.Duration(curPacket.reqTimeout) * time.Second)
					go func() {
						<-curTicker.C
						curTicker.Stop()
						curPacket.respFunc(false, nil)
					}()
				}
				if hFunc, isExist := _self.mapMsgFunc[curPacket.msgType]; isExist {
					isOK, respObject := hFunc(curPacket.msgObject)
					curRespPacket := serviceRespPacket{}
					curRespPacket.isOK = isOK
					curRespPacket.msgObject = respObject
					if nil != curPacket.respCh {
						curPacket.respCh <- curRespPacket
					}
					if curPacket.respFunc != nil && curTicker != nil {
						curTicker.Stop()
						curPacket.respFunc(curRespPacket.isOK, curRespPacket.msgObject)
					}

				}
				break
			}

			if isQuit {
				break
			}
		}
		log.Info("service normal quit:", _self.serviceName)
	}()

	for {
		if _self.isInitFinish {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	log.Info("service start finish,name:", _self.getServiceBaseInfo())
	notifyFramework.addService(_self)
	tryCount := 0
	for {
		if _self.isWaitStartedFinish {
			break
		}
		time.Sleep(time.Millisecond * 50)
		tryCount++
		if tryCount > 100 {
			tryCount = 0
			log.Error("service start fail,forget call setServiceFinish? name:", _self.getServiceBaseInfo())
		}
	}
	return true
}

func (_self *GoService) StopService() bool {
	_self.quitCh <- true
	_self.notifyFramework.removeService(_self)
	return true
}

func (_self *GoService) SetTimer(timerID uint32, timeMs int64, timerFunc OnTimerFunc) {
	_self.PostTask(func() {
		curInfo := timerInfo{}
		curInfo.timerFunc = timerFunc
		curInfo.lastTimeMs = time.Now().UnixMilli()
		curInfo.durationMs = timeMs
		_, isExist := _self.mapTimerTask[timerID]
		if isExist {
			log.Error("SetTimer fail,timer id repeated")
			return
		}
		_self.mapTimerTask[timerID] = &curInfo
	})
}
func (_self *GoService) StopTimer(timerID uint32) {
	_self.PostTask(func() {
		delete(_self.mapTimerTask, timerID)
	})
}

func (_self *GoService) RegisterServiceMsgFunc(msgType uint32, msgFunc HandleMsgFunc) {
	if !_self.isInitFinish {
		log.Info("RegisterServiceMsgFunc not init finish", _self.getServiceBaseInfo())
		return
	}
	if _, isExist := _self.mapMsgFunc[msgType]; isExist {
		log.Info("register func already exist", msgType)
		return
	}
	_self.mapMsgFunc[msgType] = msgFunc
	_self.notifyFramework.bindRegMsgService(msgType, _self)
	log.Info("register msg success,msgType", msgType, " serviceInfo", _self.getServiceBaseInfo())
}

func (_self *GoService) SyncRequest(msgType uint32, reqObject MsgObject, timeout uint32) (bool, MsgObject) {
	if !_self.isInitFinish {
		log.Info("SyncRequest fail, not init finish", _self.getServiceBaseInfo())
		return false, nil
	}
	curRoutineID := GetRoutineID()
	if _self.routineID == curRoutineID {
		if hFunc, isExist := _self.mapMsgFunc[msgType]; isExist {
			return hFunc(reqObject)
		} else {
			return false, nil
		}
	}
	if _self.isSyncWait {
		PrintException("not allow call circle sync wait req")
		return false, nil
	}
	_self.isSyncWait = true
	packet := serviceReqPacket{}
	packet.msgObject = reqObject
	packet.msgType = msgType
	packet.respCh = make(chan serviceRespPacket, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	_self.svrPacketCh <- packet
	select {
	case curRespPacket := <-packet.respCh:
		_self.isSyncWait = false
		return curRespPacket.isOK, curRespPacket.msgObject
	case <-ctx.Done():
		_self.isSyncWait = false
		return false, nil
	}
}

// AsyncRequest 异步请求,请求别的服务,只允许一个服务注册,回调在自己的协程
func (_self *GoService) AsyncRequest(msgType uint32, reqObject MsgObject, respFunc AsyncHandleMsgFunc, timeout uint32) {
	if respFunc == nil {
		log.Error("AsyncRequest fail,respFunc is nil", msgType)
		return
	}
	packet := serviceReqPacket{}
	_self.notifyFramework.mapMutex.Lock()

	mapService, isExist := _self.notifyFramework.mapMsgTypeRegService[msgType]
	if !isExist {
		log.Error("AsyncRequest fail,not reg msgType", msgType)
		_self.notifyFramework.mapMutex.Unlock()
		respFunc(false, nil)
		return
	}
	if len(mapService) != 1 {
		log.Error("AsyncRequest fail,only support 1 service reg,curNum:", len(mapService))
		_self.notifyFramework.mapMutex.Unlock()
		respFunc(false, nil)
		return
	}
	var nServiceID int64
	for serviceID, _ := range mapService {
		nServiceID = serviceID
	}
	serviceObjectI, isExist := _self.notifyFramework.mapServiceObject[nServiceID]
	if !isExist {
		_self.notifyFramework.mapMutex.Unlock()
		log.Error("AsyncRequest fail,service not exist:", nServiceID)
		respFunc(false, nil)
		return
	}
	_self.notifyFramework.mapMutex.Unlock()
	packet.msgType = msgType
	packet.msgObject = reqObject
	packet.respFunc = func(isSuccess bool, respObject MsgObject) {
		_self.PostTask(func() {
			respFunc(isSuccess, respObject)
		})
	}
	packet.reqTimeout = timeout
	packet.respCh = nil
	serviceObject := serviceObjectI.(*GoService)
	serviceObject.svrPacketCh <- packet
}

func (_self *GoService) PublishMsg(msgType uint32, reqObject MsgObject) {
	if !_self.isInitFinish {
		log.Info("PublishMsg fail, not init finish", _self.getServiceBaseInfo())
		return
	}
	_self.notifyFramework.mapMutex.Lock()
	defer _self.notifyFramework.mapMutex.Unlock()
	mapServiceID, isExistServiceID := _self.notifyFramework.mapMsgTypeRegService[msgType]
	if isExistServiceID {
		for key, _ := range mapServiceID {
			serviceObject, isExistService := _self.notifyFramework.mapServiceObject[key]
			if isExistService {
				if serviceObject.GetServiceID() == _self.serviceId {
					continue
				}
				serviceObject.sendMsgObject(msgType, reqObject)
			}
		}
	}
}

func (_self *GoService) PostTask(taskFunc PostTaskFunc) bool {
	if !_self.isInitFinish {
		log.Info("PostTask fail, not init finish", _self.getServiceBaseInfo())
		return false
	}
	routineID := GetRoutineID()
	if routineID == _self.routineID {
		taskFunc()
		return true
	}
	_self.funcTaskCh <- taskFunc
	return true
}
func (_self *GoService) SyncTask(taskFunc SyncTaskFunc, timeout uint32) (bool, interface{}) {
	if !_self.isInitFinish {
		log.Info("SyncTask fail, not init finish", _self.getServiceBaseInfo())
		return false, nil
	}
	curRoutineID := GetRoutineID()
	if _self.routineID == curRoutineID {
		return taskFunc()
	}
	//if _self.isSyncWait {
	//	PrintException("not allow call circle sync wait req")
	//	return false, nil
	//}
	_self.isSyncWait = true
	packet := serviceSyncTaskPacket{}
	packet.funcTask = taskFunc
	packet.respCh = make(chan serviceRespPacket, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	_self.syncTaskCh <- packet
	select {
	case curRespPacket := <-packet.respCh:
		_self.isSyncWait = false
		return curRespPacket.isOK, curRespPacket.msgObject
	case <-ctx.Done():
		_self.isSyncWait = false
		return false, nil
	}
}

// 内部函数
func (_self *GoService) getServiceBaseInfo() string {
	return _self.serviceName + "----serviceID:" + strconv.FormatInt(_self.serviceId, 10)
}

func (_self *GoService) sendMsgObject(msgType uint32, reqObject MsgObject) {
	packet := serviceReqPacket{}
	packet.msgObject = reqObject
	packet.msgType = msgType
	if len(_self.svrPacketCh) > int(_self.maxPacketNum/2) {
		if _self.isNeedPrintPacketLog {
			_self.isNeedPrintPacketLog = false
			log.Info("sendMsgObject warn,queue over half", _self.serviceName, " maxPacket:", _self.maxPacketNum)
		}
	} else {
		_self.isNeedPrintPacketLog = true
	}
	if uint32(len(_self.svrPacketCh)) > _self.maxPacketNum-1 {
		_self.lostPacketNum++
		if _self.lostPacketNum%1000 == 0 {
			log.Error("sendMsgObject lost,queue over half ,serviceName:", _self.serviceName, " maxPacket:", _self.maxPacketNum, " lostNum:", _self.lostPacketNum, " msgType:", msgType)
		}
		return
	}
	_self.lostPacketNum = 0
	_self.svrPacketCh <- packet
}
