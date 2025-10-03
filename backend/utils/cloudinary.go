package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func getCloudinaryInstance() (*cloudinary.Cloudinary, error) {
	return cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
}

// ✅ Upload to "events" folder
func UploadToCloudinary(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	cld, err := getCloudinaryInstance()
	if err != nil {
		return "", fmt.Errorf("cloudinary config error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	uploadResp, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "events",
	})
	if err != nil {
		return "", fmt.Errorf("upload error: %v", err)
	}

	return uploadResp.SecureURL, nil
}

// ✅ Upload to "damages" folder
func UploadDamagesToCloudinary(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	cld, err := getCloudinaryInstance()
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

// ✅ Delete image from Cloudinary using full URL
func DeleteFromCloudinary(imageURL string) error {
	cld, err := getCloudinaryInstance()
	if err != nil {
		return fmt.Errorf("cloudinary config error: %v", err)
	}

	// Extract public ID from URL
	publicID, err := extractPublicID(imageURL)
	if err != nil {
		return fmt.Errorf("could not extract public ID: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("delete error: %v", err)
	}

	return nil
}

// 🔹 Helper: Extract Cloudinary public ID from full URL
func extractPublicID(imageURL string) (string, error) {
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return "", err
	}

	// Example: https://res.cloudinary.com/demo/image/upload/v1234567890/events/abc123.jpg
	parts := strings.Split(parsedURL.Path, "/")

	if len(parts) < 3 {
		return "", fmt.Errorf("invalid cloudinary URL format")
	}

	// Remove version part (e.g., v1234567890)
	cleanPath := parts[len(parts)-2:]
	if strings.HasPrefix(cleanPath[0], "v") {
		parts = append(parts[:len(parts)-2], parts[len(parts)-1])
	}

	// Get everything after /upload/ (folder + filename without extension)
	publicID := strings.TrimSuffix(path.Join(parts[3:]...), path.Ext(parts[len(parts)-1]))

	return publicID, nil
}
