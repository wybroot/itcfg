package validate

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ConfigValidator 配置值校验器
type ConfigValidator struct{}

// NewConfigValidator 创建校验器
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// Rule 校验规则
type Rule struct {
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	Required  bool     `json:"required,omitempty"`
}

// Validate 根据变量类型和校验规则验证配置值
func (v *ConfigValidator) Validate(varType, value, validationRule string) error {
	// 空值校验
	if strings.TrimSpace(value) == "" {
		return nil // 空值不校验，由 required 字段控制
	}

	var rule Rule
	if validationRule != "" {
		if err := json.Unmarshal([]byte(validationRule), &rule); err != nil {
			// 如果规则解析失败，跳过校验
			return nil
		}
	}

	switch varType {
	case "string":
		return v.validateString(value, rule)
	case "number":
		return v.validateNumber(value, rule)
	case "password":
		return v.validatePassword(value, rule)
	case "boolean":
		return v.validateBoolean(value, rule)
	case "select":
		return v.validateSelect(value, rule)
	case "url":
		return v.validateURL(value)
	case "ip":
		return v.validateIP(value)
	case "port":
		return v.validatePort(value)
	}
	return nil
}

func (v *ConfigValidator) validateString(value string, rule Rule) error {
	if rule.MinLength != nil && len(value) < *rule.MinLength {
		return fmt.Errorf("字符串长度不能小于 %d", *rule.MinLength)
	}
	if rule.MaxLength != nil && len(value) > *rule.MaxLength {
		return fmt.Errorf("字符串长度不能大于 %d", *rule.MaxLength)
	}
	if rule.Pattern != "" {
		// 简单的包含匹配
		if !strings.Contains(value, rule.Pattern) {
			return fmt.Errorf("值不匹配要求的模式: %s", rule.Pattern)
		}
	}
	return nil
}

func (v *ConfigValidator) validateNumber(value string, rule Rule) error {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("不是有效的数字: %s", value)
	}
	if rule.Min != nil && num < *rule.Min {
		return fmt.Errorf("值不能小于 %v", *rule.Min)
	}
	if rule.Max != nil && num > *rule.Max {
		return fmt.Errorf("值不能大于 %v", *rule.Max)
	}
	return nil
}

func (v *ConfigValidator) validatePassword(value string, rule Rule) error {
	if rule.MinLength != nil && len(value) < *rule.MinLength {
		return fmt.Errorf("密码长度不能小于 %d 位", *rule.MinLength)
	}
	return nil
}

func (v *ConfigValidator) validateBoolean(value string, rule Rule) error {
	val := strings.ToLower(value)
	if val != "true" && val != "false" {
		return fmt.Errorf("布尔值必须为 true 或 false")
	}
	return nil
}

func (v *ConfigValidator) validateSelect(value string, rule Rule) error {
	if len(rule.Enum) == 0 {
		return nil
	}
	for _, opt := range rule.Enum {
		if opt == value {
			return nil
		}
	}
	return fmt.Errorf("值必须为以下之一: %s", strings.Join(rule.Enum, ", "))
}

func (v *ConfigValidator) validateURL(value string) error {
	if _, err := url.ParseRequestURI(value); err != nil {
		return fmt.Errorf("无效的 URL: %s", value)
	}
	return nil
}

func (v *ConfigValidator) validateIP(value string) error {
	if net.ParseIP(value) == nil {
		return fmt.Errorf("无效的 IP 地址: %s", value)
	}
	return nil
}

func (v *ConfigValidator) validatePort(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("无效的端口号: %s", value)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("端口号必须在 1-65535 之间")
	}
	return nil
}
