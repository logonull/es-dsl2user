package utils

import (
	"fmt"
	"strings"
)

func RemoveDuplicatesAndEmpty(a []string) (ret []string) {
	a_len := len(a)
	for i := 0; i < a_len; i++ {
		if (i > 0 && a[i-1] == a[i]) || len(a[i]) == 0 {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}

func GetInterfaceString(v interface{}) string {

	switch v.(type) {
	case string:
		return v.(string)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func GetClearnType(str string) string {

	result := strings.Index(str, "(")
	prefix := "string"
	if result >= 0 {
		prefix = str[0:result]
	}
	return prefix
}
