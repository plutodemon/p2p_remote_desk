package network

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func Test_NatCheck(t *testing.T) {
	const serverAddr = "127.0.0.1:3478" // 需要填入服务器的公网 IP

	localAddr, err := net.ResolveUDPAddr("udp", ":6666") // 绑定本地随机端口
	if err != nil {
		fmt.Println("Error resolving local address:", err)
		return
	}

	conn, err := net.DialUDP("udp", localAddr, &net.UDPAddr{
		IP:   net.ParseIP(strings.Split(serverAddr, ":")[0]),
		Port: 3478,
	})
	if err != nil {
		fmt.Println("Error dialing server:", err)
		return
	}
	defer conn.Close()

	// 发送 NAT 检测请求
	message := "NAT Type Detection Request"
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	// 读取服务器的响应
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	externalAddr := strings.TrimSpace(string(buffer[:n]))
	fmt.Println("服务器检测到你的公网地址为:", externalAddr)

	// 解析 NAT 类型
	localPort := conn.LocalAddr().(*net.UDPAddr).Port
	externalIP, externalPortStr, _ := net.SplitHostPort(externalAddr)
	externalPort := 0
	fmt.Sscanf(externalPortStr, "%d", &externalPort)

	if externalIP == conn.LocalAddr().(*net.UDPAddr).IP.String() && localPort == externalPort {
		fmt.Println("NAT 类型: Full Cone NAT")
	} else if localPort == externalPort {
		fmt.Println("NAT 类型: Restricted Cone NAT 或 Port Restricted Cone NAT")
	} else {
		fmt.Println("NAT 类型: Symmetric NAT（端口已改变）")
	}
}
