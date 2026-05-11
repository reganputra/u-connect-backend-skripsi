package utils

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// UploadImage mengunggah file gambar ke Cloudinary (hanya jpg/jpeg/png/webp).
func UploadImage(cld *cloudinary.Cloudinary, file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %w", err)
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

// UploadFile mengunggah file jenis apa pun ke Cloudinary (tanpa batasan format).
func UploadFile(cld *cloudinary.Cloudinary, file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %w", err)
	}
	defer src.Close()

	ctx := context.Background()
	uploadResult, err := cld.Upload.Upload(ctx, src, uploader.UploadParams{
		Folder:       folder,
		ResourceType: "raw",
		Type:         api.Upload,
	})
	if err != nil {
		// Beberapa akun Cloudinary mengklasifikasikan PDF sebagai aset berbasis gambar.
		// Coba lagi dengan resource_type auto untuk menghindari false negatives pada file yang valid.
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
