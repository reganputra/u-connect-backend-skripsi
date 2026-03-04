package utils

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// UploadImage uploads a multipart file to Cloudinary under the given folder.
// Returns the secure URL of the uploaded image.
func UploadImage(cld *cloudinary.Cloudinary, file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	ctx := context.Background()
	uploadResult, err := cld.Upload.Upload(ctx, src, uploader.UploadParams{
		Folder:         folder,
		AllowedFormats: []string{"jpg", "jpeg", "png", "webp"},
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return uploadResult.SecureURL, nil
}
