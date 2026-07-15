package repository

import (
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