package routers

import (
	"member_es_basic_api/controllers"
	"github.com/astaxie/beego"
	"net/http"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSRouter("/getList", &controllers.SearchController{}, "POST:GetList"), // get member info list
	)
	beego.AddNamespace(ns)

	ns1 := beego.NewNamespace("/v2",
		beego.NSRouter("/getList", &controllers.SearchController{}, "POST:GetListV2"), // get member info list
	)

	beego.AddNamespace(ns1)

	ns2 := beego.NewNamespace("/v3",
		beego.NSRouter("/getList", &controllers.SearchController{}, "POST:GetListV3"), // get member info list
	)

	beego.AddNamespace(ns2)

	//ns2:= beego.NewNamespace("/v3",
	//	beego.NSRouter("/getList", &controllers.SearchController{}, "POST:GetListV3"), // get member info list
	//)
	//beego.AddNamespace(ns2)

	beego.ErrorHandler("404", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
}
