package models

import (
	jt_session "jt809_server/internal/jt_service/session_mananage"
)

const VS_JT_VEHCLE_NO_LEN = 21 //车牌号长度

type VsJTDataType uint16

const (
	//主链路-------------------------------------------------------------------

	//链路管理-----------------------------------------------------------------
	VS_JT_TYPE_UP_LOGIN_REQ            VsJTDataType = 0x1001 //主链路登录请求消息 主链路发送
	VS_JT_TYPE_UP_LOGIN_RESP                        = 0x1002 //主链路登录应答消息 主链路发送
	VS_JT_TYPE_UP_LOGOUT_REQ                        = 0x1003 //主链路注销请求消息 主链路发送
	VS_JT_TYPE_UP_LOGOUT_RESP                       = 0x1004 //主链路注销应答消息 主链路发送
	VS_JT_TYPE_UP_LINK_KEEP_ALIVE_REQ               = 0x1005 //主链路连接保持请求消息 主链路发送
	VS_JT_TYPE_UP_LINK_KEEP_ALIVE_RESP              = 0x1006 //主链路连接保持应答消息 主链路发送
	VS_JT_TYPE_UP_DISCONNECT_NOTIFY                 = 0x1007 //主链路断开通知消息 从链路发送
	VS_JT_TYPE_UP_CLOSE_LINK_NOTIFY                 = 0x1008 //下级平台主动关闭链路通知消息 从链路发送

	//主链路动态信息交换消息
	VS_JT_TYPE_UP_EXG_MSG                             = 0x1200
	VS_JT_TYPE_UP_EXG_MSG_REGISTER                    = 0x1201 //上传车辆注册信息
	VS_JT_TYPE_UP_EXG_MSG_REALTIME                    = 0x1202 //实时上传车辆定位信息
	VS_JT_TYPE_UP_EXG_MSG_UP_EXG_MSG_HISTORY_LOCATION = 0x1203 //车辆定位信息自动补报
	VS_JT_TYPE_UP_EXG_MSG_REPORT_DRIVER_INFO_ACK      = 0x120A //上报车辆驾驶员身份识别信息应答
	VS_JT_TYPE_UP_EXG_MSG_TAKE_EWAYBILL_ACK           = 0x120B //上报车辆电子运单应答
	VS_JT_TYPE_UP_EXG_MSG_REPORT_DRIVER_INFO          = 0x120C //主动上报驾驶员身份
	VS_JT_TYPE_UP_EXG_MSG_REPORT_EWAYBILL_INFO        = 0x120D //主动上报车辆电子运单
	//主链路平台间信息交互消息
	VS_JT_TYPE_UP_PLAFORM_MSG                = 0x1300
	VS_JT_TYPE_UP_PLAFORM_MSG_POST_QUERY_ACK = 0x1301 //平台查岗应答
	VS_JT_TYPE_UP_PLAFORM_MSG_INFO_ACK       = 0x1302 //下发平台间报文应答

	//车辆报警信息交互类
	VS_JT_TYPE_UP_WARN_MSG               = 0x1400
	VS_JT_TYPE_UP_WARN_MSG_URGE_TODO_ACK = 0x1401 //报警督办应答
	VS_JT_TYPE_UP_WARN_MSG_ADPT_INFO     = 0x1402 //上报报警信息

	//主链路车辆监管消息
	VS_JT_TYPE_UP_CTRL_MSG                     = 0x1500
	VS_JT_TYPE_UP_CTRL_MSG_MONITOR_VEHICLE_ACK = 0x1501 //车辆单向监听应答
	VS_JT_TYPE_UP_CTRL_MSG_TAKE_PHOTO_ACK      = 0x1502 //车辆拍照应答

	//主链路静态信息交换消息
	VS_JT_TYPE_UP_BASE_MSG                   = 0x1600
	VS_JT_TYPE_UP_BASE_MSG_VEHICLE_ADDED_ACK = 0x1601 //补报车辆静态信息应答

	//从链路-------------------------------------------------------------------
	//链路管理
	VS_JT_TYPE_DOWN_LOGIN_REQ            = 0x9001 //从链路登录请求消息 从链路发送
	VS_JT_TYPE_DOWN_LOGIN_RESP           = 0x9002 //从链路登录应答消息 从链路发送
	VS_JT_TYPE_DOWN_LOGOUT_REQ           = 0x9003 //从链路注销请求消息 从链路发送
	VS_JT_TYPE_DOWN_LOGOUT_RESP          = 0x9004 //从链路注销应答消息 从链路发送
	VS_JT_TYPE_DOWN_LINK_KEEP_ALIVE_REQ  = 0x9005 //从链路连接保持请求消息 从链路发送
	VS_JT_TYPE_DOWN_LINK_KEEP_ALIVE_RESP = 0x9006 //从链路连接保持应答消息 从链路发送
	VS_JT_TYPE_DOWN_DISCONNECT_NOTIFY    = 0x9007 //从链路断开通知消息 从链路发送
	VS_JT_TYPE_DOWN_CLOSE_LINK_NOTIFY    = 0x9008 //上级平台主动关闭链路通知消息 主链路发送

	//信息统计类
	VS_JT_TYPE_DOWN_DOWN_TOTAL_MSG_NUM = 0x9101 //接收定位信息数量通知消息 从链路发送
	//从链路动态信息交换消息
	VS_JT_TYPE_DOWN_EXG_MSG                    = 0x9200
	VS_JT_TYPE_DOWN_EXG_MSG_REPORT_DRIVER_INFO = 0x920A //上报车辆驾驶员身份识别信息请求 从链路发送
	VS_JT_TYPE_DOWN_EXG_MSG_TAKE_EWAYBILL_REQ  = 0x920B //上报车辆电子运单请求 从链路发送
	//从链路平台间信息交互消息
	VS_JT_TYPE_DOWN_PLAFORM_MSG                 = 0x9300
	VS_JT_TYPE_DOWN_PLATFORM_MSG_POST_QUERY_REQ = 0x9301 //平台查岗请求
	VS_JT_TYPE_DOWN_PLATFORM_MSG_INFO_REQ       = 0x9302 //下发平台间报文请求

	//车辆报警信息交互类
	VS_JT_TYPE_DOWN_WARN_MSG               = 0x9400
	VS_JT_TYPE_DOWN_WARN_MSG_URGE_TODO_REQ = 0x9401 //报警督办请求
	VS_JT_TYPE_DOWN_WARN_MSG_INFORM_TIPS   = 0x9402 //报警预警
	//从链路车辆监管消息
	VS_JT_TYPE_DOWN_CTRL_MSG                     = 0x9500
	VS_JT_TYPE_DOWN_CTRL_MSG_MONITOR_VEHICLE_REQ = 0x9501 //车辆单向监听请求
	VS_JT_TYPE_DOWN_CTRL_MSG_TAKE_PHOTO_REQ      = 0x9502 //车辆拍照请求
	//从链路静态信息交换消息
	VS_JT_TYPE_DOWN_BASE_MSG               = 0x9600
	VS_JT_TYPE_DOWN_BASE_MSG_VEHICLE_ADDED = 0x9601 //补报车辆静态信息请求
)

type VS_JT_ALARM_CODE uint16

const (
	VS_JT_ALARM_CODE_NONE = 0x0 // 无
	//与位置相关
	VS_JT_ALARM_CODE_LOCATION_OVER_SPEED                = 0x0001 //超速报警
	VS_JT_ALARM_CODE_LOCATION_TIRED                     = 0x0002 //疲劳报警
	VS_JT_ALARM_CODE_LOCATION_EMERGENCY                 = 0x0003 //紧急报警
	VS_JT_ALARM_CODE_LOCATION_IN_AREA                   = 0x0004 //进入指定区域报警
	VS_JT_ALARM_CODE_LOCATION_OUT_AREA                  = 0x0005 //离开指定区域报警
	VS_JT_ALARM_CODE_LOCATION_ROAD_SECTIONBUSY_V1       = 0x0006 //路段堵塞报警 JT2011
	VS_JT_ALARM_CODE_LOCATION_DANGEROUS_ROAD_SECTION_V1 = 0x0007 //危险路段报警 JT2011
	//以下只有2019有效
	VS_JT_ALARM_CODE_LOCATION_OVER_AREA         = 0x0008 //越界报警
	VS_JT_ALARM_CODE_LOCATION_THIEF             = 0x0009 //盗警报警
	VS_JT_ALARM_CODE_LOCATION_ROBBIT            = 0x000A //劫警报警
	VS_JT_ALARM_CODE_LOCATION_LINE_OFFSET       = 0x000B //偏离路线报警
	VS_JT_ALARM_CODE_LOCATION_VEHICLE_MOVE      = 0x000C //车辆移动报警
	VS_JT_ALARM_CODE_LOCATION_OVERTIME          = 0x000D //超时驾驶报警
	VS_JT_ALARM_CODE_LOCATION_BREAK_DRIVE       = 0x0010 //违规行驶报警
	VS_JT_ALARM_CODE_LOCATION_FORWARD_COLLISION = 0x0011 //前撞报警
	VS_JT_ALARM_CODE_LOCATION_LANE_DEPARTURE    = 0x0012 //车道偏离报警
	VS_JT_ALARM_CODE_LOCATION_PRESSURE          = 0x0013 //胎压异常报警
	VS_JT_ALARM_CODE_LOCATION_DYNAMIC_EXCEPTION = 0x0014 //动态信息异常报警
	VS_JT_ALARM_CODE_LOCATION_OTHER             = 0x00FF //其他报警
	//非位置相关
	VS_JT_ALARM_CODE_OVER_TIME_STOP           = 0xA001 //超时停车
	VS_JT_ALARM_CODE_LOCATION_UPLOAD_INTERVAL = 0xA002 //车辆定位信息上报时间间隔异常
	VS_JT_ALARM_CODE_LOCATION_UPLOAD_DISTANCE = 0xA003 //车辆定位信息上报距离间隔异常
	VS_JT_ALARM_CODE_PLATFORM_DISCONNECT      = 0xA004 //下级平台异常断线
	VS_JT_ALARM_CODE_PLATFORM_DATA_SEND_ERR   = 0xA005 //下级平台数据传输异常
	VS_JT_ALARM_CODE_ROAD_SECTIONBUSY         = 0xA006 //路段堵塞报警
	VS_JT_ALARM_CODE_DANGEROUS_ROAD_SECTION   = 0xA007 //危险路段报警
	VS_JT_ALARM_CODE_RAINSNOW                 = 0xA008 //雨雪天气报警
	VS_JT_ALARM_CODE_DRIVER_IDENTIFICATION    = 0xA009 //驾驶员身份识别异常
	VS_JT_ALARM_CODE_DEVICE_ERR               = 0xA00A //终端异常(含线路连接异常)
	VS_JT_ALARM_CODE_PLATFORM_ACCESS          = 0xA00B //平台接入异常
	VS_JT_ALARM_CODE_IMPORT_DATA_ERR          = 0xA00C //核心数据异常
	VS_JT_ALARM_CODE_OTHER                    = 0xA0FF //其它报警
)

// ToJT809AlarmCode 应用的报警类型转成JT809的报警类型
func ToJT809AlarmCode(appAlarmType VS_APP_ALARM_TYPE, ver jt_session.JT_VER_TYPE) (VS_JT_ALARM_CODE, string) {
	var jtType VS_JT_ALARM_CODE = VS_JT_ALARM_CODE_NONE
	switch appAlarmType {
	case VS_APP_ALARM_TYPE_DSM_TIRED:
		jtType = VS_JT_ALARM_CODE_LOCATION_TIRED
		break
	case VS_APP_ALARM_TYPE_DSM_DRIVER_EXCEPTION:
		jtType = VS_JT_ALARM_CODE_DRIVER_IDENTIFICATION
		break
	case VS_APP_ALARM_TYPE_ADAS_IMPACT:
		jtType = VS_JT_ALARM_CODE_LOCATION_FORWARD_COLLISION
		break
	case VS_APP_ALARM_TYPE_ADAS_LANE_DEPARTURE:
		jtType = VS_JT_ALARM_CODE_LOCATION_LANE_DEPARTURE
		break
	case VS_APP_ALARM_TYPE_TPMS_ALARM:
		jtType = VS_JT_ALARM_CODE_LOCATION_PRESSURE
		break
	case VS_APP_ALARM_TYPE_OVER_SPEED:
		jtType = VS_JT_ALARM_CODE_LOCATION_OVER_SPEED
		break
	case VS_APP_ALARM_TYPE_TIMEOUT:
		jtType = VS_JT_ALARM_CODE_LOCATION_OVERTIME
		break
	case VS_APP_ALARM_TYPE_GNSS_MODULE_FAULT,
		VS_APP_ALARM_TYPE_ANT_SHORT,
		VS_APP_ALARM_TYPE_ANT_NOT_CONNECT,
		VS_APP_ALARM_TYPE_POWER_DOWN,
		VS_APP_ALARM_TYPE_POWER_LOW,
		VS_APP_ALARM_TYPE_IC_CARD,
		VS_APP_ALARM_TYPE_VSS_ERR,
		VS_APP_ALARM_TYPE_OIL_ERR,
		VS_APP_ALARM_TYPE_TTS:
		jtType = VS_JT_ALARM_CODE_DEVICE_ERR
		break

	case VS_APP_ALARM_TYPE_ILLEGAL_MOVE:
		jtType = VS_JT_ALARM_CODE_LOCATION_VEHICLE_MOVE
		break
	case VS_APP_ALARM_TYPE_DANGEROUS:
		if ver == jt_session.JT_VER_V3 {
			jtType = VS_JT_ALARM_CODE_DANGEROUS_ROAD_SECTION
		} else {
			jtType = VS_JT_ALARM_CODE_LOCATION_DANGEROUS_ROAD_SECTION_V1
		}
		break
	case VS_APP_ALARM_TYPE_EMERGENCY:
		jtType = VS_JT_ALARM_CODE_LOCATION_EMERGENCY
		break
	case VS_APP_ALARM_TYPE_VEHICLE_STOLEN:
		jtType = VS_JT_ALARM_CODE_LOCATION_THIEF
		break
	default:
		break
	}
	alarmString := GetAlarmNameByAlarmType(appAlarmType)
	return jtType, alarmString
}

func ToAppAlarmCode(jtAlarmCode VS_JT_ALARM_CODE, ver jt_session.JT_VER_TYPE) (bool, VS_APP_ALARM_TYPE) {
	var appType VS_APP_ALARM_TYPE
	switch jtAlarmCode {
	case VS_JT_ALARM_CODE_LOCATION_TIRED:
		appType = VS_APP_ALARM_TYPE_DSM_TIRED
		break
	case VS_JT_ALARM_CODE_DRIVER_IDENTIFICATION:
		appType = VS_APP_ALARM_TYPE_DSM_DRIVER_EXCEPTION
		break
	case VS_JT_ALARM_CODE_LOCATION_FORWARD_COLLISION:
		appType = VS_APP_ALARM_TYPE_ADAS_IMPACT
		break
	case VS_JT_ALARM_CODE_LOCATION_LANE_DEPARTURE:
		appType = VS_APP_ALARM_TYPE_ADAS_LANE_DEPARTURE
		break
	case VS_JT_ALARM_CODE_LOCATION_PRESSURE:
		appType = VS_APP_ALARM_TYPE_TPMS_ALARM
		break
	case VS_JT_ALARM_CODE_LOCATION_OVER_SPEED:
		appType = VS_APP_ALARM_TYPE_OVER_SPEED
		break
	case VS_JT_ALARM_CODE_LOCATION_OVERTIME:
		appType = VS_APP_ALARM_TYPE_TIMEOUT
		break
	case VS_JT_ALARM_CODE_DEVICE_ERR:
		appType = VS_APP_ALARM_TYPE_POWER_DOWN
		break

	case VS_JT_ALARM_CODE_LOCATION_VEHICLE_MOVE:
		appType = VS_APP_ALARM_TYPE_ILLEGAL_MOVE
		break
	case VS_JT_ALARM_CODE_LOCATION_DANGEROUS_ROAD_SECTION_V1:
		appType = VS_APP_ALARM_TYPE_DANGEROUS
		break
	case VS_JT_ALARM_CODE_DANGEROUS_ROAD_SECTION:
		appType = VS_APP_ALARM_TYPE_DANGEROUS
		break
	case VS_JT_ALARM_CODE_LOCATION_EMERGENCY:
		appType = VS_APP_ALARM_TYPE_EMERGENCY
		break
	case VS_JT_ALARM_CODE_LOCATION_THIEF:
		appType = VS_APP_ALARM_TYPE_VEHICLE_STOLEN
		break
	default:
		return false, appType
	}
	return true, appType
}

// VsAppMsgLoginInfoReq 链路登录请求
type VsAppMsgLoginInfoReq struct {
	UserID       uint32
	Password     string
	DownLinkIP   string
	DownLinkPort uint16
	AccessID     uint32
}

// VsAppMsgLoginResp 登录应答
type VsAppMsgLoginResp struct {
	Result     uint8  //验证结果 0:成功,1:IP地址不正确,2:接入码不正确,3:用户没有注册,4:密码错误,5:资源紧张,FF:其它
	VerifyCode uint32 //校验码
}

// VsAppMsgCommonSubBuInfo  通用的子业务头
type VsAppMsgCommonSubBuInfo struct {
	VehicleNo      string //车牌号
	VehicleNoColor uint8  //车牌颜色
	SubDataType    uint16 //子业务类型 VsJTDataType
	SubDataLen     uint32 //后续长度
}

// VsAppMsgPlatformTextResp 平台报文应答
type VsAppMsgPlatformTextResp struct {
	SrcMsgSN       uint32 //源消息序列号
	SrcSubDataType uint16 //源子业务类型
}
