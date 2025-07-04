package models

import (
	"gorm.io/gorm"
)

// SeedData 初始化种子数据
func SeedData(db *gorm.DB) error {
	// 创建默认知识域
	var count int64
	db.Model(&Domain{}).Count(&count)
	if count == 0 {
		defaultDomains := []Domain{
			{
				DomainName:  "general",
				Description: "通用知识域",
				Config: map[string]interface{}{
					"vector_dimension": 1536,
					"index_type":       "HNSW",
					"metric_type":      "IP",
				},
			},
			{
				DomainName:  "code-review",
				Description: "代码评审领域",
				Config: map[string]interface{}{
					"vector_dimension": 1536,
					"index_type":       "HNSW",
					"metric_type":      "IP",
				},
			},
			{
				DomainName:  "documentation",
				Description: "文档管理领域",
				Config: map[string]interface{}{
					"vector_dimension": 1536,
					"index_type":       "HNSW",
					"metric_type":      "IP",
				},
			},
		}

		for _, domain := range defaultDomains {
			db.Create(&domain)
		}
	}

	// 创建默认用户
	db.Model(&User{}).Count(&count)
	if count == 0 {
		defaultUsers := []User{
			{
				UserID:   "admin",
				Username: "Administrator",
				Email:    "admin@ino.com",
				Preferences: map[string]interface{}{
					"theme":    "light",
					"language": "zh",
				},
			},
			{
				UserID:   "system",
				Username: "System User",
				Email:    "system@ino.com",
				Preferences: map[string]interface{}{
					"theme":    "light",
					"language": "zh",
				},
			},
		}

		for _, user := range defaultUsers {
			db.Create(&user)
		}
	}

	return nil
}
