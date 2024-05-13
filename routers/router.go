package routers

import (
	"github.com/gin-gonic/gin"
	"jt809_server/apis"
	"net/http"
)

// CrosHandler 跨域访问：cross  origin resource share
func CrosHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		method := ctx.Request.Method
		origin := ctx.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			// 接收客户端发送的origin
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			// 服务器支持的所有跨域请求的方法
			ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			// 允许跨域设置可以返回其他子段，可以自定义字段
			ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, Content-Type, X-CSRF-Token, Token, session")
			// 允许浏览器（客户端）可以解析的头部
			ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			// 设置缓存时间
			ctx.Header("Access-Control-Max-Age", "172800")
			// 允许客户端传递校验信息比如 cookie
			ctx.Header("Access-Control-Allow-Credentials", "true")
		}
		// 允许类型校验，不经过网关，放行所有 OPTIONS 请求
		if method == "OPTIONS" {
			ctx.JSON(http.StatusOK, "")
			//ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func InitRouter() *gin.Engine {
	router := gin.New()
	router.Use(CrosHandler(), gin.Logger(), gin.Recovery())
	//lyztest TODO 开放时, 要去掉
	webApi := router.Group("/jt809_api/v1/")
	//	webApi.POST("/update_user_basic_info", apis.OnHandleUpdateUserBasicInfo)
	webApi.POST("/platform_post_query_resp", apis.OnHandleWebRequestPlatformPostQueryResp)
	webApi.POST("/warn_urge_todo_resp", apis.OnHandleWebRequestWarnUrgeTodoResp)
	return router
}
