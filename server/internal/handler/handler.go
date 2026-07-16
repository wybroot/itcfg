package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"itcfg/server/internal/model"
	"itcfg/server/internal/service"
	"itcfg/server/internal/template"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler 统一处理器
type Handler struct {
	customerSvc      *service.CustomerService
	envSvc           *service.EnvService
	componentSvc     *service.ComponentService
	configSvc        *service.ConfigService
	deployRecordSvc  *service.DeployRecordService
	templateEngine   *template.Engine
	packageExporter  *service.PackageExporter
	versionSvc       *service.ConfigVersionService
}

// NewHandler 创建处理器
func NewHandler(
	customerSvc *service.CustomerService,
	envSvc *service.EnvService,
	componentSvc *service.ComponentService,
	configSvc *service.ConfigService,
	deployRecordSvc *service.DeployRecordService,
	templateEngine *template.Engine,
	packageExporter *service.PackageExporter,
	versionSvc *service.ConfigVersionService,
) *Handler {
	return &Handler{
		customerSvc:     customerSvc,
		envSvc:          envSvc,
		componentSvc:    componentSvc,
		configSvc:       configSvc,
		deployRecordSvc: deployRecordSvc,
		templateEngine:  templateEngine,
		packageExporter: packageExporter,
		versionSvc:      versionSvc,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		// 客户管理
		api.GET("/customers", h.ListCustomers)
		api.POST("/customers", h.CreateCustomer)
		api.PUT("/customers/:id", h.UpdateCustomer)
		api.DELETE("/customers/:id", h.DeleteCustomer)

		// 环境管理
		api.GET("/customers/:id/envs", h.ListEnvs)
		api.POST("/customers/:id/envs", h.CreateEnv)
		api.DELETE("/customers/:id/envs/:envId", h.DeleteEnv)

		// 组件管理
		api.GET("/components", h.ListComponents)
		api.GET("/components/:id/variables", h.GetComponentVariables)

		// 配置管理
		api.GET("/envs/:envId/configs", h.GetEnvConfigs)
		api.PUT("/envs/:envId/configs/:componentId", h.UpdateEnvConfigs)
		api.POST("/envs/:envId/configs/preview", h.PreviewConfigs)
		api.POST("/envs/:envId/export", h.ExportPackage)

		// 部署记录
		api.GET("/envs/:envId/deploy-records", h.ListDeployRecords)
		api.POST("/envs/:envId/deploy-records", h.CreateDeployRecord)

		// 配置版本管理
		api.GET("/envs/:envId/versions", h.ListVersions)
		api.POST("/envs/:envId/versions/snapshot", h.CreateSnapshot)
		api.GET("/envs/:envId/versions/diff", h.DiffVersions)
		api.POST("/envs/:envId/versions/rollback", h.RollbackVersion)

		// 配置克隆
		api.POST("/envs/:envId/configs/clone", h.CloneConfigs)

		// Agent 接口
		api.POST("/agent/auth", h.AgentAuth)
		api.GET("/agent/envs/:envKey/configs", h.AgentGetConfigs)
	}
}

// ==================== 客户 Handler ====================

type CreateCustomerRequest struct {
	Name    string `json:"name" binding:"required"`
	Code    string `json:"code" binding:"required"`
	Contact string `json:"contact"`
}

func (h *Handler) ListCustomers(c *gin.Context) {
	customers, err := h.customerSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": customers})
}

func (h *Handler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer := &model.Customer{
		Name:    req.Name,
		Code:    req.Code,
		Contact: req.Contact,
		Status:  "active",
	}

	if err := h.customerSvc.Create(customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": customer})
}

func (h *Handler) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.customerSvc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "客户不存在"})
		return
	}

	existing.Name = req.Name
	existing.Code = req.Code
	existing.Contact = req.Contact

	if err := h.customerSvc.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": existing})
}

func (h *Handler) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	if err := h.customerSvc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 环境 Handler ====================

type CreateEnvRequest struct {
	EnvName     string `json:"env_name" binding:"required"`
	EnvKey      string `json:"env_key" binding:"required"`
	Description string `json:"description"`
}

func (h *Handler) ListEnvs(c *gin.Context) {
	customerID := c.Param("id")
	envs, err := h.envSvc.ListByCustomer(customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": envs})
}

func (h *Handler) CreateEnv(c *gin.Context) {
	customerID := c.Param("id")
	var req CreateEnvRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	env := &model.CustomerEnv{
		CustomerID:  uuid.MustParse(customerID),
		EnvName:     req.EnvName,
		EnvKey:      req.EnvKey,
		Description: req.Description,
	}

	if err := h.envSvc.Create(env); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": env})
}

func (h *Handler) DeleteEnv(c *gin.Context) {
	envID := c.Param("envId")
	if err := h.envSvc.Delete(envID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 组件 Handler ====================

func (h *Handler) ListComponents(c *gin.Context) {
	components, err := h.componentSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": components})
}

func (h *Handler) GetComponentVariables(c *gin.Context) {
	componentID := c.Param("id")
	component, err := h.componentSvc.GetByID(componentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "组件不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": component.Variables})
}

// ==================== 配置 Handler ====================

func (h *Handler) GetEnvConfigs(c *gin.Context) {
	envID := c.Param("envId")
	configs, err := h.configSvc.GetByEnv(envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": configs})
}

type UpdateConfigRequest struct {
	Values    map[string]string `json:"values" binding:"required"`
	UpdatedBy string            `json:"updated_by"`
}

func (h *Handler) UpdateEnvConfigs(c *gin.Context) {
	envID := c.Param("envId")
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.configSvc.BatchUpsert(envID, req.Values, req.UpdatedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置更新成功"})
}

func (h *Handler) PreviewConfigs(c *gin.Context) {
	envID := c.Param("envId")
	var req struct {
		ComponentName string `json:"component_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取配置值
	configs, err := h.configSvc.GetByEnv(envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构建变量值 Map
	values := make(map[string]string)
	for _, cfg := range configs {
		values[cfg.VariableID.String()] = cfg.VarValue
	}

	// 渲染模板
	rendered, err := h.templateEngine.RenderAll(req.ComponentName, values)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rendered})
}

// ==================== 部署包导出 Handler ====================

func (h *Handler) ExportPackage(c *gin.Context) {
	envID := c.Param("envId")
	createdBy := c.Query("created_by")
	if createdBy == "" {
		createdBy = "admin"
	}

	// 导出部署包
	packagePath, meta, err := h.packageExporter.Export(envID, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("导出失败: %v", err)})
		return
	}

	// 设置下载响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(packagePath)))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("X-Package-Checksum", meta.Checksum)
	c.Header("X-Package-Version", meta.Version)

	// 发送文件
	c.File(packagePath)
}

// ==================== 部署记录 Handler ====================

func (h *Handler) ListDeployRecords(c *gin.Context) {
	envID := c.Param("envId")
	records, err := h.deployRecordSvc.ListByEnv(envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": records})
}

type CreateDeployRecordRequest struct {
	VersionTag string `json:"version_tag" binding:"required"`
	DeployedBy string `json:"deployed_by"`
	Status     string `json:"status"`
	Notes      string `json:"notes"`
}

func (h *Handler) CreateDeployRecord(c *gin.Context) {
	envID := c.Param("envId")
	var req CreateDeployRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record := &model.DeployRecord{
		CustomerEnvID: uuid.MustParse(envID),
		VersionTag:    req.VersionTag,
		DeployedBy:    req.DeployedBy,
		Status:        req.Status,
		Notes:         req.Notes,
		DeployedAt:    time.Now(),
	}

	if err := h.deployRecordSvc.Create(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": record})
}

// ==================== Agent 接口 ====================

type AgentAuthRequest struct {
	EnvKey string `json:"env_key" binding:"required"`
}

func (h *Handler) AgentAuth(c *gin.Context) {
	var req AgentAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	env, err := h.envSvc.GetByKey(req.EnvKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的环境密钥"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"env_id":     env.ID,
			"env_name":   env.EnvName,
			"customer_id": env.CustomerID,
			"authenticated": true,
		},
	})
}

// ==================== 配置版本管理 Handler ====================

func (h *Handler) ListVersions(c *gin.Context) {
	envID := c.Param("envId")
	versions, err := h.versionSvc.ListByEnv(envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": versions})
}

type CreateSnapshotRequest struct {
	CreatedBy     string `json:"created_by"`
	ChangeSummary string `json:"change_summary"`
}

func (h *Handler) CreateSnapshot(c *gin.Context) {
	envID := c.Param("envId")
	var req CreateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.CreatedBy == "" {
		req.CreatedBy = "system"
	}

	version, err := h.versionSvc.SaveSnapshot(envID, req.CreatedBy, req.ChangeSummary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": version})
}

func (h *Handler) DiffVersions(c *gin.Context) {
	envID := c.Param("envId")
	fromVersion, _ := strconv.Atoi(c.Query("from"))
	toVersion, _ := strconv.Atoi(c.Query("to"))

	if fromVersion == 0 || toVersion == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请指定 from 和 to 版本号"})
		return
	}

	diff, err := h.versionSvc.Diff(envID, fromVersion, toVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": diff})
}

type RollbackRequest struct {
	TargetVersion int    `json:"target_version" binding:"required"`
	Operator      string `json:"operator"`
}

func (h *Handler) RollbackVersion(c *gin.Context) {
	envID := c.Param("envId")
	var req RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Operator == "" {
		req.Operator = "system"
	}

	if err := h.versionSvc.Rollback(envID, req.TargetVersion, req.Operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("已回滚到版本 %d", req.TargetVersion)})
}

// ==================== 配置克隆 Handler ====================

type CloneConfigsRequest struct {
	FromEnvID string `json:"from_env_id" binding:"required"`
	UpdatedBy string `json:"updated_by"`
}

func (h *Handler) CloneConfigs(c *gin.Context) {
	toEnvID := c.Param("envId")
	var req CloneConfigsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.UpdatedBy == "" {
		req.UpdatedBy = "system"
	}

	// 克隆配置
	if err := h.configSvc.CloneConfigs(req.FromEnvID, toEnvID, req.UpdatedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("克隆失败: %v", err)})
		return
	}

	// 自动保存快照
	_, _ = h.versionSvc.SaveSnapshot(toEnvID, req.UpdatedBy, fmt.Sprintf("从环境 %s 克隆配置", req.FromEnvID))

	c.JSON(http.StatusOK, gin.H{"message": "配置克隆成功"})
}

func (h *Handler) AgentGetConfigs(c *gin.Context) {
	envKey := c.Param("envKey")
	env, err := h.envSvc.GetByKey(envKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的环境密钥"})
		return
	}

	// 获取配置值
	configs, err := h.configSvc.GetByEnv(env.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构建配置值映射
	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.VariableID.String()] = cfg.VarValue
	}

	// 渲染所有组件配置
	allConfigs := make(map[string]map[string]string)
	components, err := h.componentSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, comp := range components {
		variables, err := h.templateEngine.LoadVariables(comp.Name)
		if err != nil {
			continue
		}

		renderValues := make(map[string]string)
		for _, v := range variables.Variables {
			for _, varModel := range comp.Variables {
				if varModel.VarName == v.Name {
					if val, ok := configMap[varModel.ID.String()]; ok && val != "" {
						renderValues[v.Name] = val
					} else {
						renderValues[v.Name] = v.Default
					}
					break
				}
			}
			if _, ok := renderValues[v.Name]; !ok {
				renderValues[v.Name] = v.Default
			}
		}

		rendered, err := h.templateEngine.RenderAll(comp.Name, renderValues)
		if err != nil {
			continue
		}
		allConfigs[comp.Name] = rendered
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"env_id":     env.ID,
			"env_name":   env.EnvName,
			"configs":    allConfigs,
		},
	})
}