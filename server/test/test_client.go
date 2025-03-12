package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"p2p_remote_desk/server/internal/ice"
)

var (
	serverAddr = flag.String("server", "localhost:8080", "服务器地址")
	peerID     = flag.String("id", "", "本地peer ID")
	targetID   = flag.String("target", "", "目标peer ID")
)

func main() {
	flag.Parse()

	if *peerID == "" {
		fmt.Println("必须指定peer ID")
		os.Exit(1)
	}

	// 连接服务器
	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		fmt.Printf("连接服务器失败: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("已连接到服务器: %s\n", *serverAddr)

	// 获取本地地址信息
	localAddr := conn.LocalAddr().(*net.TCPAddr)

	// 发送注册消息
	registerMsg := ice.RegisterPayload{
		PeerID:    *peerID,
		LocalIP:   localAddr.IP.String(),
		LocalPort: localAddr.Port,
	}

	if err := ice.WriteMessage(conn, ice.MsgRegister, registerMsg); err != nil {
		fmt.Printf("发送注册消息失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已注册到服务器，ID: %s\n", *peerID)

	// 如果指定了目标ID，发送连接请求
	if *targetID != "" {
		connectMsg := ice.ConnectPayload{
			TargetID: *targetID,
		}

		if err := ice.WriteMessage(conn, ice.MsgConnect, connectMsg); err != nil {
			fmt.Printf("发送连接请求失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("已发送连接请求到: %s\n", *targetID)
	}

	// 启动UDP监听器用于P2P通信
	udpAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		fmt.Printf("创建UDP地址失败: %v\n", err)
		os.Exit(1)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Printf("创建UDP监听器失败: %v\n", err)
		os.Exit(1)
	}
	defer udpConn.Close()

	fmt.Printf("UDP监听器已启动: %s\n", udpConn.LocalAddr())

	// 处理来自服务器的消息
	go handleServerMessages(conn, udpConn)

	// 处理UDP消息
	handleUDPMessages(udpConn)
}

func handleServerMessages(conn net.Conn, udpConn *net.UDPConn) {
	for {
		msg, err := ice.ReadMessage(conn)
		if err != nil {
			fmt.Printf("读取服务器消息失败: %v\n", err)
			os.Exit(1)
		}

		switch msg.Type {
		case ice.MsgPunch:
			handlePunchMessage(msg, udpConn)
		case ice.MsgNATDetect:
			handleNATDetection(msg)
		default:
			fmt.Printf("未知的消息类型: %s\n", msg.Type)
		}
	}
}

func handlePunchMessage(msg *ice.Message, udpConn *net.UDPConn) {
	var payload ice.PunchPayload
	data, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(data, &payload); err != nil {
		fmt.Printf("解析打洞消息失败: %v\n", err)
		return
	}

	fmt.Printf("收到打洞消息，来自: %s\n", payload.PeerID)
	fmt.Printf("对方地址: 公网(%s:%d), 内网(%s:%d)\n",
		payload.PublicIP, payload.PublicPort,
		payload.LocalIP, payload.LocalPort)

	// 尝试UDP打洞
	go startUDPPunching(udpConn, payload)
}

func startUDPPunching(udpConn *net.UDPConn, target ice.PunchPayload) {
	// 首先尝试公网地址
	publicAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", target.PublicIP, target.PublicPort))
	if err != nil {
		fmt.Printf("解析公网地址失败: %v\n", err)
		return
	}

	// 然后尝试内网地址
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", target.LocalIP, target.LocalPort))
	if err != nil {
		fmt.Printf("解析内网地址失败: %v\n", err)
		return
	}

	// 开始打洞
	message := []byte("PUNCH:" + *peerID)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := 0; i < 10; i++ {
		// 发送到公网地址
		_, _ = udpConn.WriteToUDP(message, publicAddr)
		// 发送到内网地址
		_, _ = udpConn.WriteToUDP(message, localAddr)

		<-ticker.C
	}
}

func handleUDPMessages(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("读取UDP消息失败: %v\n", err)
			continue
		}

		message := string(buffer[:n])
		fmt.Printf("收到UDP消息: %s，来自: %s\n", message, remoteAddr)

		// 如果收到打洞消息，回复
		if len(message) > 6 && message[:6] == "PUNCH:" {
			response := []byte("PONG:" + *peerID)
			_, _ = conn.WriteToUDP(response, remoteAddr)
		}
	}
}

func handleNATDetection(msg *ice.Message) {
	var natInfo ice.NATInfo
	data, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(data, &natInfo); err != nil {
		fmt.Printf("解析NAT信息失败: %v\n", err)
		return
	}

	fmt.Printf("NAT类型: %s, 是否对称: %v\n", natInfo.Type, natInfo.Symmetric)
}
