package text

import (
	"fmt"
	"io"
	"krikati/src/db"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	storage_go "github.com/supabase-community/storage-go"
	"gorm.io/gorm"
)

func (h *Handler) uploadFilesService(cover *multipart.FileHeader) (string, string, error) {
	bucket, err := db.Storage.Storage.GetBucket("texts_covers")

	if err != nil {
		bucket, err = db.Storage.Storage.CreateBucket("texts_covers", storage_go.BucketOptions{
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
			fmt.Println("Error creating bucket", err)
			return "", "", err
		}
	}

	f, err := cover.Open()

	if err != nil {
		fmt.Println("Error opening file", err)
		return "", "", err
	}

	defer f.Close()

	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil {
		fmt.Println("Error reading file header", err)
		return "", "", err
	}

	f.Seek(0, 0)

	contentType := http.DetectContentType(buffer)

	extensions := strings.Split(cover.Filename, ".")
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

	if err != nil {
		fmt.Println("Error reading file", err)
		return "", "", err
	}

	tempFile, err := os.CreateTemp("", "file-*."+extension)

	if err != nil {
		fmt.Println("Error creating file", err)
		return "", "", err
	}

	defer tempFile.Close()

	os.WriteFile(tempFile.Name(), bytesContainer, 0644)

	name := strings.TrimPrefix(tempFile.Name(), "/tmp/")
	client := db.InitializeStorage()
	if client == nil {
		fmt.Println("Storage client is not initialized")
		return "", "", fmt.Errorf("storage client is not initialized")
	}
	_, err = client.Storage.UploadFile(bucket.Id, name, tempFile, storage_go.FileOptions{
		ContentType: &contentType,
	})

	if err != nil {
		fmt.Println("Error uploading file", err)
		return "", "", err
	}

	return name, bucket.Id, nil
}

func (h *Handler) createTextService(data text, coverName, bucketID string) (db.Text, error) {
	text := db.Text{
		Title:    data.Title,
		Subtitle: data.Subtitle,
		Content:  data.Content,
		CoverURL: db.Storage.Storage.GetPublicUrl(bucketID, coverName).SignedURL,
	}

	err := db.Database.Create(&text)

	return text, err.Error
}

func (h *Handler) getTextsService() ([]db.Text, error) {
	texts := []db.Text{}
	err := db.Database.Find(&texts)

	return texts, err.Error
}

func (h *Handler) deleteTextService(id uint) error {

	err := db.Database.Transaction(func(tx *gorm.DB) error {
		var text db.Text
		err := tx.First(&text, "id = ?", id).Error
		if err != nil {
			return err
		}

		// Try to remove the file
		err = h.removeFileService(text.CoverURL)

		if err != nil {
			return fmt.Errorf("failed to remove cover file: %w", err)
		}

		dbErr := tx.Delete(&text).Error
		return dbErr
	})

	return err
}

func (h *Handler) replaceFileService(id string, file *multipart.FileHeader) error {
	if file == nil {
		return nil
	}

	coverName, bucketID, err := h.uploadFilesService(file)
	if err != nil {
		return err
	}

	err = db.Database.Transaction(func(tx *gorm.DB) error {
		var text db.Text
		err := tx.First(&text, "id = ?", id).Error
		if err != nil {
			return err
		}

		err = h.removeFileService(text.CoverURL)

		if err != nil {
			err = h.removeFileService(coverName)
			if err != nil {
				return fmt.Errorf("failed to remove new cover file after failure: %w", err)
			}
			return fmt.Errorf("failed to remove old cover file: %w", err)
		}

		text.CoverURL = db.Storage.Storage.GetPublicUrl(bucketID, coverName).SignedURL
		return tx.Save(&text).Error
	})

	return err
}

func (h *Handler) removeFileService(coverURL string) error {

	if coverURL == "" {
		return fmt.Errorf("cover url cannot be empty")
	}

	parts := strings.Split(coverURL, "/")
	fileName := parts[len(parts)-1]

	if strings.Contains(fileName, "?") {
		fileName = strings.Split(fileName, "?")[0]
	}

	client := db.InitializeStorage()
	if client == nil {
		return fmt.Errorf("storage client is not initialized")
	}
	_, err := client.Storage.RemoveFile("texts_covers", []string{fileName})

	if err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}

func (h *Handler) updateTextService(id string, data updateText) (db.Text, error) {
	var text db.Text

	err := db.Database.Transaction(func(tx *gorm.DB) error {
		err := tx.First(&text, "id = ?", id).Error
		if err != nil {
			return err
		}

		text.Title = data.Title
		text.Subtitle = data.Subtitle
		text.Content = data.Content

		return tx.Save(&text).Error
	})

	if err != nil {
		fmt.Println("Error updating text:", err)
	}

	return text, err
}
