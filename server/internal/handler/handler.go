package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"itcfg/server/internal/auth"
	"itcfg/server/internal/model"
	"itcfg/server/internal/notify"
	"itcfg/server/internal/service"
	"itcfg/server/internal/template"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler 统一处理器
type Handler struct {
	customerSvc        *service.CustomerService
	envSvc             *service.EnvService
	envComponentSvc    *service.EnvironmentComponentService
	componentSvc       *service.ComponentService
	configSvc          *service.ConfigService
	deployRecordSvc    *service.DeployRecordService
	templateEngine     *template.Engine
	packageExporter    *service.PackageExporter
	versionSvc         *service.ConfigVersionService
	artifactVersionSvc *service.ArtifactVersionService
	authSvc            *auth.UserService
	notifySvc          *notify.Service
}

// NewHandler 创建处理器
func NewHandler(
	customerSvc *service.CustomerService,
	envSvc *service.EnvService,
	envComponentSvc *service.EnvironmentComponentService,
	componentSvc *service.ComponentService,
	configSvc *service.ConfigService,
	deployRecordSvc *service.DeployRecordService,
	templateEngine *template.Engine,
	packageExporter *service.PackageExporter,
	versionSvc *service.ConfigVersionService,
	artifactVersionSvc *service.ArtifactVersionService,
	authSvc *auth.UserService,
	notifySvc *notify.Service,
) *Handler {
	return &Handler{
		customerSvc:        customerSvc,
		envSvc:             envSvc,
		envComponentSvc:    envComponentSvc,
		componentSvc:       componentSvc,
		configSvc:          configSvc,
		deployRecordSvc:    deployRecordSvc,
		templateEngine:     templateEngine,
		packageExporter:    packageExporter,
		versionSvc:         versionSvc,
		artifactVersionSvc: artifactVersionSvc,
		authSvc:            authSvc,
		notifySvc:          notifySvc,
	}
}

// RegisterRoutes 注册路由
// jwtMiddleware: JWT 认证中间件（nil 表示不启用）
func (h *Handler) RegisterRoutes(r *gin.Engine, jwtMiddleware gin.HandlerFunc) {
	// 公开接口组
	public := r.Group("/api/v1")
	{
		public.POST("/auth/login", h.Login)
		public.POST("/agent/auth", h.AgentAuth)
		public.GET("/agent/envs/:envKey/configs", h.AgentGetConfigs)
		public.POST("/agent/envs/:envKey/deploy-report", h.AgentReportDeploy)
	}

	// 受保护的 API 组
	api := r.Group("/api/v1")
	if jwtMiddleware != nil {
		api.Use(jwtMiddleware)
	}
	{
		// 用户管理
		api.GET("/users", h.ListUsers)
		api.POST("/users", h.CreateUser)
		api.PUT("/users/:id", h.UpdateUser)
		api.DELETE("/users/:id", h.DeleteUser)

		// 仪表盘统计
		api.GET("/dashboard/stats", h.GetDashboardStats)

		// 客户管理
		api.GET("/customers", h.ListCustomers)
		api.POST("/customers", h.CreateCustomer)
		api.PUT("/customers/:id", h.UpdateCustomer)
		api.DELETE("/customers/:id", h.DeleteCustomer)

		// 环境管理
		api.GET("/customers/:id/envs", h.ListEnvs)
		api.POST("/customers/:id/envs", h.CreateEnv)
		api.PUT("/customers/:id/envs/:envId", h.UpdateEnv)
		api.DELETE("/customers/:id/envs/:envId", h.DeleteEnv)

		// 环境组件选择
		api.GET("/envs/:envId/components", h.ListEnvComponents)
		api.PUT("/envs/:envId/components", h.ReplaceEnvComponents)

		// 模板管理
		api.GET("/templates", h.ListTemplates)
		api.GET("/templates/:name/variables", h.GetTemplateVariables)
		api.POST("/templates/sync", h.SyncTemplates)

		// 组件管理
		api.GET("/components", h.ListComponents)
		api.POST("/components", h.CreateComponent)
		api.PUT("/components/:id", h.UpdateComponent)
		api.DELETE("/components/:id", h.DeleteComponent)
		api.GET("/components/:id/variables", h.GetComponentVariables)
		api.POST("/components/:id/variables", h.CreateComponentVariable)
		api.PUT("/components/:id/variables/:variableId", h.UpdateComponentVariable)
		api.DELETE("/components/:id/variables/:variableId", h.DeleteComponentVariable)
		api.POST("/components/:id/variables/import-template", h.ImportComponentVariables)

		// 配置管理
		api.GET("/envs/:envId/configs", h.GetEnvConfigs)
		api.PUT("/envs/:envId/components/:componentId/configs", h.UpdateEnvConfigs)
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

		// 环境完整克隆
		api.POST("/envs/:envId/clone-env", h.CloneEnv)

		// 制品版本管理
		api.GET("/envs/:envId/artifacts", h.ListArtifacts)
		api.POST("/envs/:envId/artifacts", h.CreateArtifact)
		api.PUT("/envs/:envId/artifacts/:id", h.UpdateArtifact)
		api.DELETE("/envs/:envId/artifacts/:id", h.DeleteArtifact)

		// 通知配置
		api.GET("/notify-configs", h.ListNotifyConfigs)
		api.POST("/notify-configs", h.CreateNotifyConfig)
		api.PUT("/notify-configs/:id", h.UpdateNotifyConfig)
		api.DELETE("/notify-configs/:id", h.DeleteNotifyConfig)
		api.POST("/notify-configs/:id/test", h.TestNotifyConfig)
	}
}

// ==================== 仪表盘 Handler ====================

func (h *Handler) GetDashboardStats(c *gin.Context) {
	customers, _ := h.customerSvc.List()
	components, _ := h.componentSvc.List()
	deployRecords, _ := h.deployRecordSvc.ListAll()

	todayCount := 0
	today := time.Now().Format("2006-01-02")
	successCount := 0
	for _, r := range deployRecords {
		if r.DeployedAt.Format("2006-01-02") == today {
			todayCount++
		}
		if r.Status == "success" {
			successCount++
		}
	}
	successRate := 100
	if len(deployRecords) > 0 {
		successRate = successCount * 100 / len(deployRecords)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"customers":    len(customers),
			"components":   len(components),
			"todayDeploys": todayCount,
			"totalDeploys": len(deployRecords),
			"successRate":  successRate,
		},
	})
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

type UpdateEnvRequest struct {
	EnvName     string `json:"env_name"`
	EnvKey      string `json:"env_key"`
	Description string `json:"description"`
}

func (h *Handler) UpdateEnv(c *gin.Context) {
	envID := c.Param("envId")
	env, err := h.envSvc.GetByID(envID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "环境不存在"})
		return
	}
	var req UpdateEnvRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.EnvName != "" {
		env.EnvName = req.EnvName
	}
	if req.EnvKey != "" {
		env.EnvKey = req.EnvKey
	}
	if req.Description != "" {
		env.Description = req.Description
	}
	if err := h.envSvc.Update(env); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": env})
}

// ==================== 环境组件 Handler ====================

type EnvComponentRequest struct {
	ComponentID string `json:"component_id" binding:"required"`
	Enabled     *bool  `json:"enabled"`
	DeployOrder int    `json:"deploy_order"`
}

type ReplaceEnvComponentsRequest struct {
	Components []EnvComponentRequest `json:"components" binding:"required"`
}

func (h *Handler) ListEnvComponents(c *gin.Context) {
	envID := c.Param("envId")
	components, err := h.envComponentSvc.ListByEnv(envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": components})
}

func (h *Handler) ReplaceEnvComponents(c *gin.Context) {
	envID := c.Param("envId")
	if _, err := h.envSvc.GetByID(envID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "环境不存在"})
		return
	}
	var req ReplaceEnvComponentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	components := make([]model.EnvironmentComponent, 0, len(req.Components))
	for i, item := range req.Components {
		componentID, err := uuid.Parse(item.ComponentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的组件 ID"})
			return
		}
		enabled := true
		if item.Enabled != nil {
			enabled = *item.Enabled
		}
		deployOrder := item.DeployOrder
		if deployOrder == 0 {
			deployOrder = i + 1
		}
		components = append(components, model.EnvironmentComponent{
			CustomerEnvID: uuid.MustParse(envID),
			ComponentID:   componentID,
			Enabled:       enabled,
			DeployOrder:   deployOrder,
		})
	}

	if err := h.envComponentSvc.Replace(envID, components); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "环境组件已更新"})
}

// ==================== 模板 Handler ====================

func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.templateEngine.ListTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	components, err := h.componentSvc.List()
	if err == nil {
		componentByTemplate := map[string]model.Component{}
		for _, component := range components {
			componentByTemplate[component.TemplateDir] = component
			componentByTemplate[component.Name] = component
		}
		for i := range templates {
			if component, ok := componentByTemplate[templates[i].TemplateDir]; ok {
				templates[i].Registered = true
				templates[i].RegisteredComponentID = component.ID.String()
				templates[i].Active = component.IsActive
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": templates})
}

func (h *Handler) GetTemplateVariables(c *gin.Context) {
	variables, err := h.templateEngine.LoadVariables(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": variables.Variables})
}

type SyncTemplatesRequest struct {
	Templates []string `json:"templates"`
	Overwrite bool     `json:"overwrite"`
}

func (h *Handler) SyncTemplates(c *gin.Context) {
	var req SyncTemplatesRequest
	_ = c.ShouldBindJSON(&req)
	templates, err := h.templateEngine.ListTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	selected := map[string]bool{}
	for _, name := range req.Templates {
		selected[name] = true
	}
	results := []gin.H{}
	for _, tpl := range templates {
		if len(selected) > 0 && !selected[tpl.TemplateDir] && !selected[tpl.Name] {
			continue
		}
		component, err := h.componentSvc.SyncTemplate(tpl.TemplateDir, req.Overwrite)
		if err != nil {
			results = append(results, gin.H{"template": tpl.TemplateDir, "status": "error", "error": err.Error()})
			continue
		}
		results = append(results, gin.H{"template": tpl.TemplateDir, "status": "synced", "component_id": component.ID})
	}
	c.JSON(http.StatusOK, gin.H{"data": results})
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
	variables, err := h.componentSvc.ListVariables(componentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": variables})
}

type ComponentVariableRequest struct {
	VarName        string `json:"var_name" binding:"required"`
	VarLabel       string `json:"var_label" binding:"required"`
	VarType        string `json:"var_type"`
	DefaultValue   string `json:"default_value"`
	Required       bool   `json:"required"`
	ValidationRule string `json:"validation_rule"`
	VarGroup       string `json:"var_group"`
	SortOrder      int    `json:"sort_order"`
	Description    string `json:"description"`
	Options        string `json:"options"`
	LinkedTo       string `json:"linked_to"`
}

func (req ComponentVariableRequest) toModel() *model.ComponentVariable {
	return &model.ComponentVariable{
		VarName:        req.VarName,
		VarLabel:       req.VarLabel,
		VarType:        req.VarType,
		DefaultValue:   req.DefaultValue,
		Required:       req.Required,
		ValidationRule: req.ValidationRule,
		VarGroup:       req.VarGroup,
		SortOrder:      req.SortOrder,
		Description:    req.Description,
		Options:        req.Options,
		LinkedTo:       req.LinkedTo,
	}
}

func (h *Handler) CreateComponentVariable(c *gin.Context) {
	componentID := c.Param("id")
	var req ComponentVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	variable := req.toModel()
	if err := h.componentSvc.CreateVariable(componentID, variable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": variable})
}

func (h *Handler) UpdateComponentVariable(c *gin.Context) {
	componentID := c.Param("id")
	variableID := c.Param("variableId")
	var req ComponentVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	variable := req.toModel()
	if err := h.componentSvc.UpdateVariable(componentID, variableID, variable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, _ := h.componentSvc.GetVariable(componentID, variableID)
	c.JSON(http.StatusOK, gin.H{"data": updated})
}

func (h *Handler) DeleteComponentVariable(c *gin.Context) {
	componentID := c.Param("id")
	variableID := c.Param("variableId")
	if err := h.componentSvc.DeleteVariable(componentID, variableID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "变量已删除"})
}

type ImportComponentVariablesRequest struct {
	Overwrite bool `json:"overwrite"`
}

func (h *Handler) ImportComponentVariables(c *gin.Context) {
	componentID := c.Param("id")
	var req ImportComponentVariablesRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.componentSvc.ImportVariablesFromTemplate(componentID, req.Overwrite); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	variables, _ := h.componentSvc.ListVariables(componentID)
	c.JSON(http.StatusOK, gin.H{"data": variables})
}

type CreateComponentRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
	TemplateDir string `json:"template_dir"`
	IsActive    bool   `json:"is_active"`
}

func (h *Handler) CreateComponent(c *gin.Context) {
	var req CreateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TemplateDir == "" {
		req.TemplateDir = req.Name
	}
	comp := &model.Component{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Category:    req.Category,
		TemplateDir: req.TemplateDir,
		IsActive:    req.IsActive,
	}
	if err := h.componentSvc.Create(comp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": comp})
}

func (h *Handler) UpdateComponent(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.componentSvc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "组件不存在"})
		return
	}
	var req CreateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.DisplayName != "" {
		existing.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.TemplateDir != "" {
		existing.TemplateDir = req.TemplateDir
	}
	existing.IsActive = req.IsActive

	if err := h.componentSvc.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": existing})
}

func (h *Handler) DeleteComponent(c *gin.Context) {
	id := c.Param("id")
	if err := h.componentSvc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
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
	componentID := c.Param("componentId")
	component, err := h.componentSvc.GetByID(componentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "组件不存在"})
		return
	}
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allowed := make(map[string]bool, len(component.Variables))
	for _, variable := range component.Variables {
		allowed[variable.ID.String()] = true
	}
	for varID := range req.Values {
		if !allowed[varID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "配置项不属于当前组件"})
			return
		}
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
		ComponentID   string            `json:"component_id"`
		ComponentName string            `json:"component_name"`
		Values        map[string]string `json:"values"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	component, err := h.findPreviewComponent(req.ComponentID, req.ComponentName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "组件不存在"})
		return
	}

	// 获取配置值（预览使用解密版本）
	configs, err := h.configSvc.GetByEnvDecrypted(envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构建变量名 -> 配置值 Map，模板渲染使用变量名而不是变量 UUID
	savedValues := make(map[string]string)
	for _, cfg := range configs {
		savedValues[cfg.VariableID.String()] = cfg.VarValue
	}
	for variableID, value := range req.Values {
		savedValues[variableID] = value
	}
	renderValues := make(map[string]string)
	for _, variable := range component.Variables {
		value := variable.DefaultValue
		if saved, ok := savedValues[variable.ID.String()]; ok && saved != "" {
			value = saved
		}
		renderValues[variable.VarName] = value
	}

	templateDir := component.TemplateDir
	if templateDir == "" {
		templateDir = component.Name
	}

	// 渲染模板
	rendered, err := h.templateEngine.RenderAll(templateDir, renderValues)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rendered})
}

func (h *Handler) findPreviewComponent(componentID, componentName string) (*model.Component, error) {
	if componentID != "" {
		return h.componentSvc.GetByID(componentID)
	}
	components, err := h.componentSvc.List()
	if err != nil {
		return nil, err
	}
	for i := range components {
		if components[i].Name == componentName || components[i].TemplateDir == componentName {
			return &components[i], nil
		}
	}
	return nil, fmt.Errorf("component not found")
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
			"env_id":        env.ID,
			"env_name":      env.EnvName,
			"customer_id":   env.CustomerID,
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

// ==================== 环境完整克隆 Handler ====================

type CloneEnvRequest struct {
	FromEnvID string `json:"from_env_id" binding:"required"`
	Operator  string `json:"operator"`
}

func (h *Handler) CloneEnv(c *gin.Context) {
	toEnvID := c.Param("envId")
	var req CloneEnvRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Operator == "" {
		req.Operator = "system"
	}

	// 完整克隆（组件 + 配置 + 制品版本）
	if err := service.CloneEnv(h.configSvc, h.artifactVersionSvc, h.envComponentSvc, req.FromEnvID, toEnvID, req.Operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("环境克隆失败: %v", err)})
		return
	}

	// 自动保存快照
	_, _ = h.versionSvc.SaveSnapshot(toEnvID, req.Operator, fmt.Sprintf("从环境完整克隆 (来源: %s)", req.FromEnvID))

	c.JSON(http.StatusOK, gin.H{"message": "环境克隆成功，配置和制品版本已同步"})
}

// ==================== 制品版本管理 Handler ====================

func (h *Handler) ListArtifacts(c *gin.Context) {
	envID := c.Param("envId")
	artifacts, err := h.artifactVersionSvc.ListByEnv(envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": artifacts})
}

type CreateArtifactRequest struct {
	ComponentID     string `json:"component_id" binding:"required"`
	ArtifactType    string `json:"artifact_type" binding:"required"`
	ArtifactName    string `json:"artifact_name" binding:"required"`
	ArtifactVersion string `json:"artifact_version" binding:"required"`
	RegistryURL     string `json:"registry_url"`
}

func (h *Handler) CreateArtifact(c *gin.Context) {
	envID := c.Param("envId")
	var req CreateArtifactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	artifact := &model.ComponentArtifactVersion{
		CustomerEnvID:   uuid.MustParse(envID),
		ComponentID:     uuid.MustParse(req.ComponentID),
		ArtifactType:    req.ArtifactType,
		ArtifactName:    req.ArtifactName,
		ArtifactVersion: req.ArtifactVersion,
		RegistryURL:     req.RegistryURL,
	}

	if err := h.artifactVersionSvc.Create(artifact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": artifact})
}

func (h *Handler) UpdateArtifact(c *gin.Context) {
	envID := c.Param("envId")
	artifactID := c.Param("id")

	var req CreateArtifactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	artifact := &model.ComponentArtifactVersion{
		BaseModel:       model.BaseModel{ID: uuid.MustParse(artifactID)},
		CustomerEnvID:   uuid.MustParse(envID),
		ComponentID:     uuid.MustParse(req.ComponentID),
		ArtifactType:    req.ArtifactType,
		ArtifactName:    req.ArtifactName,
		ArtifactVersion: req.ArtifactVersion,
		RegistryURL:     req.RegistryURL,
	}

	if err := h.artifactVersionSvc.Update(artifact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": artifact})
}

func (h *Handler) DeleteArtifact(c *gin.Context) {
	artifactID := c.Param("id")
	if err := h.artifactVersionSvc.Delete(artifactID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *Handler) AgentGetConfigs(c *gin.Context) {
	envKey := c.Param("envKey")
	env, err := h.envSvc.GetByKey(envKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的环境密钥"})
		return
	}

	// 获取配置值（Agent 使用解密版本）
	configs, err := h.configSvc.GetByEnvDecrypted(env.ID.String())
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
			"env_id":   env.ID,
			"env_name": env.EnvName,
			"configs":  allConfigs,
		},
	})
}

type AgentDeployReportRequest struct {
	VersionTag string `json:"version_tag" binding:"required"`
	Status     string `json:"status"`
	Notes      string `json:"notes"`
	DeployedBy string `json:"deployed_by"`
}

func (h *Handler) AgentReportDeploy(c *gin.Context) {
	envKey := c.Param("envKey")
	env, err := h.envSvc.GetByKey(envKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的环境密钥"})
		return
	}

	var req AgentDeployReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status == "" {
		req.Status = "success"
	}
	if req.DeployedBy == "" {
		req.DeployedBy = "agent"
	}

	record := &model.DeployRecord{
		CustomerEnvID: env.ID,
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

	// 发送通知
	go func() {
		if record.Status == "success" {
			h.notifySvc.SendDeploySuccess("", env.EnvName, record.VersionTag, record.DeployedBy)
		} else {
			h.notifySvc.SendDeployFailed("", env.EnvName, record.VersionTag, record.DeployedBy, record.Notes)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"data": record, "message": "部署状态已记录"})
}

// ==================== 认证 Handler ====================

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"token":    token,
			"user_id":  user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"role":     user.Role,
		},
	})
}

// ==================== 用户管理 Handler ====================

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.authSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}

	user := &model.User{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Role:     req.Role,
		Status:   "active",
	}

	if err := h.authSvc.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &model.User{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Role:     req.Role,
	}
	// 确保 ID 可解析
	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}
	user.BaseModel.ID = uid

	if err := h.authSvc.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.authSvc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 通知配置 Handler ====================

func (h *Handler) ListNotifyConfigs(c *gin.Context) {
	configs, err := h.notifySvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": configs})
}

type NotifyConfigRequest struct {
	Name       string `json:"name" binding:"required"`
	Type       string `json:"type" binding:"required"`
	WebhookURL string `json:"webhook_url" binding:"required"`
	Secret     string `json:"secret"`
	Events     string `json:"events"`
	IsActive   *bool  `json:"is_active"`
}

func (h *Handler) CreateNotifyConfig(c *gin.Context) {
	var req NotifyConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Events == "" {
		req.Events = "deploy_success,deploy_failed"
	}

	cfg := &model.NotifyConfig{
		Name:       req.Name,
		Type:       req.Type,
		WebhookURL: req.WebhookURL,
		Secret:     req.Secret,
		Events:     req.Events,
		IsActive:   true,
	}
	if req.IsActive != nil {
		cfg.IsActive = *req.IsActive
	}

	if err := h.notifySvc.Create(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": cfg})
}

func (h *Handler) UpdateNotifyConfig(c *gin.Context) {
	id := c.Param("id")
	var req NotifyConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := &model.NotifyConfig{
		Name:       req.Name,
		Type:       req.Type,
		WebhookURL: req.WebhookURL,
		Secret:     req.Secret,
		Events:     req.Events,
		IsActive:   true,
	}
	if req.IsActive != nil {
		cfg.IsActive = *req.IsActive
	}

	if err := h.notifySvc.Update(id, cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cfg})
}

func (h *Handler) DeleteNotifyConfig(c *gin.Context) {
	id := c.Param("id")
	if err := h.notifySvc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *Handler) TestNotifyConfig(c *gin.Context) {
	id := c.Param("id")
	cfg, err := h.notifySvc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知配置不存在"})
		return
	}

	if err := h.notifySvc.SendTestMessage(cfg.WebhookURL, cfg.Type, cfg.Secret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("测试发送失败: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "测试消息发送成功"})
}
