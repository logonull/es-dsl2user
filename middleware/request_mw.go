package middleware

import (
	"encoding/json"
	"strconv"
)

const (
	TYPE_STRING = iota
	TYPE_INT
	TYPE_FLOAT
	TYPE_BOOLEAN
	TYPE_ARRAY
	TYPE_JSON_OBJECT
)
const (
	POS_QUERY = iota
	POS_HEADER
	POS_BODY_FORM
	POS_BODY_JSON
	POS_PATH
	POS_COOKIE
)

type InputInterface interface {
	GetQuery(key string) string
	GetForm(key string) string
	GetHeader(key string) string
	GetCookie(key string) string
	GetBody() []byte
	GetPath(key string) string
	GetArray(key string) []string
	RenderOutput(r *RenderStruct)
}

type InputContainer struct {
	simpleMap map[string]interface{}
	objectMap map[string][]byte
	arrayMap  map[string][]string
}

type InputRule struct {
	Name        string
	Description string
	Position    int
	Type        int
	//此处有蹊跷
	Limit     []string
	Necessary bool
	Max       int
	Min       int
	Default   interface{}
}

type CheckError struct {
	Missing string `json:"missing" xml:"missing"`
	Illegal string `json:"illegal" xml:"illegal"`
}

func (c *CheckError) Error() string {
	return "input error"
}

//must use this func to create a new container
func NewContainer() *InputContainer {
	return &InputContainer{
		simpleMap: make(map[string]interface{}),
		objectMap: make(map[string][]byte),
		arrayMap:  make(map[string][]string),
	}
}

func (i *InputContainer) CheckInput(input InputInterface, rules []*InputRule) *CheckError {
	for _, rule := range rules {
		//check special input type
		switch rule.Type {
		case TYPE_ARRAY:
			arr := input.GetArray(rule.Name)
			if rule.Necessary && len(arr) <= 0 {
				return &CheckError{
					Missing: rule.Name,
				}
			}
			if rule.Max > 0 {
				if len(arr) > rule.Max {
					return &CheckError{
						Illegal: rule.Name,
					}
				}
			}
			if rule.Min > 0 {
				if len(arr) < rule.Min {
					return &CheckError{
						Illegal: rule.Name,
					}
				}
			}
			i.arrayMap[rule.Name] = arr
		case TYPE_JSON_OBJECT:
			var bytes []byte
			switch rule.Position {
			default:
				fallthrough
			case POS_BODY_JSON:
				bytes = input.GetBody()
			case POS_QUERY:
				bytes = []byte(input.GetQuery(rule.Name))
			case POS_BODY_FORM:
				bytes = []byte(input.GetForm(rule.Name))
			case POS_HEADER:
				bytes = []byte(input.GetHeader(rule.Name))
			case POS_COOKIE:
				bytes = []byte(input.GetCookie(rule.Name))
			}
			if rule.Necessary && len(bytes) <= 0 {
				return &CheckError{
					Missing: rule.Name,
				}
			} else if len(bytes) <= 0 {
				continue
			}
			err := json.Unmarshal(bytes, &map[string]interface{}{})
			if err != nil {
				return &CheckError{
					Illegal: rule.Name,
				}
			}
			i.objectMap[rule.Name] = bytes
		default:
			//check normal input type
			var val string
			switch rule.Position {
			case POS_PATH:
				val = input.GetPath(rule.Name)
			default:
				fallthrough
			case POS_QUERY:
				val = input.GetQuery(rule.Name)
			case POS_BODY_FORM:
				val = input.GetForm(rule.Name)
			case POS_HEADER:
				val = input.GetHeader(rule.Name)
			case POS_COOKIE:
				val = input.GetCookie(rule.Name)
			}
			if rule.Necessary && len(val) <= 0 {
				return &CheckError{
					Missing: rule.Name,
				}
			} else if len(val) <= 0 {
				i.simpleMap[rule.Name] = rule.Default
			} else {
				res, err := checkValue(rule, val)
				if err != nil {
					return err
				}
				i.simpleMap[rule.Name] = res
			}
		}
	}
	return nil
}

func checkValue(rule *InputRule, val string) (interface{}, *CheckError) {
	limit := len(rule.Limit) <= 0
	if !limit {
		for _, i := range rule.Limit {
			if val == i {
				limit = true
				break
			}
		}
		if !limit {
			return nil, &CheckError{
				Illegal: rule.Name,
			}
		}
	}
	switch rule.Type {
	default:
		fallthrough
	case TYPE_STRING:
		if rule.Max > 0 {
			if len(val) > rule.Max {
				return nil, &CheckError{
					Illegal: rule.Name,
				}
			}
		}
		if rule.Min > 0 {
			if len(val) < rule.Min {
				return nil, &CheckError{
					Illegal: rule.Name,
				}
			}
		}
		return val, nil
	case TYPE_INT:
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return nil, &CheckError{
				Illegal: rule.Name,
			}
		}
		if rule.Max > 0 {
			if intVal > rule.Max {
				return nil, &CheckError{
					Illegal: rule.Name,
				}
			}
		}
		if rule.Min > 0 {
			if intVal < rule.Min {
				return nil, &CheckError{
					Illegal: rule.Name,
				}
			}
		}
		return intVal, nil
	case TYPE_FLOAT:
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, &CheckError{
				Illegal: rule.Name,
			}
		}
		if rule.Max > 0 {
			if floatVal > float64(rule.Max) {
				return nil, &CheckError{
					Illegal: rule.Name,
				}
			}
		}
		if rule.Min > 0 {
			if floatVal < float64(rule.Min) {
				return nil, &CheckError{
					Illegal: rule.Name,
				}
			}
		}
		return floatVal, nil
	case TYPE_BOOLEAN:
		booleanVal, err := strconv.ParseBool(val)
		if err != nil {
			return nil, &CheckError{
				Illegal: rule.Name,
			}
		}
		return booleanVal, nil
	}
}

func (i *InputContainer) GetInt(key string) int {
	if res, ok := i.simpleMap[key].(int); ok {
		return res
	}
	return 0
}

func (i *InputContainer) GetBoolean(key string) bool {
	if res, ok := i.simpleMap[key].(bool); ok {
		return res
	}
	return false
}

func (i *InputContainer) GetFloat(key string) float64 {
	if res, ok := i.simpleMap[key].(float64); ok {
		return res
	}
	return 0
}

func (i *InputContainer) GetString(key string) string {
	if res, ok := i.simpleMap[key].(string); ok {
		return res
	}
	return ""
}

func (i *InputContainer) GetArray(key string) []string {
	return i.arrayMap[key]
}
func (i *InputContainer) GetObject(key string, obj interface{}) error {
	bytes := i.objectMap[key]
	return json.Unmarshal(bytes, obj)
}
