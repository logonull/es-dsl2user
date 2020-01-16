package models

import (
	"github.com/astaxie/beego"
	"github.com/olivere/elastic"
)

var EsClient *elastic.Client

func init() {

	var err error

	EsClient, err = elastic.NewClient(
		elastic.SetURL(beego.AppConfig.String("EsHost")),
		elastic.SetHealthcheck(false),
		elastic.SetSniff(false),
		//elastic.SetBasicAuth("user", "secret"),

	)
	if err != nil {
		// Handle error
		panic(err)
	}
}
