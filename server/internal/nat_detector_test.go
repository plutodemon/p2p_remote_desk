package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
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

func Test_SignalingServer(t *testing.T) {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("Signaling server starting on :28080")
	log.Fatal(http.ListenAndServe(":28080", nil))
}

type Room struct {
	Clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	rooms  = make(map[string]*Room)
	roomMu sync.Mutex
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	log.Println("New WebSocket connection from:", remoteAddr)

	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		roomID = fmt.Sprintf("room_%d", len(rooms)+1)
		log.Printf("Created new room: %s\n", roomID)
	}

	roomMu.Lock()
	room, exists := rooms[roomID]
	if !exists {
		room = &Room{Clients: make(map[*websocket.Conn]bool)}
		rooms[roomID] = room
	}
	roomMu.Unlock()

	room.mu.Lock()
	room.Clients[conn] = true
	room.mu.Unlock()

	log.Printf("Client[%v] joined room %s\n", remoteAddr, roomID)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Error unmarshaling message:", err)
			continue
		}

		msg["roomId"] = roomID
		updatedMessage, _ := json.Marshal(msg)

		room.mu.Lock()
		for client := range room.Clients {
			if client != conn {
				clientAddr := client.RemoteAddr().String()
				if err := client.WriteMessage(messageType, updatedMessage); err != nil {
					log.Println("Error writing message:", err)
				} else {
					log.Printf("writing message to client[%v] ok\n", clientAddr)
				}
			}
		}
		room.mu.Unlock()
	}

	room.mu.Lock()
	delete(room.Clients, conn)
	room.mu.Unlock()
	log.Printf("Client[%v] left room %s\n", remoteAddr, roomID)
}
