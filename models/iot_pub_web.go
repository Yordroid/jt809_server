package models

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
)

// 所有服务之间通信的 MQ消息订阅定义
// 格式前面是数据源服务,后面是消息类型
// iot_access:表示从消息源于前端接入服务
// iot_rt:消息源于实时分析服务
// iot_schedule:消息源于调度服务
const (
	//接入服务推送
	IOT_TOPIC_IOT_APP_DATA_ACCESS_ONOFFLINE                  = "iot_access.onoffline"                  //接入终端上下线 VsFrontPbMsgOnline
	IOT_TOPIC_IOT_APP_DATA_ACCESS_GPS                        = "iot_access.gps_data"                   //接入实时位置 VsFrontPbBatchGpsData
	IOT_TOPIC_IOT_APP_DATA_ACCESS_ALARM                      = "iot_access.alarm_data"                 //接入报警数据 VsPbStorageEventInfo
	IOT_TOPIC_IOT_APP_DATA_ACCESS_ATTACH                     = "iot_access.attach_info"                //接入附件数据 VsFrontAttachInfoMsg
	IOT_TOPIC_IOT_APP_DATA_ACCESS_ATTACH_COMPLETE            = "iot_access.attach_complete"            //接入附件上传完成 VsFrontAttachUploadCompleteMsg
	IOT_TOPIC_IOT_APP_DATA_ACCESS_SENSOR_DATA_UPLOAD         = "iot_access.sensor_data"                //接入传感器数据上传 VsFrontPbStorageSensorDataBatchUpload
	IOT_TOPIC_IOT_APP_DATA_ACCESS_MULTI_MEDIA_INFO_UPLOAD    = "iot_access.multi_media_info"           //接入多媒体信息上传 VsPbStorageMultiMediaEventInfo
	IOT_TOPIC_IOT_APP_DATA_ACCESS_SWIPE_CARD                 = "iot_access.swipe_card"                 //接入刷卡信息上传 VsPbStorageSwipeCardInfo
	IOT_TOPIC_IOT_APP_DATA_ACCESS_CALL_STOP_UPLOAD           = "iot_access.call_stop"                  //接入报站信息上报 VsPbStorageCallStopsInfo
	IOT_TOPIC_IOT_APP_DATA_ACCESS_SCHEDULE_INFO_DOWNLOAD_REQ = "iot_access.schedule_info_download_req" //接入调度信息下载请求 VsFrontMsgDevScheduleTimeTableInfoReq,mq转到调度服务,通过调度http下发结果

	//实时服务推送
	IOT_TOPIC_IOT_APP_DATA_RT_ONOFFLINE               = "iot_rt.onoffline"        //实时终端上下线,多管理员GUID VsFrontPbMsgOnline
	IOT_TOPIC_IOT_APP_DATA_RT_GPS                     = "iot_rt.gps_data"         //实时位置 VsPbStorageGpsData
	IOT_TOPIC_IOT_APP_DATA_RT_ALARM                   = "iot_rt.alarm_data"       //报警数据 VsPbStorageEventInfo
	IOT_TOPIC_IOT_APP_DATA_RT_ATTACH                  = "iot_rt.attach_data"      //附件数据
	IOT_TOPIC_IOT_APP_DATA_RT_MULTI_MEDIA_INFO_UPLOAD = "iot_rt.multi_media_info" //多媒体信息上传 VsPbStorageMultiMediaEventInfo
	IOT_TOPIC_IOT_APP_DATA_RT_SWIPE_CARD              = "iot_rt.swipe_card"       //刷卡信息上传 VsPbStorageSwipeCardInfo
	IOT_TOPIC_IOT_APP_DATA_RT_CALL_STOP_UPLOAD        = "iot_rt.call_stop"        //报站信息上报 VsPbStorageCallStopsInfo
	//调度消息
	IOT_TOPIC_IOT_APP_DATA_SCHEDULE_CALL_STOP_UPLOAD = "iot_schedule.call_stop"     //调度报站信息上报 VsPbStorageCallStopsInfo
	IOT_TOPIC_IOT_APP_DATA_SCHEDULE_INFO_UPLOAD      = "iot_schedule.schedule_info" //调度信息上报 VsPbStorageScheduleInfo
	//本地存储
	IOT_TOPIC_IOT_APP_DATA_LOCAL_STORAGE_ATTACH_COMPLETE = "iot_local_storage.attach_complete" //接入附件上传完成 VsFrontAttachUploadCompleteMsg
	//存储
	IOT_TOPIC_IOT_APP_DATA_STORAGE_USER_LOG = "iot_storage.user_log" //用户日志数据 IotPbCommonUserLogInfo
)

type IOT_WEB_MSG_TYPE int32

const (
	IOT_WEB_MSG_TYPE_WEB_TO_FRONT_END_PLATFORM_POST_QUERY_PUBLISH IOT_WEB_MSG_TYPE = 8011 //平台查岗信息推送 IotWebMsgWebSocketPlatformPostQueryPublish
	IOT_WEB_MSG_TYPE_WEB_TO_FRONT_END_PLATFORM_TEXT_PUBLISH       IOT_WEB_MSG_TYPE = 8012 //平台报文信息推送 IotWebMsgWebSocketPlatformTextPublish
	IOT_WEB_MSG_TYPE_WEB_TO_FRONT_END_WARN_URGE_TODO_PUBLISH      IOT_WEB_MSG_TYPE = 8013 //报警督办信息推送 IotWebMsgWebSocketWarnUrgeTodoPublish
)

type IOT_HTTP_RESP_CODE int32

const (
	IOT_HTTP_RESP_CODE_NONE               IOT_HTTP_RESP_CODE = 0
	IOT_HTTP_RESP_CODE_OK                 IOT_HTTP_RESP_CODE = 200   //成功
	IOT_HTTP_RESP_CODE_ERR                IOT_HTTP_RESP_CODE = 10000 //通用错误
	IOT_HTTP_RESP_CODE_USER_PWD_ERR       IOT_HTTP_RESP_CODE = 10001 //用户或者密码不对
	IOT_HTTP_RESP_CODE_TOKEN_INVALID      IOT_HTTP_RESP_CODE = 10002 //token失效
	IOT_HTTP_RESP_CODE_INPUT_PARAM        IOT_HTTP_RESP_CODE = 10003 //输入参数有误
	IOT_HTTP_RESP_CODE_GEN_TOKEN_FAIL     IOT_HTTP_RESP_CODE = 10004 //生成token失败
	IOT_HTTP_RESP_CODE_SERVER_INNER_ERR   IOT_HTTP_RESP_CODE = 10005 //服务器内部出错
	IOT_HTTP_RESP_CODE_NOT_PERMISSION     IOT_HTTP_RESP_CODE = 10006 //用户无权限
	IOT_HTTP_RESP_CODE_DB_ERR             IOT_HTTP_RESP_CODE = 10007 //数据库操作失败
	IOT_HTTP_RESP_CODE_DATA_REPEATED      IOT_HTTP_RESP_CODE = 10008 //数据存在
	IOT_HTTP_RESP_CODE_DATA_NOT_FOUND     IOT_HTTP_RESP_CODE = 10009 //数据不存在
	IOT_HTTP_RESP_CODE_OLD_PWD_NOT_MATCH  IOT_HTTP_RESP_CODE = 10010 //旧密码不匹配
	IOT_HTTP_RESP_CODE_NUM_MAX_LIMIT      IOT_HTTP_RESP_CODE = 10011 //数量达到最大
	IOT_HTTP_RESP_CODE_PORT_NOT_ENOUGH    IOT_HTTP_RESP_CODE = 10012 //端口分配不足
	IOT_HTTP_RESP_CODE_RESOURCE_BUSY      IOT_HTTP_RESP_CODE = 10013 //资源忙
	IOT_HTTP_RESP_CODE_DEVICE_OFFLINE     IOT_HTTP_RESP_CODE = 10014 //终端离线
	IOT_HTTP_RESP_CODE_REQ_TIMEOUT        IOT_HTTP_RESP_CODE = 10015 //请求超时
	IOT_HTTP_RESP_CODE_SERVICE_NO_VALID   IOT_HTTP_RESP_CODE = 10016 //服务不可用
	IOT_HTTP_RESP_CODE_DEVICE_NOT_SUPPORT IOT_HTTP_RESP_CODE = 10017 //终端不支持
)

// GetAlarmNameByAlarmType 获取事件名称
func GetAlarmNameByAlarmType(eventType VS_APP_ALARM_TYPE) string {
	eventName := "无"
	switch eventType {
	case VS_APP_ALARM_TYPE_DSM_TIRED:
		eventName = "疲劳"
	case VS_APP_ALARM_TYPE_DSM_CALL:
		eventName = "打电话"
	case VS_APP_ALARM_TYPE_DSM_SMOKE:
		eventName = "抽烟"
	case VS_APP_ALARM_TYPE_DSM_DISTRICT:
		eventName = "分神"
	case VS_APP_ALARM_TYPE_DSM_DRIVER_EXCEPTION:
		eventName = "驾驶员异常"
	case VS_APP_ALARM_TYPE_DSM_AUTO_CAPTURE:
		eventName = "自动抓拍"
	case VS_APP_ALARM_TYPE_DSM_DRIVER_CHANGE_EVENT:
		eventName = "驾驶员变更事件"
	case VS_APP_ALARM_TYPE_ADAS_IMPACT:
		eventName = "向前碰撞"
	case VS_APP_ALARM_TYPE_ADAS_LANE_DEPARTURE:
		eventName = "车道偏离报警"
	case VS_APP_ALARM_TYPE_ADAS_VEHICLE_NEAR:
		eventName = "车距过近报警"
	case VS_APP_ALARM_TYPE_ADAS_PEDESTRIAN_IMPACT:
		eventName = "行人碰撞报警"
	case VS_APP_ALARM_TYPE_ADAS_FREQUENT_LANE_CHANGE:
		eventName = "频繁变道报警"
	case VS_APP_ALARM_TYPE_ADAS_ROAD_SIGN_LIMIT:
		eventName = "道路标识超限报警"
	case VS_APP_ALARM_TYPE_ADAS_BARRIER:
		eventName = "障碍物报警"
	case VS_APP_ALARM_TYPE_ADAS_ROAD_SIGN:
		eventName = "道路标识识别事件"
	case VS_APP_ALARM_TYPE_ADAS_ACTIVE_CAPTURE:
		eventName = "主动抓拍事件"
	case VS_APP_ALARM_TYPE_BSD_BACK:
		eventName = "后方"
	case VS_APP_ALARM_TYPE_BSD_LEFT_BACK:
		eventName = "左后方"
	case VS_APP_ALARM_TYPE_BSD_RIGHT_BACK:
		eventName = "右后方"
	case VS_APP_ALARM_TYPE_TPMS_ALARM:
		eventName = "胎压报警"
	case VS_APP_ALARM_TYPE_OVER_SPEED:
		eventName = "超速"
	case VS_APP_ALARM_TYPE_TIMEOUT:
		eventName = "超时"
	case VS_APP_ALARM_TYPE_GNSS_MODULE_FAULT:
		eventName = "GNSS模块异常"
	case VS_APP_ALARM_TYPE_ANT_SHORT:
		eventName = "天线短路"
	case VS_APP_ALARM_TYPE_ANT_NOT_CONNECT:
		eventName = "天线未接"
	case VS_APP_ALARM_TYPE_POWER_LOW:
		eventName = "低电压异常"
	case VS_APP_ALARM_TYPE_POWER_DOWN:
		eventName = "主电源掉电报警"
	case VS_APP_ALARM_TYPE_TTS:
		eventName = "TTS故障"
	case VS_APP_ALARM_TYPE_IC_CARD:
		eventName = "IC卡故障"
	case VS_APP_ALARM_TYPE_VSS_ERR:
		eventName = "VSS故障"
	case VS_APP_ALARM_TYPE_OIL_ERR:
		eventName = "油量异常"
	case VS_APP_ALARM_TYPE_ILLEGAL_FIRE:
		eventName = "非法点火"
	case VS_APP_ALARM_TYPE_ILLEGAL_MOVE:
		eventName = "非法位移"
	case VS_APP_ALARM_TYPE_ILLEGAL_OPEN_DOOR:
		eventName = "非法开门"
	case VS_APP_ALARM_TYPE_DANGEROUS:
		eventName = "危险预警"
	case VS_APP_ALARM_TYPE_EMERGENCY:
		eventName = "紧急报警"
	case VS_APP_ALARM_TYPE_VEHICLE_STOLEN:
		eventName = "车辆被盗"
	case VS_APP_ALARM_TYPE_HIGH_TEMP:
		eventName = "高温报警"
	case VS_APP_ALARM_TYPE_NET_MODULE_ERR:
		eventName = "通信模块异常"
	case VS_APP_ALARM_TYPE_VIDEO_LOST:
		eventName = "视频丢失"
	case VS_APP_ALARM_TYPE_STORAGE_FAULT:
		eventName = "存储介质故障"
	case VS_APP_ALARM_TYPE_VIDEO_MASK:
		eventName = "视频遮挡"
	case VS_APP_ALARM_TYPE_OFFLINE_TIMEOUT:
		eventName = "离线超时"
	case VS_APP_ALARM_TYPE_ROAD_ANALYZE:
		eventName = "巡检分析"
	case VS_APP_ALARM_TYPE_ROAD_INFO_COLLECT:
		eventName = "道路信息采集"
	case VS_APP_ALARM_TYPE_ONE_KEY_ALARM:
		eventName = "一键报警"
	}
	return eventName
}

// VsFrontAttachUploadCompleteMsg 附件完成通知
type VsFrontAttachUploadCompleteMsg struct {
	AlarmCode    uint64 `json:"alarmCode"`    //事件唯一ID
	DeviceNo     string `json:"deviceNo"`     //终端号
	FileName     string `json:"fileName"`     //文件名
	FileFullPath string `json:"fileFullPath"` //文件全路径
	FileSize     uint32 `json:"fileSize"`     //文件大小
	//以下3个目前只支持mqtt接入,部标接入时为空
	EventTime uint32 `json:"eventTime"` //事件时间
	EventType int    `json:"eventType"` //事件类型
	FileSN    int    `json:"fileSN"`    //文件序列号
}

// apis 相关定义
type IotPbWebCommonRespHead struct {
	Code    IOT_HTTP_RESP_CODE `json:"code,omitempty"` //应答码
	Message string             `json:"message,omitempty"`
}

type IotHttpRespObject struct {
	ctx      *gin.Context
	Hdr      *IotPbWebCommonRespHead `json:"hdr"`
	Data     interface{}             `json:"data"`
	isRespPb bool                    //是否应答PB,默认json,根据输入的参数来决定
}

// GetPbWebCommonRespHead 初始化内存
func GetPbWebCommonRespHead() *IotPbWebCommonRespHead {
	pbHead := &IotPbWebCommonRespHead{}
	return pbHead
}

func (_self *IotHttpRespObject) RespFromResult(code IOT_HTTP_RESP_CODE, message string) {
	_self.Hdr = GetPbWebCommonRespHead()
	_self.Hdr.Code = code
	if message != "" {
		_self.Hdr.Message = message
	}
	if _self.ctx != nil {
		_self.ctx.JSON(http.StatusOK, _self)
		return
	}
	log.Error("RespFromResult ctx not set", string(debug.Stack()))
}

func (_self *IotHttpRespObject) RespFromResultData(code IOT_HTTP_RESP_CODE, message string, respObject interface{}) {
	_self.Hdr = GetPbWebCommonRespHead()
	_self.Hdr.Code = code
	if message != "" {
		_self.Hdr.Message = message
	}
	_self.Data = respObject
	if _self.ctx != nil {
		_self.ctx.JSON(http.StatusOK, _self)
		return
	}
	log.Error("RespFromResult ctx not set", string(debug.Stack()))
}

func (_self *IotHttpRespObject) RespError(code IOT_HTTP_RESP_CODE, message string) {
	_self.Hdr = GetPbWebCommonRespHead()
	_self.Hdr.Code = code
	_self.Hdr.Message = message
	if _self.ctx != nil {
		_self.ctx.JSON(http.StatusOK, _self)
		return
	}
	log.Error("respNormal ctx not set", string(debug.Stack()))
}

func (_self *IotHttpRespObject) SetCtx(ctx *gin.Context) {
	_self.ctx = ctx
}

func (_self *IotHttpRespObject) RespNormal(respObject interface{}) {
	if _self.ctx == nil {
		log.Error("respNormal ctx not set", string(debug.Stack()))
		return
	}
	_self.Hdr = GetPbWebCommonRespHead()
	_self.Hdr.Message = "success"
	_self.Hdr.Code = IOT_HTTP_RESP_CODE_OK
	if respObject == nil {
		if !_self.isRespPb {
			_self.ctx.JSON(http.StatusOK, _self)
		} else {
			_self.ctx.ProtoBuf(http.StatusOK, _self)
		}
		return
	}
	rRespObject := reflect.ValueOf(respObject)
	kind := rRespObject.Kind()
	if reflect.Ptr == kind {
		pbObject := rRespObject.Elem()
		if !pbObject.IsValid() {
			_self.Hdr.Code = IOT_HTTP_RESP_CODE_SERVER_INNER_ERR
			_self.Hdr.Message = "server inner err,pb type convert fail,type err"
			_self.ctx.JSON(http.StatusOK, _self)
			return
		}
		rHdr := pbObject.FieldByName("Hdr")
		if !rHdr.IsValid() {
			//不存在头,则默认返回json
			_self.Hdr.Code = IOT_HTTP_RESP_CODE_OK
			_self.Hdr.Message = "success"
			_self.Data = respObject //传进来的是一个json实体数据
			_self.ctx.JSON(http.StatusOK, _self)
			return
		}
		rHdrPtr := rHdr.Elem()
		if !rHdrPtr.IsValid() {
			_self.Hdr.Code = IOT_HTTP_RESP_CODE_SERVER_INNER_ERR
			_self.Hdr.Message = "hdr no malloc"
			log.Error("hdr no malloc", debug.Stack())
			_self.ctx.JSON(http.StatusOK, _self)
			return
		}
		rCode := rHdrPtr.FieldByName("Code")
		rMessage := rHdrPtr.FieldByName("Message")
		if rCode.IsValid() && rMessage.IsValid() {
			rCode.SetInt(int64(IOT_HTTP_RESP_CODE_OK))
			rMessage.SetString("success")
		}
		if !_self.isRespPb {
			_self.ctx.JSON(http.StatusOK, respObject)
		} else {
			_self.ctx.ProtoBuf(http.StatusOK, respObject)
		}
	} else {
		_self.Hdr.Code = IOT_HTTP_RESP_CODE_OK
		_self.Hdr.Message = "success"
		_self.Data = respObject //传进来的是一个json实体数据
		_self.ctx.JSON(http.StatusOK, _self)
	}
}

// CheckParseFormAndGetRespObject 检查表单 `form:"tagName" binding:"required"`
func CheckParseFormAndGetRespObject(ctx *gin.Context, targetObject interface{}) (bool, *IotHttpRespObject) {
	respObject := IotHttpRespObject{}
	respObject.Hdr = GetPbWebCommonRespHead()
	err := ctx.ShouldBindQuery(targetObject)
	if err != nil {
		respObject.Hdr.Code = IOT_HTTP_RESP_CODE_INPUT_PARAM
		respObject.Hdr.Message = "input param err:" + err.Error()
		ctx.JSON(http.StatusOK, respObject)
		return false, nil
	}
	accept := ctx.GetHeader("Accept")
	if strings.Contains(accept, "protobuf") {
		respObject.isRespPb = true
	}
	respObject.ctx = ctx
	return true, &respObject
}

// CheckParseJsonForFront 检查json 参数 `json:"tagName" validate:"required"`
func CheckParseJsonForFront(ctx *gin.Context, targetObject interface{}) (IOT_HTTP_RESP_CODE, string) {
	err := ctx.Bind(targetObject)
	if err != nil {
		log.Error("CheckParseJsonForSimple bind fail", err.Error())
		return IOT_HTTP_RESP_CODE_INPUT_PARAM, " bind fail" + err.Error()
	}
	validate := validator.New()
	err = validate.Struct(targetObject)
	if err != nil {
		log.Error("CheckParseJsonForSimple validate fail", err.Error())
		return IOT_HTTP_RESP_CODE_INPUT_PARAM, " validate fail" + err.Error()
	}
	return IOT_HTTP_RESP_CODE_OK, ""
}

// GetRespObject 不需要较验参数
func GetRespObject(ctx *gin.Context) *IotHttpRespObject {
	respObject := IotHttpRespObject{}
	respObject.Hdr = GetPbWebCommonRespHead()
	accept := ctx.GetHeader("Accept")
	if strings.Contains(accept, "protobuf") {
		respObject.isRespPb = true
	}
	respObject.ctx = ctx
	return &respObject
}

// CheckParseJsonOrFormDataAndGetRespObject 检查json 参数 `json:"tagName" validate:"required"`
func CheckParseJsonOrFormDataAndGetRespObject(ctx *gin.Context, targetObject interface{}) (bool, *IotHttpRespObject) {
	respObject := IotHttpRespObject{}
	respObject.Hdr = GetPbWebCommonRespHead()
	err := ctx.Bind(targetObject)
	if err != nil {
		respObject.Hdr.Code = IOT_HTTP_RESP_CODE_INPUT_PARAM
		respObject.Hdr.Message = "input param err:" + err.Error()
		ctx.JSON(http.StatusOK, respObject)
		return false, nil
	}
	validate := validator.New()
	err = validate.Struct(targetObject)
	if err != nil {
		respObject.Hdr.Code = IOT_HTTP_RESP_CODE_INPUT_PARAM
		respObject.Hdr.Message = "input param err:" + err.Error()
		ctx.JSON(http.StatusOK, respObject)
		return false, nil
	}
	accept := ctx.GetHeader("Accept")
	if strings.Contains(accept, "protobuf") {
		respObject.isRespPb = true
	}
	respObject.ctx = ctx
	return true, &respObject
}

// CheckParseXml 解析body 的xml
func CheckParseXml(ctx *gin.Context, targetObject interface{}) bool {
	err := ctx.Bind(targetObject)
	if err != nil {
		log.Error("CheckParseXml fail,", err.Error())
		return false
	}
	return true
}

//web 相关的定义

// IotWebMsgWebSocketHdr 推送公共头
type IotWebMsgWebSocketHdr struct {
	Code    IOT_HTTP_RESP_CODE `json:"code"`
	Message string             `json:"message"`
	MsgType IOT_WEB_MSG_TYPE   `json:"msgType"`
	ReqSN   uint32             `json:"reqSN"`  //请求SN
	RespSN  uint32             `json:"respSN"` //应答SN 为0时为推送消息
}

// IotWebMsgWebSocketPlatformPostQueryPublish 平台查岗推送
type IotWebMsgWebSocketPlatformPostQueryPublish struct {
	Hdr  IotWebMsgWebSocketHdr         `json:"hdr"`
	Data *VsAppMsgPlatformPostQueryReq `json:"data"` //查岗信息
}

// IotWebMsgWebSocketPlatformTextPublish 平台报文推送
type IotWebMsgWebSocketPlatformTextPublish struct {
	Hdr  IotWebMsgWebSocketHdr    `json:"hdr"`
	Data *VsAppMsgPlatformTextReq `json:"data"` //报文信息
}

// IotWebMsgWebSocketWarnUrgeTodoPublish 报警督办推送
type IotWebMsgWebSocketWarnUrgeTodoPublish struct {
	Hdr  IotWebMsgWebSocketHdr    `json:"hdr"`
	Data *VsAppMsgWarnUrgeTodoReq `json:"data"` //督办信息
}
