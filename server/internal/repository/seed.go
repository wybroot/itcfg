package repository

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"itcfg/server/internal/model"
	ittemplate "itcfg/server/internal/template"

	"gorm.io/gorm"
)

// SeedComponents 从模板目录同步组件与变量定义。
func SeedComponents(db *gorm.DB, templateDir string) error {
	engine := ittemplate.NewEngine(templateDir)
	templateDirs, err := resolveSeedTemplateDirs(templateDir)
	if err != nil {
		return err
	}
	if len(templateDirs) == 0 {
		log.Printf("组件模板同步已跳过")
		return nil
	}

	for _, dir := range templateDirs {
		manifest, err := engine.LoadManifest(dir)
		if err != nil {
			log.Printf("读取组件模板 %s 失败: %v", dir, err)
			continue
		}
		variables, err := engine.LoadVariables(dir)
		if err != nil {
			log.Printf("读取组件变量 %s 失败: %v", dir, err)
			continue
		}

		component := model.Component{
			Name:        manifest.Name,
			DisplayName: manifest.DisplayName,
			Description: manifest.Description,
			Category:    manifest.Category,
			TemplateDir: dir,
			IsActive:    true,
		}

		var existing model.Component
		if err := db.Where("name = ?", manifest.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&component).Error; err != nil {
					log.Printf("插入组件 %s 失败: %v", manifest.Name, err)
					continue
				}
			} else {
				return err
			}
		} else {
			existing.DisplayName = component.DisplayName
			existing.Description = component.Description
			existing.Category = component.Category
			existing.TemplateDir = component.TemplateDir
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				log.Printf("更新组件 %s 失败: %v", manifest.Name, err)
				continue
			}
			component = existing
		}

		for i, v := range variables.Variables {
			variable := model.ComponentVariable{
				ComponentID:    component.ID,
				VarName:        v.Name,
				VarLabel:       v.Label,
				VarType:        normalizeTemplateVarType(v.Type),
				DefaultValue:   v.Default,
				Required:       v.Required,
				ValidationRule: marshalValidationRule(v.Min, v.Max, v.Regex),
				VarGroup:       v.Group,
				SortOrder:      i + 1,
				Description:    v.Description,
				Options:        marshalStringSlice(v.Options),
				LinkedTo:       v.LinkedTo,
			}

			var existing model.ComponentVariable
			err := db.Where("component_id = ? AND var_name = ?", component.ID, v.Name).First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&variable).Error; err != nil {
					log.Printf("插入变量 %s.%s 失败: %v", manifest.Name, v.Name, err)
				}
				continue
			}
			if err != nil {
				return err
			}
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
			if err := db.Save(&existing).Error; err != nil {
				log.Printf("更新变量 %s.%s 失败: %v", manifest.Name, v.Name, err)
			}
		}
	}

	log.Printf("组件模板同步完成: %v", templateDirs)
	return nil
}

func resolveSeedTemplateDirs(templateDir string) ([]string, error) {
	seedMode := strings.TrimSpace(os.Getenv("SEED_TEMPLATES"))
	if seedMode == "" || seedMode == "all" {
		entries, err := os.ReadDir(templateDir)
		if err != nil {
			return nil, err
		}
		dirs := make([]string, 0, len(entries))
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if _, err := os.Stat(filepath.Join(templateDir, entry.Name(), "manifest.yaml")); err == nil {
				dirs = append(dirs, entry.Name())
			}
		}
		sort.Strings(dirs)
		return dirs, nil
	}
	if seedMode == "none" {
		return nil, nil
	}
	parts := strings.Split(seedMode, ",")
	dirs := make([]string, 0, len(parts))
	for _, part := range parts {
		dir := strings.TrimSpace(part)
		if dir != "" {
			dirs = append(dirs, dir)
		}
	}
	return dirs, nil
}

func normalizeTemplateVarType(varType string) string {
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
