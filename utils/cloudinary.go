package utils

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// UploadImage uploads an image file to Cloudinary (jpg/jpeg/png/webp only).
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
	if uploadResult.Error.Message != "" {
		return "", fmt.Errorf("cloudinary upload failed: %s", uploadResult.Error.Message)
	}

	return uploadResult.SecureURL, nil
}

// UploadFile uploads any file type to Cloudinary (no format restriction).
func UploadFile(cld *cloudinary.Cloudinary, file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	ctx := context.Background()
	uploadResult, err := cld.Upload.Upload(ctx, src, uploader.UploadParams{
		Folder:       folder,
		ResourceType: "raw",
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary upload failed: %w", err)
	}
	if uploadResult.Error.Message != "" {
		return "", fmt.Errorf("cloudinary upload failed: %s", uploadResult.Error.Message)
	}

	return uploadResult.SecureURL, nil
}
