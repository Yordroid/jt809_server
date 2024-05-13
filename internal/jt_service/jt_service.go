package jt_service

import (
	log "github.com/sirupsen/logrus"
	"jt809_server/config"
	"jt809_server/internal/bu_service"
	"jt809_server/internal/data_manage_service"
	"jt809_server/models"
	"jt809_server/util"
	"time"
)

// 用来存上级平台请求的信息,应答时用
type userReqDataContext struct {
	srcMsgSN       uint32 //源消息序列号
	srcSubDataType uint16 //源子业务类型
	vehicleNo      string //车牌号
	vehicleNoColor uint8  //车牌颜色
}

type userTaskContext struct {
	startTime int64 //用来超时处理
	userKey   int64
	userData  userReqDataContext //用户数据
}

type JTServiceContext struct {
	util.GoService
	mapUserContext     map[int64]*jtUserInfoContext //JT用户,key:用户key
	mapUserTaskContext map[int64]*userTaskContext   //用户任务ID映射,key:任务ID,value:用户key
}

func (_self *JTServiceContext) initService() {
	_self.mapUserContext = make(map[int64]*jtUserInfoContext)
	_self.mapUserTaskContext = make(map[int64]*userTaskContext)
	_self.SetTimer(1, 1000, func(uint32) {
		for _, userCtx := range _self.mapUserContext {
			userCtx.onTimerTaskSecond()
		}
		curTime := time.Now().Unix()
		var delTaskIDs []int64
		for taskID, taskInfo := range _self.mapUserTaskContext {
			if curTime < taskInfo.startTime {
				taskInfo.startTime = curTime
			}
			if curTime-taskInfo.startTime > 300 { //300秒任务超时
				delTaskIDs = append(delTaskIDs, taskID)
			}
		}
		for idx := 0; idx < len(delTaskIDs); idx++ {
			delete(_self.mapUserTaskContext, delTaskIDs[idx])
			log.Error("task timeout,id:", delTaskIDs[idx])
		}
	})
	_self.loadUserConfig()
	_self.initRegisterAppMsgHandle()
	_self.SetServiceStartedFinish()
}

func (_self *JTServiceContext) loadUserConfig() {
	jtUserConfigs := config.GetJTUserConfig()
	for idx := 0; idx < len(jtUserConfigs); idx++ {
		curUserConf := jtUserConfigs[idx]
		_self.addJTUserContext(curUserConf)
	}

}

func (_self *JTServiceContext) addJTUserContext(conf config.JTUserConfigInfo) {
	ctx, isExist := _self.mapUserContext[conf.UserKey]
	if isExist {
		//先释放原来的,再重新创建
		ctx.deInit()
	}
	ctx = &jtUserInfoContext{}
	ctx.initContext(_self, conf)
	_self.mapUserContext[conf.UserKey] = ctx
	bu_service.MsgServiceIns().AddBuUserInfo(conf.BuUserName, conf.BuUserPwdMD5, conf.UserKey)
}

func (_self *JTServiceContext) addTaskMap(taskID, userKey int64, userData userReqDataContext) {
	userInfo := userTaskContext{}
	userInfo.userKey = userKey
	userInfo.startTime = time.Now().Unix()
	userInfo.userData = userData
	_self.mapUserTaskContext[taskID] = &userInfo
	log.Info("addTaskMap taskID:", taskID, " userKey:", userKey, " userData:", userData)
}

func (_self *JTServiceContext) initRegisterAppMsgHandle() {
	_self.RegisterServiceMsgFunc(models.VS_APP_SERVICE_MSG_ID_DEV_ONOFF_LINE, _self.onHandleServiceMsgDevOnOffline)
	_self.RegisterServiceMsgFunc(models.VS_APP_SERVICE_MSG_ID_GPS_DATA, _self.onHandleServiceMsgDevGpsData)
	_self.RegisterServiceMsgFunc(models.VS_APP_SERVICE_MSG_ID_ALARM_DATA, _self.onHandleServiceMsgAlarmData)
	_self.RegisterServiceMsgFunc(models.VS_APP_SERVICE_MSG_ID_SWIPE_CARD, _self.onHandleServiceMsgSwipeCardData)
	_self.RegisterServiceMsgFunc(models.VS_APP_SERVICE_MSG_ID_MULTI_MEDIA_INFO, _self.onHandleServiceMsgMultiMediaInfo)
	_self.RegisterServiceMsgFunc(models.VS_APP_SERVICE_MSG_ID_DEV_ATTACH_INFO_COMPLETE, _self.onHandleServiceMsgAttachmentComplete)
}

// 根据车辆ID找到需要处理的用户处理部标对象
func (_self *JTServiceContext) foreachToUserJTHandleByVehicleID(vehicleID uint32, handleCtxFunc func(ctx *MsgHandleContext)) {
	data_manage_service.DataManageServiceIns().ForeachUserVehicleByVehicleID(vehicleID, func(userKey int64) {
		jtUserCtx, isExist := _self.mapUserContext[userKey]
		if isExist {
			if handleCtxFunc != nil {
				handleCtxFunc(jtUserCtx.msgHandleCtx)
			}
		}
	})
}

// 根据车辆ID找到需要处理的用户处理对象
func (_self *JTServiceContext) foreachToUserContextByVehicleID(vehicleID uint32, handleCtxFunc func(ctx *jtUserInfoContext)) {
	data_manage_service.DataManageServiceIns().ForeachUserVehicleByVehicleID(vehicleID, func(userKey int64) {
		if handleCtxFunc != nil {
			jtUserCtx, isExist := _self.mapUserContext[userKey]
			if isExist {
				handleCtxFunc(jtUserCtx)
			}

		}
	})
}

// 根据任务ID找到需要处理的用户处理对象
func (_self *JTServiceContext) foreachToUserContextByTaskID(taskID int64, handleCtxFunc func(ctx *jtUserInfoContext, userData userReqDataContext)) {
	taskCtx, isExist := _self.mapUserTaskContext[taskID]
	if isExist {
		if handleCtxFunc != nil {
			jtUserCtx, isExist := _self.mapUserContext[taskCtx.userKey]
			if isExist {
				handleCtxFunc(jtUserCtx, taskCtx.userData)
			}

		}
		delete(_self.mapUserTaskContext, taskID)
	}
}

func (_self *JTServiceContext) onHandleServiceMsgDevOnOffline(reqObject util.MsgObject) (bool, util.MsgObject) {
	pbObject, isOK := reqObject.(*models.VsFrontPbMsgOnline)
	if !isOK {
		log.Error("onHandleServiceMsgDevOnOffline convert fail")
		return false, nil
	}
	_self.foreachToUserJTHandleByVehicleID(pbObject.VehicleGuid, func(ctx *MsgHandleContext) {
		ctx.SendExgMsgUpVehicleOnlineUpload(pbObject)
	})
	return true, nil
}

func (_self *JTServiceContext) onHandleServiceMsgDevGpsData(reqObject util.MsgObject) (bool, util.MsgObject) {
	pbObject, isOK := reqObject.(*models.VsPbStorageGpsData)
	if !isOK {
		log.Error("onHandleServiceMsgDevGpsData convert fail")
		return false, nil
	}
	_self.foreachToUserContextByVehicleID(pbObject.VehicleGuid, func(ctx *jtUserInfoContext) {
		ctx.handleGpsData(pbObject)
	})
	return true, nil
}

func (_self *JTServiceContext) onHandleServiceMsgAlarmData(reqObject util.MsgObject) (bool, util.MsgObject) {
	pbObject, isOK := reqObject.(*models.VsPbStorageEventInfo)
	if !isOK {
		log.Error("onHandleServiceMsgAlarmData convert fail")
		return false, nil
	}
	_self.foreachToUserJTHandleByVehicleID(pbObject.VehicleGuid, func(ctx *MsgHandleContext) {
		ctx.SendWarnMsgEventDataUpload(pbObject)
	})
	return true, nil
}

func (_self *JTServiceContext) onHandleServiceMsgAttachmentComplete(reqObject util.MsgObject) (bool, util.MsgObject) {
	jsonObject, isOK := reqObject.(*models.VsFrontAttachUploadCompleteMsg)
	if !isOK {
		log.Error("onHandleServiceMsgAttachmentComplete convert fail")
		return false, nil
	}
	_self.foreachToUserContextByTaskID(int64(jsonObject.AlarmCode), func(ctx *jtUserInfoContext, userData userReqDataContext) {
		ctx.handleAttachmentComplete(jsonObject)
	})

	return true, nil
}

func (_self *JTServiceContext) onHandleServiceMsgSwipeCardData(reqObject util.MsgObject) (bool, util.MsgObject) {
	pbObject, isOK := reqObject.(*models.VsPbStorageSwipeCardInfo)
	if !isOK {
		log.Error("onHandleServiceMsgSwipeCardData convert fail")
		return false, nil
	}
	if pbObject.GpsData == nil {
		log.Error("onHandleServiceMsgSwipeCardData convert fail,gpsData is nil")
		return false, nil
	}
	_self.foreachToUserJTHandleByVehicleID(pbObject.GpsData.VehicleGuid, func(ctx *MsgHandleContext) {
		ctx.SendExgMsgUpDriverInfoUpload(pbObject)
	})
	return true, nil
}
func (_self *JTServiceContext) onHandleServiceMsgMultiMediaInfo(reqObject util.MsgObject) (bool, util.MsgObject) {
	pbObject, isOK := reqObject.(*models.VsPbStorageMultiMediaEventInfo)
	if !isOK {
		log.Error("onHandleServiceMsgMultiMediaInfo convert fail")
		return false, nil
	}
	if pbObject.GpsData == nil {
		log.Error("onHandleServiceMsgMultiMediaInfo convert fail,gpsData is nil")
		return false, nil
	}
	_self.foreachToUserContextByVehicleID(pbObject.GpsData.VehicleGuid, func(ctx *jtUserInfoContext) {
		ctx.handleMultiMediaInfo(pbObject)
		_self.addTaskMap(pbObject.EventID, ctx.conf.UserKey, userReqDataContext{})
	})

	return true, nil
}

func (_self *JTServiceContext) onHandleApiMsgPostQueryResp(jsonObject *models.VsAppMsgPlatformPostQueryResp) {
	_self.foreachToUserContextByTaskID(models.GetTaskIDFromString(jsonObject.TaskID), func(ctx *jtUserInfoContext, userData userReqDataContext) {
		ctx.msgHandleCtx.SendPlatformMsgPostQueryAck(jsonObject, userData)
	})
}

func (_self *JTServiceContext) onHandleApiMsgWarnUrgeTodoResp(jsonObject *models.VsAppMsgWarnUrgeTodoResp) {
	_self.foreachToUserContextByTaskID(models.GetTaskIDFromString(jsonObject.TaskID), func(ctx *jtUserInfoContext, userData userReqDataContext) {
		ctx.msgHandleCtx.SendWarnMsgUrgeTodoResp(jsonObject, userData)
	})
}
