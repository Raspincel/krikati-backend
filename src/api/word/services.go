package word

import (
	"fmt"
	"io"
	"krikati/src/db"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	storage_go "github.com/supabase-community/storage-go"
	"gorm.io/gorm"
)

func (h *Handler) uploadFilesService(attachments []attachment) (string, []attachment) {
	bucket, err := db.Storage.Storage.GetBucket("attachments")

	if err != nil {
		bucket, err = db.Storage.Storage.CreateBucket("attachments", storage_go.BucketOptions{
			FileSizeLimit: "50mb",
			Public:        true,
			AllowedMimeTypes: []string{
				"image/jpeg",
				"image/png",
				"image/gif",
				"image/webp",
			},
		})

		if err != nil {
			panic(err.Error())
		}
	}

	for i, att := range attachments {
		f, err := att.file.Open()

		if err != nil {
			fmt.Println("Error opening file", err)
			continue
		}

		buffer := make([]byte, 512)
		_, err = f.Read(buffer)
		if err != nil {
			fmt.Println("Error reading file header", err)
			f.Close()
			continue
		}

		f.Seek(0, 0)

		contentType := http.DetectContentType(buffer)

		extensions := strings.Split(att.file.Filename, ".")
		extension := extensions[len(extensions)-1]

		if contentType == "application/octet-stream" || contentType == "application/json" {
			switch strings.ToLower(extension) {
			case "jpg", "jpeg":
				contentType = "image/jpeg"
			case "png":
				contentType = "image/png"
			case "gif":
				contentType = "image/gif"
			case "webp":
				contentType = "image/webp"
			}
		}

		bytesContainer, err := io.ReadAll(f)
		f.Close()

		if err != nil {
			fmt.Println("Error reading file", err)
			continue
		}

		tempFile, err := os.CreateTemp("", "file-*."+extension)

		if err != nil {
			fmt.Println("Error creating file", err)
			continue
		}

		os.WriteFile(tempFile.Name(), bytesContainer, 0644)

		name := strings.TrimPrefix(tempFile.Name(), "/tmp/")
		_, err = db.Storage.Storage.UploadFile(bucket.Id, name, tempFile, storage_go.FileOptions{
			ContentType: &contentType,
		})

		tempFile.Close()

		if err != nil {
			fmt.Println("Error uploading file", err)
			continue
		}

		attachments[i].Name = name
	}

	return bucket.Id, attachments
}

func (h Handler) deleteFilesService(files []attachment) {
	bucket, err := db.Storage.Storage.GetBucket("attachments")

	if err != nil {
		panic(err.Error())
	}

	for _, file := range files {
		fmt.Println("file to be deleted", file)
		_, err := db.Storage.Storage.RemoveFile(bucket.Id, []string{file.Name})

		if err != nil {
			fmt.Println("Error deleting file", err)
			continue
		}
	}
}

func (h Handler) createWordService(bucketID string, files []attachment, body word) error {
	err := db.Database.Transaction(func(tx *gorm.DB) error {
		w := db.Word{
			Word:       body.Name,
			Meaning:    body.Meaning,
			CategoryID: body.CategoryID,
		}

		err := tx.Create(&w)

		if err.Error != nil {
			return err.Error
		}

		attachments := []db.Attachment{}

		for _, f := range files {
			attachments = append(attachments, db.Attachment{
				URL:    db.Storage.Storage.GetPublicUrl(bucketID, f.Name).SignedURL,
				Source: f.Source,
				WordID: w.ID,
			})
		}

		err = tx.CreateInBatches(&attachments, len(files))

		return err.Error
	})

	return err
}

type full_attachment struct {
	ID     uint   `json:"id"`
	Source string `json:"source"`
	URL    string `json:"url"`
}

type full_data struct {
	ID          uint              `json:"id"`
	Word        string            `json:"word"`
	Meaning     string            `json:"meaning"`
	Attachments []full_attachment `json:"attachments,omitempty"`
}

func (h Handler) getWordsService() map[string][]full_data {
	result := []struct {
		CategoryID   uint   `gorm:"column:category_id"`
		WordID       uint   `gorm:"column:word_id"`
		AttachmentID uint   `gorm:"column:attachment_id"`
		Name         string `gorm:"column:category_name"`
		Word         string `gorm:"column:word"`
		Meaning      string `gorm:"column:meaning"`
		URL          string `gorm:"column:url"`
		Source       string `gorm:"column:source"`
	}{}

	db.Database.Model(&db.Category{}).Select("categories.id as category_id", "categories.name as category_name", "words.word", "words.meaning", "attachments.url", "attachments.source", "words.id as word_id", "attachments.id as attachment_id").Joins("INNER JOIN words ON words.category_id = categories.id").Joins("LEFT JOIN attachments ON attachments.word_id = words.id").Scan(&result)

	d := make(map[string][]full_data)

	for _, r := range result {
		id := strconv.Itoa(int(r.CategoryID))
		if _, ok := d[id]; !ok {
			d[id] = []full_data{}
		}

		exists := slices.ContainsFunc(d[id], func(d full_data) bool {
			return d.Word == r.Word
		})

		if !exists {
			d[id] = append(d[id], full_data{
				ID:          r.WordID,
				Word:        r.Word,
				Meaning:     r.Meaning,
				Attachments: []full_attachment{},
			})
		}

		if r.URL != "" {
			words := d[id]

			wordIndex := slices.IndexFunc(words, func(d full_data) bool {
				return d.Word == r.Word
			})

			words[wordIndex].Attachments = append(words[wordIndex].Attachments, full_attachment{
				ID:     r.AttachmentID,
				URL:    r.URL,
				Source: r.Source,
			})

			d[id] = words
		}
	}

	return d
}

func (h Handler) updateWordService(id, name, meaning string) error {
	newData := map[string]any{}

	if name != "" {
		newData["word"] = name
	}

	if meaning != "" {
		newData["meaning"] = meaning
	}

	err := db.Database.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&db.Word{}).Where("id = ?", id).Updates(newData).Error
	})

	return err
}

func (h Handler) addAttachmentService(id, bucketID string, files []attachment) error {
	err := db.Database.Transaction(func(tx *gorm.DB) error {
		word := db.Word{}
		err := tx.Where("id = ?", id).First(&word).Error

		if err != nil {
			return err
		}

		attachments := []db.Attachment{}

		for _, f := range files {
			attachments = append(attachments, db.Attachment{
				URL:    db.Storage.Storage.GetPublicUrl(bucketID, f.Name).SignedURL,
				Source: f.Source,
				WordID: word.ID,
			})
		}

		err = tx.CreateInBatches(&attachments, len(files)).Error

		return err
	})

	return err
}

func (h Handler) deleteAttachmentService(id string) error {
	err := db.Database.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(db.Attachment{}, id)

		return err.Error
	})

	return err
}

func (h Handler) updateAttachmentService(id, source string) error {
	err := db.Database.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&db.Attachment{}).Where("id = ?", id).Update("source", source)

		return err.Error
	})

	return err
}
