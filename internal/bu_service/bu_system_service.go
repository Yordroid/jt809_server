package bu_service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"jt809_server/config"
	"jt809_server/internal/data_manage_service"
	"jt809_server/models"
	"jt809_server/util"
	"time"
)

type buUserInfo struct {
	UserName             string
	UserPwdMd5           string
	Token                string
	lastRefreshTokenTime int64 //最后刷新Token时间
	userKey              int64
	isLoadBaseFinish     bool
}

type buSystemServiceContext struct {
	util.GoService
	mqSubscribe *util.VsRabbitMQClient
	mapUserInfo map[int64]*buUserInfo
}

func checkIsUserVehicle(vehicleGuid uint32) bool {
	vehicleInfo := data_manage_service.DataManageServiceIns().GetVehicleInfoByID(vehicleGuid)
	if vehicleInfo == nil {
		return false
	}
	return true
}

func (_self *buSystemServiceContext) initService() {
	_self.mapUserInfo = make(map[int64]*buUserInfo)
	broken := config.GetMQBroken()
	topicList := []string{models.IOT_TOPIC_IOT_APP_DATA_RT_GPS, models.IOT_TOPIC_IOT_APP_DATA_RT_ONOFFLINE,
		models.IOT_TOPIC_IOT_APP_DATA_RT_ALARM, models.IOT_TOPIC_IOT_APP_DATA_RT_SWIPE_CARD,
		models.IOT_TOPIC_IOT_APP_DATA_RT_MULTI_MEDIA_INFO_UPLOAD, models.IOT_TOPIC_IOT_APP_DATA_LOCAL_STORAGE_ATTACH_COMPLETE}
	log.Info("initService,broken:", broken, "topic:", topicList)

	queueName := "jt809_server_" + util.GetOsHostName()
	_self.mqSubscribe = &util.VsRabbitMQClient{}
	_self.mqSubscribe.InitClient(util.WithBrokenInfo(broken),
		util.WithExchangeNames(topicList),
		util.WithQueueName(queueName),
		util.WithIsProducer(false),
		util.WithDurable(config.IsDurable()),
		util.WithOnMessage(func(exchangeName string, msg []byte) {
			if models.IOT_TOPIC_IOT_APP_DATA_RT_GPS == exchangeName {
				_self.onHandleMQMsgGpsInfo(msg)
			} else if models.IOT_TOPIC_IOT_APP_DATA_RT_ONOFFLINE == exchangeName {
				_self.onHandleMQMsgDevOnOfflineInfo(msg)
			} else if models.IOT_TOPIC_IOT_APP_DATA_RT_ALARM == exchangeName {
				_self.onHandleMQMsgAlarmInfo(msg)
			} else if models.IOT_TOPIC_IOT_APP_DATA_RT_SWIPE_CARD == exchangeName {
				_self.onHandleMQMsgSwipeCardInfo(msg)
			} else if models.IOT_TOPIC_IOT_APP_DATA_RT_MULTI_MEDIA_INFO_UPLOAD == exchangeName {
				_self.onHandleMQMsgMultiMediaInfo(msg)
			} else if models.IOT_TOPIC_IOT_APP_DATA_LOCAL_STORAGE_ATTACH_COMPLETE == exchangeName {
				_self.onHandleMQMsgAttachmentComplete(msg)
			} else {
				log.Info("register msg ,but not handle", exchangeName)
			}
		}))
	_self.SetTimer(1, 1000, _self.onSecondTimerTask)
	_self.SetServiceStartedFinish()
}
func (_self *buSystemServiceContext) onSecondTimerTask(uint32) {
	curTime := time.Now().Unix()
	for _, userInfo := range _self.mapUserInfo {
		if curTime < userInfo.lastRefreshTokenTime {
			userInfo.lastRefreshTokenTime = curTime
		}
		if curTime-userInfo.lastRefreshTokenTime > 3600*10 { //超时10小时,重新获取token
			userInfo.lastRefreshTokenTime = curTime
			go func() {
				token := _self.getToken(userInfo.UserName, userInfo.UserPwdMd5)
				if token != "" {
					_self.PostTask(func() {
						userInfo.Token = token
						log.Info("refresh token success:", userInfo.UserName, " token:", token)
					})

				}
			}()
		}
		if !userInfo.isLoadBaseFinish && userInfo.Token != "" {
			userInfo.isLoadBaseFinish = true
			go func() {
				isOK := _self.loadBaseInfo(userInfo.Token, userInfo.UserName, userInfo.userKey)
				if !isOK {
					//失败,需要再次获取
					userInfo.isLoadBaseFinish = false
				}
			}()
		}

	}
}

func (_self *buSystemServiceContext) addBuUserInfo(userName, userPwdMd5 string, userKey int64) {
	_self.PostTask(func() {
		curUserInfo := buUserInfo{}
		curUserInfo.UserName = userName
		curUserInfo.UserPwdMd5 = userPwdMd5
		curUserInfo.userKey = userKey
		curUserInfo.lastRefreshTokenTime = 0
		curUserInfo.isLoadBaseFinish = false
		curUserInfo.Token = ""
		_self.mapUserInfo[userKey] = &curUserInfo
	})
}

func (_self *buSystemServiceContext) deleteBuUserInfo(userKey int64) {
	_self.PostTask(func() {
		delete(_self.mapUserInfo, userKey)
	})
}

func (_self *buSystemServiceContext) onHandleMQMsgDevOnOfflineInfo(msg []byte) {
	curPbObject := models.VsFrontPbMsgOnline{}
	err := proto.Unmarshal(msg, &curPbObject)
	if err != nil {
		log.Error("onHandleMQMsgDevOnOfflineInfo:", err.Error())
		return
	}
	isOK := checkIsUserVehicle(curPbObject.VehicleGuid)
	if !isOK {
		return
	}
	_self.PublishMsg(models.VS_APP_SERVICE_MSG_ID_DEV_ONOFF_LINE, &curPbObject)
}

func (_self *buSystemServiceContext) onHandleMQMsgGpsInfo(msg []byte) {
	pbObject := models.VsPbStorageGpsData{}
	err := proto.Unmarshal(msg, &pbObject)
	if err != nil {
		log.Error("onHandleMQMsgGpsInfo:", err.Error())
		return
	}
	isOK := checkIsUserVehicle(pbObject.VehicleGuid)
	if !isOK {
		return
	}
	_self.PublishMsg(models.VS_APP_SERVICE_MSG_ID_GPS_DATA, &pbObject)

}

// onHandleKafkaMsgAlarmAttach 处理报警信息
func (_self *buSystemServiceContext) onHandleMQMsgAlarmInfo(msg []byte) {
	curObject := models.VsPbStorageEventInfo{}
	err := proto.Unmarshal(msg, &curObject)
	if err != nil {
		log.Error("onHandleMQMsgAlarmInfo:", err.Error())
		return
	}
	isOK := checkIsUserVehicle(curObject.VehicleGuid)
	if !isOK {
		return
	}
	_self.PublishMsg(models.VS_APP_SERVICE_MSG_ID_ALARM_DATA, &curObject)
}

// onHandleMQMsgSwipeCardInfo 处理刷卡信息
func (_self *buSystemServiceContext) onHandleMQMsgSwipeCardInfo(msg []byte) {
	reqObject := models.VsPbStorageSwipeCardInfo{}
	err := proto.Unmarshal(msg, &reqObject)
	if err != nil {
		log.Error("onHandleMQMsgSwipeCardInfo pb convert fail")
		return
	}
	if reqObject.GpsData == nil {
		log.Error("onHandleMQMsgSwipeCardInfo GpsData data is nil,")
		return
	}
	isOK := checkIsUserVehicle(reqObject.GpsData.VehicleGuid)
	if !isOK {
		return
	}
	_self.PublishMsg(models.VS_APP_SERVICE_MSG_ID_SWIPE_CARD, &reqObject)
}

func (_self *buSystemServiceContext) onHandleMQMsgMultiMediaInfo(msg []byte) {
	curObject := models.VsPbStorageMultiMediaEventInfo{}
	err := proto.Unmarshal(msg, &curObject)
	if err != nil {
		log.Error("onHandleMQMsgMultiMediaInfo:", err.Error())
		return
	}
	isOK := checkIsUserVehicle(curObject.GpsData.VehicleGuid)
	if !isOK {
		return
	}
	_self.PublishMsg(models.VS_APP_SERVICE_MSG_ID_MULTI_MEDIA_INFO, &curObject)
}

func (_self *buSystemServiceContext) onHandleMQMsgAttachmentComplete(msg []byte) {
	curObject := models.VsFrontAttachUploadCompleteMsg{}
	isOK := util.ParseJson(msg, &curObject)
	if !isOK {
		log.Error("onHandleMQMsgAttachmentComplete parse fail")
		return
	}
	_self.PublishMsg(models.VS_APP_SERVICE_MSG_ID_DEV_ATTACH_INFO_COMPLETE, &curObject)
}

//api 数据接入///////////////////////////////////////////////////////////////////////////////

// IotApiMsgUserLoginData 登录应答数据
type IotApiMsgUserLoginData struct {
	Token     string `json:"token"`     //JWT token
	DeadTime  int64  `json:"deadTime"`  //token失效时间
	AdminGuid int    `json:"adminGuid"` //管理员GUID
	UserGuid  int    `json:"userGuid"`  //用户GUID
}

type IotApiMsgUserLoginResp struct {
	Hdr  *models.IotPbWebCommonRespHead `json:"hdr"`
	Data IotApiMsgUserLoginData         `json:"data"`
}

// IotPbWebStaffInfoBase 人员基本信息
type IotPbWebStaffInfoBase struct {
	StaffGuid       int32  `json:"staffGuid,omitempty"`       //人员Guid
	StaffName       string `json:"staffName,omitempty"`       //人员名称
	StaffPhone      string `json:"staffPhone,omitempty"`      //人员联系方式
	IdCard          string `json:"idCard,omitempty"`          //唯一ID,如果是校车方案,则为卡号,一般为身份证
	StaffType       uint32 `json:"staffType,omitempty"`       //人员类型 VsStaffType
	ImageSuffix     string `json:"imageSuffix,omitempty"`     // 人员头像文件后缀
	ImageUpdateTime int64  `json:"imageUpdateTime,omitempty"` //人员头像更新时间
	ClassName       string `json:"className,omitempty"`       //班级名称
	AdminGuid       int32  `json:"adminGuid,omitempty"`       //所在的管理员
}

// 车辆绑定对象的基本信息
type IotPbWebUserVehicleBindBaseInfo struct {
	VehicleGuid      int32                    `json:"vehicleGuid,omitempty"`      //车辆GUID
	VehicleNo        string                   `json:"vehicleNo,omitempty"`        //车牌号
	VehicleNoColor   int32                    `json:"vehicleNoColor,omitempty"`   //车牌号颜色
	DeviceGuid       int32                    `json:"deviceGuid,omitempty"`       //设备Guid
	DeviceNo         string                   `json:"deviceNo,omitempty"`         //设备号
	OrgGuid          int32                    `json:"orgGuid,omitempty"`          //分组ID
	OrgName          string                   `json:"orgName,omitempty"`          //分组名称
	StaffInfos       []*IotPbWebStaffInfoBase `json:"staffInfos,omitempty"`       //人员信息列表
	ChannelNum       uint32                   `json:"channelNum,omitempty"`       //通道数量
	ChannelValidFlag uint32                   `json:"channelValidFlag,omitempty"` //通道有效标志位
	IconID           uint32                   `json:"iconID,omitempty"`           //车图标
	Note             string                   `json:"note,omitempty"`             //备注
	DeadlineTime     uint32                   `json:"deadlineTime,omitempty"`     //过期时间
	VehicleType      uint32                   `json:"vehicleType,omitempty"`      //车辆类型
	SimNo            string                   `json:"simNo,omitempty"`            //设备绑定的SIMNo
	OrgType          uint32                   `json:"orgType,omitempty"`          //分组类型 0:公司,1:车队,2:线路
}

type IotApiMsgUserVehicleBaseResp struct {
	Hdr  *models.IotPbWebCommonRespHead     `json:"hdr"`
	Data []*IotPbWebUserVehicleBindBaseInfo `json:"data"`
}

func (_self *buSystemServiceContext) getTokenByUserKey(userKey int64) string {
	isOK, dataI := _self.SyncTask(func() (bool, interface{}) {
		userInfo, isExist := _self.mapUserInfo[userKey]
		if isExist {
			return true, userInfo.Token
		}
		return false, ""
	}, 2)
	if isOK {
		token := dataI.(string)
		return token
	}
	return ""
}

func (_self *buSystemServiceContext) getToken(userName, userPwdMD5 string) string {

	httpClient := util.ClientInfo{}
	httpClient.InitHTTPClient()
	strUrl := fmt.Sprintf("%s/basic_api/v1/login?userName=%s&&password=%s", config.GetDataUrlBase(), userName, userPwdMD5)
	strResp := httpClient.SendURLRequest(strUrl, "")
	respLoginObject := IotApiMsgUserLoginResp{}
	isOK := util.ParseJson([]byte(strResp), &respLoginObject)
	if !isOK {
		log.Error("loadBaseInfo fail", strUrl)
		return ""
	}
	if respLoginObject.Hdr.Code == models.IOT_HTTP_RESP_CODE_OK {
		return respLoginObject.Data.Token
	}
	log.Error("getToken fail", strResp)
	return ""

}

func (_self *buSystemServiceContext) loadBaseInfo(token, userName string, userKey int64) bool {
	httpClient := util.ClientInfo{}
	httpClient.InitHTTPClient()
	httpClient.AddHeader("Authorization", token)
	strUrl := fmt.Sprintf("%s/basic_api/v1/user/get_user_vehicle_info_base_info", config.GetDataUrlBase())
	strResp := httpClient.SendURLRequest(strUrl, "")
	respVehicleInfoObject := IotApiMsgUserVehicleBaseResp{}
	isOK := util.ParseJson([]byte(strResp), &respVehicleInfoObject)
	if !isOK {
		log.Error("loadBaseInfo fail", strUrl)
		return false
	}
	if respVehicleInfoObject.Hdr.Code == models.IOT_HTTP_RESP_CODE_OK {
		vehicleIDs := make([]uint32, len(respVehicleInfoObject.Data))
		for idx := 0; idx < len(respVehicleInfoObject.Data); idx++ {
			curSrcVehicleInfo := respVehicleInfoObject.Data[idx]
			curDstInfo := models.VsVehicleBaseInfo{}
			curDstInfo.VehicleID = uint32(curSrcVehicleInfo.VehicleGuid)
			curDstInfo.DeviceNo = curSrcVehicleInfo.DeviceNo
			curDstInfo.SimNo = curSrcVehicleInfo.SimNo
			curDstInfo.VehicleNo = curSrcVehicleInfo.VehicleNo
			curDstInfo.VehicleNoColor = uint8(curSrcVehicleInfo.VehicleNoColor)
			curDstInfo.LineGuid = uint32(curSrcVehicleInfo.OrgGuid)
			curDstInfo.DeviceIMEI = ""
			vehicleIDs[idx] = curDstInfo.VehicleID
			data_manage_service.DataManageServiceIns().UpdateVehicleInfo(&curDstInfo)

		}
		data_manage_service.DataManageServiceIns().UpdateUserVehicleMapInfo(userKey, vehicleIDs)

		log.Info("load vehicle num:", len(vehicleIDs), " userName:", userName)
		return true
	}
	log.Info("load vehicle fail num:0", " userName:", userName)
	return false
}

func (_self *buSystemServiceContext) sendDevTakePhotoReq(userKey int64, reqObject *models.VsAppTakePhotoReq) bool {
	token := _self.getTokenByUserKey(userKey)
	httpClient := util.ClientInfo{}
	httpClient.InitHTTPClient()
	httpClient.SetMethod("GET")
	httpClient.AddHeader("Authorization", token)
	httpClient.SetURLParam("vehicleNo", reqObject.VehicleNo)
	httpClient.SetURLParam("channelNo", fmt.Sprintf("%d", reqObject.ChannelNo))
	httpClient.SetURLParam("captureNum", "1")
	httpClient.SetURLParam("captureInterval", "10")
	httpClient.SetURLParam("flag", "0")
	httpClient.SetURLParam("resolution", fmt.Sprintf("%d", reqObject.PicType))
	httpClient.SetURLParam("quality", "5")
	httpClient.SetURLParam("brightness", "127")
	httpClient.SetURLParam("contrast", "127")
	httpClient.SetURLParam("chrominance", "127")
	httpClient.SetURLParam("saturation", "127")

	strUrl := fmt.Sprintf("%s/web_api/v1/media/dev_capture", config.GetWebUrlBase())

	respJsonData := httpClient.SendURLRequest(strUrl, "")
	respObject := models.IotHttpRespObject{}
	isOK := util.ParseJson([]byte(respJsonData), &respObject)
	if !isOK {
		log.Error("onHandleServiceMsgTakePhotoReq request bu json parse fail", respJsonData)
		return false
	}
	if respObject.Hdr != nil && respObject.Hdr.Code == models.IOT_HTTP_RESP_CODE_OK {
		return true
	}
	return false
}

func (_self *buSystemServiceContext) sendPlatformPostQueryReq(userKey int64, reqObject *models.VsAppMsgPlatformPostQueryReq) {
	token := _self.getTokenByUserKey(userKey)
	httpClient := util.ClientInfo{}
	httpClient.InitHTTPClient()
	httpClient.SetMethod("POST")
	httpClient.SetTimeout(10)
	httpClient.AddHeader("Authorization", token)
	url := fmt.Sprintf("%s/web_api/v1/jt809/platform_post_query_req", config.GetWebUrlBase())
	strResp := httpClient.SendURLRequest(url, util.ToJson(reqObject))
	if len(strResp) == 0 {
		log.Error("onHandlePlatformMsgPostQueryReq resp fail:", url)
		return
	}
}

func (_self *buSystemServiceContext) sendPlatformMsgTextInfoReq(userKey int64, reqObject *models.VsAppMsgPlatformTextReq) {
	token := _self.getTokenByUserKey(userKey)
	httpClient := util.ClientInfo{}
	httpClient.InitHTTPClient()
	httpClient.SetMethod("POST")
	httpClient.SetTimeout(10)
	httpClient.AddHeader("Authorization", token)
	url := fmt.Sprintf("%s/web_api/v1/jt809/platform_msg_send_text", config.GetWebUrlBase())
	strResp := httpClient.SendURLRequest(url, util.ToJson(reqObject))
	if len(strResp) == 0 {
		log.Error("sendPlatformMsgTextInfoReq resp fail:", url)
		return
	}
}

func (_self *buSystemServiceContext) sendPlatformWarnUrgeTodoReq(userKey int64, reqObject *models.VsAppMsgWarnUrgeTodoReq) {
	token := _self.getTokenByUserKey(userKey)
	httpClient := util.ClientInfo{}
	httpClient.InitHTTPClient()
	httpClient.SetMethod("POST")
	httpClient.SetTimeout(10)
	httpClient.AddHeader("Authorization", token)
	url := fmt.Sprintf("%s/web_api/v1/jt809/warn_urge_todo_req", config.GetWebUrlBase())
	strResp := httpClient.SendURLRequest(url, util.ToJson(reqObject))
	if len(strResp) == 0 {
		log.Error("sendPlatformMsgTextInfoReq resp fail:", url)
		return
	}
}
