package utils

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"io"
	"net/http"
	"strconv"
	"time"
)

var HASH_ALGORITHM_256 = sha256.New
var HASH_ALGORITHM_512 = sha512.New

type ApiResponse struct {
	Message string      `json:"m"`
	Status  int         `json:"s"`
	Data    interface{} `json:"d"`
}

const (
	IV_LEN         = 16
	HASH_COUNT     = 1024
	OUTPUT_KEY_LEN = 64
)

func GenerateToken256(payload []byte, salt []byte) string {
	if salt == nil {
		saltStr, _ := AlphaString(IV_LEN)
		salt = []byte(saltStr)
	}
	token := Encrypt(payload, salt, HASH_COUNT, (OUTPUT_KEY_LEN-IV_LEN)/2, HASH_ALGORITHM_256)
	return hex.EncodeToString(token) + string(salt)
}

func GenerateToken512(payload []byte, salt []byte) string {
	if salt == nil {
		saltStr, _ := AlphaString(IV_LEN)
		salt = []byte(saltStr)
	}
	token := Encrypt(payload, salt, HASH_COUNT, (OUTPUT_KEY_LEN-IV_LEN)/2, HASH_ALGORITHM_512)
	return hex.EncodeToString(token) + string(salt)
}

func GeneratePayloadNow(payload map[string]interface{}, publicKey string) string {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	now := time.Now()
	nowUnix := now.Unix() - int64(now.Second())
	return string(payloadBytes) + publicKey + strconv.FormatInt(nowUnix, 10)
}

func GeneratePayload3Minutes(payload map[string]interface{}, publicKey string) []string {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	payloadStr := string(append(payloadBytes, []byte(publicKey)...))
	now := time.Now()
	nowUnix := now.Unix() - int64(now.Second())
	result := make([]string, 3)
	result[0] = payloadStr + strconv.FormatInt(nowUnix-60, 10)
	result[1] = payloadStr + strconv.FormatInt(nowUnix, 10)
	result[2] = payloadStr + strconv.FormatInt(nowUnix+60, 10)
	return result
}

func CheckAccessToken(userID string, clientID string, token string) (int, error) {
	return 1, nil

	checkUrl := beego.AppConfig.String("authCenter") + "/v1/token/check?token=%s&client_id=%s&user_id=%s"
	checkUrl = fmt.Sprintf(checkUrl, token, clientID, userID)
	client := new(http.Client)
	req, _ := http.NewRequest(http.MethodGet, checkUrl, nil)
	req.Header.Set("X-test-Src", "user_like")
	resp, err := client.Do(req)
	if err != nil {
		beego.Error(err)
		return 0, errors.New("auth error")
	}
	var buffer bytes.Buffer
	io.Copy(&buffer, resp.Body)
	apiResponse := new(ApiResponse)
	err = json.Unmarshal(buffer.Bytes(), apiResponse)
	if err != nil {
		beego.Error(err)
		return 0, errors.New("response error")
	}
	if apiResponse.Status == 1 {
		return 1, nil
	}
	return apiResponse.Status, errors.New(apiResponse.Message)
}

func IllegalToken(token string, src string, inputMap map[string]interface{}) bool {
	beego.Debug(token, src)
	if len(token) != 64 {
		return true
	}
	if len(src) <= 0 {
		return true
	}
	publicKey := PublicKeys[src]
	salt := token[48:]
	payloads := GeneratePayload3Minutes(inputMap, publicKey)
	beego.Debug(payloads)
	for i := 0; i < 3; i++ {
		genToken := GenerateToken256([]byte(payloads[i]), []byte(salt))
		beego.Debug("GenToken:", genToken)
		if token == genToken {
			return false
		}
	}
	return true
}
