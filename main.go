package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	_ "member_es_basic_api/routers"
)

func main() {
	beego.SetLogger("file", `{"filename":"`+beego.AppConfig.String("logPath")+beego.AppConfig.String("appname")+`_runtime.log","level":7,"daily":true,"maxdays":7}`)
	if beego.AppConfig.String("runmode") != "dev" {
		beego.BeeLogger.DelLogger("console")
	}
	beego.SetLogFuncCall(true)
	beego.BeeLogger.Async()

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		//允许访问所有源
		AllowAllOrigins: true,
		//可选参数"GET", "POST", "PUT", "DELETE", "OPTIONS" (*为所有)
		//其中Options跨域复杂请求预检
		AllowMethods: []string{"*"},
		//指的是允许的Header的种类
		AllowHeaders: []string{"*"},
		//公开的HTTP标头列表
		ExposeHeaders: []string{"Content-Length"},
		//如果设置，则允许共享身份验证凭据，例如cookie
		AllowCredentials: true,
	}))

	beego.Run()
}