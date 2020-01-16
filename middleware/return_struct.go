package middleware

import (
	"github.com/astaxie/beego"
	"net/http"
)

const (
	S_UNKNOWN_ERROR = iota
	S_OK
	S_NO_RESULT
	S_ILLEGAL_TOKEN
	S_WRONG_INPUT
	S_WRONG_EXEC_RESULT
	S_TOKEN_EXPIRE
)
const (
	M_OK                = "success"
	M_ILLEGAL_TOKEN     = "illegal token"
	M_TOKEN_EXPIRE      = "token expired"
	M_ILLEGAL_SESSION   = "illegal session"
	M_WRONG_INPUT       = "wrong input"
	M_WRONG_EXEC_RESULT = "exec error"
)

type RenderStruct struct {
	Status int         `json:"-"`
	S      int         `json:"code"`
	M      string      `json:"msg"`
	D      interface{} `json:"data"`
}

func Success(list interface{}) (r *RenderStruct) {
	r = new(RenderStruct)
	r.Status = http.StatusOK
	r.S = 0
	r.M = M_OK
	if list == nil {
		list = []interface{}{}
	}
	r.D = list
	return
}

func InputErr(data interface{}) (r *RenderStruct) {
	r = new(RenderStruct)
	r.Status = http.StatusBadRequest
	if data == nil {
		data = []interface{}{}
	}
	r.S = S_WRONG_INPUT
	r.M = M_WRONG_INPUT
	r.D = data
	if beego.BConfig.RunMode != "dev" {
		r.D = []interface{}{}
	}
	return
}

func NormalErr(status int, s int, msg string, d interface{}) (r *RenderStruct) {
	r = new(RenderStruct)
	r.Status = status
	r.S = s
	r.M = msg
	if d == nil {
		d = []interface{}{}
	}
	r.D = d
	return
}

func ServerErr(message string) (r *RenderStruct) {
	r = new(RenderStruct)
	r.Status = http.StatusInternalServerError
	r.S = S_WRONG_EXEC_RESULT
	r.M = message
	if len(message) <= 0 {
		r.M = M_WRONG_EXEC_RESULT
	}
	r.D = []interface{}{}
	return
}

func TokenErr(errorType int) (r *RenderStruct) {
	r = new(RenderStruct)
	r.Status = http.StatusForbidden
	switch errorType {
	case S_TOKEN_EXPIRE:
		r.S = S_TOKEN_EXPIRE
		r.M = M_TOKEN_EXPIRE
	default:
		fallthrough
	case S_ILLEGAL_TOKEN:
		r.S = S_ILLEGAL_TOKEN
		r.M = M_ILLEGAL_TOKEN
	}
	r.D = []interface{}{}
	return
}
