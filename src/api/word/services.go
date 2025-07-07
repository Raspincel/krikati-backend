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
	"time"

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

type fullWord struct {
	ID          uint            `json:"id"`
	Word        string          `json:"word"`
	Meaning     string          `json:"meaning"`
	Category    db.Category     `json:"category"`
	Attachments []db.Attachment `json:"attachments,omitempty"`
	Translation string          `json:"translation,omitempty"`
	CreatedAt   time.Time       `json:"created_at,omitempty"`
}

func (h Handler) createWordService(bucketID string, files []attachment, body word) (fullWord, error) {
	w := db.Word{
		Word:        body.Name,
		Meaning:     body.Meaning,
		CategoryID:  body.CategoryID,
		Translation: body.Translation,
	}

	err := db.Database.Transaction(func(tx *gorm.DB) error {

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

	if err != nil {
		return fullWord{}, fmt.Errorf("error creating word: %s", err.Error())
	}

	var category db.Category
	err = db.Database.First(&category, "id = ?", w.CategoryID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fullWord{}, fmt.Errorf("category with ID %d not found", w.CategoryID)
		}
		return fullWord{}, fmt.Errorf("error finding category: %w", err)
	}

	var attachments []db.Attachment

	err = db.Database.Model(&db.Attachment{}).Where("word_id = ?", w.ID).Select("id", "url", "source").Find(&attachments).Error

	if err != nil {
		return fullWord{}, fmt.Errorf("error finding attachments: %w", err)
	}

	fullWord := fullWord{
		ID:          w.ID,
		Word:        w.Word,
		Meaning:     w.Meaning,
		Category:    category,
		Translation: w.Translation,
		CreatedAt:   w.CreatedAt,
		Attachments: attachments,
	}

	return fullWord, err
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
	Translation string            `json:"translation,omitempty"`
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
		Translation  string `gorm:"column:translation"`
		URL          string `gorm:"column:url"`
		Source       string `gorm:"column:source"`
	}{}

	db.Database.Model(&db.Category{}).Select("categories.id as category_id", "categories.name as category_name", "words.word", "words.meaning", "attachments.url", "attachments.source", "words.id as word_id", "attachments.id as attachment_id", "words.translation").Joins("INNER JOIN words ON words.category_id = categories.id").Joins("LEFT JOIN attachments ON attachments.word_id = words.id").Scan(&result)

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
				Translation: r.Translation,
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

func (h Handler) updateWordService(id string, w updateWord) (fullWord, error) {
	var word db.Word

	err := db.Database.Transaction(func(tx *gorm.DB) error {
		err := tx.First(&word, "id = ?", id).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("word with ID %s not found", id)
			}
			return fmt.Errorf("error finding word: %w", err)
		}

		if w.Name != "" {
			word.Word = w.Name
		}

		if w.Meaning != "" {
			word.Meaning = w.Meaning
		}

		word.Translation = w.Translation
		word.CategoryID = w.CategoryID

		err = tx.Save(&word).Error
		if err != nil {
			return fmt.Errorf("error updating word: %w", err)
		}

		return nil
	})

	if err != nil {
		return fullWord{}, fmt.Errorf("error updating word: %s", err.Error())
	}

	var category db.Category

	err = db.Database.First(&category, "id = ?", word.CategoryID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fullWord{}, fmt.Errorf("category with ID %d not found", word.CategoryID)
		}
		return fullWord{}, fmt.Errorf("error finding category: %w", err)
	}

	var attachments []db.Attachment

	err = db.Database.Model(&db.Attachment{}).Where("word_id = ?", word.ID).Select("id", "url", "source").Find(&attachments).Error

	if err != nil {
		return fullWord{}, fmt.Errorf("error finding attachments: %w", err)
	}

	var fullWord = fullWord{
		ID:          word.ID,
		Word:        word.Word,
		Meaning:     word.Meaning,
		Category:    category,
		Translation: word.Translation,
		CreatedAt:   word.CreatedAt,
		Attachments: attachments,
	}

	return fullWord, err
}

func (h Handler) addAttachmentService(id, bucketID string, files []attachment) ([]db.Attachment, error) {
	attachments := []db.Attachment{}
	err := db.Database.Transaction(func(tx *gorm.DB) error {
		word := db.Word{}
		err := tx.Where("id = ?", id).First(&word).Error

		if err != nil {
			return err
		}

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

	return attachments, err
}

func (h Handler) deleteAttachmentService(id string) error {
	var attachment db.Attachment
	err := db.Database.First(&attachment, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("attachment with ID %s not found", id)
		}
		return fmt.Errorf("error finding attachment: %w", err)
	}

	// Delete the file from storage
	fmt.Println("attachment bah", attachment)
	if attachment.URL != "" {

		client := db.InitializeStorage()
		bucket, err := client.Storage.GetBucket("attachments")
		if err != nil {
			return fmt.Errorf("error getting bucket: %w", err)
		}

		parts := strings.Split(attachment.URL, "/")
		fileName := parts[len(parts)-1]

		_, err = client.Storage.RemoveFile(bucket.Id, []string{fileName})

		if err != nil {
			return fmt.Errorf("error deleting file from storage: %w", err)
		}
	}

	err = db.Database.Transaction(func(tx *gorm.DB) error {

		err := tx.Delete(db.Attachment{}, id)

		return err.Error
	})

	return err
}

func (h Handler) updateAttachmentService(id, source string) (db.Attachment, error) {
	var attachment db.Attachment

	err := db.Database.Transaction(func(tx *gorm.DB) error {
		err := tx.First(&attachment, "id = ?", id).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("attachment with ID %s not found", id)
			}
			return fmt.Errorf("error finding attachment: %w", err)
		}

		if source != "" {
			attachment.Source = source
		}

		err = tx.Save(&attachment).Error
		if err != nil {
			return fmt.Errorf("error updating attachment: %w", err)
		}

		return nil
	})

	return attachment, err
}

func (h Handler) deleteWordService(id string) error {
	wordID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid word ID: %w", err)
	}

	err = db.Database.Transaction(func(tx *gorm.DB) error {
		var word db.Word
		err := tx.First(&word, "id = ?", wordID).Error
		if err != nil {
			return fmt.Errorf("word not found: %w", err)
		}

		fmt.Println("word id", wordID)

		// Delete attachments associated with the word
		var attachments []db.Attachment
		err = tx.Where("word_id = ?", wordID).Find(&attachments).Error
		if err != nil {
			return fmt.Errorf("failed to find attachments: %w", err)
		}

		fmt.Println("attachments", len(attachments))

		for _, attachment := range attachments {
			fmt.Println("what", attachment)
			err = h.deleteAttachmentService(strconv.Itoa(int(attachment.ID)))

			if err != nil {
				return fmt.Errorf("failed to delete attachment %d: %w", attachment.ID, err)
			}
		}

		err = tx.Delete(&word).Error
		if err != nil {
			return fmt.Errorf("failed to delete word: %w", err)
		}

		return nil
	})

	return err
}
