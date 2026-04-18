package controllers

import (
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/utils"
)

// uploadFilesIfPresent handles multiple file uploads for form-data requests
// Returns a slice of URLs. If no files found or error occurs, returns empty slice or error
func uploadFilesIfPresent(c *fiber.Ctx, fieldName, folder string) ([]string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return []string{}, nil
	}

	files := findMultipartFiles(form, fieldName)
	if len(files) == 0 {
		return []string{}, nil
	}

	var urls []string
	for _, file := range files {
		url, err := utils.UploadImage(config.Cloudinary, file, folder)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func findMultipartFiles(form *multipart.Form, fieldName string) []*multipart.FileHeader {
	if form == nil || form.File == nil {
		return []*multipart.FileHeader{}
	}

	// Priority order avoids accidental duplicates when clients submit multiple
	// equivalent key styles in a single request.
	if files, ok := form.File[fieldName]; ok && len(files) > 0 {
		return files
	}

	if files, ok := form.File[fieldName+"[]"]; ok && len(files) > 0 {
		return files
	}

	files := make([]*multipart.FileHeader, 0)
	seen := make(map[*multipart.FileHeader]struct{})
	for key, fileHeaders := range form.File {
		isIndexed := strings.HasPrefix(key, fieldName+"[") && strings.HasSuffix(key, "]")
		if !isIndexed {
			continue
		}
		for _, fh := range fileHeaders {
			if _, exists := seen[fh]; exists {
				continue
			}
			seen[fh] = struct{}{}
			files = append(files, fh)
		}
	}

	return files
}
