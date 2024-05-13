package jt_service

import (
	"fmt"
	"github.com/edwingeng/deque"
	log "github.com/sirupsen/logrus"
	"jt809_server/config"
	jt_session "jt809_server/internal/jt_service/session_mananage"
	"jt809_server/models"
	"jt809_server/util"
	"time"
)

//管理每个用户的会话信息

type jtUserInfoContext struct {
	msgHandleCtx           *MsgHandleContext
	jtSession              *jt_session.UserSession
	mapJTMsgPacketHandle   map[uint16]jt_session.OnNetMsgPacketFunc
	vehicleAutoReUploadMap map[uint32]deque.Deque
	conf                   config.JTUserConfigInfo
	parent                 *JTServiceContext
	lastCheckReUploadTime  int64 //上次重传时间
}

func (_self *jtUserInfoContext) postToService(asyncFunc util.PostTaskFunc) {
	_self.parent.PostTask(asyncFunc)
}

func (_self *jtUserInfoContext) initContext(parent *JTServiceContext, conf config.JTUserConfigInfo) {
	_self.conf = conf
	_self.parent = parent
	_self.vehicleAutoReUploadMap = make(map[uint32]deque.Deque, 10000)
	isPrintHex := false
	if conf.IsPrintHex > 0 {
		isPrintHex = true
	}
	_self.msgHandleCtx = NewMsgHandleContext(_self)
	sessionConf := jt_session.UserSessionConfig{
		BindPort:   conf.BindPort,
		RemoteAddr: conf.RemoteAddr,
		RemotePort: conf.RemotePort,
		Version:    jt_session.JT_VER_TYPE(conf.JTVer),
		M1:         conf.M1,
		IA1:        conf.IA1,
		IC1:        conf.IC1,
		Key:        conf.Key,
		AccessID:   conf.AccessID,
		IsPrintHex: isPrintHex,
	}
	_self.jtSession = jt_session.NewUserSession(sessionConf, func(isMainLink, isConnect bool, sessionID int64) {
		_self.postToService(func() {
			log.Info("session status change:isMainLink", isMainLink, " isConnect:", isConnect, " sessionID:", sessionID)
			if isMainLink && isConnect {
				//到时根据配置来
				reqObject := models.VsAppMsgLoginInfoReq{
					UserID:       conf.JTUserID,
					Password:     conf.JTUserPwd,
					DownLinkIP:   conf.PublicAddr,
					DownLinkPort: conf.BindPort,
					AccessID:     conf.AccessID,
				}
				_self.msgHandleCtx.SendLinkManageUpLoginReq(&reqObject)
			}
		})
	}, func(hdrObject jt_session.VsAppMsgHeader, bodyData []byte) bool {
		_self.msgHandleCtx.updateLastAliveTime()
		_self.postToService(func() {

			if _self.mapJTMsgPacketHandle == nil {
				log.Error("received handle fail,mapJTMsgPacketHandle is nil ,hdrObject", util.ToJson(hdrObject))
				return
			}
			cbFunc, isExist := _self.mapJTMsgPacketHandle[hdrObject.MsgID]
			if !isExist {
				log.Error("received handle fail,dataType no reg", fmt.Sprintf("0x%X", hdrObject.MsgID))
				return
			}
			cbFunc(hdrObject, bodyData)
		})
		return true
	})
	_self.initRegisterJTMsgHandle()
	_self.jtSession.StartMainConnect()

}

func (_self *jtUserInfoContext) deInit() {
	if _self.jtSession != nil {
		_self.jtSession.DeInitUserSession()
	}
}

// onTimerTaskSecond 秒定时任务
func (_self *jtUserInfoContext) onTimerTaskSecond() {
	_self.msgHandleCtx.onTimerTaskSecond()
	hasTask := util.HasTimerTask(&_self.lastCheckReUploadTime, 10*1000)
	if hasTask {
		_self.onTimerHandleAutoUploadGps()
	}
}

func (_self *jtUserInfoContext) initRegisterJTMsgHandle() {
	_self.mapJTMsgPacketHandle = make(map[uint16]jt_session.OnNetMsgPacketFunc)
	//链路管理
	_self.addRegisterNetRouter(models.VS_JT_TYPE_UP_LOGIN_RESP, _self.msgHandleCtx.onHandleLinkManageUpLoginResp)
	_self.addRegisterNetRouter(models.VS_JT_TYPE_UP_LINK_KEEP_ALIVE_RESP, _self.msgHandleCtx.onHandleLinkManageUpKeepLiveResp)
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_LOGIN_REQ, _self.msgHandleCtx.onHandleLinkManageDownLoginReq)
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_LINK_KEEP_ALIVE_REQ, _self.msgHandleCtx.onHandleLinkManageDownKeepLiveReq)

	//平台间的交互
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_PLAFORM_MSG, _self.onHandleSubBuMsg)
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_PLATFORM_MSG_POST_QUERY_REQ, _self.msgHandleCtx.onHandlePlatformMsgPostQueryReq)
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_PLATFORM_MSG_INFO_REQ, _self.msgHandleCtx.onHandlePlatformMsgTextInfoReq)
	//报警交互
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_WARN_MSG, _self.onHandleSubBuMsg)
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_WARN_MSG_URGE_TODO_REQ, _self.msgHandleCtx.onHandleWarnMsgUrgeTodoReq)
	//监管类
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_CTRL_MSG, _self.onHandleSubBuMsg)
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_CTRL_MSG_TAKE_PHOTO_REQ, _self.msgHandleCtx.onHandleCtrlMsgTakePhotoReq)
	//静态类
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_BASE_MSG, _self.onHandleSubBuMsg)
	_self.addRegisterNetRouter(models.VS_JT_TYPE_DOWN_BASE_MSG_VEHICLE_ADDED, _self.msgHandleCtx.onHandleBaseMsgVehicleStaticInfoReq)
}

// 消息体是否存在车辆信息
func (_self *jtUserInfoContext) bodyDataHasVehicleInfo(dataType uint16) bool {
	if dataType == models.VS_JT_TYPE_DOWN_CTRL_MSG ||
		dataType == models.VS_JT_TYPE_DOWN_BASE_MSG {
		return true
	}
	if _self.conf.JTVer == jt_session.JT_VER_V3 {
		return false
	} else if _self.conf.JTVer == jt_session.JT_VER_V2 {
		if dataType == models.VS_JT_TYPE_DOWN_WARN_MSG {
			return true
		}
	} else if _self.conf.JTVer == jt_session.JT_VER_V1 {
		if dataType == models.VS_JT_TYPE_DOWN_WARN_MSG {
			return true
		}
	}
	return false
}

// 处理存在子业务消息
func (_self *jtUserInfoContext) onHandleSubBuMsg(hdrObject jt_session.VsAppMsgHeader, bodyMsg []byte) bool {
	bodyLen := uint32(len(bodyMsg))
	if bodyLen < 6 {
		log.Error("onHandleSubBuMsg body msg len < 6 err")
		return false
	}
	var headLen uint32 = 6
	respObject := models.VsAppMsgCommonSubBuInfo{}
	readObject := util.NewBinaryRead(bodyMsg)
	hasVehicleInfo := _self.bodyDataHasVehicleInfo(hdrObject.MsgID)
	if hasVehicleInfo {
		respObject.VehicleNo = readObject.ReadStringToUTF8(models.VS_JT_VEHCLE_NO_LEN)
		respObject.VehicleNoColor = readObject.ReadByte()
		hdrObject.VehicleNoColor = respObject.VehicleNoColor
		hdrObject.VehicleNo = respObject.VehicleNo
		headLen += models.VS_JT_VEHCLE_NO_LEN + 1
	}
	respObject.SubDataType = readObject.ReadInt16()
	respObject.SubDataLen = readObject.ReadInt32()
	if respObject.SubDataLen != bodyLen-headLen {
		log.Error("onHandleSubBuMsg body msg len no match,bodyLen:", bodyLen, " SubDataLen:", respObject.SubDataLen)
		return false
	}
	cbFunc, isExist := _self.mapJTMsgPacketHandle[respObject.SubDataType]
	if !isExist {
		log.Error("received handle sub bu fail,dataType no reg", fmt.Sprintf("0x%X", respObject.SubDataType))
		return false
	}
	hdrObject.MsgID = respObject.SubDataType
	cbFunc(hdrObject, bodyMsg[headLen:])
	return true
}

func (_self *jtUserInfoContext) addRegisterNetRouter(dataType uint16, onNetMsgPacketFunc jt_session.OnNetMsgPacketFunc) {
	if _self.mapJTMsgPacketHandle == nil {
		log.Error("addRegisterNetRouter fail,mapJTMsgPacketHandle is nil")
		return
	}
	_, isExist := _self.mapJTMsgPacketHandle[dataType]
	if isExist {
		log.Error("addRegisterNetRouter fail,dataType add repeated:", dataType)
		return
	}
	_self.mapJTMsgPacketHandle[dataType] = onNetMsgPacketFunc
}

func (_self *jtUserInfoContext) SendDataToServer(dataType models.VsJTDataType, bodyData []byte, isMainLink bool) bool {
	if _self.jtSession != nil {
		return _self.jtSession.BuildJTDataAndSend(uint16(dataType), bodyData, isMainLink)
	}
	return false
}

func (_self *jtUserInfoContext) onTimerHandleAutoUploadGps() {
	for _, curQueue := range _self.vehicleAutoReUploadMap {
		var gpsInfos []*models.VsPbStorageGpsData
		for {
			curGpsInfo := curQueue.PopFront()
			if curGpsInfo == nil {
				if len(gpsInfos) > 0 {
					_self.msgHandleCtx.sendExgMsgUpAutoGpsDataUpload(gpsInfos)
				}
				gpsInfos = nil
				break
			}
			gpsInfos = append(gpsInfos, curGpsInfo.(*models.VsPbStorageGpsData))
			if len(gpsInfos) > 4 {
				_self.msgHandleCtx.sendExgMsgUpAutoGpsDataUpload(gpsInfos)
				gpsInfos = nil
			}
		}
	}

}
func (_self *jtUserInfoContext) handleGpsData(pbObject *models.VsPbStorageGpsData) {
	//是否为补传
	if pbObject.AppStatusFlag&(1<<models.VS_PB_APP_STATUS_FLAG_ASF_IS_BLIND_AREA) > 0 {
		curQueue, isExist := _self.vehicleAutoReUploadMap[pbObject.VehicleGuid]
		if !isExist {
			curQueue = deque.NewDeque()
			_self.vehicleAutoReUploadMap[pbObject.VehicleGuid] = curQueue
		}
		curQueue.PushBack(pbObject)
		gpsLen := curQueue.Len()
		if gpsLen > 1000 {
			var gpsInfos []*models.VsPbStorageGpsData
			for {
				curGpsInfo := curQueue.PopFront()
				if curGpsInfo == nil {
					if len(gpsInfos) > 0 {
						_self.msgHandleCtx.sendExgMsgUpAutoGpsDataUpload(gpsInfos)
					}
					gpsInfos = nil
					break
				}
				gpsInfos = append(gpsInfos, curGpsInfo.(*models.VsPbStorageGpsData))
				if len(gpsInfos) > 4 {
					_self.msgHandleCtx.sendExgMsgUpAutoGpsDataUpload(gpsInfos)
					gpsInfos = nil
				}
			}

		}
	} else {
		_self.msgHandleCtx.SendExgMsgUpGpsDataUpload(pbObject)
	}
}

func (_self *jtUserInfoContext) handleMultiMediaInfo(pbObject *models.VsPbStorageMultiMediaEventInfo) {
	if pbObject.GpsData == nil {
		log.Error("handleMultiMediaInfo fail,gpsData is nil", pbObject.EventID)
		return
	}
	eventVehicle := eventVehicleMapInfo{}
	eventVehicle.vehicleNo = pbObject.GpsData.VehicleNo
	eventVehicle.startTime = time.Now().Unix()
	_self.msgHandleCtx.mapEventIDVehicleNo[uint64(pbObject.EventID)] = &eventVehicle
	log.Info("handleMultiMediaInfo", pbObject.EventID, " vehicleNo:", pbObject.GpsData.VehicleNo)
}

func (_self *jtUserInfoContext) handleAttachmentComplete(reqObject *models.VsFrontAttachUploadCompleteMsg) {
	vehicleTaskObject := _self.msgHandleCtx.getVehicleTaskInfoByEventID(reqObject.AlarmCode)
	if vehicleTaskObject == nil {
		//log.Error("onHandleServiceMsgAttachmentComplete not found alarmCode:", jsonObject.AlarmCode)
		return
	}
	go util.GoTaskFunc("getAttachment", func() {
		httpClient := util.ClientInfo{}
		httpClient.InitHTTPClient()
		httpClient.SetTimeout(10)
		url := fmt.Sprintf("%s/web_api/v1/media/get_file?alarmID=%d&fileName=%s", config.GetWebUrlBase(),
			reqObject.AlarmCode, reqObject.FileName)
		strResp := httpClient.SendURLRequest(url, "")
		if len(strResp) == 0 {
			log.Error("onHandleServiceMsgAttachmentComplete not found fail:", url)
			return
		}
		_self.msgHandleCtx.sendCtrlMsgTakePhotoResp(vehicleTaskObject.takePhotoInfo, VS_TAKE_PHOTO_RESULT_FINISH, []byte(strResp))
		_self.postToService(func() {
			_self.msgHandleCtx.deleteTakePhotoTask(vehicleTaskObject.takePhotoInfo.vehicleNo)
		})

	})
}
