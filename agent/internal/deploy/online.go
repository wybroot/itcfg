package deploy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AgentConfig Agent 配置
type AgentConfig struct {
	ServerURL string `json:"server_url"`
	EnvKey    string `json:"env_key"`
	ConfigDir string `json:"config_dir"`
}

// AuthResult 认证结果
type AuthResult struct {
	EnvID       string `json:"env_id"`
	EnvName     string `json:"env_name"`
	CustomerID  string `json:"customer_id"`
	AuthSuccess bool   `json:"authenticated"`
}

// OnlineClient 在线客户端
type OnlineClient struct {
	serverURL string
	envKey    string
	verbose   bool
	client    *http.Client
}

// NewOnlineClient 创建在线客户端
func NewOnlineClient(serverURL, envKey string, verbose bool) *OnlineClient {
	return &OnlineClient{
		serverURL: strings.TrimRight(serverURL, "/"),
		envKey:    envKey,
		verbose:   verbose,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Auth 认证
func (c *OnlineClient) Auth() (*AuthResult, error) {
	url := c.serverURL + "/api/v1/agent/auth"
	body := fmt.Sprintf(`{"env_key":"%s"}`, c.envKey)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("认证失败 (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data AuthResult `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result.Data, nil
}

// PullConfigs 拉取配置
func (c *OnlineClient) PullConfigs() (map[string]map[string]string, error) {
	url := fmt.Sprintf("%s/api/v1/agent/envs/%s/configs", c.serverURL, c.envKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Env-Key", c.envKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("拉取失败 (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			EnvID   string                        `json:"env_id"`
			EnvName string                        `json:"env_name"`
			Configs map[string]map[string]string  `json:"configs"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if c.verbose {
		fmt.Printf("  环境: %s\n", result.Data.EnvName)
		componentCount := len(result.Data.Configs)
		fmt.Printf("  组件数: %d\n", componentCount)
	}

	return result.Data.Configs, nil
}

// SaveConfig 保存 Agent 配置
func SaveConfig(cfg *AgentConfig) error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".itcfg", "agent.json")
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

// LoadConfig 加载 Agent 配置
func LoadConfig() (*AgentConfig, error) {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".itcfg", "agent.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg AgentConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// WriteConfigs 写入配置文件
func WriteConfigs(configsDir string, configs map[string]map[string]string) error {
	for compName, files := range configs {
		compDir := filepath.Join(configsDir, compName)
		if err := os.MkdirAll(compDir, 0755); err != nil {
			return err
		}

		for path, content := range files {
			fileName := filepath.Base(path)
			filePath := filepath.Join(compDir, fileName)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("写入 %s/%s 失败: %w", compName, fileName, err)
			}
		}
	}
	return nil
}