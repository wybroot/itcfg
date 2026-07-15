package service

import (
	"itcfg/server/internal/model"
	"itcfg/server/internal/repository"
)

// CustomerService 客户服务
type CustomerService struct {
	repo *repository.CustomerRepo
}

func NewCustomerService(repo *repository.CustomerRepo) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) List() ([]model.Customer, error) {
	return s.repo.List()
}

func (s *CustomerService) GetByID(id string) (*model.Customer, error) {
	return s.repo.GetByID(id)
}

func (s *CustomerService) Create(customer *model.Customer) error {
	return s.repo.Create(customer)
}

func (s *CustomerService) Update(customer *model.Customer) error {
	return s.repo.Update(customer)
}

func (s *CustomerService) Delete(id string) error {
	return s.repo.Delete(id)
}

// EnvService 环境服务
type EnvService struct {
	repo *repository.EnvRepo
}

func NewEnvService(repo *repository.EnvRepo) *EnvService {
	return &EnvService{repo: repo}
}

func (s *EnvService) ListByCustomer(customerID string) ([]model.CustomerEnv, error) {
	return s.repo.ListByCustomer(customerID)
}

func (s *EnvService) GetByKey(envKey string) (*model.CustomerEnv, error) {
	return s.repo.GetByKey(envKey)
}

func (s *EnvService) Create(env *model.CustomerEnv) error {
	return s.repo.Create(env)
}

func (s *EnvService) Delete(id string) error {
	return s.repo.Delete(id)
}

// ComponentService 组件服务
type ComponentService struct {
	repo *repository.ComponentRepo
}

func NewComponentService(repo *repository.ComponentRepo) *ComponentService {
	return &ComponentService{repo: repo}
}

func (s *ComponentService) List() ([]model.Component, error) {
	return s.repo.List()
}

func (s *ComponentService) GetByID(id string) (*model.Component, error) {
	return s.repo.GetByID(id)
}

func (s *ComponentService) Create(component *model.Component) error {
	return s.repo.Create(component)
}

// ConfigService 配置服务
type ConfigService struct {
	repo *repository.ConfigValueRepo
}

func NewConfigService(repo *repository.ConfigValueRepo) *ConfigService {
	return &ConfigService{repo: repo}
}

func (s *ConfigService) GetByEnv(envID string) ([]model.CustomerConfigValue, error) {
	return s.repo.GetByEnv(envID)
}

func (s *ConfigService) Upsert(envID, variableID, value, updatedBy string) error {
	return s.repo.Upsert(envID, variableID, value, updatedBy)
}

func (s *ConfigService) BatchUpsert(envID string, values map[string]string, updatedBy string) error {
	return s.repo.BatchUpsert(envID, values, updatedBy)
}

// DeployRecordService 部署记录服务
type DeployRecordService struct {
	repo *repository.DeployRecordRepo
}

func NewDeployRecordService(repo *repository.DeployRecordRepo) *DeployRecordService {
	return &DeployRecordService{repo: repo}
}

func (s *DeployRecordService) ListByEnv(envID string) ([]model.DeployRecord, error) {
	return s.repo.ListByEnv(envID)
}

func (s *DeployRecordService) Create(record *model.DeployRecord) error {
	return s.repo.Create(record)
}