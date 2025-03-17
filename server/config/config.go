package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"p2p_remote_desk/llog"

	"github.com/pelletier/go-toml/v2"
)

// Config 服务器配置结构
type Config struct {
	// 环境配置
	Environment string `toml:"environment"` // development 或 production

	// 服务器基本配置
	Server *ServerConfig `toml:"server"`

	// 日志配置
	LogConfig *llog.LogSetting `toml:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host       string `toml:"host"`
	AuthPort   int    `toml:"auth_port"`
	SignalPort int    `toml:"signal_port"`
	IcePort    int    `toml:"ice_port"`

	MaxConnections    int   `toml:"max_connections"`     // 最大并发连接数
	IdleTimeout       int64 `toml:"idle_timeout"`        // 连接空闲超时时间
	GoroutinePoolSize int   `toml:"goroutine_pool_size"` // goroutine池大小
	MessageBufferSize int   `toml:"message_buffer_size"` // 消息缓冲区大小
	CleanupInterval   int64 `toml:"cleanup_interval"`    // 清理间隔

	ExpiryDuration   int64 `toml:"expiry_duration"`    // 空闲worker的过期时间
	MaxBlockingTasks int   `toml:"max_blocking_tasks"` // 最大阻塞任务数
}

var (
	config     *Config
	once       sync.Once
	configPath string
)

func init() {
	// 设置项目内部配置路径
	configPath = filepath.Join("config", "server.toml")
}

// Init 初始化配置
func Init() error {
	var err error
	once.Do(func() {
		err = initConfig()
	})
	return err
}

// initConfig 初始化配置
func initConfig() error {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置
	cfg := &Config{}
	if err = toml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	config = cfg

	return nil
}

// GetConfig 获取配置实例
func GetConfig() Config {
	if config == nil {
		panic("配置未初始化")
	}
	return *config
}

// IsDevelopment 判断是否为开发环境
func IsDevelopment() bool {
	return GetConfig().Environment == Development
}
