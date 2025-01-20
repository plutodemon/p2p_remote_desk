package config

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"github.com/plutodemon/slog"
	"os"
	"path/filepath"
	"sync"
)

// RemoteConfig 配置结构体
type RemoteConfig struct {
	Version string `toml:"version"`

	// 环境设置
	Environment string `toml:"environment"`

	// 服务器设置
	ServerConfig *Server `toml:"server"`

	// 屏幕设置
	ScreenConfig *Screen `toml:"screen"`

	// UI设置
	UIConfig *UI `toml:"ui"`

	// 性能设置
	PerformanceConfig *Performance `toml:"performance"`

	// 日志设置
	LogConfig *slog.LogSetting `toml:"log"`
}

type Server struct {
	Address string `toml:"address"`
	Port    string `toml:"port"`
}

type Screen struct {
	DefaultQuality   string            `toml:"default_quality"`
	DefaultFrameRate int               `toml:"default_frame_rate"`
	FrameRates       []int             `toml:"frame_rates"`
	QualityList      []*QualitySetting `toml:"quality"`
}

type QualitySetting struct {
	Name        string  `toml:"name"`
	Scale       float64 `toml:"scale"`
	Compression int     `toml:"compression"`
}

type UI struct {
	Theme                   string      `toml:"theme"`
	Language                string      `toml:"language"`
	HideToolbarInFullscreen bool        `toml:"hide_toolbar_in_fullscreen"`
	ShowPerformanceStats    bool        `toml:"show_performance_stats"`
	Development             *EnvSetting `toml:"development"`
	Production              *EnvSetting `toml:"production"`
}

type EnvSetting struct {
	ShowToolbar      bool `toml:"show_toolbar"`
	ShowFPS          bool `toml:"show_fps"`
	ShowStatus       bool `toml:"show_status"`
	ShowRemoteScreen bool `toml:"show_remote_screen"`
	AllowModeSwitch  bool `toml:"allow_mode_switch"`
}

type Performance struct {
	MaxGoroutines       int  `toml:"max_goroutines"`
	FrameBufferSize     int  `toml:"frame_buffer_size"`
	SkipIdenticalFrames bool `toml:"skip_identical_frames"`
}

type LogSetting struct {
	Console    bool   `toml:"console"`
	File       bool   `toml:"file"`
	FilePath   string `toml:"file_path"`
	MaxSize    int    `toml:"max_size"`
	MaxAge     int    `toml:"max_age"`
	MaxBackups int    `toml:"max_backups"`
	Compress   bool   `toml:"compress"`
	LocalTime  bool   `toml:"local_time"`
	Format     string `toml:"format"`
}

var (
	Config     *RemoteConfig
	configLock sync.RWMutex
	configPath string
)

func init() {
	// 设置项目内部配置路径
	configPath = filepath.Join("config", "config.toml")
}

// Init 初始化配置
func Init() error {
	return Load()
}

// Load 加载配置文件
func Load() error {
	configLock.Lock()
	defer configLock.Unlock()

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置
	cfg := &RemoteConfig{}
	if err = toml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	Config = cfg
	return nil
}

// Save 保存配置到文件
func Save() error {
	configLock.Lock()
	defer configLock.Unlock()

	if Config == nil {
		return fmt.Errorf("配置未初始化")
	}

	// 将配置转换为TOML格式
	data, err := toml.Marshal(Config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	return nil
}

// GetConfig 获取配置
func GetConfig() *RemoteConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return Config
}

// SetConfig 设置配置
func SetConfig(cfg *RemoteConfig) {
	configLock.Lock()
	defer configLock.Unlock()
	Config = cfg
}

// IsDevelopment 判断是否为开发环境
func IsDevelopment() bool {
	cfg := GetConfig()
	if cfg == nil {
		return false
	}
	return cfg.Environment == "development"
}

// createDefaultConfig 创建默认配置
func createDefaultConfig() error {
	// 将默认配置写入文件
	//data, err := toml.Marshal(defaultConfig)
	//if err != nil {
	//	return fmt.Errorf("序列化默认配置失败: %v", err)
	//}
	//
	//if err := os.WriteFile(configPath, data, 0644); err != nil {
	//	return fmt.Errorf("写入默认配置失败: %v", err)
	//}
	//
	//Config = defaultConfig
	return nil
}
