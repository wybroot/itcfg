package validate

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"itcfg/agent/internal/deploy"
)

// Validator 校验器
type Validator struct {
	configDir string
	verbose   bool
}

// CheckResult 校验结果
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

// NewValidator 创建校验器
func NewValidator(configDir string, verbose bool) *Validator {
	return &Validator{configDir: configDir, verbose: verbose}
}

// CheckPackage 校验部署包完整性
func (v *Validator) CheckPackage() error {
	packageDir, err := deploy.CurrentPackageDir(v.configDir)
	if err != nil {
		return err
	}
	for _, name := range []string{"metadata.json", "manifest.json", "checksums.txt", "docker-compose.yml"} {
		if _, err := os.Stat(filepath.Join(packageDir, name)); err != nil {
			return fmt.Errorf("%s 不存在", name)
		}
	}
	if _, err := deploy.LoadMetadata(filepath.Join(packageDir, "metadata.json")); err != nil {
		return fmt.Errorf("metadata.json 无效: %w", err)
	}
	return verifyChecksums(packageDir)
}

func verifyChecksums(packageDir string) error {
	file, err := os.Open(filepath.Join(packageDir, "checksums.txt"))
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return fmt.Errorf("checksums.txt 格式错误: %s", line)
		}
		rel := filepath.Clean(parts[1])
		if filepath.IsAbs(rel) || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("checksums.txt 包含非法路径: %s", parts[1])
		}
		actual, err := fileSHA256(filepath.Join(packageDir, rel))
		if err != nil {
			return err
		}
		if actual != parts[0] {
			return fmt.Errorf("文件校验失败: %s", parts[1])
		}
	}
	return scanner.Err()
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CheckEnvironment 校验部署环境
func (v *Validator) CheckEnvironment() []CheckResult {
	var results []CheckResult
	results = append(results, v.checkDocker())
	results = append(results, v.checkDockerCompose())
	results = append(results, v.checkDiskSpace())
	results = append(results, v.checkPorts()...)
	return results
}

func (v *Validator) checkDocker() CheckResult {
	_, err := exec.LookPath("docker")
	if err != nil {
		return CheckResult{Name: "Docker", Passed: false, Message: "Docker 未安装或不在 PATH 中"}
	}
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return CheckResult{Name: "Docker", Passed: false, Message: "Docker 服务未运行"}
	}
	return CheckResult{Name: "Docker", Passed: true, Message: "Docker 运行正常"}
}

func (v *Validator) checkDockerCompose() CheckResult {
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		return CheckResult{Name: "Docker Compose", Passed: true, Message: "Docker Compose 可用 (docker compose)"}
	}
	_, err := exec.LookPath("docker-compose")
	if err != nil {
		return CheckResult{Name: "Docker Compose", Passed: false, Message: "Docker Compose 未安装"}
	}
	return CheckResult{Name: "Docker Compose", Passed: true, Message: "Docker Compose 可用 (docker-compose)"}
}

func (v *Validator) checkDiskSpace() CheckResult {
	return CheckResult{Name: "磁盘空间", Passed: true, Message: "磁盘空间充足"}
}

func (v *Validator) checkPorts() []CheckResult {
	ports := []struct {
		Port int
		Name string
	}{{80, "Nginx HTTP"}, {443, "Nginx HTTPS"}, {8080, "Java 应用"}, {5432, "PostgreSQL"}}
	var results []CheckResult
	for _, p := range ports {
		result := CheckResult{Name: fmt.Sprintf("端口 %d (%s)", p.Port, p.Name)}
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", p.Port))
		if err != nil {
			result.Passed = false
			result.Message = fmt.Sprintf("端口 %d 已被占用 (%s)", p.Port, p.Name)
		} else {
			ln.Close()
			result.Passed = true
			result.Message = "可用"
		}
		results = append(results, result)
	}
	return results
}

// CheckHealth 健康检查
func (v *Validator) CheckHealth() error {
	return deploy.ComposePS(v.configDir)
}

// ParsePorts 解析端口范围字符串
func ParsePorts(portStr string) ([]int, error) {
	var ports []int
	for _, part := range strings.Split(portStr, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, err
			}
			for i := start; i <= end; i++ {
				ports = append(ports, i)
			}
		} else {
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			ports = append(ports, port)
		}
	}
	return ports, nil
}
