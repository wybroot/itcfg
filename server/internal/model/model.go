package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate UUID 自动生成
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// Customer 客户
type Customer struct {
	BaseModel
	Name    string        `gorm:"type:varchar(128);not null" json:"name"`
	Code    string        `gorm:"type:varchar(32);uniqueIndex;not null" json:"code"`
	Contact string        `gorm:"type:varchar(64)" json:"contact"`
	Status  string        `gorm:"type:varchar(16);default:active" json:"status"`
	Envs    []CustomerEnv `gorm:"foreignKey:CustomerID" json:"envs,omitempty"`
}

// CustomerEnv 客户环境
type CustomerEnv struct {
	BaseModel
	CustomerID  uuid.UUID              `gorm:"type:uuid;not null;index" json:"customer_id"`
	EnvName     string                 `gorm:"type:varchar(32);not null" json:"env_name"`
	EnvKey      string                 `gorm:"type:varchar(64);uniqueIndex;not null" json:"env_key"`
	Description string                 `gorm:"type:text" json:"description"`
	Components  []EnvironmentComponent `gorm:"foreignKey:CustomerEnvID" json:"components,omitempty"`
}

// EnvironmentComponent 环境启用组件
type EnvironmentComponent struct {
	BaseModel
	CustomerEnvID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_env_component" json:"customer_env_id"`
	ComponentID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_env_component" json:"component_id"`
	Enabled       bool      `gorm:"default:true" json:"enabled"`
	DeployOrder   int       `gorm:"default:0" json:"deploy_order"`
	Component     Component `gorm:"foreignKey:ComponentID" json:"component,omitempty"`
}

// Component 组件定义
type Component struct {
	BaseModel
	Name        string              `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	DisplayName string              `gorm:"type:varchar(128);not null" json:"display_name"`
	Description string              `gorm:"type:text" json:"description"`
	Category    string              `gorm:"type:varchar(32)" json:"category"`
	TemplateDir string              `gorm:"type:varchar(256);not null" json:"template_dir"`
	IsActive    bool                `gorm:"default:true" json:"is_active"`
	Variables   []ComponentVariable `gorm:"foreignKey:ComponentID" json:"variables,omitempty"`
}

// ComponentVariable 组件变量定义
type ComponentVariable struct {
	BaseModel
	ComponentID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_component_var_name" json:"component_id"`
	VarName        string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_component_var_name" json:"var_name"`
	VarLabel       string    `gorm:"type:varchar(128);not null" json:"var_label"`
	VarType        string    `gorm:"type:varchar(32);default:string" json:"var_type"`
	DefaultValue   string    `gorm:"type:text" json:"default_value"`
	Required       bool      `gorm:"default:false" json:"required"`
	ValidationRule string    `gorm:"type:jsonb" json:"validation_rule"`
	VarGroup       string    `gorm:"type:varchar(64)" json:"var_group"`
	SortOrder      int       `gorm:"default:0" json:"sort_order"`
	Description    string    `gorm:"type:text" json:"description"`
	Options        string    `gorm:"type:jsonb" json:"options"`
	LinkedTo       string    `gorm:"type:varchar(256)" json:"linked_to"`
}

// CustomerConfigValue 客户配置值
type CustomerConfigValue struct {
	BaseModel
	CustomerEnvID uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_env_id"`
	VariableID    uuid.UUID `gorm:"type:uuid;not null;index" json:"variable_id"`
	VarValue      string    `gorm:"type:text" json:"var_value"`
	UpdatedBy     string    `gorm:"type:varchar(64)" json:"updated_by"`
}

// ComponentArtifactVersion 制品版本关联
type ComponentArtifactVersion struct {
	BaseModel
	CustomerEnvID   uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_env_id"`
	ComponentID     uuid.UUID `gorm:"type:uuid;not null;index" json:"component_id"`
	ArtifactType    string    `gorm:"type:varchar(32);not null" json:"artifact_type"`
	ArtifactName    string    `gorm:"type:varchar(128);not null" json:"artifact_name"`
	ArtifactVersion string    `gorm:"type:varchar(64);not null" json:"artifact_version"`
	RegistryURL     string    `gorm:"type:varchar(256)" json:"registry_url"`
}

// DeployRecord 部署记录
type DeployRecord struct {
	BaseModel
	CustomerEnvID    uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_env_id"`
	VersionTag       string    `gorm:"type:varchar(64);not null" json:"version_tag"`
	ConfigSnapshot   string    `gorm:"type:jsonb" json:"config_snapshot"`
	ArtifactSnapshot string    `gorm:"type:jsonb" json:"artifact_snapshot"`
	PackageChecksum  string    `gorm:"type:varchar(128)" json:"package_checksum"`
	DeployedAt       time.Time `json:"deployed_at"`
	DeployedBy       string    `gorm:"type:varchar(64)" json:"deployed_by"`
	Status           string    `gorm:"type:varchar(32);default:pending" json:"status"`
	Notes            string    `gorm:"type:text" json:"notes"`
}

// ConfigVersion 配置版本快照
type ConfigVersion struct {
	BaseModel
	CustomerEnvID  uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_env_id"`
	Version        int       `gorm:"not null" json:"version"`
	ConfigSnapshot string    `gorm:"type:jsonb;not null" json:"config_snapshot"`
	CreatedBy      string    `gorm:"type:varchar(64)" json:"created_by"`
	ChangeSummary  string    `gorm:"type:text" json:"change_summary"`
}

// User 系统用户
type User struct {
	BaseModel
	Username string `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Password string `gorm:"type:varchar(256);not null" json:"-"`
	Nickname string `gorm:"type:varchar(64)" json:"nickname"`
	Role     string `gorm:"type:varchar(16);default:user" json:"role"`
	Status   string `gorm:"type:varchar(16);default:active" json:"status"`
}

// NotifyConfig 通知配置
type NotifyConfig struct {
	BaseModel
	Name       string `gorm:"type:varchar(64);not null" json:"name"`
	Type       string `gorm:"type:varchar(16);not null" json:"type"` // dingtalk / wecom / webhook
	WebhookURL string `gorm:"type:varchar(512);not null" json:"webhook_url"`
	Secret     string `gorm:"type:varchar(128)" json:"-"`
	Events     string `gorm:"type:varchar(256);default:deploy_success,deploy_failed" json:"events"`
	IsActive   bool   `gorm:"default:true" json:"is_active"`
}
