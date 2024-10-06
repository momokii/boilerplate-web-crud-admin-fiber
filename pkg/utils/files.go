package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	basePath = "./web/uploads/"
	maxSize  = 8 * 1024 * 1024 // 8MB
)

func IsAllowedFileTypes(fileType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"image/gif":       true,
		"application/pdf": true,

		// Word Documents
		"application/msword": true, // .doc
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx

		// Excel Spreadsheets
		"application/vnd.ms-excel": true, // .xls
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true, // .xlsx

		// PowerPoint Presentations
		"application/vnd.ms-powerpoint":                                             true, // .ppt
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true, // .pptx

		// ZIP and RAR Files
		"application/zip":              true, // .zip
		"application/x-zip-compressed": true, // .zip
		"application/x-rar-compressed": true, // .rar
		"application/vnd.rar":          true, // .rar (alternative MIME type)
	}

	return allowedTypes[fileType]
}

func GenerateNameLogsFiles(date string, project_id int) (string, error) {
	uuid := uuid.New().String()
	id := strconv.Itoa(project_id)
	filename := fmt.Sprintf("logs_%s_%s_%s", date, id, uuid)

	return filename, nil
}

func FileUpload(c *fiber.Ctx, path string, filename string) (string, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return "", err
	}

	// check file type
	if !IsAllowedFileTypes(file.Header.Get("Content-Type")) {
		return "", fmt.Errorf("tipe file tidak diizinkan (jpeg, jpg, png, gif, pdf, doc, docx, xls, xlsx, ppt, pptx, zip, rar)")
	}

	// check size
	if file.Size > int64(maxSize) {
		return "", fmt.Errorf("file maksimal berukuran 8MB")
	}

	// change file name
	filename = filename + filepath.Ext(file.Filename)
	filepath := basePath + path + filename // add to file path

	if err := c.SaveFile(file, filepath); err != nil {
		return "", err
	}

	return filepath, nil
}

func DeleteFile(pathfile string) error {
	if err := os.Remove(pathfile); err != nil {
		return err
	}

	return nil
}
