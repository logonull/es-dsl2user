package controllers

import (
	. "member_es_basic_api/middleware"
	"member_es_basic_api/process"
	"member_es_basic_api/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/bitly/go-simplejson"
	"strconv"
	"strings"
)

type SearchController struct {
	beego.Controller
}

func (controller *SearchController) GetList() {
	ic := NewContainer()
	ctl := &BeegoController{Controller: controller.Controller}
	if check := ic.CheckInput(ctl, []*InputRule{
		{
			Position:  POS_BODY_JSON,
			Necessary: true,
			Type:      TYPE_JSON_OBJECT,
		},
	}); check != nil {
		ctl.RenderOutput(InputErr(check))
		return
	}

	searchParams := make(map[string]map[string]interface{})

	orderParams := make([]map[string]string, 0)

	limitParams := make(map[string]int)

	orderParamsTmp := make([]interface{}, 0)

	limitParamsTmp := make(map[string]interface{})

	input := controller.Ctx.Input.RequestBody

	tmpJson, err := simplejson.NewJson(input)

	tmpReq := tmpJson.MustMap()

	mainId := ""

	for k, v := range tmpReq {
		beego.Debug(k, v)

		switch k {

		case "order":
			orderParamsTmp = v.([]interface{})
		case "limit":
			limitParamsTmp = v.(map[string]interface{})
		case "params":
			params := v.(map[string]interface{})
			for kk, vv := range params {
				beego.Debug(kk, vv)
				searchParams[kk] = vv.(map[string]interface{})
			}
		case "main_id":

			mainId = utils.GetInterfaceString(v)

		default:
			beego.Debug(k, v)
		}
	}

	if len(strings.Trim(mainId, " ")) == 0 {
		beego.Debug()
		ctl.RenderOutput(NormalErr(200, 0, "main_id不能为空", ""))
	}

	beego.Debug(searchParams)

	beego.Debug(orderParamsTmp)
	for _, v := range orderParamsTmp {

		tmpOrder:=make(map[string]string)

		tmpOrderV := v.(map[string]interface{})

		for kk, vv:=range tmpOrderV  {
			tmpOrder[kk] = vv.(string)
		}

		orderParams = append(orderParams, tmpOrder)
	}

	for k, v := range limitParamsTmp {
		tmpV := int(0)
		switch v.(type) {
		case string:
			tmpV, err = strconv.Atoi(v.(string))
		case json.Number:
			tmpV, err = strconv.Atoi(v.(json.Number).String())
		}
		limitParams[k] = tmpV
		if err != nil {
			limitParams[k] = 0
		}
	}

	reMap := make(map[string]interface{})

	resList, total, err := process.GetList(mainId, searchParams, orderParams, limitParams)

	reMap["list"] = resList
	reMap["total"] = total

	if err != nil {
		ctl.RenderOutput(NormalErr(200, 0, err.Error(), ""))
	}

	ctl.RenderOutput(Success(reMap))
}


func (controller *SearchController) GetListV2() {
	ic := NewContainer()
	ctl := &BeegoController{Controller: controller.Controller}
	if check := ic.CheckInput(ctl, []*InputRule{
		{
			Position:  POS_BODY_JSON,
			Necessary: true,
			Type:      TYPE_JSON_OBJECT,
		},
	}); check != nil {
		ctl.RenderOutput(InputErr(check))
		return
	}

	searchParams := make(map[string]map[string]interface{})

	orderParams := make([]map[string]string, 0)

	limitParams := make(map[string]int)

	isScrollTmp := "false"

	isScroll := false

	scrollId  := ""

	orderParamsTmp := make([]interface{}, 0)

	limitParamsTmp := make(map[string]interface{})

	sourceParams := make([]interface{}, 0)

	input := controller.Ctx.Input.RequestBody

	tmpJson, err := simplejson.NewJson(input)

	tmpReq := tmpJson.MustMap()

	mainId := ""

//	isDeepPage:=false

	for k, v := range tmpReq {
		//beego.Debug(k, v)

		switch k {

		case "order":
			orderParamsTmp = v.([]interface{})
		case "limit":
			limitParamsTmp = v.(map[string]interface{})
		case "params":
			params := v.(map[string]interface{})
			for kk, vv := range params {
				beego.Debug(kk, vv)
				searchParams[kk] = vv.(map[string]interface{})
			}
		case "main_id":

			mainId = utils.GetInterfaceString(v)
		case "source":
			sourceParams = v.([]interface{})
		case "is_scroll":
			isScrollTmp = v.(string)

		case "scroll_id":
			scrollId = v.(string)

		default:
			beego.Debug(k, v)
		}
	}

	if isScrollTmp == "true" {
		isScroll = true
	}

	if len(strings.Trim(mainId, " ")) == 0 {
		beego.Debug()
		ctl.RenderOutput(NormalErr(200, 0, "main_id不能为空", ""))
	}

	beego.Debug(searchParams)

	for _, v := range orderParamsTmp {

		tmpOrder:=make(map[string]string)

		tmpOrderV := v.(map[string]interface{})

		for kk, vv:=range tmpOrderV  {
			tmpOrder[kk] = vv.(string)
		}

		orderParams = append(orderParams, tmpOrder)
	}

	for k, v := range limitParamsTmp {
		tmpV := int(0)
		switch v.(type) {
		case string:
			tmpV, err = strconv.Atoi(v.(string))
		case json.Number:
			tmpV, err = strconv.Atoi(v.(json.Number).String())
		}
		limitParams[k] = tmpV
		if err != nil {
			limitParams[k] = 0
		}
	}

	reMap := make(map[string]interface{})

	//is_scroll

	//scroll_id

	//

	//

	resList, total, reScrollId, err := process.GetListV2(mainId, sourceParams, searchParams, orderParams, isScroll, scrollId,  limitParams)

	reMap["list"] 		= resList
	reMap["total"] 		= total
	reMap["scrollId"] 	= reScrollId

	if err != nil {
		ctl.RenderOutput(NormalErr(200, 0, err.Error(), ""))
	}

	ctl.RenderOutput(Success(reMap))
}


func (controller *SearchController) GetListV3() {
	ic := NewContainer()
	ctl := &BeegoController{Controller: controller.Controller}
	if check := ic.CheckInput(ctl, []*InputRule{
		{
			Position:  POS_BODY_JSON,
			Necessary: true,
			Type:      TYPE_JSON_OBJECT,
		},
	}); check != nil {
		ctl.RenderOutput(InputErr(check))
		return
	}

	reMap  := make(map[string]interface{}, 0)

	reMap["list"] 		= "asdfasd"
	reMap["total"] 		= 20
	reMap["scrollId"] 	= "sdfasdfasdf"


	ctl.RenderOutput(Success(reMap))
}
