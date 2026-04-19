package utils

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
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
		Type:         api.Upload,
	})
	if err != nil {
		// Some Cloudinary accounts classify PDFs as image-based assets.
		// Retry with auto resource_type to avoid false negatives on valid files.
		_, seekErr := src.Seek(0, 0)
		if seekErr != nil {
			return "", fmt.Errorf("cloudinary upload failed: %w", err)
		}

		uploadResult, err = cld.Upload.Upload(ctx, src, uploader.UploadParams{
			Folder:       folder,
			ResourceType: "auto",
			Type:         api.Upload,
		})
	}
	if err != nil {
		return "", fmt.Errorf("cloudinary upload failed: %w", err)
	}
	if uploadResult.Error.Message != "" {
		return "", fmt.Errorf("cloudinary upload failed: %s", uploadResult.Error.Message)
	}
	if uploadResult.Type != api.Upload.String() {
		return "", fmt.Errorf("cloudinary upload invalid delivery (type=%s): expected upload", uploadResult.Type)
	}

	return uploadResult.SecureURL, nil
}
