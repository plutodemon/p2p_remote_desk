package internal

import (
	"fmt"
	"net"
	"testing"
)

func Test_NatDetector(t *testing.T) {
	const serverPort = ":3478" // STUN 服务器监听的端口

	addr, err := net.ResolveUDPAddr("udp", serverPort)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("NAT 类型检测服务器已启动，监听端口", serverPort)

	buffer := make([]byte, 1024)
	for {
		_, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}

		fmt.Printf("收到来自 %s 的请求\n", clientAddr.String())

		// 将服务器看到的客户端 IP 和端口返回给客户端
		response := fmt.Sprintf("Your external address is %s", clientAddr.String())
		_, err = conn.WriteToUDP([]byte(response), clientAddr)
		if err != nil {
			fmt.Println("Error sending response:", err)
		}
	}
}
