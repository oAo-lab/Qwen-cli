package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type ModelConfig struct {
	Name string `json:"name"`
}

type Config struct {
	APIURL string                 `json:"api_url"`
	APIKey string                 `json:"api_key"`
	Models map[string]ModelConfig `json:"models"`
	Roles  map[string]string      `json:"roles"`
}

// GetConfigDir 获取跨平台配置目录
func GetConfigDir() string {
	var configDir string
	
	switch runtime.GOOS {
	case "windows":
		// Windows: %USERPROFILE%\.config\ask
		userProfile := os.Getenv("USERPROFILE")
		if userProfile == "" {
			userProfile = os.Getenv("HOME")
		}
		configDir = filepath.Join(userProfile, ".config", "ask")
	case "darwin", "linux":
		// macOS/Linux: ~/.config/ask
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = os.Getenv("USERPROFILE") // 备用方案
		}
		configDir = filepath.Join(homeDir, ".config", "ask")
	default:
		// 其他系统，使用当前目录
		configDir = "."
	}
	
	return configDir
}

// GetConfigPath 获取配置文件完整路径
func GetConfigPath() string {
	configDir := GetConfigDir()
	return filepath.Join(configDir, "config.json")
}

// LoadConfig 加载配置文件，支持环境变量覆盖
func LoadConfig() (Config, error) {
	var config Config
	configPath := GetConfigPath()
	
	// 首先尝试从文件加载配置
	file, err := os.Open(configPath)
	if err != nil {
		// 如果文件不存在，使用默认配置
		if os.IsNotExist(err) {
			config, err = LoadDefaultConfig()
			if err != nil {
				return config, fmt.Errorf("failed to load default config: %w", err)
			}
		} else {
			return config, fmt.Errorf("failed to open config file: %w", err)
		}
	} else {
		defer file.Close()
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		if err != nil {
			return config, fmt.Errorf("failed to decode config file: %w", err)
		}
	}
	
	// 应用环境变量覆盖
	applyEnvOverrides(&config)
	
	return config, nil
}

// LoadDefaultConfig 从嵌入的默认配置加载
func LoadDefaultConfig() (Config, error) {
	var config Config
	err := json.Unmarshal(defaultConfigJSON, &config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal default config: %w", err)
	}
	return config, nil
}

// applyEnvOverrides 应用环境变量覆盖配置
func applyEnvOverrides(config *Config) {
	if apiURL := os.Getenv("ASK_API_URL"); apiURL != "" {
		config.APIURL = apiURL
	}
	
	if apiKey := os.Getenv("ASK_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}
}

// SaveConfig 保存配置到文件
func SaveConfig(config Config) error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)
	
	// 确保配置目录存在
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(config)
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}
	
	return nil
}

// InitConfig 初始化配置文件
func InitConfig() error {
	configPath := GetConfigPath()
	
	// 检查配置文件是否已存在
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}
	
	// 加载默认配置
	config, err := LoadDefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
	}
	
	// 保存配置文件
	err = SaveConfig(config)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	return nil
}
