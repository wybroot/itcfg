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
	"sort"
	"strings"
	"time"

	"itcfg/server/internal/model"
	"itcfg/server/internal/template"

	"github.com/google/uuid"
)

// PackageExporter 部署包导出服务
type PackageExporter struct {
	configSvc       *ConfigService
	componentSvc    *ComponentService
	envSvc          *EnvService
	customerSvc     *CustomerService
	envComponentSvc *EnvironmentComponentService
	artifactSvc     *ArtifactVersionService
	deployRecordSvc *DeployRecordService
	templateEngine  *template.Engine
	outputDir       string
}

// PackageMetadata 部署包元信息
type PackageMetadata struct {
	Customer   string             `json:"customer"`
	Env        string             `json:"env"`
	Version    string             `json:"version"`
	CreatedAt  string             `json:"created_at"`
	CreatedBy  string             `json:"created_by"`
	Checksum   string             `json:"checksum"`
	Components []PackageComponent `json:"components"`
}

// PackageComponent 部署包组件信息
type PackageComponent struct {
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	TemplateDir     string `json:"template_dir"`
	ArtifactType    string `json:"artifact_type"`
	ArtifactName    string `json:"artifact_name"`
	ArtifactVersion string `json:"artifact_version"`
	Image           string `json:"image"`
	ImageTar        string `json:"image_tar,omitempty"`
}

type packageManifest struct {
	Version    string                  `json:"version"`
	Customer   string                  `json:"customer"`
	Env        string                  `json:"env"`
	CreatedAt  string                  `json:"created_at"`
	Components []packageManifestComp   `json:"components"`
	Files      []packageManifestConfig `json:"files"`
	Images     []packageManifestImage  `json:"images"`
}

type packageManifestComp struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	TemplateDir string `json:"template_dir"`
	DeployOrder int    `json:"deploy_order"`
}

type packageManifestConfig struct {
	Component string `json:"component"`
	Path      string `json:"path"`
	Target    string `json:"target"`
	Owner     string `json:"owner"`
	Mode      string `json:"mode"`
}

type packageManifestImage struct {
	Component string `json:"component"`
	Image     string `json:"image"`
	Tar       string `json:"tar,omitempty"`
}

// NewPackageExporter 创建导出服务
func NewPackageExporter(
	configSvc *ConfigService,
	componentSvc *ComponentService,
	envSvc *EnvService,
	customerSvc *CustomerService,
	envComponentSvc *EnvironmentComponentService,
	artifactSvc *ArtifactVersionService,
	deployRecordSvc *DeployRecordService,
	templateEngine *template.Engine,
	outputDir string,
) *PackageExporter {
	return &PackageExporter{
		configSvc:       configSvc,
		componentSvc:    componentSvc,
		envSvc:          envSvc,
		customerSvc:     customerSvc,
		envComponentSvc: envComponentSvc,
		artifactSvc:     artifactSvc,
		deployRecordSvc: deployRecordSvc,
		templateEngine:  templateEngine,
		outputDir:       outputDir,
	}
}

// Export 导出完整部署包
func (e *PackageExporter) Export(envID string, createdBy string) (string, *PackageMetadata, error) {
	env, err := e.envSvc.GetByID(envID)
	if err != nil {
		return "", nil, fmt.Errorf("获取环境信息失败: %w", err)
	}

	customer, err := e.customerSvc.GetByID(env.CustomerID.String())
	if err != nil {
		return "", nil, fmt.Errorf("获取客户信息失败: %w", err)
	}

	envComponents, err := e.envComponentSvc.ListByEnv(envID)
	if err != nil {
		return "", nil, fmt.Errorf("获取环境组件失败: %w", err)
	}
	if len(envComponents) == 0 {
		return "", nil, fmt.Errorf("当前环境未启用任何组件")
	}

	configs, err := e.configSvc.GetByEnvDecrypted(envID)
	if err != nil {
		return "", nil, fmt.Errorf("获取配置值失败: %w", err)
	}
	configMap := make(map[string]string, len(configs))
	for _, cfg := range configs {
		configMap[cfg.VariableID.String()] = cfg.VarValue
	}

	artifacts, err := e.artifactSvc.ListByEnv(envID)
	if err != nil {
		return "", nil, fmt.Errorf("获取制品版本失败: %w", err)
	}
	artifactMap := make(map[string]model.ComponentArtifactVersion, len(artifacts))
	for _, artifact := range artifacts {
		artifactMap[artifact.ComponentID.String()] = artifact
	}

	version := fmt.Sprintf("v%s", time.Now().Format("20060102-150405"))
	workDir := filepath.Join(e.outputDir, fmt.Sprintf("export-%s", uuid.New().String()))
	defer os.RemoveAll(workDir)

	for _, dir := range []string{"configs", "images", "scripts"} {
		if err := os.MkdirAll(filepath.Join(workDir, dir), 0755); err != nil {
			return "", nil, fmt.Errorf("创建临时目录失败: %w", err)
		}
	}

	createdAt := time.Now().Format(time.RFC3339)
	meta := &PackageMetadata{
		Customer:   customer.Name,
		Env:        env.EnvName,
		Version:    version,
		CreatedAt:  createdAt,
		CreatedBy:  createdBy,
		Components: []PackageComponent{},
	}
	manifest := packageManifest{
		Version:    version,
		Customer:   customer.Name,
		Env:        env.EnvName,
		CreatedAt:  createdAt,
		Components: []packageManifestComp{},
		Files:      []packageManifestConfig{},
		Images:     []packageManifestImage{},
	}

	composeImages := map[string]string{}
	configSnapshot := map[string]string{}
	artifactSnapshot := map[string]string{}

	for _, envComponent := range envComponents {
		comp := envComponent.Component
		templateDir := comp.TemplateDir
		if templateDir == "" {
			templateDir = comp.Name
		}

		artifact, ok := artifactMap[comp.ID.String()]
		if !ok {
			return "", nil, fmt.Errorf("组件 %s 缺少制品版本", comp.DisplayName)
		}
		image := artifact.RegistryURL
		if image == "" {
			image = fmt.Sprintf("%s:%s", artifact.ArtifactName, artifact.ArtifactVersion)
		}
		composeImages[templateDir] = image
		artifactSnapshot[comp.Name] = image

		manifestDef, err := e.templateEngine.LoadManifest(templateDir)
		if err != nil {
			return "", nil, fmt.Errorf("读取组件模板 %s 失败: %w", templateDir, err)
		}
		variables, err := e.templateEngine.LoadVariables(templateDir)
		if err != nil {
			return "", nil, fmt.Errorf("读取组件变量 %s 失败: %w", templateDir, err)
		}

		renderValues := make(map[string]string, len(variables.Variables))
		for _, variable := range variables.Variables {
			value := variable.Default
			for _, dbVar := range comp.Variables {
				if dbVar.VarName == variable.Name {
					if configured, ok := configMap[dbVar.ID.String()]; ok && configured != "" {
						value = configured
					}
					configSnapshot[dbVar.VarName] = value
					break
				}
			}
			if variable.Required && strings.TrimSpace(value) == "" {
				return "", nil, fmt.Errorf("组件 %s 缺少必填配置 %s", comp.DisplayName, variable.Label)
			}
			renderValues[variable.Name] = value
		}

		compConfigDir := filepath.Join(workDir, "configs", templateDir)
		if err := os.MkdirAll(compConfigDir, 0755); err != nil {
			return "", nil, fmt.Errorf("创建组件配置目录失败: %w", err)
		}
		for _, cfg := range manifestDef.ConfigFiles {
			if !strings.HasSuffix(cfg.Path, ".tmpl") {
				continue
			}
			content, err := e.templateEngine.RenderFile(templateDir, cfg.Path, renderValues)
			if err != nil {
				return "", nil, fmt.Errorf("渲染组件 %s 配置 %s 失败: %w", comp.DisplayName, cfg.Path, err)
			}
			cleanPath := filepath.Clean(strings.TrimSuffix(cfg.Path, ".tmpl"))
			if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
				return "", nil, fmt.Errorf("模板输出路径非法: %s", cfg.Path)
			}
			filePath := filepath.Join(compConfigDir, cleanPath)
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return "", nil, fmt.Errorf("创建配置文件目录失败: %w", err)
			}
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return "", nil, fmt.Errorf("写入配置文件失败 %s: %w", filePath, err)
			}
			manifest.Files = append(manifest.Files, packageManifestConfig{
				Component: templateDir,
				Path:      filepath.ToSlash(filepath.Join("configs", templateDir, cleanPath)),
				Target:    cfg.Target,
				Owner:     cfg.Owner,
				Mode:      cfg.Mode,
			})
		}

		imageTar := copyImageTar(workDir, templateDir, artifact.RegistryURL)
		meta.Components = append(meta.Components, PackageComponent{
			Name:            comp.Name,
			DisplayName:     comp.DisplayName,
			TemplateDir:     templateDir,
			ArtifactType:    artifact.ArtifactType,
			ArtifactName:    artifact.ArtifactName,
			ArtifactVersion: artifact.ArtifactVersion,
			Image:           image,
			ImageTar:        imageTar,
		})
		manifest.Components = append(manifest.Components, packageManifestComp{
			Name:        comp.Name,
			DisplayName: comp.DisplayName,
			TemplateDir: templateDir,
			DeployOrder: envComponent.DeployOrder,
		})
		manifest.Images = append(manifest.Images, packageManifestImage{
			Component: templateDir,
			Image:     image,
			Tar:       imageTar,
		})
	}

	if err := os.WriteFile(filepath.Join(workDir, "docker-compose.yml"), []byte(generateCompose(composeImages)), 0644); err != nil {
		return "", nil, fmt.Errorf("写入 docker-compose.yml 失败: %w", err)
	}
	if err := writeExecutable(filepath.Join(workDir, "scripts", "deploy.sh"), GenerateInstallScript()); err != nil {
		return "", nil, err
	}
	if err := writeExecutable(filepath.Join(workDir, "scripts", "rollback.sh"), GenerateRollbackScript()); err != nil {
		return "", nil, err
	}
	if err := writeExecutable(filepath.Join(workDir, "scripts", "healthcheck.sh"), GenerateHealthcheckScript()); err != nil {
		return "", nil, err
	}

	if err := writeJSON(filepath.Join(workDir, "metadata.json"), meta); err != nil {
		return "", nil, err
	}
	if err := writeJSON(filepath.Join(workDir, "manifest.json"), manifest); err != nil {
		return "", nil, err
	}
	if err := writeChecksums(workDir); err != nil {
		return "", nil, fmt.Errorf("生成 checksums.txt 失败: %w", err)
	}

	packageName := fmt.Sprintf("%s-%s-%s.tar.gz", customer.Code, env.EnvKey, version)
	outputPath := filepath.Join(e.outputDir, packageName)
	if err := e.createTarGz(workDir, outputPath); err != nil {
		return "", nil, fmt.Errorf("打包失败: %w", err)
	}

	checksum, err := e.calculateChecksum(outputPath)
	if err != nil {
		return "", nil, fmt.Errorf("计算校验和失败: %w", err)
	}
	meta.Checksum = checksum

	configData, _ := json.Marshal(configSnapshot)
	artifactData, _ := json.Marshal(artifactSnapshot)
	_ = e.deployRecordSvc.Create(&model.DeployRecord{
		CustomerEnvID:    uuid.MustParse(envID),
		VersionTag:       version,
		ConfigSnapshot:   string(configData),
		ArtifactSnapshot: string(artifactData),
		PackageChecksum:  checksum,
		DeployedAt:       time.Now(),
		DeployedBy:       createdBy,
		Status:           "exported",
		Notes:            "部署包已导出",
	})

	return outputPath, meta, nil
}

func copyImageTar(workDir, component, source string) string {
	if source == "" {
		return ""
	}
	info, err := os.Stat(source)
	if err != nil || info.IsDir() {
		return ""
	}
	fileName := fmt.Sprintf("%s.tar", component)
	targetRel := filepath.ToSlash(filepath.Join("images", fileName))
	targetPath := filepath.Join(workDir, "images", fileName)
	in, err := os.Open(source)
	if err != nil {
		return ""
	}
	defer in.Close()
	out, err := os.Create(targetPath)
	if err != nil {
		return ""
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return ""
	}
	return targetRel
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func writeExecutable(path, content string) error {
	return os.WriteFile(path, []byte(content), 0755)
}

func writeChecksums(workDir string) error {
	var lines []string
	if err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Base(path) == "checksums.txt" {
			return nil
		}
		rel, err := filepath.Rel(workDir, path)
		if err != nil {
			return err
		}
		checksum, err := fileSHA256(path)
		if err != nil {
			return err
		}
		lines = append(lines, fmt.Sprintf("%s  %s", checksum, filepath.ToSlash(rel)))
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(lines)
	return os.WriteFile(filepath.Join(workDir, "checksums.txt"), []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

func fileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
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
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(tw, file)
		return err
	})
}

// calculateChecksum 计算文件 SHA256 校验和
func (e *PackageExporter) calculateChecksum(filePath string) (string, error) {
	checksum, err := fileSHA256(filePath)
	if err != nil {
		return "", err
	}
	return "sha256:" + checksum, nil
}

func generateCompose(images map[string]string) string {
	image := func(component, fallback string) string {
		if value := images[component]; value != "" {
			return value
		}
		return fallback
	}
	var b strings.Builder
	b.WriteString("version: '3.8'\n\n")
	b.WriteString("networks:\n  itcfg-network:\n    driver: bridge\n\n")
	b.WriteString("services:\n")
	if _, ok := images["postgresql"]; ok {
		b.WriteString(fmt.Sprintf(`  postgresql:
    image: %s
    container_name: itcfg-postgresql
    restart: unless-stopped
    ports:
      - "${POSTGRES_PORT:-5432}:5432"
    volumes:
      - ${DATA_BASE_DIR:-/data}/postgresql:/var/lib/postgresql/data
      - ./configs/postgresql/postgresql.conf:/etc/postgresql/postgresql.conf:ro
      - ./configs/postgresql/pg_hba.conf:/etc/postgresql/pg_hba.conf:ro
    networks:
      - itcfg-network
    environment:
      - TZ=${TIMEZONE:-Asia/Shanghai}
      - POSTGRES_DB=${POSTGRES_DB:-itcfg}
      - POSTGRES_USER=${POSTGRES_USER:-itcfg}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-itcfg123}
    command: postgres -c config_file=/etc/postgresql/postgresql.conf

`, image("postgresql", "postgres:16-alpine")))
	}
	if _, ok := images["java-app"]; ok {
		depends := ""
		if _, hasPostgres := images["postgresql"]; hasPostgres {
			depends = "    depends_on:\n      - postgresql\n"
		}
		b.WriteString(fmt.Sprintf(`  java-app:
    image: %s
    container_name: itcfg-java-app
    restart: unless-stopped
    ports:
      - "${JAVA_APP_PORT:-8080}:8080"
    volumes:
      - ./configs/java-app:/opt/java-app/config:ro
      - ${LOG_BASE_DIR:-/var/log}/java-app:/opt/java-app/logs
%s    networks:
      - itcfg-network
    environment:
      - TZ=${TIMEZONE:-Asia/Shanghai}
      - JAVA_OPTS=${JAVA_OPTS:--Xms512m -Xmx1g}

`, image("java-app", "java-app:latest"), depends))
	}
	if _, ok := images["nginx"]; ok {
		depends := ""
		if _, hasJava := images["java-app"]; hasJava {
			depends = "    depends_on:\n      - java-app\n"
		}
		b.WriteString(fmt.Sprintf(`  nginx:
    image: %s
    container_name: itcfg-nginx
    restart: unless-stopped
    ports:
      - "${NGINX_HTTP_PORT:-80}:80"
      - "${NGINX_HTTPS_PORT:-443}:443"
    volumes:
      - ./configs/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./configs/nginx/conf.d:/etc/nginx/conf.d:ro
      - ${DATA_BASE_DIR:-/data}/nginx/logs:/var/log/nginx
%s    networks:
      - itcfg-network
    environment:
      - TZ=${TIMEZONE:-Asia/Shanghai}

`, image("nginx", "nginx:1.25-alpine"), depends))
	}
	return b.String()
}

// GenerateInstallScript 生成安装脚本
func GenerateInstallScript() string {
	return `#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"

echo "[1/4] 校验部署包文件"
sha256sum -c checksums.txt

echo "[2/4] 导入 Docker 镜像"
if compgen -G "images/*.tar" > /dev/null; then
  for image in images/*.tar; do
    echo "  docker load -i $image"
    docker load -i "$image"
  done
else
  echo "  未发现本地镜像 tar，跳过导入"
fi

echo "[3/4] 启动服务"
docker compose -f docker-compose.yml up -d

echo "[4/4] 查看服务状态"
docker compose -f docker-compose.yml ps
`
}

func GenerateRollbackScript() string {
	return `#!/bin/bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"
docker compose -f docker-compose.yml down
echo "服务已停止，请使用上一版本部署包执行 scripts/deploy.sh 完成回滚"
`
}

func GenerateHealthcheckScript() string {
	return `#!/bin/bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"
docker compose -f docker-compose.yml ps
`
}

// GenerateStartScript 生成启动脚本
func GenerateStartScript() string {
	return `#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"
docker compose -f docker-compose.yml up -d
`
}

// GenerateStopScript 生成停止脚本
func GenerateStopScript() string {
	return `#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"
docker compose -f docker-compose.yml down
`
}

// GenerateUninstallScript 生成卸载脚本
func GenerateUninstallScript() string {
	return `#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"

echo "警告: 此操作将停止并删除所有服务容器和数据卷！"
read -p "确认卸载? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "已取消"
    exit 0
fi

docker compose -f docker-compose.yml down -v
`
}
