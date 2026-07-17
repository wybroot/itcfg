package repository

import (
	"encoding/json"
	"log"

	"itcfg/server/internal/model"
	ittemplate "itcfg/server/internal/template"

	"gorm.io/gorm"
)

var mvpTemplateDirs = []string{"postgresql", "java-app", "nginx"}

// SeedComponents 从模板目录同步 MVP 组件与变量定义。
func SeedComponents(db *gorm.DB, templateDir string) error {
	engine := ittemplate.NewEngine(templateDir)

	for _, dir := range mvpTemplateDirs {
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

		if err := db.Where("component_id = ?", component.ID).Delete(&model.ComponentVariable{}).Error; err != nil {
			return err
		}

		for i, v := range variables.Variables {
			varType := v.Type
			if varType == "boolean" {
				varType = "bool"
			}
			variable := model.ComponentVariable{
				ComponentID:    component.ID,
				VarName:        v.Name,
				VarLabel:       v.Label,
				VarType:        varType,
				DefaultValue:   v.Default,
				Required:       v.Required,
				ValidationRule: marshalValidationRule(v.Min, v.Max, v.Regex),
				VarGroup:       v.Group,
				SortOrder:      i + 1,
				Description:    v.Description,
				Options:        marshalStringSlice(v.Options),
				LinkedTo:       v.LinkedTo,
			}
			if err := db.Create(&variable).Error; err != nil {
				log.Printf("插入变量 %s.%s 失败: %v", manifest.Name, v.Name, err)
			}
		}
	}

	log.Printf("MVP 组件模板同步完成: %v", mvpTemplateDirs)
	return nil
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
