package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// ComponentManifest 组件元信息
type ComponentManifest struct {
	Name         string       `yaml:"name" json:"name"`
	DisplayName  string       `yaml:"display_name" json:"display_name"`
	Description  string       `yaml:"description" json:"description"`
	Category     string       `yaml:"category" json:"category"`
	OutputDir    string       `yaml:"output_dir" json:"output_dir"`
	ConfigFiles  []ConfigFile `yaml:"config_files" json:"config_files"`
	Dependencies []Dep        `yaml:"dependencies" json:"dependencies"`
}

type ConfigFile struct {
	Path   string `yaml:"path" json:"path"`
	Target string `yaml:"target" json:"target"`
	Owner  string `yaml:"owner" json:"owner"`
	Mode   string `yaml:"mode" json:"mode"`
}

type Dep struct {
	Component string `yaml:"component" json:"component"`
	Reason    string `yaml:"reason" json:"reason"`
}

// VariableDefinition 变量定义
type VariableDefinition struct {
	Name         string   `yaml:"name" json:"name"`
	Label        string   `yaml:"label" json:"label"`
	Type         string   `yaml:"type" json:"type"`
	Default      string   `yaml:"default" json:"default"`
	Required     bool     `yaml:"required" json:"required"`
	Min          int      `yaml:"min" json:"min"`
	Max          int      `yaml:"max" json:"max"`
	Regex        string   `yaml:"regex" json:"regex"`
	Options      []string `yaml:"options" json:"options"`
	Group        string   `yaml:"group" json:"group"`
	Description  string   `yaml:"description" json:"description"`
	LinkedTo     string   `yaml:"linked_to" json:"linked_to"`
}

// VariablesFile 变量定义文件
type VariablesFile struct {
	Variables []VariableDefinition `yaml:"variables" json:"variables"`
}

// Engine 模板引擎
type Engine struct {
	baseDir string
}

// NewEngine 创建模板引擎
func NewEngine(baseDir string) *Engine {
	return &Engine{baseDir: baseDir}
}

// LoadManifest 加载组件元信息
func (e *Engine) LoadManifest(componentName string) (*ComponentManifest, error) {
	manifestPath := filepath.Join(e.baseDir, componentName, "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("读取 manifest 失败: %w", err)
	}
	var manifest ComponentManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("解析 manifest 失败: %w", err)
	}
	return &manifest, nil
}

// LoadVariables 加载变量定义
func (e *Engine) LoadVariables(componentName string) (*VariablesFile, error) {
	variablesPath := filepath.Join(e.baseDir, componentName, "variables.yaml")
	data, err := os.ReadFile(variablesPath)
	if err != nil {
		return nil, fmt.Errorf("读取 variables 失败: %w", err)
	}
	var vars VariablesFile
	if err := yaml.Unmarshal(data, &vars); err != nil {
		return nil, fmt.Errorf("解析 variables 失败: %w", err)
	}
	return &vars, nil
}

// RenderFile 渲染单个模板文件
func (e *Engine) RenderFile(componentName, templatePath string, values map[string]string) (string, error) {
	fullPath := filepath.Join(e.baseDir, componentName, "files", templatePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("读取模板文件失败 %s: %w", fullPath, err)
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(data))
	if err != nil {
		return "", fmt.Errorf("解析模板失败 %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, values); err != nil {
		return "", fmt.Errorf("渲染模板失败 %s: %w", templatePath, err)
	}

	return buf.String(), nil
}

// RenderAll 渲染组件的所有模板文件
func (e *Engine) RenderAll(componentName string, values map[string]string) (map[string]string, error) {
	manifest, err := e.LoadManifest(componentName)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, cf := range manifest.ConfigFiles {
		// 跳过静态文件（不需要模板渲染）
		if !strings.HasSuffix(cf.Path, ".tmpl") {
			continue
		}

		content, err := e.RenderFile(componentName, cf.Path, values)
		if err != nil {
			return nil, fmt.Errorf("渲染 %s 失败: %w", cf.Path, err)
		}

		// 输出文件名去掉 .tmpl 后缀
		outputPath := strings.TrimSuffix(cf.Target, ".tmpl")
		result[outputPath] = content
	}

	return result, nil
}

// BuildValues 构建渲染用的变量值 Map
func (e *Engine) BuildValues(componentName string, configValues map[string]string) (map[string]string, error) {
	variables, err := e.LoadVariables(componentName)
	if err != nil {
		return nil, err
	}

	values := make(map[string]string)
	for _, v := range variables.Variables {
		if val, ok := configValues[v.Name]; ok && val != "" {
			values[v.Name] = val
		} else {
			values[v.Name] = v.Default
		}
	}
	return values, nil
}