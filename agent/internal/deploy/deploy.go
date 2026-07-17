package deploy

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Metadata 部署包元信息
type Metadata struct {
	Customer   string      `json:"customer"`
	Env        string      `json:"env"`
	Version    string      `json:"version"`
	CreatedAt  string      `json:"created_at"`
	CreatedBy  string      `json:"created_by"`
	Checksum   string      `json:"checksum"`
	Components []Component `json:"components"`
}

// Component 组件信息
type Component struct {
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	TemplateDir     string `json:"template_dir"`
	ArtifactType    string `json:"artifact_type"`
	ArtifactName    string `json:"artifact_name"`
	ArtifactVersion string `json:"artifact_version"`
	Image           string `json:"image"`
	ImageTar        string `json:"image_tar,omitempty"`
}

type State struct {
	Status          string `json:"status"`
	CurrentVersion  string `json:"current_version"`
	PreviousVersion string `json:"previous_version,omitempty"`
	CurrentPackage  string `json:"current_package,omitempty"`
	PreviousPackage string `json:"previous_package,omitempty"`
	UpdatedAt       string `json:"updated_at"`
	Error           string `json:"error,omitempty"`
}

// Importer 部署包导入器
type Importer struct {
	configDir string
	verbose   bool
}

// NewImporter 创建导入器
func NewImporter(configDir string, verbose bool) *Importer {
	return &Importer{configDir: configDir, verbose: verbose}
}

// Import 导入部署包
func (i *Importer) Import(packagePath string) (*Metadata, error) {
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("部署包不存在: %s", packagePath)
	}
	if err := os.MkdirAll(i.packagesDir(), 0755); err != nil {
		return nil, fmt.Errorf("创建部署包目录失败: %w", err)
	}

	version, err := readPackageVersion(packagePath)
	if err != nil {
		return nil, err
	}
	packageDir := filepath.Join(i.packagesDir(), version)
	if err := os.RemoveAll(packageDir); err != nil {
		return nil, fmt.Errorf("清理旧部署包目录失败: %w", err)
	}
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return nil, fmt.Errorf("创建部署包目录失败: %w", err)
	}

	if i.verbose {
		fmt.Printf("解压部署包: %s -> %s\n", packagePath, packageDir)
	}
	if err := extractTarGz(packagePath, packageDir, i.verbose); err != nil {
		return nil, fmt.Errorf("解压失败: %w", err)
	}

	meta, err := LoadMetadata(GetMetadataPath(packageDir))
	if err != nil {
		return nil, fmt.Errorf("读取元信息失败: %w", err)
	}
	if meta.Version != version {
		return nil, fmt.Errorf("部署包版本不一致: %s != %s", meta.Version, version)
	}

	state, _ := LoadState(i.configDir)
	state.PreviousVersion = state.CurrentVersion
	state.PreviousPackage = state.CurrentPackage
	state.CurrentVersion = meta.Version
	state.CurrentPackage = packageDir
	state.Status = "imported"
	state.Error = ""
	state.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := SaveState(i.configDir, state); err != nil {
		return nil, fmt.Errorf("写入状态失败: %w", err)
	}

	return meta, nil
}

func (i *Importer) packagesDir() string {
	return filepath.Join(i.configDir, "packages")
}

func readPackageVersion(packagePath string) (string, error) {
	tmp, err := os.MkdirTemp("", "itcfg-package-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	if err := extractSingleFile(packagePath, "metadata.json", filepath.Join(tmp, "metadata.json")); err != nil {
		return "", fmt.Errorf("读取部署包 metadata 失败: %w", err)
	}
	meta, err := LoadMetadata(filepath.Join(tmp, "metadata.json"))
	if err != nil {
		return "", err
	}
	if meta.Version == "" {
		return "", fmt.Errorf("metadata.json 缺少 version")
	}
	return meta.Version, nil
}

func extractSingleFile(packagePath, name, target string) error {
	f, err := os.Open(packagePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if filepath.ToSlash(header.Name) != name || header.Typeflag != tar.TypeReg {
			continue
		}
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, tr)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	}
	return fmt.Errorf("%s 不存在", name)
}

func extractTarGz(packagePath, targetDir string, verbose bool) error {
	f, err := os.Open(packagePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	base := filepath.Clean(targetDir) + string(os.PathSeparator)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := filepath.Clean(header.Name)
		if name == "." || filepath.IsAbs(name) || strings.HasPrefix(name, "..") {
			return fmt.Errorf("非法路径: %s", header.Name)
		}
		target := filepath.Join(targetDir, name)
		cleanTarget := filepath.Clean(target)
		if cleanTarget != filepath.Clean(targetDir) && !strings.HasPrefix(cleanTarget, base) {
			return fmt.Errorf("非法路径: %s", header.Name)
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(cleanTarget, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(cleanTarget), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(cleanTarget, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, tr)
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		default:
			return fmt.Errorf("不支持的 tar 条目类型: %s", header.Name)
		}
		if verbose {
			fmt.Printf("  解压: %s\n", header.Name)
		}
	}
	return nil
}

// GetMetadataPath 获取元信息文件路径
func GetMetadataPath(configDir string) string {
	return filepath.Join(configDir, "metadata.json")
}

// LoadMetadata 加载元信息
func LoadMetadata(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func LoadState(configDir string) (*State, error) {
	data, err := os.ReadFile(filepath.Join(configDir, "state.json"))
	if err != nil {
		return &State{}, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func SaveState(configDir string, state *State) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir, "state.json"), data, 0644)
}

func CurrentPackageDir(configDir string) (string, error) {
	state, err := LoadState(configDir)
	if err != nil || state.CurrentPackage == "" {
		return configDir, nil
	}
	return state.CurrentPackage, nil
}

// Deployer 部署执行器
type Deployer struct {
	configDir string
	dryRun    bool
	verbose   bool
}

// NewDeployer 创建部署执行器
func NewDeployer(configDir string, dryRun, verbose bool) *Deployer {
	return &Deployer{configDir: configDir, dryRun: dryRun, verbose: verbose}
}

// Deploy 执行部署
func (d *Deployer) Deploy() error {
	packageDir, err := CurrentPackageDir(d.configDir)
	if err != nil {
		return err
	}
	if _, err := os.Stat(GetMetadataPath(packageDir)); err != nil {
		return fmt.Errorf("未找到部署包元信息，请先执行 'config-agent import'")
	}

	meta, err := LoadMetadata(GetMetadataPath(packageDir))
	if err != nil {
		return err
	}
	if d.verbose {
		fmt.Printf("部署包: %s\n", packageDir)
		fmt.Printf("  客户: %s  环境: %s  版本: %s\n", meta.Customer, meta.Env, meta.Version)
	}

	if err := d.loadImages(packageDir); err != nil {
		d.markFailed(meta.Version, err)
		return fmt.Errorf("导入镜像失败: %w", err)
	}
	if err := d.StartServices(); err != nil {
		d.markFailed(meta.Version, err)
		return fmt.Errorf("启动服务失败: %w", err)
	}

	state, _ := LoadState(d.configDir)
	state.Status = "running"
	state.CurrentVersion = meta.Version
	state.CurrentPackage = packageDir
	state.Error = ""
	state.UpdatedAt = time.Now().Format(time.RFC3339)
	return SaveState(d.configDir, state)
}

func (d *Deployer) markFailed(version string, cause error) {
	state, _ := LoadState(d.configDir)
	state.Status = "failed"
	state.CurrentVersion = version
	state.Error = cause.Error()
	state.UpdatedAt = time.Now().Format(time.RFC3339)
	_ = SaveState(d.configDir, state)
}

func (d *Deployer) loadImages(packageDir string) error {
	imagesDir := filepath.Join(packageDir, "images")
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar") {
			continue
		}
		imagePath := filepath.Join(imagesDir, entry.Name())
		fmt.Printf("  导入镜像: %s\n", entry.Name())
		if d.dryRun {
			fmt.Printf("    [DRY-RUN] docker load -i %s\n", imagePath)
			continue
		}
		cmd := exec.Command("docker", "load", "-i", imagePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("docker load %s: %w", entry.Name(), err)
		}
	}
	return nil
}

// StartServices 启动服务
func (d *Deployer) StartServices() error {
	packageDir, err := CurrentPackageDir(d.configDir)
	if err != nil {
		return err
	}
	composeFile := filepath.Join(packageDir, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err != nil {
		return fmt.Errorf("docker-compose.yml 不存在: %w", err)
	}
	fmt.Println("  启动服务...")
	if d.dryRun {
		fmt.Printf("    [DRY-RUN] docker compose -f %s up -d\n", composeFile)
		return nil
	}
	cmd := dockerComposeCommand(composeFile, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// StopServices 停止服务
func (d *Deployer) StopServices() error {
	packageDir, err := CurrentPackageDir(d.configDir)
	if err != nil {
		return err
	}
	return d.stopPackage(packageDir)
}

func (d *Deployer) stopPackage(packageDir string) error {
	composeFile := filepath.Join(packageDir, "docker-compose.yml")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return nil
	}
	fmt.Println("  停止服务...")
	if d.dryRun {
		fmt.Printf("    [DRY-RUN] docker compose -f %s down\n", composeFile)
		return nil
	}
	cmd := dockerComposeCommand(composeFile, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d *Deployer) Rollback() error {
	state, err := LoadState(d.configDir)
	if err != nil {
		return fmt.Errorf("未找到部署状态")
	}
	if state.PreviousPackage == "" || state.PreviousVersion == "" {
		return fmt.Errorf("没有可回滚的上一版本")
	}
	currentPackage := state.CurrentPackage
	if currentPackage != "" {
		if err := d.stopPackage(currentPackage); err != nil {
			return err
		}
	}
	state.CurrentVersion, state.PreviousVersion = state.PreviousVersion, state.CurrentVersion
	state.CurrentPackage, state.PreviousPackage = state.PreviousPackage, state.CurrentPackage
	state.Status = "imported"
	state.Error = ""
	state.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := SaveState(d.configDir, state); err != nil {
		return err
	}
	return d.Deploy()
}

func ComposePS(configDir string) error {
	packageDir, err := CurrentPackageDir(configDir)
	if err != nil {
		return err
	}
	composeFile := filepath.Join(packageDir, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err != nil {
		return err
	}
	cmd := dockerComposeCommand(composeFile, "ps")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func dockerComposeCommand(composeFile string, args ...string) *exec.Cmd {
	fullArgs := append([]string{"compose", "-f", composeFile}, args...)
	cmd := exec.Command("docker", fullArgs...)
	if err := cmd.Err; err == nil {
		return cmd
	}
	return exec.Command("docker-compose", append([]string{"-f", composeFile}, args...)...)
}
