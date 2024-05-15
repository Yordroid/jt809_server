package bu_service

import "jt809_server/models"

//业务服务对接接口

type buSystemBase interface {
	addBuUserInfo(userName, userPwdMd5 string, userKey int64)                               //添加业务系统的登录信息
	deleteBuUserInfo(userKey int64)                                                         //删除用户信息
	sendDevTakePhotoReq(userKey int64, reqObject *models.VsAppTakePhotoReq) bool            //设备抓拍
	sendPlatformPostQueryReq(userKey int64, reqObject *models.VsAppMsgPlatformPostQueryReq) //发送平台查岗
	sendPlatformMsgTextInfoReq(userKey int64, reqObject *models.VsAppMsgPlatformTextReq)    //报文下发
	sendPlatformWarnUrgeTodoReq(userKey int64, reqObject *models.VsAppMsgWarnUrgeTodoReq)   //报警督办
}
