package notify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"itcfg/server/internal/model"
	"itcfg/server/internal/repository"
)

// Service 通知服务
type Service struct {
	repo *repository.NotifyConfigRepo
}

// NewService 创建通知服务
func NewService(repo *repository.NotifyConfigRepo) *Service {
	return &Service{repo: repo}
}

// List 通知配置列表
func (s *Service) List() ([]model.NotifyConfig, error) {
	return s.repo.List()
}

// GetByID 获取通知配置
func (s *Service) GetByID(id string) (*model.NotifyConfig, error) {
	return s.repo.GetByID(id)
}

// Create 创建通知配置
func (s *Service) Create(cfg *model.NotifyConfig) error {
	return s.repo.Create(cfg)
}

// Update 更新通知配置
func (s *Service) Update(id string, cfg *model.NotifyConfig) error {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	existing.Name = cfg.Name
	existing.Type = cfg.Type
	existing.WebhookURL = cfg.WebhookURL
	if cfg.Secret != "" {
		existing.Secret = cfg.Secret
	}
	existing.Events = cfg.Events
	existing.IsActive = cfg.IsActive
	return s.repo.Update(existing)
}

// Delete 删除通知配置
func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}

// Send 根据事件类型发送通知
func (s *Service) Send(event string, data map[string]interface{}) error {
	configs, err := s.repo.ListActive()
	if err != nil {
		return err
	}

	for _, cfg := range configs {
		if !matchEvent(cfg.Events, event) {
			continue
		}
		switch cfg.Type {
		case "dingtalk":
			s.sendDingTalk(cfg, event, data)
		case "wecom":
			s.sendWeCom(cfg, event, data)
		case "webhook":
			s.sendWebhook(cfg, event, data)
		}
	}
	return nil
}

// ==================== 钉钉 ====================

type dingTalkMessage struct {
	MsgType  string             `json:"msgtype"`
	Markdown dingTalkMarkdown   `json:"markdown"`
	At       dingTalkAt         `json:"at,omitempty"`
}

type dingTalkMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type dingTalkAt struct {
	IsAtAll bool `json:"isAtAll"`
}

func (s *Service) sendDingTalk(cfg model.NotifyConfig, event string, data map[string]interface{}) {
	title := notificationTitle(event, data)
	text := buildMarkdown(event, data)

	msg := dingTalkMessage{
		MsgType: "markdown",
		Markdown: dingTalkMarkdown{
			Title: title,
			Text:  text,
		},
	}

	body, _ := json.Marshal(msg)
	url := cfg.WebhookURL
	if cfg.Secret != "" {
		url = addDingTalkSign(url, cfg.Secret)
	}
	http.Post(url, "application/json", bytes.NewReader(body))
}

// ==================== 企业微信 ====================

type weComMessage struct {
	MsgType  string        `json:"msgtype"`
	Markdown weComMarkdown `json:"markdown"`
}

type weComMarkdown struct {
	Content string `json:"content"`
}

func (s *Service) sendWeCom(cfg model.NotifyConfig, event string, data map[string]interface{}) {
	msg := weComMessage{
		MsgType: "markdown",
		Markdown: weComMarkdown{
			Content: buildMarkdown(event, data),
		},
	}
	body, _ := json.Marshal(msg)
	http.Post(cfg.WebhookURL, "application/json", bytes.NewReader(body))
}

// ==================== 通用 Webhook ====================

func (s *Service) sendWebhook(cfg model.NotifyConfig, event string, data map[string]interface{}) {
	payload := map[string]interface{}{
		"event":     event,
		"timestamp": time.Now().Format(time.RFC3339),
		"data":      data,
	}
	body, _ := json.Marshal(payload)
	http.Post(cfg.WebhookURL, "application/json", bytes.NewReader(body))
}

// ==================== 工具函数 ====================

func matchEvent(events, event string) bool {
	for _, e := range strings.Split(events, ",") {
		if strings.TrimSpace(e) == event {
			return true
		}
	}
	return false
}

func notificationTitle(event string, data map[string]interface{}) string {
	switch event {
	case "deploy_success":
		return fmt.Sprintf("✅ 部署成功 - %v", data["customer"])
	case "deploy_failed":
		return fmt.Sprintf("❌ 部署失败 - %v", data["customer"])
	case "config_updated":
		return fmt.Sprintf("📝 配置更新 - %v", data["customer"])
	default:
		return fmt.Sprintf("ITCFG 通知: %s", event)
	}
}

func buildMarkdown(event string, data map[string]interface{}) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("### %s\n\n", notificationTitle(event, data)))
	buf.WriteString("| 字段 | 值 |\n|------|------|\n")
	for k, v := range data {
		buf.WriteString(fmt.Sprintf("| %s | %v |\n", k, v))
	}
	buf.WriteString(fmt.Sprintf("\n> 时间: %s", time.Now().Format("2006-01-02 15:04:05")))
	return buf.String()
}

func addDingTalkSign(url, secret string) string {
	timestamp := time.Now().UnixMilli()
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%stimestamp=%d&sign=%s", url, sep, timestamp, sign)
}

// ==================== 便捷发送方法 ====================

// SendDeploySuccess 发送部署成功通知
func (s *Service) SendDeploySuccess(customer, env, version, operator string) {
	s.Send("deploy_success", map[string]interface{}{
		"客户":   customer,
		"环境":   env,
		"版本":   version,
		"操作人": operator,
	})
}

// SendDeployFailed 发送部署失败通知
func (s *Service) SendDeployFailed(customer, env, version, operator, reason string) {
	s.Send("deploy_failed", map[string]interface{}{
		"客户":   customer,
		"环境":   env,
		"版本":   version,
		"操作人": operator,
		"失败原因": reason,
	})
}

// SendConfigUpdated 发送配置更新通知
func (s *Service) SendConfigUpdated(customer, env, operator, summary string) {
	s.Send("config_updated", map[string]interface{}{
		"客户":   customer,
		"环境":   env,
		"操作人": operator,
		"变更说明": summary,
	})
}

// SendTestMessage 发送测试消息（验证配置）
func (s *Service) SendTestMessage(webhookURL, notifyType, secret string) error {
	cfg := model.NotifyConfig{
		WebhookURL: webhookURL,
		Type:       notifyType,
		Secret:     secret,
		Events:     "test",
	}
	switch notifyType {
	case "dingtalk":
		resp, err := s.sendDingTalkRaw(cfg, "🧪 ITCFG 通知测试", "通知配置验证成功，您将收到后续的部署和配置变更通知。")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("发送失败 (%d): %s", resp.StatusCode, string(body))
		}
	case "wecom":
		resp, err := s.sendWeComRaw(cfg, "🧪 ITCFG 通知测试\n\n通知配置验证成功")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("发送失败 (%d): %s", resp.StatusCode, string(body))
		}
	case "webhook":
		resp, err := s.sendWebhookRaw(cfg, map[string]string{"event": "test", "message": "通知配置验证成功"})
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("发送失败 (%d): %s", resp.StatusCode, string(body))
		}
	}
	return nil
}

func (s *Service) sendDingTalkRaw(cfg model.NotifyConfig, title, text string) (*http.Response, error) {
	msg := dingTalkMessage{
		MsgType: "markdown",
		Markdown: dingTalkMarkdown{Title: title, Text: text},
	}
	body, _ := json.Marshal(msg)
	url := cfg.WebhookURL
	if cfg.Secret != "" {
		url = addDingTalkSign(url, cfg.Secret)
	}
	return http.Post(url, "application/json", bytes.NewReader(body))
}

func (s *Service) sendWeComRaw(cfg model.NotifyConfig, content string) (*http.Response, error) {
	msg := weComMessage{
		MsgType: "markdown",
		Markdown: weComMarkdown{Content: content},
	}
	body, _ := json.Marshal(msg)
	return http.Post(cfg.WebhookURL, "application/json", bytes.NewReader(body))
}

func (s *Service) sendWebhookRaw(cfg model.NotifyConfig, payload interface{}) (*http.Response, error) {
	body, _ := json.Marshal(payload)
	return http.Post(cfg.WebhookURL, "application/json", bytes.NewReader(body))
}
