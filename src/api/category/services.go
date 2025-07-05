package category

import (
	"krikati/src/db"

	"gorm.io/gorm"
)

func (h *Handler) createCategoryService(name string) (db.Category, error) {
	category := db.Category{Name: name}

	err := db.Database.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&category).Error
	})

	return category, err
}

func (h *Handler) getCategoriesService() []db.Category {
	var categories []db.Category

	db.Database.Select("ID", "Name").Find(&categories)

	return categories
}

func (h *Handler) updateCategoryService(id string, name string) (db.Category, error) {
	var category db.Category
	err := db.Database.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&category, "id = ?", id).Error; err != nil {
			return err
		}

		category.Name = name

		if err := tx.Save(&category).Error; err != nil {
			return err
		}

		return nil
	})

	return category, err
}

func (h *Handler) deleteCategoryService(id string) error {
	return db.Database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&db.Category{}).Error; err != nil {
			return err
		}
		return nil
	})
}
