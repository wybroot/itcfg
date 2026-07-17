package repository

import (
	"time"

	"itcfg/server/internal/model"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库连接管理器
type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建数据库连接
func NewDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}

// AutoMigrate 自动迁移
func (d *Database) AutoMigrate() error {
	return d.DB.AutoMigrate(
		&model.Customer{},
		&model.CustomerEnv{},
		&model.Component{},
		&model.ComponentVariable{},
		&model.CustomerConfigValue{},
		&model.ComponentArtifactVersion{},
		&model.DeployRecord{},
		&model.ConfigVersion{},
		&model.User{},
		&model.NotifyConfig{},
	)
}

// ==================== 客户 Repository ====================

type CustomerRepo struct {
	db *gorm.DB
}

func NewCustomerRepo(db *gorm.DB) *CustomerRepo {
	return &CustomerRepo{db: db}
}

func (r *CustomerRepo) List() ([]model.Customer, error) {
	var customers []model.Customer
	err := r.db.Order("created_at DESC").Find(&customers).Error
	return customers, err
}

func (r *CustomerRepo) GetByID(id string) (*model.Customer, error) {
	var customer model.Customer
	err := r.db.Preload("Envs").First(&customer, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepo) Create(customer *model.Customer) error {
	return r.db.Create(customer).Error
}

func (r *CustomerRepo) Update(customer *model.Customer) error {
	return r.db.Save(customer).Error
}

func (r *CustomerRepo) Delete(id string) error {
	return r.db.Delete(&model.Customer{}, "id = ?", id).Error
}

// ==================== 环境 Repository ====================

type EnvRepo struct {
	db *gorm.DB
}

func NewEnvRepo(db *gorm.DB) *EnvRepo {
	return &EnvRepo{db: db}
}

func (r *EnvRepo) ListByCustomer(customerID string) ([]model.CustomerEnv, error) {
	var envs []model.CustomerEnv
	err := r.db.Where("customer_id = ?", customerID).Order("created_at DESC").Find(&envs).Error
	return envs, err
}

func (r *EnvRepo) GetByID(id string) (*model.CustomerEnv, error) {
	var env model.CustomerEnv
	err := r.db.First(&env, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *EnvRepo) GetByKey(envKey string) (*model.CustomerEnv, error) {
	var env model.CustomerEnv
	err := r.db.First(&env, "env_key = ?", envKey).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *EnvRepo) Create(env *model.CustomerEnv) error {
	return r.db.Create(env).Error
}

func (r *EnvRepo) Update(env *model.CustomerEnv) error {
	return r.db.Save(env).Error
}

func (r *EnvRepo) Delete(id string) error {
	return r.db.Delete(&model.CustomerEnv{}, "id = ?", id).Error
}

// ==================== 组件 Repository ====================

type ComponentRepo struct {
	db *gorm.DB
}

func NewComponentRepo(db *gorm.DB) *ComponentRepo {
	return &ComponentRepo{db: db}
}

func (r *ComponentRepo) List() ([]model.Component, error) {
	var components []model.Component
	err := r.db.Order("created_at DESC").Find(&components).Error
	return components, err
}

func (r *ComponentRepo) GetByID(id string) (*model.Component, error) {
	var component model.Component
	err := r.db.Preload("Variables").First(&component, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

func (r *ComponentRepo) Create(component *model.Component) error {
	return r.db.Create(component).Error
}

func (r *ComponentRepo) Update(component *model.Component) error {
	return r.db.Save(component).Error
}

func (r *ComponentRepo) Delete(id string) error {
	return r.db.Delete(&model.Component{}, "id = ?", id).Error
}

// GetAllVariables 获取所有组件的变量定义（返回 map[variableID]*ComponentVariable）
func (r *ComponentRepo) GetAllVariables() (map[string]*model.ComponentVariable, error) {
	var variables []model.ComponentVariable
	if err := r.db.Find(&variables).Error; err != nil {
		return nil, err
	}
	result := make(map[string]*model.ComponentVariable, len(variables))
	for i := range variables {
		result[variables[i].ID.String()] = &variables[i]
	}
	return result, nil
}

// ==================== 配置值 Repository ====================

type ConfigValueRepo struct {
	db *gorm.DB
}

func NewConfigValueRepo(db *gorm.DB) *ConfigValueRepo {
	return &ConfigValueRepo{db: db}
}

func (r *ConfigValueRepo) GetByEnv(envID string) ([]model.CustomerConfigValue, error) {
	var values []model.CustomerConfigValue
	err := r.db.Where("customer_env_id = ?", envID).Find(&values).Error
	return values, err
}

func (r *ConfigValueRepo) Upsert(envID string, variableID string, value string, updatedBy string) error {
	var existing model.CustomerConfigValue
	err := r.db.Where("customer_env_id = ? AND variable_id = ?", envID, variableID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(&model.CustomerConfigValue{
			CustomerEnvID: uuid.MustParse(envID),
			VariableID:    uuid.MustParse(variableID),
			VarValue:      value,
			UpdatedBy:     updatedBy,
		}).Error
	}
	if err != nil {
		return err
	}
	existing.VarValue = value
	existing.UpdatedBy = updatedBy
	return r.db.Save(&existing).Error
}

func (r *ConfigValueRepo) BatchUpsert(envID string, values map[string]string, updatedBy string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for varID, val := range values {
			var existing model.CustomerConfigValue
			err := tx.Where("customer_env_id = ? AND variable_id = ?", envID, varID).First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				if createErr := tx.Create(&model.CustomerConfigValue{
					CustomerEnvID: uuid.MustParse(envID),
					VariableID:    uuid.MustParse(varID),
					VarValue:      val,
					UpdatedBy:     updatedBy,
				}).Error; createErr != nil {
					return createErr
				}
			} else if err != nil {
				return err
			} else {
				existing.VarValue = val
				existing.UpdatedBy = updatedBy
				if saveErr := tx.Save(&existing).Error; saveErr != nil {
					return saveErr
				}
			}
		}
		return nil
	})
}

// ==================== 部署记录 Repository ====================

type DeployRecordRepo struct {
	db *gorm.DB
}

func NewDeployRecordRepo(db *gorm.DB) *DeployRecordRepo {
	return &DeployRecordRepo{db: db}
}

func (r *DeployRecordRepo) ListByEnv(envID string) ([]model.DeployRecord, error) {
	var records []model.DeployRecord
	err := r.db.Where("customer_env_id = ?", envID).Order("deployed_at DESC").Find(&records).Error
	return records, err
}

func (r *DeployRecordRepo) Create(record *model.DeployRecord) error {
	return r.db.Create(record).Error
}

func (r *DeployRecordRepo) ListAll() ([]model.DeployRecord, error) {
	var records []model.DeployRecord
	err := r.db.Order("deployed_at DESC").Find(&records).Error
	return records, err
}

// ==================== 配置版本 Repository ====================

type ConfigVersionRepo struct {
	db *gorm.DB
}

func NewConfigVersionRepo(db *gorm.DB) *ConfigVersionRepo {
	return &ConfigVersionRepo{db: db}
}

func (r *ConfigVersionRepo) ListByEnv(envID string) ([]model.ConfigVersion, error) {
	var versions []model.ConfigVersion
	err := r.db.Where("customer_env_id = ?", envID).
		Order("version DESC").Limit(50).Find(&versions).Error
	return versions, err
}

func (r *ConfigVersionRepo) GetByEnvAndVersion(envID string, version int) (*model.ConfigVersion, error) {
	var v model.ConfigVersion
	err := r.db.Where("customer_env_id = ? AND version = ?", envID, version).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ConfigVersionRepo) GetLatestVersion(envID string) (int, error) {
	var v model.ConfigVersion
	err := r.db.Where("customer_env_id = ?", envID).
		Order("version DESC").First(&v).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return v.Version, nil
}

func (r *ConfigVersionRepo) Create(version *model.ConfigVersion) error {
	return r.db.Create(version).Error
}

// ==================== 配置克隆 Repository ====================

func (r *ConfigValueRepo) CloneConfigs(fromEnvID, toEnvID string, updatedBy string) error {
	// 获取源环境的配置
	fromConfigs, err := r.GetByEnv(fromEnvID)
	if err != nil {
		return err
	}

	values := make(map[string]string)
	for _, cfg := range fromConfigs {
		values[cfg.VariableID.String()] = cfg.VarValue
	}

	return r.BatchUpsert(toEnvID, values, updatedBy)
}

// ==================== 制品版本 Repository ====================

type ArtifactVersionRepo struct {
	db *gorm.DB
}

func NewArtifactVersionRepo(db *gorm.DB) *ArtifactVersionRepo {
	return &ArtifactVersionRepo{db: db}
}

func (r *ArtifactVersionRepo) ListByEnv(envID string) ([]model.ComponentArtifactVersion, error) {
	var artifacts []model.ComponentArtifactVersion
	err := r.db.Where("customer_env_id = ?", envID).
		Order("created_at DESC").Find(&artifacts).Error
	return artifacts, err
}

func (r *ArtifactVersionRepo) GetByID(id string) (*model.ComponentArtifactVersion, error) {
	var artifact model.ComponentArtifactVersion
	err := r.db.First(&artifact, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &artifact, nil
}

func (r *ArtifactVersionRepo) Create(artifact *model.ComponentArtifactVersion) error {
	return r.db.Create(artifact).Error
}

func (r *ArtifactVersionRepo) Update(artifact *model.ComponentArtifactVersion) error {
	return r.db.Save(artifact).Error
}

func (r *ArtifactVersionRepo) Delete(id string) error {
	return r.db.Delete(&model.ComponentArtifactVersion{}, "id = ?", id).Error
}

// CloneArtifacts 克隆制品版本关联
func (r *ArtifactVersionRepo) CloneArtifacts(fromEnvID, toEnvID string) error {
	var sourceArtifacts []model.ComponentArtifactVersion
	if err := r.db.Where("customer_env_id = ?", fromEnvID).Find(&sourceArtifacts).Error; err != nil {
		return err
	}

	for _, artifact := range sourceArtifacts {
		artifact.ID = uuid.Nil // 让数据库生成新 ID
		artifact.CustomerEnvID = uuid.MustParse(toEnvID)
		artifact.CreatedAt = time.Now()
		artifact.UpdatedAt = time.Now()
		if err := r.db.Create(&artifact).Error; err != nil {
			return err
		}
	}
	return nil
}

// ==================== 用户 Repository ====================

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) List() ([]model.User, error) {
	var users []model.User
	err := r.db.Order("created_at DESC").Find(&users).Error
	return users, err
}

func (r *UserRepo) GetByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepo) Delete(id string) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

// ==================== 通知配置 Repository ====================

type NotifyConfigRepo struct {
	db *gorm.DB
}

func NewNotifyConfigRepo(db *gorm.DB) *NotifyConfigRepo {
	return &NotifyConfigRepo{db: db}
}

func (r *NotifyConfigRepo) List() ([]model.NotifyConfig, error) {
	var configs []model.NotifyConfig
	err := r.db.Order("created_at DESC").Find(&configs).Error
	return configs, err
}

func (r *NotifyConfigRepo) ListActive() ([]model.NotifyConfig, error) {
	var configs []model.NotifyConfig
	err := r.db.Where("is_active = ?", true).Find(&configs).Error
	return configs, err
}

func (r *NotifyConfigRepo) GetByID(id string) (*model.NotifyConfig, error) {
	var cfg model.NotifyConfig
	err := r.db.First(&cfg, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *NotifyConfigRepo) Create(cfg *model.NotifyConfig) error {
	return r.db.Create(cfg).Error
}

func (r *NotifyConfigRepo) Update(cfg *model.NotifyConfig) error {
	return r.db.Save(cfg).Error
}

func (r *NotifyConfigRepo) Delete(id string) error {
	return r.db.Delete(&model.NotifyConfig{}, "id = ?", id).Error
}