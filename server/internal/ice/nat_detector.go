package ice

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

// NATType NAT类型
type NATType string

const (
	NATUnknown        NATType = "Unknown"
	NATNone           NATType = "None" // 公网IP
	NATFullCone       NATType = "Full Cone"
	NATRestrictedCone NATType = "Restricted Cone"
	NATPortRestricted NATType = "Port Restricted"
	NATSymmetric      NATType = "Symmetric"
)

// NATDetector NAT检测器
type NATDetector struct {
	mainAddr      *net.UDPAddr // 主检测地址
	alternateAddr *net.UDPAddr // 备用检测地址
}

// NewNATDetector 创建NAT检测器
func NewNATDetector(mainIP string, mainPort int, alternateIP string, alternatePort int) (*NATDetector, error) {
	mainAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", mainIP, mainPort))
	if err != nil {
		return nil, fmt.Errorf("解析主检测地址失败: %w", err)
	}

	alternateAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", alternateIP, alternatePort))
	if err != nil {
		return nil, fmt.Errorf("解析备用检测地址失败: %w", err)
	}

	return &NATDetector{
		mainAddr:      mainAddr,
		alternateAddr: alternateAddr,
	}, nil
}

// DetectNATType 检测NAT类型
func (d *NATDetector) DetectNATType(conn *net.UDPConn) (NATType, error) {
	// 步骤1: 检测是否是公网IP
	if isPublicIP(conn.LocalAddr().(*net.UDPAddr).IP) {
		return NATNone, nil
	}

	// 步骤2: 发送测试1 - 检测基本连通性
	mappedAddr1, err := d.test1(conn)
	if err != nil {
		return NATUnknown, fmt.Errorf("测试1失败: %w", err)
	}

	// 步骤3: 发送测试2 - 使用相同IP不同端口
	mappedAddr2, err := d.test2(conn)
	if err != nil {
		return NATUnknown, fmt.Errorf("测试2失败: %w", err)
	}

	// 步骤4: 发送测试3 - 使用不同IP不同端口
	mappedAddr3, err := d.test3(conn)
	if err != nil {
		return NATUnknown, fmt.Errorf("测试3失败: %w", err)
	}

	// 分析结果
	if mappedAddr1.Port != mappedAddr2.Port {
		return NATSymmetric, nil
	}

	if mappedAddr1.String() != mappedAddr3.String() {
		return NATSymmetric, nil
	}

	// 步骤5: 发送测试4 - 检测限制类型
	isPortRestricted, err := d.test4(conn, mappedAddr1)
	if err != nil {
		return NATUnknown, fmt.Errorf("测试4失败: %w", err)
	}

	if isPortRestricted {
		return NATPortRestricted, nil
	}

	return NATFullCone, nil
}

// test1 测试1: 基本连通性测试
func (d *NATDetector) test1(conn *net.UDPConn) (*net.UDPAddr, error) {
	return d.sendTest(conn, d.mainAddr, "test1")
}

// test2 测试2: 相同IP不同端口测试
func (d *NATDetector) test2(conn *net.UDPConn) (*net.UDPAddr, error) {
	altPort := &net.UDPAddr{
		IP:   d.mainAddr.IP,
		Port: d.mainAddr.Port + 1,
	}
	return d.sendTest(conn, altPort, "test2")
}

// test3 测试3: 不同IP不同端口测试
func (d *NATDetector) test3(conn *net.UDPConn) (*net.UDPAddr, error) {
	return d.sendTest(conn, d.alternateAddr, "test3")
}

// test4 测试4: 限制类型测试
func (d *NATDetector) test4(conn *net.UDPConn, mappedAddr *net.UDPAddr) (bool, error) {
	// 发送测试消息
	testMsg := []byte("test4")
	if _, err := conn.WriteToUDP(testMsg, d.alternateAddr); err != nil {
		return false, err
	}

	// 设置读取超时
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	// 尝试接收响应
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		if isTimeout(err) {
			return true, nil // 超时表示是端口限制型
		}
		return false, err
	}

	return string(buffer[:n]) != "test4", nil
}

// sendTest 发送测试消息并等待响应
func (d *NATDetector) sendTest(conn *net.UDPConn, addr *net.UDPAddr, msg string) (*net.UDPAddr, error) {
	// 发送测试消息
	if _, err := conn.WriteToUDP([]byte(msg), addr); err != nil {
		return nil, err
	}

	// 设置读取超时
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	// 接收响应
	buffer := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}

	if string(buffer[:n]) != msg {
		return nil, fmt.Errorf("收到未知响应: %s", string(buffer[:n]))
	}

	return remoteAddr, nil
}

// isPublicIP 检查是否是公网IP
func isPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}

	// 检查私有IP范围
	privateIPRanges := []struct {
		start net.IP
		end   net.IP
	}{
		{
			net.ParseIP("10.0.0.0"),
			net.ParseIP("10.255.255.255"),
		},
		{
			net.ParseIP("172.16.0.0"),
			net.ParseIP("172.31.255.255"),
		},
		{
			net.ParseIP("192.168.0.0"),
			net.ParseIP("192.168.255.255"),
		},
	}

	for _, r := range privateIPRanges {
		if bytes.Compare(ip, r.start) >= 0 && bytes.Compare(ip, r.end) <= 0 {
			return false
		}
	}

	return true
}

// isTimeout 检查错误是否是超时
func isTimeout(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	return false
}
