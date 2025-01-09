package cvt

import (
	"encoding/json"
	"fmt"
	"io"
	"moredoc/pkg/logger"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

var log, _ = logger.NewLogger(logger.SetDebug(true), logger.SetLevel(logger.DebugLevel))

func fn(s string) string {
	return s[strings.LastIndex(s, "/")+1:]
}

// CallerName 获取调用者名称.
func CallerName() string {
	var s string
	for i := 1; i < 6; i++ {
		if pc, file, line, ok := runtime.Caller(i); ok {
			s += fmt.Sprintf("%s:%d %s", fn(file), line, fn(runtime.FuncForPC(pc).Name())) + " <--- "
			continue
		}
		break
	}
	return strings.TrimSuffix(s, " <--- ")
}

// JSONImpl 只需令 "var JSONImpl = nil" 就可以切换JSON库.
var JSONImpl jsoniter.API = jsoniter.ConfigCompatibleWithStandardLibrary

type JSONEncoder interface {
	Encode(v interface{}) error
}

type JSONDecoder interface {
	Decode(v interface{}) error
}

// Marshal 实质是调用开源库"github.com/json-iterator/go".
func Marshal(data interface{}) ([]byte, error) {
	if JSONImpl != nil {
		return JSONImpl.Marshal(data)
	}
	return json.Marshal(data)
}

// MarshalIndent 实质是调用开源库"github.com/json-iterator/go".
func MarshalIndent(data interface{}, prefix, indent string) ([]byte, error) {
	if JSONImpl != nil {
		return JSONImpl.MarshalIndent(data, prefix, indent)
	}
	return json.MarshalIndent(data, prefix, indent)
}

// Unmarshal 实质是调用开源库"github.com/json-iterator/go".
func Unmarshal(data []byte, v interface{}) error {
	if l := len(data); l < 2 {
		// log.Errorf(ErrFailedCvtJSON, "Bad JSON: %s.", string(data))
		return nil
	}
	if JSONImpl != nil {
		return JSONImpl.Unmarshal(data, v)
	}
	return json.Unmarshal(data, v)
}

// NewDecoder 实质是调用开源库"github.com/json-iterator/go".
func NewDecoder(reader io.Reader) JSONDecoder {
	if JSONImpl != nil {
		return JSONImpl.NewDecoder(reader)
	}
	return json.NewDecoder(reader)
}

// NewDecoder 实质是调用开源库"github.com/json-iterator/go".
func NewEncoder(writer io.Writer) JSONEncoder {
	if JSONImpl != nil {
		return JSONImpl.NewEncoder(writer)
	}
	return json.NewEncoder(writer)
}

// ToNumber 将interface安全转换为number.
func ToNumber(v interface{}) float64 {
	if v == nil {
		return 0
	}

	switch a := v.(type) {
	case []interface{}:
		if len(a) > 0 {
			log.Warnf("参数错误,ToNumber: your input is an array, caller: %s\n", CallerName())
			return ToNumber(a[0])
		}
	case []map[string]interface{}:
		if len(a) > 0 {
			log.Warnf("参数错误ToNumber: your input is map-array, caller: %s\n", CallerName())
			return ToNumber(a[0])
		}
	case float64:
		return a
	case string:
		if a == "" {
			return 0
		}
		if b, err := strconv.ParseBool(a); err == nil {
			return map[bool]float64{true: 1, false: 0}[b]
		}
		if f, err := strconv.ParseFloat(a, 64); err == nil {
			return f
		}
	case bool:
		if a {
			return 1
		}
		return 0
	case uint:
		return float64(a)
	case uint8:
		return float64(a)
	case uint16:
		return float64(a)
	case uint32:
		return float64(a)
	case uint64:
		return float64(a)
	case int:
		return float64(a)
	case int8:
		return float64(a)
	case int16:
		return float64(a)
	case int32:
		return float64(a)
	case int64:
		return float64(a)
	case float32:
		return float64(a)
	}
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		ref := reflect.ValueOf(v)
		if ref.IsZero() {
			return 0
		}
		if elem := ref.Elem(); elem.CanInterface() {
			return ToNumber(elem.Interface())
		}
	}
	log.Debugf("[%v]ToNumber: your input '%v' is unsupported, caller: %s\n", reflect.TypeOf(v).Kind(), v, CallerName())

	return 0
}

func ToInt(v interface{}) int {
	return int(ToNumber(v))
}

// ToInt64 将interface安全转换为int64.
func ToInt64(v interface{}) int64 {
	return int64(ToNumber(v))
}

func ToInt32(v interface{}) int32 {
	return int32(ToNumber(v))
}

// ToBoolean 将interface安全转换为bool.
func ToBoolean(v interface{}) bool {
	return ToNumber(v) != 0
}

// Bytes2String 切片转换为字符串，字符串是不可变的.
func Bytes2String(slice []byte) string {
	return *(*string)(unsafe.Pointer(&slice))
}

// ToString 将interface安全转换为字符串.
// 如果v是数组将返回数组的第一个元素.
func ToString(v interface{}) string {
	if v == nil {
		return ""
	}

	switch a := v.(type) {
	case []interface{}: // 数组
		if len(a) > 0 {
			log.Debugf("ToString: your input is an array, caller: %s\n", CallerName())
			return ToString(a[0])
		}
	case []map[string]interface{}: // map数组
		if len(a) > 0 {
			log.Debugf("ToString: your input is map-array, caller: %s\n", CallerName())
			return ToString(a[0])
		}
	case map[string]interface{}: // map
		b, _ := Marshal(a)
		return string(b)
	case float64: // number
		return strconv.FormatFloat(a, 'f', -1, 64)
	case string: // string
		return a
	case bool: // bool
		if a {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32:
		return ToString(ToNumber(v))
	case []byte:
		return Bytes2String(a)
	}
	b, err := Marshal(v)
	if err != nil {
		log.Warnf("类型不支持,ToString: your input is unsupported, caller: %s\n", CallerName())
		return ""
	}
	return string(b)
}

// MapArrayToArray 将map数组[]map[string]interface{}安全转换为[]interface{}.
func MapArrayToArray(v []map[string]interface{}) []interface{} {
	r := make([]interface{}, len(v))
	for i, s := range v {
		r[i] = s
	}
	return r
}

// Valid 实质是调用开源库"github.com/json-iterator/go".
func Valid(data []byte) bool {
	var v interface{}
	return Unmarshal(data, &v) == nil
}

// MaybeJSON 判断字符串是否可能是JSON.
// 注意: 这个函数就简单的判断首尾字符。返回true，可能是json；返回false，一定不是json。
func MaybeJSON(data string) bool {
	var num = len(data)
	if num < 2 {
		return false
	}
	if num < 8192 {
		return Valid([]byte(data))
	}
	var idx int
	for i := 0; i < num; i++ {
		if data[i] != ' ' && data[i] != '\t' && data[i] != '\r' && data[i] != '\n' {
			idx = i
			break
		}
	}
	left := data[idx]
	if left != '{' && left != '[' {
		return false
	}
	for i := num - 1; i >= 0; i-- {
		if data[i] != ' ' && data[i] != '\t' && data[i] != '\r' && data[i] != '\n' {
			idx = i
			break
		}
	}
	right := data[idx]
	return (left == '{' && right == '}') || (left == '[' && right == ']')
}

// ToStringArray 将interface安全转换为string数组.
func ToStringArray(v interface{}) []string {
	arr := ToArray(v)
	r := make([]string, len(arr))
	for i, s := range arr {
		r[i] = ToString(s)
	}
	return r
}

func ToIntArray(arr []interface{}) []int {
	r := make([]int, len(arr))
	for i, s := range arr {
		r[i] = ToInt(s)
	}
	return r
}

func ToInt64Array(v interface{}) []int64 {
	arr := ToArray(v)
	r := make([]int64, len(arr))
	for i, s := range arr {
		r[i] = ToInt64(s)
	}
	return r
}

func ToInt32Array(v interface{}) []int32 {
	arr := ToArray(v)
	r := make([]int32, len(arr))
	for i, s := range arr {
		r[i] = ToInt32(s)
	}
	return r
}

func ToIntBoolArray(v interface{}) []bool {
	arr := ToArray(v)
	r := make([]bool, len(arr))
	for i, s := range arr {
		r[i] = ToBoolean(s)
	}
	return r
}

// ToArray 将interface安全转换为数组.
// 规则1：如果是基本数据类型的单个元素(int,string,bool等),则构造一个仅含有该元素的interface{}数组;
// 规则2：如果是基本数据类型的数组（[]int,[]string,[]float64等）,则转换为interface{}数组;
// 规则3：如果是数组形式的JSON字符串,则反序列化为interface{}数组.
func ToArray(v interface{}) []interface{} {
	if v == nil {
		return make([]interface{}, 0)
	}

	switch a := v.(type) {
	case string:
		if MaybeJSON(a) {
			var arr []interface{}
			if err := Unmarshal([]byte(a), &arr); err == nil {
				return arr
			}
		}
		return []interface{}{a}
	case bool, float32, float64, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, map[string]interface{}:
		return []interface{}{a}
	case []interface{}:
		return a
	case []map[string]interface{}:
		return MapArrayToArray(a)
	case []bool, []string, []float32, []float64, []int, []int8, []int16, []int32, []int64,
		[]uint, []uint8, []uint16, []uint32, []uint64:
		f := reflect.ValueOf(a)
		arr := make([]interface{}, f.Len())
		for i, n := 0, f.Len(); i < n; i++ {
			arr[i] = f.Index(i).Interface()
		}
		return arr
	default:
		log.Warnf("类型不支持,ToArray: your input is unsupported, caller: %s\n", CallerName())
	}

	return make([]interface{}, 0)
}

// ExistElem 数组arr中是否存在元素elems.存在至少一个就返回true.
func ExistElem(arr []interface{}, elems ...interface{}) bool {
	for _, v := range arr {
		for _, elem := range elems {
			if v == elem {
				return true
			}
		}
	}
	return false
}

// ExistStringElem 数组arr中是否存在元素elem.
func ExistStringElem(arr []string, elem string) bool {
	for _, v := range arr {
		if v == elem {
			return true
		}
	}
	return false
}
