package repository

import (
	"log"

	"itcfg/server/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedComponents 插入默认组件数据（仅当组件表为空时）
func SeedComponents(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Component{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	components := []model.Component{
		{Name: "nginx",       DisplayName: "Nginx",          Category: "web-server",     Description: "高性能 Web 服务器和反向代理",                      TemplateDir: "nginx"},
		{Name: "apache",      DisplayName: "Apache HTTPD",    Category: "web-server",     Description: "Apache HTTP 服务器",                              TemplateDir: "apache"},
		{Name: "tomcat",      DisplayName: "Apache Tomcat",   Category: "application",    Description: "Java Web 应用服务器",                             TemplateDir: "tomcat"},
		{Name: "springboot",  DisplayName: "Spring Boot",     Category: "application",    Description: "Spring Boot 微服务应用",                          TemplateDir: "springboot"},
		{Name: "postgresql",  DisplayName: "PostgreSQL",      Category: "database",       Description: "关系型数据库",                                    TemplateDir: "postgresql"},
		{Name: "mysql",       DisplayName: "MySQL",           Category: "database",       Description: "关系型数据库",                                    TemplateDir: "mysql"},
		{Name: "redis",       DisplayName: "Redis",           Category: "cache",          Description: "高性能缓存数据库",                                TemplateDir: "redis"},
		{Name: "rabbitmq",    DisplayName: "RabbitMQ",        Category: "message-queue",  Description: "消息队列中间件",                                  TemplateDir: "rabbitmq"},
		{Name: "minio",       DisplayName: "MinIO",           Category: "object-storage", Description: "兼容 S3 的对象存储服务",                          TemplateDir: "minio"},
		{Name: "nacos",       DisplayName: "Nacos",           Category: "coordination",   Description: "服务注册与配置中心",                              TemplateDir: "nacos"},
		{Name: "elasticsearch", DisplayName: "Elasticsearch", Category: "search-engine",  Description: "分布式搜索引擎",                                  TemplateDir: "elasticsearch"},
		{Name: "onlyoffice",  DisplayName: "OnlyOffice",      Category: "office",         Description: "在线办公文档协作套件",                             TemplateDir: "onlyoffice"},
		{Name: "sftpgo",      DisplayName: "SFTPGo",          Category: "file-service",   Description: "SFTP/FTPS/WebDAV 文件服务",                      TemplateDir: "sftpgo"},
	}

	for i := range components {
		if err := db.Create(&components[i]).Error; err != nil {
			log.Printf("插入组件 %s 失败: %v", components[i].Name, err)
			continue
		}

		// 为每个组件添加默认变量
		variables := getDefaultVariables(components[i].ID, components[i].Name)
		for j := range variables {
			if err := db.Create(&variables[j]).Error; err != nil {
				log.Printf("插入变量 %s.%s 失败: %v", components[i].Name, variables[j].VarName, err)
			}
		}
	}

	log.Printf("已插入 %d 个默认组件", len(components))
	return nil
}

func getDefaultVariables(componentID uuid.UUID, compName string) []model.ComponentVariable {
	type varDef struct {
		name, label, varType, defaultValue, description, varGroup string
		required bool
		sortOrder int
	}

	defaults := []varDef{
		{name: "port",         label: "服务端口",   varType: "number", defaultValue: "8080",     description: "服务监听端口",           varGroup: "基础配置", required: true,  sortOrder: 1},
		{name: "host",         label: "绑定地址",   varType: "string", defaultValue: "0.0.0.0", description: "服务绑定地址",           varGroup: "基础配置", required: true,  sortOrder: 2},
		{name: "log_level",    label: "日志级别",   varType: "select", defaultValue: "info",    description: "日志输出级别",           varGroup: "日志配置", required: false, sortOrder: 3},
		{name: "max_conn",     label: "最大连接数", varType: "number", defaultValue: "1000",    description: "最大并发连接数",         varGroup: "性能配置", required: false, sortOrder: 4},
		{name: "timeout",      label: "超时时间",   varType: "number", defaultValue: "30",      description: "请求超时时间（秒）",      varGroup: "性能配置", required: false, sortOrder: 5},
	}

	var variables []model.ComponentVariable
	for i, d := range defaults {
		variables = append(variables, model.ComponentVariable{
			ComponentID:  componentID,
			VarName:      d.name,
			VarLabel:     d.label,
			VarType:      d.varType,
			DefaultValue: d.defaultValue,
			Required:     d.required,
			Description:  d.description,
			VarGroup:     d.varGroup,
			SortOrder:    i + 1,
		})
	}
	return variables
}
