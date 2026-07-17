package main

import (
	"fmt"
	"log"
	"os"

	"itcfg/server/internal/auth"
	"itcfg/server/internal/handler"
	"itcfg/server/internal/middleware"
	"itcfg/server/internal/notify"
	"itcfg/server/internal/repository"
	"itcfg/server/internal/service"
	"itcfg/server/internal/template"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "itcfg/server/docs"
)

// @title           ITCFG 配置中台 API
// @version         1.0
// @description     配置管理与部署自动化系统 API 文档
// @contact.name    ITCFG Team
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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

	// 初始化模板目录
	templateDir := os.Getenv("TEMPLATE_DIR")
	if templateDir == "" {
		templateDir = "templates"
	}

	// 插入种子数据
	if err := repository.SeedComponents(db.DB, templateDir); err != nil {
		log.Printf("种子数据插入失败: %v", err)
	}

	// 初始化 Repository
	customerRepo := repository.NewCustomerRepo(db.DB)
	envRepo := repository.NewEnvRepo(db.DB)
	envComponentRepo := repository.NewEnvironmentComponentRepo(db.DB)
	componentRepo := repository.NewComponentRepo(db.DB)
	configValueRepo := repository.NewConfigValueRepo(db.DB)
	deployRecordRepo := repository.NewDeployRecordRepo(db.DB)
	configVersionRepo := repository.NewConfigVersionRepo(db.DB)
	artifactVersionRepo := repository.NewArtifactVersionRepo(db.DB)
	userRepo := repository.NewUserRepo(db.DB)
	notifyConfigRepo := repository.NewNotifyConfigRepo(db.DB)

	// 初始化 JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "itcfg-default-secret-change-in-production"
	}
	jwtManager := auth.NewJWTManager(jwtSecret, 24)

	// 初始化 Service
	authSvc := auth.NewUserService(userRepo, jwtManager)
	notifySvc := notify.NewService(notifyConfigRepo)
	customerSvc := service.NewCustomerService(customerRepo)
	envSvc := service.NewEnvService(envRepo)
	envComponentSvc := service.NewEnvironmentComponentService(envComponentRepo)
	componentSvc := service.NewComponentService(componentRepo)
	// 加密密钥（敏感配置加密存储）
	encryptionKey := os.Getenv("ENCRYPTION_KEY")

	configSvc := service.NewConfigService(configValueRepo, componentRepo, encryptionKey)
	deployRecordSvc := service.NewDeployRecordService(deployRecordRepo)
	versionSvc := service.NewConfigVersionService(configVersionRepo, configValueRepo)
	artifactVersionSvc := service.NewArtifactVersionService(artifactVersionRepo)

	// 初始化模板引擎
	templateEngine := template.NewEngine(templateDir)

	// 初始化导出服务
	exportDir := os.Getenv("EXPORT_DIR")
	if exportDir == "" {
		exportDir = "exports"
	}
	os.MkdirAll(exportDir, 0755)
	packageExporter := service.NewPackageExporter(
		configSvc, componentSvc, envSvc, customerSvc, envComponentSvc, artifactVersionSvc, deployRecordSvc, templateEngine, exportDir,
	)

	// 初始化 Handler
	h := handler.NewHandler(
		customerSvc, envSvc, envComponentSvc, componentSvc, configSvc, deployRecordSvc,
		templateEngine, packageExporter, versionSvc, artifactVersionSvc, authSvc, notifySvc,
	)

	// 初始化管理员账户
	adminUser := os.Getenv("ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("ADMIN_PASS")
	if adminPass == "" {
		adminPass = "admin123"
	}
	if err := authSvc.InitAdmin(adminUser, adminPass); err != nil {
		log.Printf("初始化管理员账户失败: %v", err)
	} else {
		log.Printf("管理员账户已就绪 (用户: %s)", adminUser)
	}

	// 初始化 Gin
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.EnvAuthMiddleware())

	// 注册路由（传入 JWT 中间件保护业务接口）
	jwtMiddleware := middleware.JWTAuthMiddleware(jwtManager)
	h.RegisterRoutes(r, jwtMiddleware)

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		status := "ok"
		dbStatus := "ok"

		// 检查数据库连接
		sqlDB, err := db.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "error"
			status = "degraded"
		}

		// 检查模板目录
		templateStatus := "ok"
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			templateStatus = "error"
			status = "degraded"
		}

		c.JSON(200, gin.H{
			"status":  status,
			"service": "itcfg-server",
			"checks": gin.H{
				"database":  dbStatus,
				"templates": templateStatus,
			},
		})
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
