package auth

import (
	"krikati/src/api"
	"krikati/src/db"

	"gorm.io/gorm"
)

func createAdmin(email, password string) error {
	pass, err := api.HashPassword(password)
	if err != nil {
		return err
	}

	admin := &db.Admin{Email: email, Password: pass}

	err = db.Database.Transaction(func(tx *gorm.DB) error {
		return tx.Create(admin).Error
	})

	return err
}
