package lkit

import (
	"fmt"
	"os"
)

// SigChan 创建一个通道来接收信号
var SigChan = make(chan os.Signal)

func GetAddr(host, port any) string {
	return fmt.Sprintf("%s:%d", host, port)
}
