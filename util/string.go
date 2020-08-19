package util

import (
	"bytes"
	"crypto/md5"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

//RandomStr 随机生成字符串
func RandomStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// convert string to specify type

type StrTo string

func (f *StrTo) Set(v string) {
	if v != "" {
		*f = StrTo(v)
	} else {
		f.Clear()
	}
}

func (f *StrTo) Clear() {
	*f = StrTo(0x1E)
}

func (f StrTo) Exist() bool {
	return string(f) != string(0x1E)
}

func (f StrTo) Bool() (bool, error) {
	if f == "on" {
		return true, nil
	}
	return strconv.ParseBool(f.String())
}

func (f StrTo) Float32() (float32, error) {
	v, err := strconv.ParseFloat(f.String(), 32)
	return float32(v), err
}

func (f StrTo) Float64() (float64, error) {
	return strconv.ParseFloat(f.String(), 64)
}

func (f StrTo) Int() (int, error) {
	v, err := strconv.ParseInt(f.String(), 10, 32)
	return int(v), err
}

func (f StrTo) Int8() (int8, error) {
	v, err := strconv.ParseInt(f.String(), 10, 8)
	return int8(v), err
}

func (f StrTo) Int16() (int16, error) {
	v, err := strconv.ParseInt(f.String(), 10, 16)
	return int16(v), err
}

func (f StrTo) Int32() (int32, error) {
	v, err := strconv.ParseInt(f.String(), 10, 32)
	return int32(v), err
}

func (f StrTo) Int64() (int64, error) {
	v, err := strconv.ParseInt(f.String(), 10, 64)
	return int64(v), err
}

func (f StrTo) Uint() (uint, error) {
	v, err := strconv.ParseUint(f.String(), 10, 32)
	return uint(v), err
}

func (f StrTo) Uint8() (uint8, error) {
	v, err := strconv.ParseUint(f.String(), 10, 8)
	return uint8(v), err
}

func (f StrTo) Uint16() (uint16, error) {
	v, err := strconv.ParseUint(f.String(), 10, 16)
	return uint16(v), err
}

func (f StrTo) Uint32() (uint32, error) {
	v, err := strconv.ParseUint(f.String(), 10, 32)
	return uint32(v), err
}

func (f StrTo) Uint64() (uint64, error) {
	v, err := strconv.ParseUint(f.String(), 10, 64)
	return uint64(v), err
}

func (f StrTo) String() string {
	if f.Exist() {
		return string(f)
	}
	return ""
}

// convert any type to string
func ToStr(value interface{}, args ...int) (s string) {
	switch v := value.(type) {
	case bool:
		s = strconv.FormatBool(v)
	case float32:
		s = strconv.FormatFloat(float64(v), 'f', argInt(args).Get(0, -1), argInt(args).Get(1, 32))
	case float64:
		s = strconv.FormatFloat(v, 'f', argInt(args).Get(0, -1), argInt(args).Get(1, 64))
	case int:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int8:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int16:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int32:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int64:
		s = strconv.FormatInt(v, argInt(args).Get(0, 10))
	case uint:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint8:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint16:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint32:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint64:
		s = strconv.FormatUint(v, argInt(args).Get(0, 10))
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		s = fmt.Sprintf("%v", v)
	}
	return s
}

// convert any numeric value to int64
func ToInt64(value interface{}) (d int64, err error) {
	val := reflect.ValueOf(value)
	switch value.(type) {
	case int, int8, int16, int32, int64:
		d = val.Int()
	case uint, uint8, uint16, uint32, uint64:
		d = int64(val.Uint())
	default:
		err = fmt.Errorf("ToInt64 need numeric not `%T`", value)
	}
	return
}

type argString []string

func (a argString) Get(i int, args ...string) (r string) {
	if i >= 0 && i < len(a) {
		r = a[i]
	} else if len(args) > 0 {
		r = args[0]
	}
	return
}

type argInt []int

func (a argInt) Get(i int, args ...int) (r int) {
	if i >= 0 && i < len(a) {
		r = a[i]
	}
	if len(args) > 0 {
		r = args[0]
	}
	return
}

type argAny []interface{}

func (a argAny) Get(i int, args ...interface{}) (r interface{}) {
	if i >= 0 && i < len(a) {
		r = a[i]
	}
	if len(args) > 0 {
		r = args[0]
	}
	return
}

func formatMapToXML(req map[string]string) (buf []byte, err error) {
	bodyBuf := textBufferPool.Get().(*bytes.Buffer)
	bodyBuf.Reset()
	defer textBufferPool.Put(bodyBuf)

	if bodyBuf == nil {
		return []byte{}, errors.New("nil xmlWriter")
	}

	if _, err = io.WriteString(bodyBuf, "<xml>"); err != nil {
		return
	}

	for k, v := range req {
		if _, err = io.WriteString(bodyBuf, "<"+k+">"); err != nil {
			return
		}
		if err = xml.EscapeText(bodyBuf, []byte(v)); err != nil {
			return
		}
		if _, err = io.WriteString(bodyBuf, "</"+k+">"); err != nil {
			return
		}
	}

	if _, err = io.WriteString(bodyBuf, "</xml>"); err != nil {
		return
	}

	return bodyBuf.Bytes(), nil
}

var textBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 16<<10)) // 16KB
	},
}

func GetMD5(args ...string) string {
	var str string
	for _, s := range args {
		str += s
	}
	value := md5.Sum([]byte(str))
	rs := []rune(fmt.Sprintf("%x", value))
	return string(rs)
}

func GetSign(args ...string) string {
	salt := "Yexhj8agldf3yaexuda7da"
	var str string
	for _, s := range args {
		str += s
	}
	str += salt
	value := md5.Sum([]byte(str))
	rs := []rune(fmt.Sprintf("%x", value))
	return string(rs)
}

//res:原字符串，sep替换的字符串，idx开始替换的位置
//如果替换的字符串超出，就加在原字符串后面
//idx从0开始
func ReplaceString(res, sep string, idx int) string {
	sepLen := len(sep)
	if sepLen == 0 {
		return res
	}

	resLen := len(res)
	if idx > resLen-1 {
		return res + sep
	}

	allLen := resLen
	if sepLen > resLen-idx {
		allLen = idx + sepLen
	}

	buf := bytes.Buffer{}
	sepIdx := 0
	for i := 0; i < allLen; i++ {
		if i < idx {
			buf.WriteByte(res[i])
		} else {
			if sepIdx < sepLen {
				buf.WriteByte(sep[sepIdx])
			} else {
				buf.WriteByte(res[i])
			}
			sepIdx++
		}
	}
	return buf.String()
}

//
func ConvFormMapToString(mData map[string]string) string {
	formBuf := bytes.Buffer{}
	l := len(mData)
	i := 0
	for k, v := range mData {
		formBuf.WriteString(k)
		formBuf.WriteString("=")
		formBuf.WriteString(v)
		if i < l {
			formBuf.WriteString("&")
			i++
		}
	}
	return string(formBuf.Bytes())
}

//检查是否包含,必须全部包含
func CheckContains(s string, subArr ...string) bool {
	for _, sub := range subArr {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

//有任何一个包含就返回true
func CheckContainsAny(s string, subArr ...string) bool {
	for _, sub := range subArr {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
