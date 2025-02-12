package internal

import (
	"fmt"
	"github.com/plutodemon/llog"
	"net"
	"p2p_remote_desk/server/config"
	"sync"
	"time"
)

// PeerInfo 存储对等端信息
type PeerInfo struct {
	ID         string
	PublicIP   string
	PublicPort int
	LocalIP    string
	LocalPort  int
	NATType    string
	Conn       net.Conn
	LastSeen   time.Time
}

// Server P2P中继服务器
type Server struct {
	listener    net.Listener
	peers       map[string]*PeerInfo // 存储所有连接的peer信息
	mutex       sync.RWMutex
	running     bool
	done        chan struct{}
	natDetector *NATDetector
}

// NewServer 创建新的服务器实例
func NewServer() *Server {
	// 创建NAT检测器
	detector, err := NewNATDetector(
		"stun1.l.google.com", 19302, // 主检测服务器
		"stun2.l.google.com", 19302, // 备用检测服务器
	)
	if err != nil {
		llog.Error("创建NAT检测器失败: ", err)
		detector = nil
	}

	return &Server{
		peers:       make(map[string]*PeerInfo),
		done:        make(chan struct{}),
		natDetector: detector,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	cfg := config.GetConfig()
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// 创建TCP监听器
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("创建监听器失败: %w", err)
	}
	s.listener = listener

	llog.Info("P2P中继服务器启动成功, 监听地址: ", addr)
	s.running = true

	// 启动清理过期连接的goroutine
	go s.cleanupExpiredPeers()

	// 开始接受连接
	go s.acceptConnections()

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	if !s.running {
		return nil
	}

	s.running = false
	close(s.done)

	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("关闭监听器失败: %w", err)
	}

	s.mutex.Lock()
	for _, peer := range s.peers {
		_ = peer.Conn.Close()
	}
	s.peers = make(map[string]*PeerInfo)
	s.mutex.Unlock()

	llog.Info("服务器已停止")
	return nil
}

// cleanupExpiredPeers 清理过期的peer连接
func (s *Server) cleanupExpiredPeers() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mutex.Lock()
			now := time.Now()
			for id, peer := range s.peers {
				// 如果peer超过5分钟没有活动，认为已断开
				if now.Sub(peer.LastSeen) > 5*time.Minute {
					llog.Info("清理过期peer: ", id)
					_ = peer.Conn.Close()
					delete(s.peers, id)
				}
			}
			s.mutex.Unlock()
		}
	}
}

// acceptConnections 接受新的连接
func (s *Server) acceptConnections() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				llog.Error("接受连接失败: ", err)
			}
			continue
		}

		// 设置连接超时
		_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

		// 启动新的goroutine处理连接
		go s.handleConnection(conn)
	}
}

// handleConnection 处理新的连接
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			llog.Error("处理连接时发生panic: ", r)
		}
		conn.Close()
	}()

	// 等待客户端注册
	msg, err := ReadMessage(conn)
	if err != nil {
		llog.Error("读取注册消息失败: ", err)
		return
	}

	if msg.Type != MsgRegister {
		llog.Error("首条消息必须是注册消息")
		return
	}

	// 解析注册信息
	var payload RegisterPayload
	if err := ParsePayload(msg, &payload); err != nil {
		llog.Error("解析注册消息失败: ", err)
		return
	}

	// 获取远程地址信息
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)

	// 创建peer信息
	peer := &PeerInfo{
		ID:         payload.PeerID,
		PublicIP:   remoteAddr.IP.String(),
		PublicPort: remoteAddr.Port,
		LocalIP:    payload.LocalIP,
		LocalPort:  payload.LocalPort,
		Conn:       conn,
		LastSeen:   time.Now(),
	}

	// 注册peer
	s.registerPeer(peer)
	defer s.unregisterPeer(peer.ID)

	// 处理连接
	s.handlePeerMessages(peer)
}

// registerPeer 注册新的peer
func (s *Server) registerPeer(peer *PeerInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 如果已存在同ID的peer，先关闭旧连接
	if oldPeer, exists := s.peers[peer.ID]; exists {
		_ = oldPeer.Conn.Close()
	}

	s.peers[peer.ID] = peer
	llog.Info("新peer注册:  (公网: :%d, 内网: :%d)",
		peer.ID, peer.PublicIP, peer.PublicPort, peer.LocalIP, peer.LocalPort)
}

// unregisterPeer 注销peer
func (s *Server) unregisterPeer(peerID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if peer, exists := s.peers[peerID]; exists {
		_ = peer.Conn.Close()
		delete(s.peers, peerID)
		llog.Info("peer注销: ", peerID)
	}
}

// handlePeerMessages 处理peer消息
func (s *Server) handlePeerMessages(peer *PeerInfo) {
	for {
		// 更新最后活动时间
		_ = peer.Conn.SetDeadline(time.Now().Add(30 * time.Second))

		msg, err := ReadMessage(peer.Conn)
		if err != nil {
			llog.Error("读取消息失败: ", err)
			return
		}

		// 更新最后活动时间
		peer.LastSeen = time.Now()

		switch msg.Type {
		case MsgConnect:
			s.handleConnectRequest(peer, msg)
		case MsgNATDetect:
			s.handleNATDetection(peer)
		default:
			llog.Warn("未知的消息类型: ", msg.Type)
		}
	}
}

// handleConnectRequest 处理连接请求
func (s *Server) handleConnectRequest(peer *PeerInfo, msg *Message) {
	var payload ConnectPayload
	if err := ParsePayload(msg, &payload); err != nil {
		llog.Error("解析连接请求失败: ", err)
		return
	}

	// 查找目标peer
	s.mutex.RLock()
	targetPeer, exists := s.peers[payload.TargetID]
	s.mutex.RUnlock()

	if !exists {
		llog.Error("目标peer不存在: ", payload.TargetID)
		return
	}

	// 向双方发送打洞消息
	punchMsg := PunchPayload{
		PeerID:     peer.ID,
		PublicIP:   peer.PublicIP,
		PublicPort: peer.PublicPort,
		LocalIP:    peer.LocalIP,
		LocalPort:  peer.LocalPort,
	}

	// 发送给目标peer
	if err := WriteMessage(targetPeer.Conn, MsgPunch, punchMsg); err != nil {
		llog.Error("发送打洞消息失败: ", err)
		return
	}

	// 发送给请求peer
	punchMsg.PeerID = targetPeer.ID
	punchMsg.PublicIP = targetPeer.PublicIP
	punchMsg.PublicPort = targetPeer.PublicPort
	punchMsg.LocalIP = targetPeer.LocalIP
	punchMsg.LocalPort = targetPeer.LocalPort

	if err := WriteMessage(peer.Conn, MsgPunch, punchMsg); err != nil {
		llog.Error("发送打洞消息失败: ", err)
		return
	}

	llog.Info("已发送打洞消息:  <-> ", peer.ID, targetPeer.ID)
}

// handleNATDetection 处理NAT类型检测
func (s *Server) handleNATDetection(peer *PeerInfo) {
	if s.natDetector == nil {
		natInfo := NATInfo{
			Type:      string(NATUnknown),
			Symmetric: true, // 默认假设是对称NAT
		}
		if err := WriteMessage(peer.Conn, MsgNATDetect, natInfo); err != nil {
			llog.Error("发送NAT检测结果失败: ", err)
		}
		return
	}

	// 创建UDP连接用于检测
	udpAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		llog.Error("创建UDP地址失败: ", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		llog.Error("创建UDP连接失败: ", err)
		return
	}
	defer udpConn.Close()

	// 执行NAT类型检测
	natType, err := s.natDetector.DetectNATType(udpConn)
	if err != nil {
		llog.Error("NAT类型检测失败: ", err)
		natType = NATUnknown
	}

	// 更新peer的NAT类型
	peer.NATType = string(natType)

	// 发送检测结果
	natInfo := NATInfo{
		Type:      string(natType),
		Symmetric: natType == NATSymmetric,
	}

	if err := WriteMessage(peer.Conn, MsgNATDetect, natInfo); err != nil {
		llog.Error("发送NAT检测结果失败: ", err)
	}

	llog.Info("Peer  的NAT类型: ", peer.ID, natType)
}
