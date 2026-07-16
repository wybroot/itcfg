package main

import (
	"fmt"
	"log"
	"os"

	"itcfg/server/internal/handler"
	"itcfg/server/internal/middleware"
	"itcfg/server/internal/repository"
	"itcfg/server/internal/service"
	"itcfg/server/internal/template"

	"github.com/gin-gonic/gin"
)

func main() {
	// 数据库配置
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=itcfg password=itcfg123 dbname=itcfg port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}

	// 连接数据库
	db, err := repository.NewDatabase(dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")

	// 初始化 Repository
	customerRepo := repository.NewCustomerRepo(db.DB)
	envRepo := repository.NewEnvRepo(db.DB)
	componentRepo := repository.NewComponentRepo(db.DB)
	configValueRepo := repository.NewConfigValueRepo(db.DB)
	deployRecordRepo := repository.NewDeployRecordRepo(db.DB)

	// 初始化 Repository
	configVersionRepo := repository.NewConfigVersionRepo(db.DB)

	// 初始化 Service
	customerSvc := service.NewCustomerService(customerRepo)
	envSvc := service.NewEnvService(envRepo)
	componentSvc := service.NewComponentService(componentRepo)
	configSvc := service.NewConfigService(configValueRepo)
	deployRecordSvc := service.NewDeployRecordService(deployRecordRepo)
	versionSvc := service.NewConfigVersionService(configVersionRepo, configValueRepo)

	// 初始化模板引擎
	templateDir := os.Getenv("TEMPLATE_DIR")
	if templateDir == "" {
		templateDir = "templates"
	}
	templateEngine := template.NewEngine(templateDir)

	// 初始化导出服务
	exportDir := os.Getenv("EXPORT_DIR")
	if exportDir == "" {
		exportDir = "exports"
	}
	os.MkdirAll(exportDir, 0755)
	packageExporter := service.NewPackageExporter(
		configSvc, componentSvc, envSvc, customerSvc, templateEngine, exportDir,
	)

	// 初始化 Handler
	h := handler.NewHandler(
		customerSvc, envSvc, componentSvc, configSvc, deployRecordSvc, templateEngine, packageExporter, versionSvc,
	)

	// 初始化 Gin
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.EnvAuthMiddleware())

	// 注册路由
	h.RegisterRoutes(r)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "itcfg-server"})
	})

	// 启动服务
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("ITCFG 配置中台启动成功，监听端口 %s", port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}