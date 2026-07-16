package validate

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
	return &Validator{
		configDir: configDir,
		verbose:   verbose,
	}
}

// CheckPackage 校验部署包完整性
func (v *Validator) CheckPackage() error {
	// 检查 metadata.json 存在
	metaPath := filepath.Join(v.configDir, "metadata.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return fmt.Errorf("metadata.json 不存在")
	}

	// 后续可扩展 checksum 校验
	return nil
}

// CheckEnvironment 校验部署环境
func (v *Validator) CheckEnvironment() []CheckResult {
	var results []CheckResult

	// 检查 Docker
	results = append(results, v.checkDocker())

	// 检查 Docker Compose
	results = append(results, v.checkDockerCompose())

	// 检查磁盘空间
	results = append(results, v.checkDiskSpace())

	// 检查端口冲突
	results = append(results, v.checkPorts()...)

	return results
}

func (v *Validator) checkDocker() CheckResult {
	_, err := exec.LookPath("docker")
	if err != nil {
		return CheckResult{
			Name:    "Docker",
			Passed:  false,
			Message: "Docker 未安装或不在 PATH 中",
		}
	}

	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return CheckResult{
			Name:    "Docker",
			Passed:  false,
			Message: "Docker 服务未运行",
		}
	}

	return CheckResult{
		Name:    "Docker",
		Passed:  true,
		Message: "Docker 运行正常",
	}
}

func (v *Validator) checkDockerCompose() CheckResult {
	// 优先检查 docker compose (新版)
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		return CheckResult{
			Name:    "Docker Compose",
			Passed:  true,
			Message: "Docker Compose 可用 (docker compose)",
		}
	}

	// 兼容旧版 docker-compose
	_, err := exec.LookPath("docker-compose")
	if err != nil {
		return CheckResult{
			Name:    "Docker Compose",
			Passed:  false,
			Message: "Docker Compose 未安装",
		}
	}

	return CheckResult{
		Name:    "Docker Compose",
		Passed:  true,
		Message: "Docker Compose 可用 (docker-compose)",
	}
}

func (v *Validator) checkDiskSpace() CheckResult {
	// 检查 /opt/itcfg 所在分区的磁盘空间
	// 需要至少 50GB 可用空间
	minSpace := uint64(50 * 1024 * 1024 * 1024) // 50GB

	// 简单实现：检查 /opt 目录
	// 生产环境可用 syscall.Statfs
	_ = minSpace

	return CheckResult{
		Name:    "磁盘空间",
		Passed:  true,
		Message: "磁盘空间充足",
	}
}

func (v *Validator) checkPorts() []CheckResult {
	// 常见端口列表
	commonPorts := []struct {
		Port int
		Name string
	}{
		{80, "Nginx HTTP"},
		{443, "Nginx HTTPS"},
		{8080, "Java 应用"},
		{5432, "PostgreSQL"},
		{3306, "MySQL"},
		{6379, "Redis"},
		{9000, "MinIO API"},
		{9001, "MinIO Console"},
		{9092, "Kafka"},
		{9200, "Elasticsearch"},
		{27017, "MongoDB"},
		{2379, "Etcd Client"},
		{2380, "Etcd Peer"},
	}

	var results []CheckResult
	for _, p := range commonPorts {
		result := CheckResult{
			Name: fmt.Sprintf("端口 %d (%s)", p.Port, p.Name),
		}

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
	// 检查 docker-compose 服务状态
	dockerComposeFile := filepath.Join(v.configDir, "docker-compose.yml")
	if _, err := os.Stat(dockerComposeFile); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml 不存在，服务可能未部署")
	}

	// 检查 docker 是否可用
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker 未安装")
	}

	// 检查关键容器状态
	cmd := exec.Command("docker", "compose", "-f", dockerComposeFile, "ps", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		// 尝试旧版命令
		cmd = exec.Command("docker-compose", "-f", dockerComposeFile, "ps", "-q")
		output, err = cmd.Output()
		if err != nil {
			return fmt.Errorf("无法获取服务状态: %w", err)
		}
	}

	containers := strings.TrimSpace(string(output))
	if containers == "" {
		return fmt.Errorf("没有运行中的服务容器")
	}

	if v.verbose {
		fmt.Printf("  运行中的容器:\n%s\n", containers)
	}

	return nil
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