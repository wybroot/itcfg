package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"itcfg/server/internal/crypto"
	"itcfg/server/internal/model"
	"itcfg/server/internal/repository"
	ittemplate "itcfg/server/internal/template"
	"itcfg/server/internal/validate"

	"github.com/google/uuid"
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

func (s *EnvService) GetByID(id string) (*model.CustomerEnv, error) {
	return s.repo.GetByID(id)
}

func (s *EnvService) GetByKey(envKey string) (*model.CustomerEnv, error) {
	return s.repo.GetByKey(envKey)
}

func (s *EnvService) Create(env *model.CustomerEnv) error {
	return s.repo.Create(env)
}

func (s *EnvService) Update(env *model.CustomerEnv) error {
	return s.repo.Update(env)
}

func (s *EnvService) Delete(id string) error {
	return s.repo.Delete(id)
}

// EnvironmentComponentService 环境组件服务
type EnvironmentComponentService struct {
	repo *repository.EnvironmentComponentRepo
}

func NewEnvironmentComponentService(repo *repository.EnvironmentComponentRepo) *EnvironmentComponentService {
	return &EnvironmentComponentService{repo: repo}
}

func (s *EnvironmentComponentService) ListByEnv(envID string) ([]model.EnvironmentComponent, error) {
	return s.repo.ListByEnv(envID)
}

func (s *EnvironmentComponentService) Replace(envID string, components []model.EnvironmentComponent) error {
	return s.repo.Replace(envID, components)
}

func (s *EnvironmentComponentService) CloneComponents(fromEnvID, toEnvID string) error {
	return s.repo.CloneComponents(fromEnvID, toEnvID)
}

// ComponentService 组件服务
type ComponentService struct {
	repo           *repository.ComponentRepo
	templateEngine *ittemplate.Engine
}

func NewComponentService(repo *repository.ComponentRepo, templateEngine *ittemplate.Engine) *ComponentService {
	return &ComponentService{repo: repo, templateEngine: templateEngine}
}

func (s *ComponentService) List() ([]model.Component, error) {
	return s.repo.List()
}

func (s *ComponentService) ListActiveWithVariables() ([]model.Component, error) {
	return s.repo.ListActiveWithVariables()
}

func (s *ComponentService) GetByID(id string) (*model.Component, error) {
	return s.repo.GetByID(id)
}

func (s *ComponentService) Create(component *model.Component) error {
	return s.repo.Create(component)
}

func (s *ComponentService) Update(component *model.Component) error {
	return s.repo.Update(component)
}

func (s *ComponentService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *ComponentService) ListVariables(componentID string) ([]model.ComponentVariable, error) {
	return s.repo.ListVariables(componentID)
}

func (s *ComponentService) GetVariable(componentID, variableID string) (*model.ComponentVariable, error) {
	return s.repo.GetVariable(componentID, variableID)
}

func (s *ComponentService) CreateVariable(componentID string, variable *model.ComponentVariable) error {
	if _, err := s.GetByID(componentID); err != nil {
		return err
	}
	variable.ComponentID = uuid.MustParse(componentID)
	if variable.SortOrder == 0 {
		variables, err := s.repo.ListVariables(componentID)
		if err != nil {
			return err
		}
		variable.SortOrder = len(variables) + 1
	}
	if err := normalizeAndValidateVariable(variable); err != nil {
		return err
	}
	return s.repo.CreateVariable(variable)
}

func (s *ComponentService) UpdateVariable(componentID, variableID string, variable *model.ComponentVariable) error {
	existing, err := s.repo.GetVariable(componentID, variableID)
	if err != nil {
		return err
	}
	existing.VarName = variable.VarName
	existing.VarLabel = variable.VarLabel
	existing.VarType = variable.VarType
	existing.DefaultValue = variable.DefaultValue
	existing.Required = variable.Required
	existing.ValidationRule = variable.ValidationRule
	existing.VarGroup = variable.VarGroup
	existing.SortOrder = variable.SortOrder
	existing.Description = variable.Description
	existing.Options = variable.Options
	existing.LinkedTo = variable.LinkedTo
	if err := normalizeAndValidateVariable(existing); err != nil {
		return err
	}
	return s.repo.UpdateVariable(existing)
}

func (s *ComponentService) DeleteVariable(componentID, variableID string) error {
	return s.repo.DeleteVariable(componentID, variableID)
}

func (s *ComponentService) ImportVariablesFromTemplate(componentID string, overwrite bool) error {
	component, err := s.GetByID(componentID)
	if err != nil {
		return err
	}
	if s.templateEngine == nil {
		return fmt.Errorf("模板引擎未初始化")
	}
	templateDir := component.TemplateDir
	if templateDir == "" {
		templateDir = component.Name
	}
	variables, err := s.buildVariablesFromTemplate(templateDir)
	if err != nil {
		return err
	}
	return s.repo.UpsertVariablesByName(componentID, variables, overwrite)
}

func (s *ComponentService) SyncTemplate(templateDir string, overwrite bool) (*model.Component, error) {
	if s.templateEngine == nil {
		return nil, fmt.Errorf("模板引擎未初始化")
	}
	manifest, err := s.templateEngine.LoadManifest(templateDir)
	if err != nil {
		return nil, err
	}
	component := &model.Component{
		Name:        manifest.Name,
		DisplayName: manifest.DisplayName,
		Description: manifest.Description,
		Category:    manifest.Category,
		TemplateDir: templateDir,
		IsActive:    true,
	}
	existingComponents, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	for i := range existingComponents {
		if existingComponents[i].Name == manifest.Name {
			existingComponents[i].DisplayName = component.DisplayName
			existingComponents[i].Description = component.Description
			existingComponents[i].Category = component.Category
			existingComponents[i].TemplateDir = component.TemplateDir
			existingComponents[i].IsActive = true
			if err := s.repo.Update(&existingComponents[i]); err != nil {
				return nil, err
			}
			component = &existingComponents[i]
			break
		}
	}
	if component.ID == uuid.Nil {
		if err := s.repo.Create(component); err != nil {
			return nil, err
		}
	}
	variables, err := s.buildVariablesFromTemplate(templateDir)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpsertVariablesByName(component.ID.String(), variables, overwrite); err != nil {
		return nil, err
	}
	return s.GetByID(component.ID.String())
}

func (s *ComponentService) buildVariablesFromTemplate(templateDir string) ([]model.ComponentVariable, error) {
	variablesFile, err := s.templateEngine.LoadVariables(templateDir)
	if err != nil {
		return nil, err
	}
	variables := make([]model.ComponentVariable, 0, len(variablesFile.Variables))
	for i, v := range variablesFile.Variables {
		variable := model.ComponentVariable{
			VarName:        v.Name,
			VarLabel:       v.Label,
			VarType:        normalizeVarType(v.Type),
			DefaultValue:   v.Default,
			Required:       v.Required,
			ValidationRule: marshalValidationRule(v.Min, v.Max, v.Regex),
			VarGroup:       v.Group,
			SortOrder:      i + 1,
			Description:    v.Description,
			Options:        marshalStringSlice(v.Options),
			LinkedTo:       v.LinkedTo,
		}
		if err := normalizeAndValidateVariable(&variable); err != nil {
			return nil, err
		}
		variables = append(variables, variable)
	}
	return variables, nil
}

func normalizeAndValidateVariable(variable *model.ComponentVariable) error {
	variable.VarName = strings.TrimSpace(variable.VarName)
	variable.VarLabel = strings.TrimSpace(variable.VarLabel)
	variable.VarType = normalizeVarType(variable.VarType)
	if variable.VarName == "" {
		return fmt.Errorf("变量名不能为空")
	}
	if variable.VarLabel == "" {
		return fmt.Errorf("变量标签不能为空")
	}
	switch variable.VarType {
	case "string", "password", "number", "bool", "select":
	default:
		return fmt.Errorf("不支持的变量类型: %s", variable.VarType)
	}
	if strings.TrimSpace(variable.Options) != "" {
		var options []string
		if err := json.Unmarshal([]byte(variable.Options), &options); err != nil {
			return fmt.Errorf("选项必须是 JSON 字符串数组")
		}
	}
	if strings.TrimSpace(variable.ValidationRule) != "" {
		var rule map[string]any
		if err := json.Unmarshal([]byte(variable.ValidationRule), &rule); err != nil {
			return fmt.Errorf("校验规则必须是 JSON 对象")
		}
	}
	return nil
}

func normalizeVarType(varType string) string {
	if varType == "" {
		return "string"
	}
	if varType == "boolean" {
		return "bool"
	}
	return varType
}

func marshalStringSlice(values []string) string {
	if len(values) == 0 {
		return ""
	}
	data, err := json.Marshal(values)
	if err != nil {
		return ""
	}
	return string(data)
}

func marshalValidationRule(min, max int, regex string) string {
	rule := map[string]any{}
	if min != 0 {
		rule["min"] = min
	}
	if max != 0 {
		rule["max"] = max
	}
	if regex != "" {
		rule["regex"] = regex
	}
	if len(rule) == 0 {
		return ""
	}
	data, err := json.Marshal(rule)
	if err != nil {
		return ""
	}
	return string(data)
}

// ConfigService 配置服务
type ConfigService struct {
	repo      *repository.ConfigValueRepo
	compRepo  *repository.ComponentRepo
	crypto    *crypto.AESGCM
	validator *validate.ConfigValidator
}

func NewConfigService(repo *repository.ConfigValueRepo, compRepo *repository.ComponentRepo, encryptionKey string) *ConfigService {
	svc := &ConfigService{
		repo:      repo,
		compRepo:  compRepo,
		validator: validate.NewConfigValidator(),
	}
	if encryptionKey != "" {
		svc.crypto = crypto.NewAESGCM(encryptionKey)
	}
	return svc
}

// maskedPassword 密码掩码标记
const maskedPassword = "***ENCRYPTED***"

// GetByEnv 获取环境配置值（Web UI 用，密码字段脱敏）
func (s *ConfigService) GetByEnv(envID string) ([]model.CustomerConfigValue, error) {
	configs, err := s.repo.GetByEnv(envID)
	if err != nil {
		return nil, err
	}
	// Web UI：密码字段返回掩码
	if s.crypto != nil {
		varMap, _ := s.compRepo.GetAllVariables()
		for i := range configs {
			if v, ok := varMap[configs[i].VariableID.String()]; ok && v.VarType == "password" {
				configs[i].VarValue = maskedPassword
			}
		}
	}
	return configs, nil
}

// GetByEnvDecrypted 获取环境配置值（Agent/导出用，解密密码字段）
func (s *ConfigService) GetByEnvDecrypted(envID string) ([]model.CustomerConfigValue, error) {
	configs, err := s.repo.GetByEnv(envID)
	if err != nil {
		return nil, err
	}
	if s.crypto != nil {
		varMap, _ := s.compRepo.GetAllVariables()
		for i := range configs {
			if v, ok := varMap[configs[i].VariableID.String()]; ok && v.VarType == "password" {
				if configs[i].VarValue != "" && configs[i].VarValue != maskedPassword {
					decrypted, decErr := s.crypto.Decrypt(configs[i].VarValue)
					if decErr == nil {
						configs[i].VarValue = decrypted
					}
				}
			}
		}
	}
	return configs, nil
}

func (s *ConfigService) Upsert(envID, variableID, value, updatedBy string) error {
	return s.BatchUpsert(envID, map[string]string{variableID: value}, updatedBy)
}

func (s *ConfigService) BatchUpsert(envID string, values map[string]string, updatedBy string) error {
	// 获取变量定义
	varMap, _ := s.compRepo.GetAllVariables()

	// 校验 + 加密处理
	processedValues := make(map[string]string, len(values))
	for varID, val := range values {
		// 跳过掩码值（用户未修改密码字段）
		if val == maskedPassword {
			continue
		}

		varDef := varMap[varID]

		// 校验值
		if varDef != nil {
			if err := s.validator.Validate(varDef.VarType, val, varDef.ValidationRule); err != nil {
				return fmt.Errorf("变量 %s(%s) 校验失败: %w", varDef.VarLabel, varDef.VarName, err)
			}
		}

		// 加密密码类型
		if varDef != nil && varDef.VarType == "password" && s.crypto != nil && strings.TrimSpace(val) != "" {
			encrypted, err := s.crypto.Encrypt(val)
			if err != nil {
				return fmt.Errorf("加密失败: %w", err)
			}
			val = encrypted
		}

		processedValues[varID] = val
	}

	return s.repo.BatchUpsert(envID, processedValues, updatedBy)
}

func (s *ConfigService) CloneConfigs(fromEnvID, toEnvID, updatedBy string) error {
	return s.repo.CloneConfigs(fromEnvID, toEnvID, updatedBy)
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

func (s *DeployRecordService) ListAll() ([]model.DeployRecord, error) {
	return s.repo.ListAll()
}

// ConfigVersionService 配置版本服务
type ConfigVersionService struct {
	repo       *repository.ConfigVersionRepo
	configRepo *repository.ConfigValueRepo
}

func NewConfigVersionService(repo *repository.ConfigVersionRepo, configRepo *repository.ConfigValueRepo) *ConfigVersionService {
	return &ConfigVersionService{repo: repo, configRepo: configRepo}
}

func (s *ConfigVersionService) ListByEnv(envID string) ([]model.ConfigVersion, error) {
	return s.repo.ListByEnv(envID)
}

func (s *ConfigVersionService) GetByEnvAndVersion(envID string, version int) (*model.ConfigVersion, error) {
	return s.repo.GetByEnvAndVersion(envID, version)
}

// SaveSnapshot 保存当前配置快照
func (s *ConfigVersionService) SaveSnapshot(envID string, createdBy string, changeSummary string) (*model.ConfigVersion, error) {
	// 获取当前配置
	configs, err := s.configRepo.GetByEnv(envID)
	if err != nil {
		return nil, err
	}

	// 序列化配置快照
	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.VariableID.String()] = cfg.VarValue
	}
	snapshotData, _ := json.Marshal(configMap)

	// 获取最新版本号
	latestVersion, err := s.repo.GetLatestVersion(envID)
	if err != nil {
		return nil, err
	}

	version := &model.ConfigVersion{
		CustomerEnvID:  uuid.MustParse(envID),
		Version:        latestVersion + 1,
		ConfigSnapshot: string(snapshotData),
		CreatedBy:      createdBy,
		ChangeSummary:  changeSummary,
	}

	if err := s.repo.Create(version); err != nil {
		return nil, err
	}
	return version, nil
}

// Rollback 回滚到指定版本
func (s *ConfigVersionService) Rollback(envID string, targetVersion int, operator string) error {
	// 获取目标版本快照
	version, err := s.repo.GetByEnvAndVersion(envID, targetVersion)
	if err != nil {
		return fmt.Errorf("版本 %d 不存在", targetVersion)
	}

	// 解析快照
	var configMap map[string]string
	if err := json.Unmarshal([]byte(version.ConfigSnapshot), &configMap); err != nil {
		return fmt.Errorf("解析快照失败: %w", err)
	}

	// 恢复到目标版本
	if err := s.configRepo.BatchUpsert(envID, configMap, operator); err != nil {
		return fmt.Errorf("恢复配置失败: %w", err)
	}

	// 记录回滚操作（创建一个新版本，指向回滚来源）
	_, err = s.SaveSnapshot(envID, operator, fmt.Sprintf("回滚到版本 %d", targetVersion))
	return err
}

// Diff 对比两个版本的差异
func (s *ConfigVersionService) Diff(envID string, fromVersion, toVersion int) (map[string]DiffEntry, error) {
	fromV, err := s.repo.GetByEnvAndVersion(envID, fromVersion)
	if err != nil {
		return nil, err
	}
	toV, err := s.repo.GetByEnvAndVersion(envID, toVersion)
	if err != nil {
		return nil, err
	}

	var fromMap, toMap map[string]string
	json.Unmarshal([]byte(fromV.ConfigSnapshot), &fromMap)
	json.Unmarshal([]byte(toV.ConfigSnapshot), &toMap)

	result := make(map[string]DiffEntry)
	allKeys := make(map[string]bool)
	for k := range fromMap {
		allKeys[k] = true
	}
	for k := range toMap {
		allKeys[k] = true
	}

	for k := range allKeys {
		oldVal := fromMap[k]
		newVal := toMap[k]
		if oldVal != newVal {
			result[k] = DiffEntry{
				OldValue: oldVal,
				NewValue: newVal,
			}
		}
	}
	return result, nil
}

// DiffEntry 差异条目
type DiffEntry struct {
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// ArtifactVersionService 制品版本服务
type ArtifactVersionService struct {
	repo *repository.ArtifactVersionRepo
}

func NewArtifactVersionService(repo *repository.ArtifactVersionRepo) *ArtifactVersionService {
	return &ArtifactVersionService{repo: repo}
}

func (s *ArtifactVersionService) ListByEnv(envID string) ([]model.ComponentArtifactVersion, error) {
	return s.repo.ListByEnv(envID)
}

func (s *ArtifactVersionService) Create(artifact *model.ComponentArtifactVersion) error {
	return s.repo.Create(artifact)
}

func (s *ArtifactVersionService) Update(artifact *model.ComponentArtifactVersion) error {
	return s.repo.Update(artifact)
}

func (s *ArtifactVersionService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *ArtifactVersionService) CloneArtifacts(fromEnvID, toEnvID string) error {
	return s.repo.CloneArtifacts(fromEnvID, toEnvID)
}

// CloneEnv 环境完整克隆（配置+制品版本）
func CloneEnv(
	configSvc *ConfigService,
	artifactSvc *ArtifactVersionService,
	envComponentSvc *EnvironmentComponentService,
	fromEnvID, toEnvID, operator string,
) error {
	// 1. 克隆环境组件
	if err := envComponentSvc.CloneComponents(fromEnvID, toEnvID); err != nil {
		return fmt.Errorf("克隆环境组件失败: %w", err)
	}

	// 2. 克隆配置
	if err := configSvc.CloneConfigs(fromEnvID, toEnvID, operator); err != nil {
		return fmt.Errorf("克隆配置失败: %w", err)
	}

	// 3. 克隆制品版本
	if err := artifactSvc.CloneArtifacts(fromEnvID, toEnvID); err != nil {
		return fmt.Errorf("克隆制品版本失败: %w", err)
	}

	return nil
}
