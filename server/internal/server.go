package internal

import (
	"fmt"
	"github.com/plutodemon/llog"
	"net"
	"p2p_remote_desk/server/config"
	"sync"
)

// Server 服务器结构体
type Server struct {
	listener net.Listener
	clients  map[string]*Client
	mutex    sync.RWMutex
	running  bool
}

// NewServer 创建新的服务器实例
func NewServer() *Server {
	return &Server{
		clients: make(map[string]*Client),
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	cfg := config.GetConfig()
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// 创建监听器
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("创建监听器失败: %w", err)
	}
	s.listener = listener

	llog.Info("服务器启动成功，监听地址: %s", addr)
	s.running = true

	// 开始接受连接
	go s.acceptConnections()

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	if !s.running {
		return nil
	}

	// 关闭监听器
	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("关闭监听器失败: %w", err)
	}

	// 断开所有客户端连接
	s.mutex.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clients = make(map[string]*Client)
	s.mutex.Unlock()

	s.running = false
	llog.Info("服务器已停止")
	return nil
}

// acceptConnections 接受新的连接
func (s *Server) acceptConnections() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				llog.Error("接受连接失败: %v", err)
			}
			continue
		}

		// 创建新的客户端连接
		client := NewClient(conn, s)
		s.addClient(client)

		// 启动客户端处理
		go client.Handle()
	}
}

// addClient 添加客户端
func (s *Server) addClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.clients[client.ID] = client
	llog.Info("新客户端连接: %s", client.ID)
}

// removeClient 移除客户端
func (s *Server) removeClient(clientID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if client, exists := s.clients[clientID]; exists {
		client.Close()
		delete(s.clients, clientID)
		llog.Info("客户端断开连接: %s", clientID)
	}
}

// GetClientCount 获取当前连接的客户端数量
func (s *Server) GetClientCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.clients)
}

// BroadcastMessage 广播消息给所有客户端
func (s *Server) BroadcastMessage(message []byte) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, client := range s.clients {
		client.SendMessage(message)
	}
}
