package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	cld, err := cloudinary.NewFromParams(
    os.Getenv("CLOUDINARY_CLOUD_NAME"),
	os.Getenv("CLOUDINARY_API_KEY"),
	os.Getenv("CLOUDINARY_API_SECRET"),
)
	if err != nil {
		return "", fmt.Errorf("cloudinary config error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	uploadResp, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "properties", 
	})
	if err != nil {
		return "", fmt.Errorf("upload error: %v", err)
	}

	return uploadResp.SecureURL, nil
}

func UploadDamagesToCloudinary(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	cld, err := cloudinary.NewFromParams(
    os.Getenv("CLOUDINARY_CLOUD_NAME"),
	os.Getenv("CLOUDINARY_API_KEY"),
	os.Getenv("CLOUDINARY_API_SECRET"),
)
	if err != nil {
		return "", fmt.Errorf("cloudinary config error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	uploadResp, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "damages", 
	})
	if err != nil {
		return "", fmt.Errorf("upload error: %v", err)
	}

	return uploadResp.SecureURL, nil
}
