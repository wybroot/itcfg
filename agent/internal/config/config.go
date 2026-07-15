package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AgentConfig Agent 配置
type AgentConfig struct {
	ServerURL string `json:"server_url"`
	EnvKey    string `json:"env_key"`
	ConfigDir string `json:"config_dir"`
}

// ConfigFile Agent 配置文件路径
func ConfigFile() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".itcfg", "agent.json")
}

// Load 加载配置
func Load() (*AgentConfig, error) {
	path := ConfigFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &AgentConfig{}, nil
		}
		return nil, err
	}

	var cfg AgentConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	return &cfg, nil
}

// Save 保存配置
func Save(cfg *AgentConfig) error {
	path := ConfigFile()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}