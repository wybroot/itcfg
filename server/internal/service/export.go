package service

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"itcfg/server/internal/model"
	"itcfg/server/internal/template"

	"github.com/google/uuid"
)

// PackageExporter 部署包导出服务
type PackageExporter struct {
	configSvc      *ConfigService
	componentSvc   *ComponentService
	envSvc         *EnvService
	customerSvc    *CustomerService
	templateEngine *template.Engine
	outputDir      string
}

// PackageMetadata 部署包元信息
type PackageMetadata struct {
	Customer    string              `json:"customer"`
	Env         string              `json:"env"`
	Version     string              `json:"version"`
	CreatedAt   string              `json:"created_at"`
	CreatedBy   string              `json:"created_by"`
	Checksum    string              `json:"checksum"`
	Components  []PackageComponent  `json:"components"`
}

// PackageComponent 部署包组件信息
type PackageComponent struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NewPackageExporter 创建导出服务
func NewPackageExporter(
	configSvc *ConfigService,
	componentSvc *ComponentService,
	envSvc *EnvService,
	customerSvc *CustomerService,
	templateEngine *template.Engine,
	outputDir string,
) *PackageExporter {
	return &PackageExporter{
		configSvc:      configSvc,
		componentSvc:   componentSvc,
		envSvc:         envSvc,
		customerSvc:    customerSvc,
		templateEngine: templateEngine,
		outputDir:      outputDir,
	}
}

// Export 导出完整部署包
func (e *PackageExporter) Export(envID string, createdBy string) (string, *PackageMetadata, error) {
	// 1. 获取环境信息
	env, err := e.envSvc.repo.GetByID(envID)
	if err != nil {
		return "", nil, fmt.Errorf("获取环境信息失败: %w", err)
	}

	// 2. 获取客户信息
	customer, err := e.customerSvc.GetByID(env.CustomerID.String())
	if err != nil {
		return "", nil, fmt.Errorf("获取客户信息失败: %w", err)
	}

	// 3. 获取所有组件
	components, err := e.componentSvc.List()
	if err != nil {
		return "", nil, fmt.Errorf("获取组件列表失败: %w", err)
	}

	// 4. 获取所有配置值
	configs, err := e.configSvc.GetByEnv(envID)
	if err != nil {
		return "", nil, fmt.Errorf("获取配置值失败: %w", err)
	}

	// 构建变量ID到值的映射
	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.VariableID.String()] = cfg.VarValue
	}

	// 5. 生成版本号
	version := fmt.Sprintf("v%s", time.Now().Format("20060102-150405"))

	// 6. 创建临时工作目录
	workDir := filepath.Join(e.outputDir, fmt.Sprintf("export-%s", uuid.New().String()))
	defer os.RemoveAll(workDir)

	configsDir := filepath.Join(workDir, "configs")
	if err := os.MkdirAll(configsDir, 0755); err != nil {
		return "", nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 7. 渲染每个组件的配置文件
	meta := &PackageMetadata{
		Customer:   customer.Name,
		Env:        env.EnvName,
		Version:    version,
		CreatedAt:  time.Now().Format(time.RFC3339),
		CreatedBy:  createdBy,
		Components: []PackageComponent{},
	}

	for _, comp := range components {
		if !comp.IsActive {
			continue
		}

		// 获取组件变量定义
		variables, err := e.templateEngine.LoadVariables(comp.Name)
		if err != nil {
			// 跳过没有模板的组件
			continue
		}

		// 构建渲染值
		renderValues := make(map[string]string)
		for _, v := range variables.Variables {
			// 通过变量ID在configMap中查找值
			// 需要找到对应变量的ID
			for _, varModel := range comp.Variables {
				if varModel.VarName == v.Name {
					if val, ok := configMap[varModel.ID.String()]; ok && val != "" {
						renderValues[v.Name] = val
					} else {
						renderValues[v.Name] = v.Default
					}
					break
				}
			}
			// fallback: 使用默认值
			if _, ok := renderValues[v.Name]; !ok {
				renderValues[v.Name] = v.Default
			}
		}

		// 渲染所有模板文件
		rendered, err := e.templateEngine.RenderAll(comp.Name, renderValues)
		if err != nil {
			continue // 跳过渲染失败的组件
		}

		// 写入配置文件
		compConfigDir := filepath.Join(configsDir, comp.Name)
		if err := os.MkdirAll(compConfigDir, 0755); err != nil {
			continue
		}

		for path, content := range rendered {
			// 生成输出路径
			fileName := filepath.Base(path)
			filePath := filepath.Join(compConfigDir, fileName)
			// 确保子目录存在
			if dir := filepath.Dir(filePath); dir != compConfigDir {
				os.MkdirAll(dir, 0755)
			}
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return "", nil, fmt.Errorf("写入配置文件失败 %s: %w", filePath, err)
			}
		}

		meta.Components = append(meta.Components, PackageComponent{
			Name:    comp.Name,
			Version: comp.Name, // 后续可从制品版本表获取
		})
	}

	// 8. 写入 metadata.json
	metaPath := filepath.Join(workDir, "metadata.json")
	metaData, _ := json.MarshalIndent(meta, "", "  ")
	os.WriteFile(metaPath, metaData, 0644)

	// 9. 打包为 tar.gz
	packageName := fmt.Sprintf("%s-%s-%s.tar.gz",
		customer.Code, env.EnvKey, version)
	outputPath := filepath.Join(e.outputDir, packageName)

	if err := e.createTarGz(workDir, outputPath); err != nil {
		return "", nil, fmt.Errorf("打包失败: %w", err)
	}

	// 10. 计算校验和
	checksum, err := e.calculateChecksum(outputPath)
	if err != nil {
		return "", nil, fmt.Errorf("计算校验和失败: %w", err)
	}
	meta.Checksum = checksum

	// 更新 metadata
	metaData, _ = json.MarshalIndent(meta, "", "  ")
	os.WriteFile(metaPath, metaData, 0644)

	return outputPath, meta, nil
}

// createTarGz 创建 tar.gz 压缩包
func (e *PackageExporter) createTarGz(sourceDir, targetFile string) error {
	f, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// 创建 tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// 写入 header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// 如果是文件，写入内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// calculateChecksum 计算文件 SHA256 校验和
func (e *PackageExporter) calculateChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

// GenerateInstallScript 生成安装脚本
func GenerateInstallScript() string {
	return `#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=========================================="
echo "  ITCFG 一键部署脚本"
echo "=========================================="

# 读取元信息
if [ -f "$SCRIPT_DIR/metadata.json" ]; then
    echo "客户: $(grep -o '"customer": *"[^"]*"' "$SCRIPT_DIR/metadata.json" | cut -d'"' -f4)"
    echo "环境: $(grep -o '"env": *"[^"]*"' "$SCRIPT_DIR/metadata.json" | cut -d'"' -f4)"
    echo "版本: $(grep -o '"version": *"[^"]*"' "$SCRIPT_DIR/metadata.json" | cut -d'"' -f4)"
fi

# 1. 导入 Docker 镜像
if [ -d "$SCRIPT_DIR/images" ] && [ -f "$SCRIPT_DIR/images/image-list.txt" ]; then
    echo "[1/5] 导入 Docker 镜像..."
    while IFS= read -r image_file; do
        [ -z "$image_file" ] && continue
        echo "  加载: $image_file"
        docker load < "$SCRIPT_DIR/images/$image_file"
    done < "$SCRIPT_DIR/images/image-list.txt"
else
    echo "[1/5] 跳过镜像导入 (无镜像目录)"
fi

# 2. 写入配置文件
echo "[2/5] 写入配置文件..."
if [ -d "$SCRIPT_DIR/configs" ]; then
    cp -r "$SCRIPT_DIR/configs"/* /opt/itcfg/configs/ 2>/dev/null || true
    echo "  配置文件已写入 /opt/itcfg/configs/"
fi

# 3. 校验环境
echo "[3/5] 校验部署环境..."
if command -v docker &> /dev/null; then
    echo "  Docker: ✓"
else
    echo "  Docker: ✗ 未安装"
    exit 1
fi

# 4. 启动服务
echo "[4/5] 启动服务..."
if [ -f "$SCRIPT_DIR/docker-compose.yml" ]; then
    cd "$SCRIPT_DIR"
    docker-compose -f docker-compose.yml up -d
    echo "  服务已启动"
else
    echo "  未找到 docker-compose.yml，跳过"
fi

# 5. 等待服务就绪
echo "[5/5] 等待服务就绪..."
sleep 10

echo "=========================================="
echo "  部署完成！"
echo "=========================================="
`
}

// GenerateStartScript 生成启动脚本
func GenerateStartScript() string {
	return `#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"
docker-compose -f docker-compose.yml up -d
echo "所有服务已启动"
`
}

// GenerateStopScript 生成停止脚本
func GenerateStopScript() string {
	return `#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"
docker-compose -f docker-compose.yml down
echo "所有服务已停止"
`
}

// GenerateUninstallScript 生成卸载脚本
func GenerateUninstallScript() string {
	return `#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "警告: 此操作将停止并删除所有服务容器和数据卷！"
read -p "确认卸载? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "已取消"
    exit 0
fi

docker-compose -f docker-compose.yml down -v
echo "服务已卸载"
`
}

// 确保导入 model 包
var _ = model.Customer{}
var _ = strings.TrimSpace