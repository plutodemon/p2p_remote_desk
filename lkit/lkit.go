package lkit

import (
	"fmt"
	"os"
)

// SigChan 创建一个通道来接收信号
var SigChan = make(chan os.Signal)

func AnyToStr(v interface{}) string {
	return fmt.Sprint(v)
}

func SliceToStrList[k comparable](v []k) []string {
	var res []string
	for _, i := range v {
		res = append(res, fmt.Sprint(i))
	}
	return res
}

func GetAddr(host, port any) string {
	return fmt.Sprintf("%s:%d", host, port)
}
