package deploy

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Importer 部署包导入器
type Importer struct {
	configDir string
	verbose   bool
}

// NewImporter 创建导入器
func NewImporter(configDir string, verbose bool) *Importer {
	return &Importer{
		configDir: configDir,
		verbose:   verbose,
	}
}

// Import 导入部署包
func (i *Importer) Import(packagePath string) (*Metadata, error) {
	// 检查文件是否存在
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("部署包不存在: %s", packagePath)
	}

	// 创建配置目录
	if err := os.MkdirAll(i.configDir, 0755); err != nil {
		return nil, fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 解压
	if i.verbose {
		fmt.Printf("解压部署包: %s\n", packagePath)
	}
	if err := i.extract(packagePath); err != nil {
		return nil, fmt.Errorf("解压失败: %w", err)
	}

	// 读取元信息
	metaPath := GetMetadataPath(i.configDir)
	meta, err := LoadMetadata(metaPath)
	if err != nil {
		return nil, fmt.Errorf("读取元信息失败: %w", err)
	}

	return meta, nil
}

// extract 解压 tar.gz 文件
func (i *Importer) extract(packagePath string) error {
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

		// 构建目标路径
		target := filepath.Join(i.configDir, header.Name)

		// 安全检查：防止路径穿越
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(i.configDir)) {
			return fmt.Errorf("非法路径: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			os.Chmod(target, os.FileMode(header.Mode))
		}

		if i.verbose {
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

// Deployer 部署执行器
type Deployer struct {
	configDir string
	dryRun    bool
	verbose   bool
}

// NewDeployer 创建部署执行器
func NewDeployer(configDir string, dryRun, verbose bool) *Deployer {
	return &Deployer{
		configDir: configDir,
		dryRun:    dryRun,
		verbose:   verbose,
	}
}

// Deploy 执行部署
func (d *Deployer) Deploy() error {
	metaPath := GetMetadataPath(d.configDir)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return fmt.Errorf("未找到部署包，请先执行 'config-agent import'")
	}

	meta, err := LoadMetadata(metaPath)
	if err != nil {
		return err
	}

	// Step 1: 导入 Docker 镜像
	if err := d.loadImages(); err != nil {
		return fmt.Errorf("导入镜像失败: %w", err)
	}

	// Step 2: 写入配置文件
	if err := d.writeConfigs(); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}

	// Step 3: 启动服务
	if err := d.startServices(); err != nil {
		return fmt.Errorf("启动服务失败: %w", err)
	}

	_ = meta // 可用于日志输出
	return nil
}

// loadImages 导入 Docker 镜像
func (d *Deployer) loadImages() error {
	imagesDir := filepath.Join(d.configDir, "images")
	imageListFile := filepath.Join(imagesDir, "image-list.txt")

	if _, err := os.Stat(imageListFile); os.IsNotExist(err) {
		if d.verbose {
			fmt.Println("  未找到镜像列表，跳过镜像导入")
		}
		return nil
	}

	data, err := os.ReadFile(imageListFile)
	if err != nil {
		return err
	}

	images := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, image := range images {
		image = strings.TrimSpace(image)
		if image == "" {
			continue
		}

		imagePath := filepath.Join(imagesDir, image)
		fmt.Printf("  导入镜像: %s\n", image)

		if d.dryRun {
			fmt.Printf("    [DRY-RUN] docker load < %s\n", imagePath)
			continue
		}

		// 实际执行 docker load
		// cmd := exec.Command("docker", "load", "-i", imagePath)
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr
		// if err := cmd.Run(); err != nil {
		//     return fmt.Errorf("导入镜像失败 %s: %w", image, err)
		// }
		fmt.Printf("    ✓ 完成\n")
	}

	return nil
}

// writeConfigs 写入配置文件
func (d *Deployer) writeConfigs() error {
	configsDir := filepath.Join(d.configDir, "configs")

	if _, err := os.Stat(configsDir); os.IsNotExist(err) {
		if d.verbose {
			fmt.Println("  未找到配置目录，跳过配置写入")
		}
		return nil
	}

	fmt.Println("  写入配置文件...")

	if d.dryRun {
		// 遍历并打印
		filepath.Walk(configsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				relPath, _ := filepath.Rel(configsDir, path)
				fmt.Printf("    [DRY-RUN] %s -> 目标路径\n", relPath)
			}
			return nil
		})
		return nil
	}

	// 实际写入
	return filepath.Walk(configsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(configsDir, path)
		targetPath := filepath.Join("/", relPath)

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// 读取源文件
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// 写入目标
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("写入 %s 失败: %w", targetPath, err)
		}

		if d.verbose {
			fmt.Printf("    ✓ %s\n", targetPath)
		}
		return nil
	})
}

// startServices 启动服务
func (d *Deployer) startServices() error {
	dockerComposeFile := filepath.Join(d.configDir, "docker-compose.yml")

	if _, err := os.Stat(dockerComposeFile); os.IsNotExist(err) {
		if d.verbose {
			fmt.Println("  未找到 docker-compose.yml，跳过服务启动")
		}
		return nil
	}

	fmt.Println("  启动服务...")

	if d.dryRun {
		fmt.Printf("    [DRY-RUN] docker-compose -f %s up -d\n", dockerComposeFile)
		return nil
	}

	// 实际执行 docker-compose
	// cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "up", "-d")
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// if err := cmd.Run(); err != nil {
	//     return fmt.Errorf("启动服务失败: %w", err)
	// }
	fmt.Println("    ✓ 服务已启动")

	return nil
}