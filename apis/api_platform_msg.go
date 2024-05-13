package apis

import (
	"github.com/gin-gonic/gin"
	"jt809_server/internal/jt_service"
	"jt809_server/models"
)

// OnHandleWebRequestPlatformPostQueryResp 平台查岗应答
func OnHandleWebRequestPlatformPostQueryResp(ctx *gin.Context) {
	reqObject := models.VsAppMsgPlatformPostQueryResp{}
	isOK, respObject := models.CheckParseJsonOrFormDataAndGetRespObject(ctx, &reqObject)
	if !isOK {
		return
	}
	jt_service.JTServiceIns().SendPlatformPostQueryResp(&reqObject)
	respObject.RespNormal(nil)
}

// OnHandleWebRequestWarnUrgeTodoResp 督办应答
func OnHandleWebRequestWarnUrgeTodoResp(ctx *gin.Context) {
	reqObject := models.VsAppMsgWarnUrgeTodoResp{}
	isOK, respObject := models.CheckParseJsonOrFormDataAndGetRespObject(ctx, &reqObject)
	if !isOK {
		return
	}
	jt_service.JTServiceIns().SendWarnUrgeTodoResp(&reqObject)
	respObject.RespNormal(nil)
}
