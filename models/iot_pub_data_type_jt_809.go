package models

import "strconv"

const (
	VS_OBJECT_TYPE_SINGLE_PLATFORM   = 0 //下级平台所属单一平台
	VS_OBJECT_TYPE_CUR_CONNECT       = 1 //当前连接的下级平台
	VS_OBJECT_TYPE_SINGLE_OWER       = 2 //下级平台所属单一业户
	VS_OBJECT_TYPE_ALL_OWER          = 3 //下级平台所属所有业户
	VS_OBJECT_TYPE_ALL_PLATFORM      = 4 //下级平台所属所有平台
	VS_OBJECT_TYPE_ALL_OWER_PLATFORM = 5 //下级平台所属所有平台和业户
	VS_OBJECT_TYPE_GOV_MONITOR       = 6 //下级平台所属所有政府监管平台(含监控端)
	VS_OBJECT_TYPE_COMPANY_MONITOR   = 7 //下级平台所属所有企业监控平台
	VS_OBJECT_TYPE_OP_MONITOR        = 8 //下级平台所属所有经营性企业监控平台
	VS_OBJECT_TYPE_NOT_OP_MONITOR    = 9 //下级平台所属所有非经营性企业监控平台
)

// VsAppMsgPlatformPostQueryReq 平台查岗
type VsAppMsgPlatformPostQueryReq struct {
	ObjectType  uint8  `json:"objectType"`                 //查岗对象类型 VS_OBJECT_TYPE_
	ObjectID    string `json:"objectID"`                   //查岗对象ID
	AnswerTime  uint8  `json:"answerTime"`                 //查岗应答时限,以本条消息的TIME开始,秒
	InfoID      uint32 `json:"infoID"`                     //信息ID
	InfoContent string `json:"infoContent"`                //信息内容
	TaskID      string `json:"taskID" validate:"required"` //任务ID
}

// VsAppMsgPlatformPostQueryResp 平台查岗应答
type VsAppMsgPlatformPostQueryResp struct {
	ObjectType uint8  `json:"objectType"`                  //查岗对象类型 VS_OBJECT_TYPE_
	ObjectID   string `json:"objectID"`                    //查岗对象ID
	AckName    string `json:"ackName"`                     //应答人名称
	AckPhone   string `json:"ackPhone"`                    //应答人联系电话
	Content    string `json:"content" validate:"required"` //应答内容
	InfoID     uint32 `json:"infoID"`                      //信息ID 和下发一致
	TaskID     string `json:"taskID" validate:"required"`  //任务ID,填入请求时的任务ID
}

// VsAppMsgPlatformTextReq 平台报文, 直接应答
type VsAppMsgPlatformTextReq struct {
	ObjectType  uint8  `json:"objectType"`  // 对象类型 VS_OBJECT_TYPE_
	ObjectID    string `json:"objectID"`    // 对象ID
	InfoID      uint32 `json:"infoID"`      // 信息ID
	InfoContent string `json:"infoContent"` // 信息内容
}

// VsAppMsgWarnUrgeTodoReq 报警督办请求
type VsAppMsgWarnUrgeTodoReq struct {
	PlatformID         string `json:"platformID"`                 //发起报警平台唯一编码 V3
	AlarmType          uint16 `json:"alarmType"`                  //报警类型 VS_APP_ALARM_TYPE V1,V2
	AlarmSrc           uint8  `json:"alarmSrc"`                   //报警来源 1:车载终端2:企业监控平台3:政府平台 9:其它 V1,V2
	AlarmTime          uint32 `json:"alarmTime"`                  //报警时间
	ReqMsgSN           uint32 `json:"reqMsgSN"`                   //督办请求消息SN V3
	SupervisionEndTime uint32 `json:"supervisionEndTime"`         //督办截止时间
	SuperVisionLevel   uint8  `json:"superVisionLevel"`           //督办级别 00:紧急,1:一般
	SuperVisor         string `json:"superVisor"`                 //督办人
	SuperVisorTel      string `json:"superVisorTel"`              //督办人联系电话
	SuperVisorEmail    string `json:"superVisorEmail"`            //督办人联系电子邮件
	SuperVisionID      uint32 `json:"superVisionID"`              //督办ID
	TaskID             string `json:"taskID" validate:"required"` //任务ID
}

// VsAppMsgWarnUrgeTodoResp 报警督办应答
type VsAppMsgWarnUrgeTodoResp struct {
	ReqMsgSN      uint32 `json:"reqMsgSN"`                   //督办请求消息SN V3
	SuperVisionID uint32 `json:"superVisionID"`              //督办ID V1,V2
	Result        uint8  `json:"result"`                     //结果,0:处理中,1:已处理完毕,2:不作处理,3:将来处理
	TaskID        string `json:"taskID" validate:"required"` //任务ID,填入请求时的任务ID
}

func GetTaskIDToString(taskID int64) string {
	return strconv.FormatInt(taskID, 10)
}
func GetTaskIDFromString(taskID string) int64 {
	nTaskID, _ := strconv.ParseInt(taskID, 10, 64)
	return nTaskID
}
