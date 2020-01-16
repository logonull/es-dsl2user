package main

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/bitly/go-simplejson"
	"github.com/olivere/elastic"
	"os"
	"fmt"
)


func main() {

	input := `{
    "params": {
        "multi_params": {
            "search_type": "multi",
 			"multi_op": "or",
            "value": [
                {
                    "search_type": "multi",
					"multi_op": "and",
                    "value": {
                        "lds_creator": {
                            "search_type": "eq",
                            "value": "443"
                        },
                        "lds_assignment_uid": {
                            "search_type": "eq",
                            "value": 0
                        },
                        "lds_status": {
                            "search_type": "eq",
                            "value": 1
                        }
                    }
                },
                {
                    "search_type": "multi",
					"multi_op": "and",
                    "value": {
                        "lds_assignment_uid": {
                            "search_type": "eq",
                            "value": "443"
                        },
                        "lds_status": {
                            "search_type": "neq",
                            "value": 5
                        }
                    }
                }
            ]
        },
        "lds_sys_id": {
            "search_type": "gt",
            "value": 0
        },
        "lds_status": {
            "search_type": "gte",
            "value": 0
        }
    }
}`

	tmpJson, err := simplejson.NewJson([]byte(input))
	if err!=nil{
		beego.Debug("不执行了?")
		beego.Debug(err)
		os.Exit(0)
	}

	tmpReq := tmpJson.MustMap()

	searchParams := make(map[string]map[string]interface{})

	for k, v := range tmpReq {
		beego.Debug(k, v)

		switch k {

		case "params":
			params := v.(map[string]interface{})
			for kk, vv := range params {
				beego.Debug(kk, vv)
				searchParams[kk] = vv.(map[string]interface{})
			}

		default:
			beego.Debug(k, v)
		}
	}

	q :=elastic.NewBoolQuery()

	ParseSearchInfo(searchParams, q)

	tmp,err:=q.Source()
	tmpS, err:=json.Marshal(tmp)
	beego.Debug(string(tmpS))


}

func Parse() {
	//tmpValue:=item["value"].(map[string]interface{})
}


func ParseSearchInfo(searchParams map[string]map[string]interface{}, q *elastic.BoolQuery) (*elastic.BoolQuery) {

	//根据参数 进行修改 如果在给定的列表中
	for k, item := range searchParams {

		if _, ok := item["search_type"]; ok {

			//beego.Debug(k, item)

			switch item["search_type"].(string) {


			case "eq":
					q = q.Must(elastic.NewMatchPhraseQuery(k, item["value"]).Slop(0))

			case "in":
				//$searchdata['filter']['terms']['tagname'] = $data['tags'];


					itemList := item["value"].([]interface{})
					beego.Debug(itemList)
					q = q.Filter(elastic.NewTermsQuery(k, itemList...))
					//q = q.Filter(elastic.NewTermsQuery(k, item["value"].([]interface{})))


			case "all_in":
				itemList := item["value"].([]interface{})
				for _, vv := range itemList {
					q = q.Must(elastic.NewTermQuery(k, vv))
				}

			case "not_in":

				itemList := item["value"].([]interface{})
				for _, vv := range itemList {
					q = q.MustNot(elastic.NewTermQuery(k, vv))
				}

			case "not_all_in":

				q = q.MustNot(elastic.NewTermsQuery(k, item["value"]))

			case "neq":

				q = q.MustNot(elastic.NewMatchPhraseQuery(k, item["value"]).Slop(0))

			case "gte":
				q = q.Filter(elastic.NewRangeQuery(k).Gte(item["value"]))

			case "lte":
				q = q.Filter(elastic.NewRangeQuery(k).Lte(item["value"]))

			case "gt":
				q = q.Filter(elastic.NewRangeQuery(k).Gte(item["value"]))

			case "lt":
				q = q.Filter(elastic.NewRangeQuery(k).Lt(item["value"]))

			case "between_and":

				//@todo 格式验证
				start := item["value"].(map[string]interface{})["start"]
				startStr := fmt.Sprintf("%v", start)

				end := item["value"].(map[string]interface{})["end"]
				endStr := fmt.Sprintf("%v", end)

				if start != nil && len(startStr) > 0 {
					q = q.Filter(elastic.NewRangeQuery(k).Gte(start))
				}

				if end != nil && len(endStr) > 0 {
					q = q.Filter(elastic.NewRangeQuery(k).Lte(end))
				}



			case "or":

				orTmpQuery := elastic.NewBoolQuery()

				for kk, vv := range item["value"].(map[string]interface{}) {
					tmpVV := vv.(map[string]interface{})
					beego.Debug(kk, vv)
					orTmpQuery.Should(elastic.NewTermsQuery(kk, tmpVV["value"].([]interface{})...))
				}

				q = q.Should(orTmpQuery)

			case "and":

				orTmpQuery := elastic.NewBoolQuery()

				for kk, vv := range item["value"].(map[string]interface{}) {
					tmpVV := vv.(map[string]interface{})
					beego.Debug(kk, vv)
					orTmpQuery.Must(elastic.NewTermsQuery(kk, tmpVV["value"].([]interface{})...))
				}

				q = q.Must(orTmpQuery)

			case "tag_in":

				//tag_ids{
				//
				// search_type:"tag_in"
				//	value:{
				// 		"start" : "156xxxx",
				//		"end" :""
				// }
				//  tag_time
				//
				//
				// }

				tagSecondQuery := elastic.NewBoolQuery()
				for _, vv := range item["value"].([]interface{}) {
					tagQuery := elastic.NewBoolQuery()
					orTmpQuery := elastic.NewHasChildQuery("member_tag", tagQuery)
					orTmpQuery.InnerHit(elastic.NewInnerHit())
					//orTmpQuery.Should(elastic.NewTermsQuery(kk, tmpVV["value"].([]interface{})...))
					tagQuery.Must(elastic.NewTermQuery("id_string", vv))

					//标签时间搜索 @todo 需要优化

					if item["tag_time"] != nil {
						timeItemMap := item["tag_time"].(map[string]interface{})
						timeItem := timeItemMap["value"].(map[string]interface{})

						start := timeItem["start"]

						startStr := fmt.Sprintf("%v", start)

						end := timeItem["end"]

						endStr := fmt.Sprintf("%v", end)

						if start != nil && len(startStr) > 0 {
							tagQuery = tagQuery.Filter(elastic.NewRangeQuery("create_time").Gte(start))
						}

						if end != nil && len(endStr) > 0 {
							tagQuery = tagQuery.Filter(elastic.NewRangeQuery("create_time").Lte(end))
						}
					}

					//标签时间搜索

					tagSecondQuery.Should(orTmpQuery)

				}
				q = q.Filter(tagSecondQuery)

			case "tag_all_in":

				tagSecondQuery := elastic.NewBoolQuery()
				for _, vv := range item["value"].([]interface{}) {
					tagQuery := elastic.NewBoolQuery()
					orTmpQuery := elastic.NewHasChildQuery("member_tag", tagQuery)
					orTmpQuery.InnerHit(elastic.NewInnerHit())
					//orTmpQuery.Should(elastic.NewTermsQuery(kk, tmpVV["value"].([]interface{})...))
					tagQuery.Must(elastic.NewTermQuery("id_string", vv))

					//标签时间搜索 @todo 需要优化

					if item["tag_time"] != nil {
						timeItemMap := item["tag_time"].(map[string]interface{})
						timeItem := timeItemMap["value"].(map[string]interface{})

						start := timeItem["start"]

						startStr := fmt.Sprintf("%v", start)

						end := timeItem["end"]

						endStr := fmt.Sprintf("%v", end)

						if start != nil && len(startStr) > 0 {
							tagQuery = tagQuery.Filter(elastic.NewRangeQuery("create_time").Gte(start))
						}

						if end != nil && len(endStr) > 0 {
							tagQuery = tagQuery.Filter(elastic.NewRangeQuery("create_time").Lte(end))
						}
					}

					//标签时间搜索

					tagSecondQuery.Must(orTmpQuery)
				}
				q = q.Must(tagSecondQuery)

			case "tag_not_in":

				tagSecondQuery := elastic.NewBoolQuery()
				for _, vv := range item["value"].([]interface{}) {
					tagQuery := elastic.NewBoolQuery()
					orTmpQuery := elastic.NewHasChildQuery("member_tag", tagQuery)
					orTmpQuery.InnerHit(elastic.NewInnerHit())
					//orTmpQuery.Should(elastic.NewTermsQuery(kk, tmpVV["value"].([]interface{})...))
					tagQuery.Must(elastic.NewTermQuery("id_string", vv))

					//标签时间搜索 @todo 需要优化

					if item["tag_time"] != nil {
						timeItemMap := item["tag_time"].(map[string]interface{})
						timeItem := timeItemMap["value"].(map[string]interface{})

						start := timeItem["start"]

						startStr := fmt.Sprintf("%v", start)

						end := timeItem["end"]

						endStr := fmt.Sprintf("%v", end)

						if start != nil && len(startStr) > 0 {
							tagQuery = tagQuery.Filter(elastic.NewRangeQuery("create_time").Gte(start))
						}

						if end != nil && len(endStr) > 0 {
							tagQuery = tagQuery.Filter(elastic.NewRangeQuery("create_time").Lte(end))
						}
					}

					//标签时间搜索

					tagSecondQuery.MustNot(orTmpQuery)
				}
				q = q.Must(tagSecondQuery)

			case "tag_not_all_in":

				tagSecondQuery := elastic.NewBoolQuery()
				for _, vv := range item["value"].([]interface{}) {
					tagQuery := elastic.NewBoolQuery()
					orTmpQuery := elastic.NewHasChildQuery("member_tag", tagQuery)
					orTmpQuery.InnerHit(elastic.NewInnerHit())
					//orTmpQuery.Should(elastic.NewTermsQuery(kk, tmpVV["value"].([]interface{})...))
					tagQuery.Must(elastic.NewTermQuery("id_string", vv))

					//标签时间搜索 @todo 需要优化

					if item["tag_time"] != nil {
						timeItemMap := item["tag_time"].(map[string]interface{})
						timeItem := timeItemMap["value"].(map[string]interface{})

						start := timeItem["start"]

						startStr := fmt.Sprintf("%v", start)

						end := timeItem["end"]

						endStr := fmt.Sprintf("%v", end)

						if start != nil && len(startStr) > 0 {
							tagQuery = tagQuery.Filter(elastic.NewRangeQuery("create_time").Gte(start))
						}

						if end != nil && len(endStr) > 0 {
							tagQuery = tagQuery.Filter(elastic.NewRangeQuery("create_time").Lte(end))
						}
					}
					//标签时间搜索

					tagSecondQuery.Must(orTmpQuery)
				}
				q = q.MustNot(tagSecondQuery)
			case "exist":
				//tmpValue := "*" + item["value"].(string) + "*"
				//q = q.Must(elastic.NewMatchPhraseQuery(k, tmpValue).Slop(0))

				q = q.Must(elastic.NewExistsQuery(k))

			case "not_exist":
				q = q.MustNot(elastic.NewExistsQuery(k))

			case "multi":

				//tq:=elastic.NewBoolQuery()
				tq1:= make([]elastic.Query, 0)
				//拆分multi_params 为对象列表
				beego.Debug(item)

				itemList	:= item["value"].([]interface{})
				multiOpt 	:= item["multi_op"].(string)

				for _, mvv:=range itemList{

					params := mvv.(map[string]interface{})

					beego.Debug(params)

					var sType, multiOp  string
					var stp = make(map[string]interface{})
					var tp  = make(map[string]map[string]interface{})
					ttq:=elastic.NewBoolQuery()

					sType = params["search_type"].(string)
					multiOp = params["multi_op"].(string)

					stp = params["value"].(map[string]interface{})
					for st, sv:=range stp{
						tp[st] = sv.(map[string]interface{})
					}

					beego.Debug(sType, multiOp, sType)

					if len(sType)> 0 && len(multiOp) > 0 && sType=="multi" {

						beego.Debug(tp)

						tmpQ := ParseSearchInfo(tp, ttq)

						ts, err:=tmpQ.Source()
						if err!=nil {
							beego.Debug(err)
						}
						tss,err:=json.Marshal(ts)
						beego.Debug(string(tss))

						tq1=append(tq1, tmpQ)
					}
				}

				switch multiOpt {

				case "and":
					//q.Must(tq)
					q.Must(tq1...)
				case "or":
					//q.Should(tq)
					q.Should(tq1...)

				case "not":
					//q.MustNot(tq)
					q.MustNot(tq1...)

				}

				//拆分对应
				//
			default:
				beego.Debug("未预料到的搜索选项:", k, item["search_type"], item["value"])
			}
		}

	}

	return q
}
