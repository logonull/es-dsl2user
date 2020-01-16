package process

import (
	"member_es_basic_api/models"
	"member_es_basic_api/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/olivere/elastic"
	"regexp"
	"strings"
)

const StrRegex = `[A-Za-z0-9]+$`

var (
	NeedMatchPhraseKey = map[string]bool{
		"name":    true,
		"open_id": true,
		"mobile":  true,
	}
)

var (
	DefaultSearchParams = map[string]map[string]interface{}{
		"identity": map[string]interface{}{
			"search_type": "eq",
			"value":       1,
		},
		"is_del": map[string]interface{}{
			"search_type": "str_eq",
			"value":       "N",
		},
	}

	ConDefaultOrder = map[string]string{
			"lds_status": "lds_status_order",
	}

	DefualutArrSearch = map[string]bool{
		"sex":                true,
		"level_id":           true,
		"from_source":        true,
		"province":           true,
		"city":               true,
		"region":             true,
		"id":                 true,
		"lds_assignment_uid": true,
	}
)

func GetList(mainId string, searchParams map[string]map[string]interface{}, orderParams []map[string]string, limitParams map[string]int) (resList []map[string]interface{}, total int64, err error) {
	//ctx := context.Background()
	client := models.EsClient

	indexStr := beego.AppConfig.String("EsIndex") + mainId

	q := elastic.NewBoolQuery()

	//文档地址:  http://wiki.martech-global.com:8086/web/#/16?page_id=674

	tmpDefaultSearchParams := DefaultSearchParams
	for k, _ := range searchParams {
		for kk, _ := range DefaultSearchParams {
			if kk == k {
				delete(tmpDefaultSearchParams, kk)
			}
		}
	}

	if len(tmpDefaultSearchParams) > 0 {
		for k, v := range tmpDefaultSearchParams {
			searchParams[k] = v
		}
	}

	q = ParseSearchInfo(searchParams, q)

	q.Must(elastic.NewExistsQuery("sex"))

	sourceStr, err := q.Source()
	sourceStrB, err := json.Marshal(sourceStr)

	beego.Debug(string(sourceStrB))


	searchNext := client.Search().
		Index(indexStr).
		Query(q)

	//分页 简单实现 @todo 修改使用深度分页 以及 其他特性
	if _, ok := limitParams["start"]; ok && limitParams["start"] > 0 {
		searchNext = searchNext.From(limitParams["start"])
	}

	if _, ok := limitParams["offset"]; ok && limitParams["offset"] > 0 {
		searchNext = searchNext.Size(limitParams["offset"])
	}


	//2019-09-12 17:49:33  处理多个order 并且控制

	if len(orderParams) > 0 {


		for _, orderParamItem:=range orderParams {

				orderByField:=orderParamItem["orderBy"]
				//判断存在给定的默认分类参数中
				if _, ok := ConDefaultOrder[orderParamItem["orderBy"]]; ok {
					//
					orderByField = ConDefaultOrder[orderParamItem["orderBy"]]
				}

				ascending := true
				if orderParamItem["sort"] == "desc" {
					ascending = false
				}

				builder := elastic.SortInfo{Field: orderByField, Ascending: ascending}
				searchNext.SortWithInfo(builder)

				src1, err := builder.Source()
				beego.Debug(err)
				data1, err := json.Marshal(src1)
				beego.Debug(string(data1))
		}
	}


	searchResult, err := searchNext.
		Pretty(true).
		Do(context.TODO())


	if err != nil {
		beego.Debug(err)
	}

	resList = make([]map[string]interface{}, 0)
	total   = 0

	if err == nil && searchResult.Error == nil {

		if searchResult.Hits == nil {
			err = errors.New("expected SearchResult.Hits != nil; got nil")
		}

		beego.Debug(len(searchResult.Hits.Hits))

		total = searchResult.Hits.TotalHits

		for _, hit := range searchResult.Hits.Hits {
			item := make(map[string]interface{})

			err := json.Unmarshal(*hit.Source, &item)

			beego.Debug(item)

			resList = append(resList, item)
			if err != nil {
				beego.Debug(err)
			}
		}
	}
	return resList, total, err
}

func GetListV2(mainId string, sourceParams []interface{},  searchParams map[string]map[string]interface{}, orderParams []map[string]string, isScroll bool, scrollId string, limitParams map[string]int) (resList []map[string]interface{}, total int64, reScrollId string, err error) {

	////is_scroll
	//
	//	//scroll_id

	beego.Debug(isScroll, len(scrollId))

	client := models.EsClient

	indexStr := beego.AppConfig.String("EsIndex") + mainId

	var searchResult * elastic.SearchResult
	//var reScrollId string

	if len(strings.Trim(scrollId, " "))==0 {
		q := elastic.NewBoolQuery()

		//文档地址:  http://wiki.martech-global.com:8086/web/#/16?page_id=674
		tmpDefaultSearchParams := DefaultSearchParams
		for k, _ := range searchParams {
			for kk, _ := range DefaultSearchParams {
				if kk == k {
					delete(tmpDefaultSearchParams, kk)
				}
			}
		}

		if len(tmpDefaultSearchParams) > 0 {
			for k, v := range tmpDefaultSearchParams {
				searchParams[k] = v
			}
		}

		//根据参数 进行修改 如果在给定的列表中

		for k, item := range searchParams {

			if _, ok := item["search_type"]; ok {

				//beego.Debug(k, item)

				switch item["search_type"].(string) {

				case "str_eq":
					//beego.Debug("可以使用的eq:", item["search_type"], item["value"])

					q = q.Must(elastic.NewMatchQuery(k, utils.GetInterfaceString(item["value"])))

				case "eq":
					//beego.Debug("可以使用的eq:", item["search_type"], item["value"])
					if _, ok := NeedMatchPhraseKey[k]; ok {
						//存在
						q = q.Must(elastic.NewMatchPhraseQuery(k, item["value"]).Slop(0))
					} else {
						q = q.Must(elastic.NewTermQuery(k, utils.GetInterfaceString(item["value"])))
					}

				case "in":
					//$searchdata['filter']['terms']['tagname'] = $data['tags'];

					if _, ok := DefualutArrSearch[k]; ok {

						itemList := item["value"].([]interface{})

						q = q.Filter(elastic.NewTermsQuery(k, itemList...))

					} else {
						itemList := item["value"].([]interface{})
						beego.Debug(itemList)
						q = q.Filter(elastic.NewTermsQuery(k, itemList...))
						//q = q.Filter(elastic.NewTermsQuery(k, item["value"].([]interface{})))
					}

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

					beego.Debug("可以使用的neq:", item["search_type"], item["value"])
					if _, ok := NeedMatchPhraseKey[k]; ok {
						//存在
						q = q.MustNot(elastic.NewMatchPhraseQuery(k, item["value"]).Slop(0))
					} else {
						q = q.MustNot(elastic.NewTermQuery(k, item["value"]))
					}

				case "gte":
					q = q.Filter(elastic.NewRangeQuery(k).Gte(item["start"]))

				case "lte":
					q = q.Filter(elastic.NewRangeQuery(k).Lte(item["end"]))

				case "gt":
					q = q.Filter(elastic.NewRangeQuery(k).Gte(item["start"]))

				case "lt":
					q = q.Filter(elastic.NewRangeQuery(k).Lt(item["end"]))

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

				case "like":

					tmpStr := fmt.Sprintf("%v", item["value"])

					reg := regexp.MustCompile(StrRegex)
					matched:= reg.FindStringSubmatch(tmpStr)

					tmpValue := "*" + tmpStr + "*"

					if len(matched) > 0 {
						q = q.Must(elastic.NewWildcardQuery(k+".keyword", tmpValue))
					}else{
						q = q.Must(elastic.NewMatchPhraseQuery(k, tmpValue).Slop(0))
					}

				case "not_like":

					tmpStr := fmt.Sprintf("%v", item["value"])

					reg := regexp.MustCompile(StrRegex)
					matched:= reg.FindStringSubmatch(tmpStr)

					tmpValue := "*" + tmpStr + "*"

					if len(matched) > 0                    {
						q = q.MustNot(elastic.NewWildcardQuery(k+".keyword", tmpValue))
					}else{
						q = q.MustNot(elastic.NewMatchPhraseQuery(k, tmpValue).Slop(0))
					}

					//q = q.MustNot(elastic.NewMatchPhraseQuery(k, tmpValue).Slop(0))

				case "or":

					orTmpQuery := elastic.NewBoolQuery()

					for kk, vv := range item["value"].(map[string]interface{}) {
						tmpVV := vv.(map[string]interface{})
						beego.Debug(kk, vv)
						orTmpQuery.Should(elastic.NewTermsQuery(kk, tmpVV["value"].([]interface{})...))
					}

					q = q.Must(orTmpQuery)

					//tmpValue:=item["value"].(map[string]interface{})
					//
					//
					//orTmpQuery:=elastic.BoolQuery{}

					//q = q.Must()

					//$searchdata['filter']['terms']['tagname'] = $data['tags'];
					//
					//if _, ok := DefualutArrSearch[k]; ok {
					//
					//	itemList:= item["value"].([]interface{})
					//	q = q.Filter(elastic.NewTermsQuery(k, itemList...))
					//}else{
					//	q = q.Filter(elastic.NewTermsQuery(k, item["value"].([]interface{})))
					//}
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

				default:
					beego.Debug("未预料到的搜索选项:", k, item["search_type"], item["value"])
				}
			}

			//switch vv {
			//
			//case "eq" :
			//	beego.Debug("可以使用的eq", k, kk, vv)
			//	//eqSearch:=elastic.NewTermQuery(k,vv)
			//	q = q.Must(elastic.NewTermQuery(k, vv))
			//default:
			//	beego.Debug("未预料到的搜索选项", k, kk, vv)
			//}
		}

		q.Must(elastic.NewExistsQuery("sex"))


		searchNext := client.Search().
			Index(indexStr).
			Query(q)

		if len(sourceParams) > 0 {

			t:=make([]string, 0)
			for _,v:=range sourceParams{
				t = append(t, v.(string))
			}

			sourceField := elastic.NewFetchSourceContext(true).Include(t...) //.Exclude("*.description")
			searchNext.FetchSourceContext(sourceField)
		}


		//beego.Debug(all.Source())

		if len(orderParams) > 0 {


			for _, orderParamItem:=range orderParams {

				orderByField:=orderParamItem["orderBy"]
				//判断存在给定的默认分类参数中
				if _, ok := ConDefaultOrder[orderParamItem["orderBy"]]; ok {
					//
					orderByField = ConDefaultOrder[orderParamItem["orderBy"]]
				}

				ascending := true
				if orderParamItem["sort"] == "desc" {
					ascending = false
				}

				builder := elastic.SortInfo{Field: orderByField, Ascending: ascending}
				searchNext.SortWithInfo(builder)

				src1, err := builder.Source()
				beego.Debug(err)
				data1, err := json.Marshal(src1)
				beego.Debug(string(data1))
			}
		}

		//分页 简单实现 @todo 修改使用深度分页 以及 其他特性
		if _, ok := limitParams["start"]; ok && limitParams["start"] > 0 {
			searchNext = searchNext.From(limitParams["start"])
		}

		if _, ok := limitParams["offset"]; ok && limitParams["offset"] > 0 {
			searchNext = searchNext.Size(limitParams["offset"])
		}

		beego.Debug(err)


		if isScroll {
			pageSize := 10

			if _, ok := limitParams["offset"]; ok && limitParams["offset"] > 0 {
				pageSize = limitParams["offset"]
			}

			qT:=client.Scroll().Index(indexStr).Scroll("5m").Query(q).Size(pageSize)

			if len(sourceParams) > 0 {

				t:=make([]string, 0)
				for _,v:=range sourceParams{
					t = append(t, v.(string))
				}

				sourceField := elastic.NewFetchSourceContext(true).Include(t...) //.Exclude("*.description")
				qT.FetchSourceContext(sourceField)
			}

			searchResult, err =qT.Do(context.Background())

		}else{
			searchResult, err = searchNext.
				Pretty(true).
				Do(context.TODO())
		}

	}


	//2019-09-12 17:49:33  处理多个order 并且控制

	if isScroll && len(scrollId)>0{
		tmp:=client.Scroll("5m").ScrollId(scrollId)
		searchResult, err = tmp.Do(context.Background())
	}



	resList = make([]map[string]interface{}, 0)
	total   = 0

	if err != nil {
		beego.Debug(err)
		return  resList, 0, reScrollId, nil
	}

	if err == nil && searchResult!=nil && searchResult.Error == nil {

		if searchResult.Hits == nil {
			err = errors.New("expected SearchResult.Hits != nil; got nil")
		}

		if len(searchResult.ScrollId) > 0 {

			reScrollId  = searchResult.ScrollId

			beego.Debug(len(reScrollId))
			beego.Debug(err)

		}

		beego.Debug(len(searchResult.Hits.Hits))

		total = searchResult.Hits.TotalHits

		for _, hit := range searchResult.Hits.Hits {
			item := make(map[string]interface{})

			err := json.Unmarshal(*hit.Source, &item)

			resList = append(resList, item)
			if err != nil {
				beego.Debug(err)
			}
		}
	}
	return resList, total, reScrollId, err
}

//进行递归查询

func ParseSearchInfo(searchParams map[string]map[string]interface{}, q *elastic.BoolQuery) (*elastic.BoolQuery) {

	//根据参数 进行修改 如果在给定的列表中
	for k, item := range searchParams {

		if _, ok := item["search_type"]; ok {

			//beego.Debug(k, item)

			switch item["search_type"].(string) {

			case "all_eq":
				//beego.Debug("可以使用的eq:", item["search_type"], item["value"])

				q = q.Must(elastic.NewMatchQuery(k, utils.GetInterfaceString(item["value"])))

			case "str_eq":
				//beego.Debug("可以使用的eq:", item["search_type"], item["value"])

				q = q.Must(elastic.NewMatchQuery(k, utils.GetInterfaceString(item["value"])))

			case "eq":
				//beego.Debug("可以使用的eq:", item["search_type"], item["value"])
				if _, ok := NeedMatchPhraseKey[k]; ok {
					//存在
					q = q.Must(elastic.NewMatchPhraseQuery(k, item["value"]).Slop(0))
				} else {
					q = q.Must(elastic.NewTermQuery(k, utils.GetInterfaceString(item["value"])))
				}

			case "in":
				//$searchdata['filter']['terms']['tagname'] = $data['tags'];

				if _, ok := DefualutArrSearch[k]; ok {

					itemList := item["value"].([]interface{})

					q = q.Filter(elastic.NewTermsQuery(k, itemList...))

				} else {
					itemList := item["value"].([]interface{})
					beego.Debug(itemList)
					q = q.Filter(elastic.NewTermsQuery(k, itemList...))
					//q = q.Filter(elastic.NewTermsQuery(k, item["value"].([]interface{})))
				}

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

				beego.Debug("可以使用的neq:", item["search_type"], item["value"])
				if _, ok := NeedMatchPhraseKey[k]; ok {
					//存在
					q = q.MustNot(elastic.NewMatchPhraseQuery(k, item["value"]).Slop(0))
				} else {
					q = q.MustNot(elastic.NewTermQuery(k, item["value"]))
				}

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

			case "like":

				tmpStr := fmt.Sprintf("%v", item["value"])

				reg := regexp.MustCompile(StrRegex)
				matched:= reg.FindStringSubmatch(tmpStr)

				tmpValue := "*" + tmpStr + "*"

				if len(matched) > 0 {
					q = q.Must(elastic.NewWildcardQuery(k+".keyword", tmpValue))
				}else{
					q = q.Must(elastic.NewMatchPhraseQuery(k, tmpValue).Slop(0))
				}

			case "not_like":

				tmpStr := fmt.Sprintf("%v", item["value"])

				reg := regexp.MustCompile(StrRegex)
				matched:= reg.FindStringSubmatch(tmpStr)

				tmpValue := "*" + tmpStr + "*"

				if len(matched) > 0                    {
					q = q.MustNot(elastic.NewWildcardQuery(k+".keyword", tmpValue))
				}else{
					q = q.MustNot(elastic.NewMatchPhraseQuery(k, tmpValue).Slop(0))
				}

				//q = q.MustNot(elastic.NewMatchPhraseQuery(k, tmpValue).Slop(0))

			case "or":

				orTmpQuery := elastic.NewBoolQuery()

				for kk, vv := range item["value"].(map[string]interface{}) {
					tmpVV := vv.(map[string]interface{})
					beego.Debug(kk, vv)
					orTmpQuery.Should(elastic.NewTermsQuery(kk, tmpVV["value"].([]interface{})...))
				}

				q = q.Must(orTmpQuery)

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
					tagQuery.Must(elastic.NewMatchQuery("is_del", "N"))

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
					tagQuery.Must(elastic.NewMatchQuery("is_del", "N"))

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
					tagQuery.Must(elastic.NewMatchQuery("is_del", "N"))

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
					tagQuery.Must(elastic.NewMatchQuery("is_del", "N"))

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

				tq:=elastic.NewBoolQuery()
				beego.Debug(tq)

				//tq1:= make([]elastic.Query, 0)
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

					beego.Debug(sType, multiOp, stp)

					if len(sType)> 0 && len(multiOp) > 0 && sType=="multi" {

						beego.Debug(tp)

						tmpQ := ParseSearchInfo(tp, ttq)

						ts, err:=tmpQ.Source()
						if err!=nil {
							beego.Debug(err)
						}
						tss,err:=json.Marshal(ts)
						beego.Debug(string(tss))

						//tq1=append(tq1, tmpQ)
						switch multiOpt {
						case "and":
							tq.Must(tmpQ)
						case "or":
							tq.Should(tmpQ)
						case "not":
							tq.MustNot(tmpQ)
						default:
							beego.Debug("错误的逻辑关系",sType)
						}
					}
				}

				q.Must(tq)

				//switch multiOpt {

				//case "and":
				//	q.Must(tq1...)

				//case "or":
				//	q.Should(tq1...)
				//
				//case "not":
				//	q.MustNot(tq1...)
				//}

			default:
				beego.Debug("未预料到的搜索选项:", k, item["search_type"], item["value"])
			}
		}

	}

	return q
}
