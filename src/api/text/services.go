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

	fmt.Println("lol", cover)
	f, err := cover.Open()
	defer f.Close()

	if err != nil {
		fmt.Println("Error opening file", err)
		return "", "", err
	}

	fmt.Println("lmao")
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
	defer tempFile.Close()

	if err != nil {
		fmt.Println("Error creating file", err)
		return "", "", err
	}

	os.WriteFile(tempFile.Name(), bytesContainer, 0644)

	name := strings.TrimPrefix(tempFile.Name(), "/tmp/")
	_, err = db.Storage.Storage.UploadFile(bucket.Id, name, tempFile, storage_go.FileOptions{
		ContentType: &contentType,
	})

	if err != nil {
		fmt.Println("Error uploading file", err)
		return "", "", err
	}

	return name, bucket.Id, nil
}

func (h *Handler) createTextService(data text, coverName, bucketID string) error {
	err := db.Database.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&db.Text{
			Title:    data.Title,
			Subtitle: data.Subtitle,
			Content:  data.Content,
			CoverURL: db.Storage.Storage.GetPublicUrl(bucketID, coverName).SignedURL,
		})

		return err.Error
	})

	return err
}

func (h *Handler) getTextsService() ([]db.Text, error) {
	texts := []db.Text{}
	err := db.Database.Find(&texts)

	return texts, err.Error
}
