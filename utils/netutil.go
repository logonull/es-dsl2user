package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func SendRequestBodyJson(reqType string, urlString string, header map[string][]string, body []byte, respModel interface{}, query url.Values) *http.Response {
	client := new(http.Client)
	if query != nil {
		var buffer bytes.Buffer
		buffer.WriteString(urlString)
		buffer.WriteString("?")
		buffer.WriteString(query.Encode())
		urlString = buffer.String()
	}
	req, _ := http.NewRequest(reqType, urlString, bytes.NewReader(body))
	if header != nil {
		header["Content-Type"] = []string{"application/json"}
	} else {
		header = map[string][]string{
			"Content-Type": {"application/json"},
		}
	}
	req.Header = header
	if query != nil {
		urlString += query.Encode()
	}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	bodyMsg, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyMsg, respModel)
	return resp
}

func SendRequestBodyForm(reqType string, urlString string, header map[string][]string, body map[string]string, respModel interface{}, query url.Values) *http.Response {
	client := new(http.Client)
	form := url.Values{}
	if body != nil {
		for key, value := range body {
			form.Add(key, value)
		}
	}
	if query != nil {
		var buffer bytes.Buffer
		buffer.WriteString(urlString)
		buffer.WriteString("?")
		buffer.WriteString(query.Encode())
		urlString = buffer.String()
	}
	req, _ := http.NewRequest(reqType, urlString, strings.NewReader(form.Encode()))
	if header != nil {
		header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	} else {
		header = map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		}
	}
	req.Header = header
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	bodyMsg, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyMsg, respModel)
	return resp
}
