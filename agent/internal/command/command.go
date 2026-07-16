package command

import (
	"fmt"
	"os"
	"time"

	"itcfg/agent/internal/config"
	"itcfg/agent/internal/deploy"
	"itcfg/agent/internal/validate"

	"github.com/spf13/cobra"
)

var (
	configDir string
	dryRun    bool
	verbose   bool
)

// NewRootCommand 创建根命令
func NewRootCommand(version, buildTime, gitCommit string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "config-agent",
		Short: "ITCFG Config Agent - 配置部署代理工具",
		Long: `ITCFG Config Agent 是配置中台的客户端工具，
用于在客户现场完成配置部署、环境校验和服务启动。

支持离线部署包导入和在线配置拉取两种模式。`,
		Version: version,
	}

	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "/opt/itcfg", "配置根目录")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "仅预览，不实际写入")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "详细输出")

	// 注册子命令
	rootCmd.AddCommand(newLoginCmd())
	rootCmd.AddCommand(newPullCmd())
	rootCmd.AddCommand(newImportCmd())
	rootCmd.AddCommand(newDeployCmd())
	rootCmd.AddCommand(newOnlineCmd())
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newRollbackCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

// newImportCmd 导入部署包命令
func newImportCmd() *cobra.Command {
	var packagePath string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "导入离线部署包",
		Long:  "解压部署包并校验完整性，准备部署环境",
		RunE: func(cmd *cobra.Command, args []string) error {
			if packagePath == "" {
				return fmt.Errorf("请指定部署包路径: --package")
			}

			fmt.Println("==========================================")
			fmt.Println("  ITCFG Config Agent - 导入部署包")
			fmt.Println("==========================================")

			importer := deploy.NewImporter(configDir, verbose)
			meta, err := importer.Import(packagePath)
			if err != nil {
				return fmt.Errorf("导入失败: %w", err)
			}

			fmt.Println()
			fmt.Println("部署包信息:")
			fmt.Printf("  客户: %s\n", meta.Customer)
			fmt.Printf("  环境: %s\n", meta.Env)
			fmt.Printf("  版本: %s\n", meta.Version)
			fmt.Printf("  创建时间: %s\n", meta.CreatedAt)
			fmt.Printf("  创建人: %s\n", meta.CreatedBy)
			fmt.Printf("  组件数量: %d\n", len(meta.Components))
			fmt.Println()
			fmt.Println("组件列表:")
			for _, comp := range meta.Components {
				fmt.Printf("  - %s (v%s)\n", comp.Name, comp.Version)
			}
			fmt.Println()
			fmt.Println("导入完成，执行 'config-agent deploy' 开始部署")
			return nil
		},
	}

	cmd.Flags().StringVar(&packagePath, "package", "", "部署包路径 (tar.gz)")
	cmd.MarkFlagRequired("package")
	return cmd
}

// newDeployCmd 部署命令
func newDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "执行部署",
		Long:  "读取配置、写入文件、导入镜像、启动服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("==========================================")
			fmt.Println("  ITCFG Config Agent - 执行部署")
			fmt.Println("==========================================")

			deployer := deploy.NewDeployer(configDir, dryRun, verbose)
			if err := deployer.Deploy(); err != nil {
				return fmt.Errorf("部署失败: %w", err)
			}

			fmt.Println()
			fmt.Println("==========================================")
			fmt.Println("  部署完成！")
			fmt.Println("==========================================")
			return nil
		},
	}

	return cmd
}

// newValidateCmd 校验命令
func newValidateCmd() *cobra.Command {
	var (
		checkPackage bool
		checkEnv     bool
		checkHealth  bool
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "校验部署环境",
		Long:  "校验部署包完整性、环境配置、服务健康状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			validator := validate.NewValidator(configDir, verbose)

			if checkPackage {
				fmt.Println("[校验] 部署包完整性...")
				if err := validator.CheckPackage(); err != nil {
					return fmt.Errorf("部署包校验失败: %w", err)
				}
				fmt.Println("  ✓ 部署包完整性通过")
			}

			if checkEnv {
				fmt.Println("[校验] 部署环境...")
				results := validator.CheckEnvironment()
				allPassed := true
				for _, r := range results {
					status := "✓"
					if !r.Passed {
						status = "✗"
						allPassed = false
					}
					fmt.Printf("  %s %s: %s\n", status, r.Name, r.Message)
				}
				if allPassed {
					fmt.Println("  ✓ 环境校验全部通过")
				} else {
					return fmt.Errorf("环境校验存在未通过项，请检查后重试")
				}
			}

			if checkHealth {
				fmt.Println("[校验] 服务健康检查...")
				if err := validator.CheckHealth(); err != nil {
					return fmt.Errorf("健康检查失败: %w", err)
				}
				fmt.Println("  ✓ 服务健康检查通过")
			}

			if !checkPackage && !checkEnv && !checkHealth {
				// 默认全部校验
				cmd.Help()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&checkPackage, "package", false, "校验部署包完整性")
	cmd.Flags().BoolVar(&checkEnv, "env", false, "校验部署环境")
	cmd.Flags().BoolVar(&checkHealth, "health", false, "检查服务健康状态")
	return cmd
}

// newStatusCmd 状态命令
func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "查看部署状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			metaPath := deploy.GetMetadataPath(configDir)
			if _, err := os.Stat(metaPath); os.IsNotExist(err) {
				fmt.Println("状态: 未部署")
				fmt.Println("未找到部署包元信息，请先执行 'config-agent import'")
				return nil
			}

			meta, err := deploy.LoadMetadata(metaPath)
			if err != nil {
				return fmt.Errorf("读取部署状态失败: %w", err)
			}

			fmt.Println("==========================================")
			fmt.Println("  部署状态")
			fmt.Println("==========================================")
			fmt.Printf("  客户: %s\n", meta.Customer)
			fmt.Printf("  环境: %s\n", meta.Env)
			fmt.Printf("  版本: %s\n", meta.Version)
			fmt.Printf("  部署时间: %s\n", meta.CreatedAt)
			fmt.Println("==========================================")
			return nil
		},
	}

	return cmd
}

// newVersionCmd 版本命令
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "查看版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Config Agent Version: %s\n", cmd.Root().Version)
		},
	}
	return cmd
}

// newLoginCmd 在线登录命令
func newLoginCmd() *cobra.Command {
	var (
		serverURL string
		envKey    string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "登录配置中台",
		Long:  "通过环境密钥认证到配置中台，获取部署权限",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverURL == "" {
				return fmt.Errorf("请指定配置中台地址: --server")
			}
			if envKey == "" {
				return fmt.Errorf("请指定环境密钥: --env-key")
			}

			fmt.Println("==========================================")
			fmt.Println("  ITCFG Config Agent - 在线登录")
			fmt.Println("==========================================")
			fmt.Printf("  服务器: %s\n", serverURL)

			client := deploy.NewOnlineClient(serverURL, envKey, verbose)
			authResult, err := client.Auth()
			if err != nil {
				return fmt.Errorf("认证失败: %w", err)
			}

			fmt.Println()
			fmt.Println("认证成功!")
			fmt.Printf("  环境ID: %s\n", authResult.EnvID)
			fmt.Printf("  环境名称: %s\n", authResult.EnvName)
			fmt.Println()

			// 保存配置
			cfg := config.AgentConfig{
				ServerURL: serverURL,
				EnvKey:    envKey,
				ConfigDir: configDir,
			}
			if err := config.Save(&cfg); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			fmt.Println("登录成功，配置已保存。执行 'config-agent pull' 拉取配置")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "配置中台地址")
	cmd.Flags().StringVar(&envKey, "env-key", "", "环境密钥")
	cmd.MarkFlagRequired("server")
	cmd.MarkFlagRequired("env-key")
	return cmd
}

// newPullCmd 在线拉取配置命令
func newPullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "拉取配置",
		Long:  "从配置中台在线拉取并渲染所有组件配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 加载配置
			cfg, err := config.Load()
			if err != nil || cfg.ServerURL == "" {
				return fmt.Errorf("未登录，请先执行 'config-agent login'")
			}

			fmt.Println("==========================================")
			fmt.Println("  ITCFG Config Agent - 在线拉取配置")
			fmt.Println("==========================================")
			fmt.Printf("  服务器: %s\n", cfg.ServerURL)

			client := deploy.NewOnlineClient(cfg.ServerURL, cfg.EnvKey, verbose)
			configs, err := client.PullConfigs()
			if err != nil {
				return fmt.Errorf("拉取配置失败: %w", err)
			}

			// 写入配置
			if dryRun {
				fmt.Println()
				fmt.Println("[DRY-RUN] 以下配置将被写入:")
				for compName, files := range configs {
					fmt.Printf("  组件: %s\n", compName)
					for path, content := range files {
						fmt.Printf("    %s (%d 字节)\n", path, len(content))
					}
				}
				return nil
			}

			configsDir := configDir + "/configs"
			if err := deploy.WriteConfigs(configsDir, configs); err != nil {
				return fmt.Errorf("写入配置失败: %w", err)
			}

			fmt.Println()
			fmt.Println("配置拉取完成，执行 'config-agent deploy' 开始部署")
			return nil
		},
	}

	return cmd
}

// newOnlineCmd 在线一键部署命令
func newOnlineCmd() *cobra.Command {
	var (
		serverURL   string
		envKey      string
		versionTag  string
	)

	cmd := &cobra.Command{
		Use:   "online",
		Short: "在线一键部署",
		Long:  "自动完成登录 → 拉取配置 → 部署 → 上报状态，一条命令完成在线部署",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverURL == "" {
				return fmt.Errorf("请指定配置中台地址: --server")
			}
			if envKey == "" {
				return fmt.Errorf("请指定环境密钥: --env-key")
			}
			if versionTag == "" {
				versionTag = fmt.Sprintf("online-%s", time.Now().Format("20060102-150405"))
			}

			fmt.Println("==========================================")
			fmt.Println("  ITCFG Config Agent - 在线一键部署")
			fmt.Println("==========================================")
			fmt.Printf("  服务器: %s\n", serverURL)

			client := deploy.NewOnlineClient(serverURL, envKey, verbose)

			// Step 1: 认证
			fmt.Println("\n[1/4] 认证...")
			authResult, err := client.Auth()
			if err != nil {
				return fmt.Errorf("认证失败: %w", err)
			}
			fmt.Printf("  ✓ 认证成功 (环境: %s)\n", authResult.EnvName)

			// Step 2: 拉取配置
			fmt.Println("\n[2/4] 拉取配置...")
			configs, err := client.PullConfigs()
			if err != nil {
				return fmt.Errorf("拉取配置失败: %w", err)
			}
			fmt.Printf("  ✓ 已拉取 %d 个组件配置\n", len(configs))

			// 写入配置
			if !dryRun {
				configsDir := configDir + "/configs"
				if err := deploy.WriteConfigs(configsDir, configs); err != nil {
					return fmt.Errorf("写入配置失败: %w", err)
				}
			} else {
				fmt.Println("  [DRY-RUN] 跳过配置写入")
			}

			// Step 3: 部署
			fmt.Println("\n[3/4] 执行部署...")
			deployer := deploy.NewDeployer(configDir, dryRun, verbose)
			if err := deployer.Deploy(); err != nil {
				// 上报失败状态
				client.ReportDeploy(versionTag, "failed", err.Error())
				return fmt.Errorf("部署失败: %w", err)
			}
			fmt.Println("  ✓ 部署完成")

			// Step 4: 上报状态
			fmt.Println("\n[4/4] 上报部署状态...")
			if err := client.ReportDeploy(versionTag, "success", "在线部署完成"); err != nil {
				fmt.Printf("  ⚠ 上报状态失败: %v\n", err)
			} else {
				fmt.Println("  ✓ 状态已上报")
			}

			// 保存配置供后续使用
			cfg := config.AgentConfig{
				ServerURL: serverURL,
				EnvKey:    envKey,
				ConfigDir: configDir,
			}
			config.Save(&cfg)

			fmt.Println()
			fmt.Println("==========================================")
			fmt.Println("  在线部署完成！")
			fmt.Println("==========================================")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "配置中台地址")
	cmd.Flags().StringVar(&envKey, "env-key", "", "环境密钥")
	cmd.Flags().StringVar(&versionTag, "version-tag", "", "部署版本标签 (默认自动生成)")
	cmd.MarkFlagRequired("server")
	cmd.MarkFlagRequired("env-key")
	return cmd
}

// newRollbackCmd 回滚命令
func newRollbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "回滚到上一版本",
		Long:  "停止当前服务，恢复到上一个部署版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("==========================================")
			fmt.Println("  ITCFG Config Agent - 版本回滚")
			fmt.Println("==========================================")

			backupDir := configDir + "/backup"
			if _, err := os.Stat(backupDir); os.IsNotExist(err) {
				return fmt.Errorf("没有可回滚的备份")
			}

			fmt.Println("正在回滚...")
			// TODO: 实现具体的回滚逻辑
			// 1. 停止当前服务
			// 2. 恢复备份配置
			// 3. 重新启动服务
			fmt.Println("回滚功能将在后续版本实现")

			return nil
		},
	}

	return cmd
}