package config

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"github.com/plutodemon/llog"
	"os"
	"path/filepath"
	"sync"
)

// Config 服务器配置结构
type Config struct {
	// 服务器基本配置
	Server struct {
		Host string `toml:"host"`
		Port int    `toml:"port"`
	} `toml:"server"`

	// 性能相关配置
	Performance struct {
		MaxConnections int `toml:"max_connections"`
		BufferSize     int `toml:"buffer_size"`
	} `toml:"performance"`

	// 日志配置
	LogConfig *llog.LogSetting `toml:"log"`

	// 环境配置
	Environment string `toml:"environment"` // development 或 production
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
