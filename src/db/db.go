package db

import (
	"fmt"
	"krikati/src/env"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Database *gorm.DB

type Admin struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Name     string `json:"name" gorm:"unique"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`

	CreatedAt time.Time `json:"created_at,omitzero"`
}

type Text struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Content  string `json:"content"`
	CoverURL string `json:"cover_url"`
}

type Category struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Name string `gorm:"unique" json:"name"`

	CreatedAt time.Time `json:"created_at,omitzero" `
}

type Attachment struct {
	ID     uint   `gorm:"primarykey" json:"id"`
	URL    string `json:"url"`
	Source string `json:"source"`
	WordID uint   `json:"-"`
	Word   Word   `json:"word" gorm:"foreignKey:WordID;references:id;constraint:OnDelete:CASCADE;"`

	CreatedAt time.Time `json:"created_at,omitzero"`
}

type Word struct {
	ID         uint     `gorm:"primarykey" json:"id"`
	Word       string   `json:"word"`
	Meaning    string   `json:"meaning"`
	Category   Category `json:"category" gorm:"foreignKey:CategoryID;references:ID;constraint:OnDelete:CASCADE;"`
	CategoryID uint     `json:"-" gorm:"OnDelete:CASCADE;"`

	CreatedAt time.Time `json:"created_at,omitzero"`
}

func Connect() {
	dbURL := env.Get("DB_URL", "")

	if dbURL == "" {
		panic("DB_URL is required")
	}

	if dbURL[len(dbURL)-1] != '?' {
		dbURL += "?"
	} else {
		dbURL += "&"
	}
	dbURL += "disable_prepared_statement=true"

	d, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		PrepareStmt: true,
	})

	if err != nil {
		panic(err)
	}
	pgDb, err := d.DB()

	pgDb.SetMaxOpenConns(100)
	pgDb.SetMaxIdleConns(10)
	pgDb.SetConnMaxLifetime(0)

	err = d.AutoMigrate(&Admin{}, &Category{}, &Word{}, &Attachment{}, &Text{})
	// fmt.Println("err", err)

	fmt.Println("Connected to database")
	Database = d
}
