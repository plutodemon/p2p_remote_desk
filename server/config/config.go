package config

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"github.com/plutodemon/slog"
	"os"
	"path/filepath"
	"sync"
)

var (
	cfg  *Config
	once sync.Once
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
	LogConfig *slog.LogSetting `toml:"log"`

	// 环境配置
	Environment string `toml:"environment"` // development 或 production
}

// Init 初始化配置
func Init() error {
	return initConfig()
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	if cfg == nil {
		panic("配置未初始化")
	}
	return cfg
}

// IsDevelopment 判断是否为开发环境
func IsDevelopment() bool {
	return GetConfig().Environment == "development"
}

// initConfig 初始化配置
func initConfig() error {
	var err error
	once.Do(func() {
		cfg = &Config{}

		// 确保配置目录存在
		configDir := "config"
		if err = os.MkdirAll(configDir, 0755); err != nil {
			return
		}

		// 配置文件路径
		configFile := filepath.Join(configDir, "server.toml")

		// 读取配置文件
		data, err := os.ReadFile(configFile)
		if err != nil {
			err = fmt.Errorf("读取配置文件失败: %w", err)
			return
		}

		// 解析配置
		if err = toml.Unmarshal(data, cfg); err != nil {
			err = fmt.Errorf("解析配置文件失败: %w", err)
			return
		}
	})

	return err
}
