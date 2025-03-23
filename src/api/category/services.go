package category

import (
	"krikati/src/db"

	"gorm.io/gorm"
)

func (h *Handler) createCategoryService(name string) error {
	category := &db.Category{Name: name}

	err := db.Database.Transaction(func(tx *gorm.DB) error {
		return tx.Create(category).Error
	})

	return err
}

func (h *Handler) getCategoriesService() []db.Category {
	var categories []db.Category

	db.Database.Select("ID", "Name").Find(&categories)

	return categories
}

func (h *Handler) updateCategoryService(id string, name string) error {
	err := db.Database.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&db.Category{}).Where("id = ?", id).UpdateColumn("name", name).Error
	})

	return err
}
