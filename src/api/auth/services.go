package auth

import (
	"krikati/src/api"
	"krikati/src/db"

	"gorm.io/gorm"
)

func createAdmin(name, email, password string) error {
	pass, err := api.HashPassword(password)
	if err != nil {
		return err
	}

	admin := &db.Admin{Name: name, Email: email, Password: pass}

	err = db.Database.Transaction(func(tx *gorm.DB) error {
		return tx.Create(admin).Error
	})

	return err
}
