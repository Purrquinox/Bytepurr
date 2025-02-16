package files

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"popkat/state"
	"popkat/types"

	"popkat/uapi"

	docs "popkat/doclib"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type Response struct {
	Key      string `json:"key"`
	UserID   string `json:"userID"`
	Platform string `json:"platform"`
	Type     string `json:"mimetype"`
	Size     int64  `json:"size"`
}

// Generate a unique file key
func generateKey(originalName string) string {
	hasher := md5.New()
	hasher.Write([]byte(originalName))
	hash := hex.EncodeToString(hasher.Sum(nil))

	milliseconds := time.Now().UnixNano() / int64(time.Millisecond)
	ext := filepath.Ext(originalName)
	return fmt.Sprintf("%s_%d%s", hash, milliseconds, ext)
}

func UploadDocs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Upload File",
		Description: "Upload a file to Popkat",
		Params: []docs.Parameter{
			{
				Name:        "userID",
				In:          "header",
				Description: "The User ID of the Platform, this should be a unique identifier to make data requests and deletions easy to process.",
				Required:    true,
				Schema:      docs.IdSchema,
			},
			{
				Name:        "platform",
				In:          "header",
				Description: "The platform that the content is being uploaded for/by.",
				Required:    true,
				Schema:      docs.IdSchema,
			},
		},
		Resp: Response{},
	}
}

func UploadRoute(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	// Get user headers
	userID := r.Header.Get("userID")
	platform := r.Header.Get("platform")

	if userID == "" || platform == "" {
		return uapi.HttpResponse{
			Status: http.StatusBadRequest,
			Json: types.ApiError{
				Message: "User ID and platform headers are required",
			},
		}
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		return uapi.HttpResponse{
			Status: http.StatusBadRequest,
			Json: types.ApiError{
				Message: "Failed to parse multipart form",
			},
		}
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		return uapi.HttpResponse{
			Status: http.StatusBadRequest,
			Json: types.ApiError{
				Message: "Failed to get file from multipart form",
			},
		}
	}
	defer file.Close()

	// Generate file key
	fileKey := generateKey(fileHeader.Filename)
	contentType := fileHeader.Header.Get("Content-Type")

	// Upload to MinIO
	_, err = state.S3.PutObject(context.Background(), "popkat", fileKey, file, fileHeader.Size, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"userID":   userID,
			"platform": platform,
		},
	})
	if err != nil {
		return uapi.HttpResponse{
			Status: http.StatusInternalServerError,
			Json: types.ApiError{
				Message: "Failed to upload file to MinIO",
			},
		}
	}

	// Store file metadata in database
	meta := types.Metadata{
		Key:      fileKey,
		UserID:   userID,
		Platform: platform,
		FileType: contentType,
		FileSize: fileHeader.Size,
	}
	_, err = state.Pool.Exec(
		context.Background(),
		`INSERT INTO "Metadata" (key, "userID", platform, "fileType", "fileSize") VALUES ($1, $2, $3, $4, $5)`,
		meta.Key, meta.UserID, meta.Platform, meta.FileType, meta.FileSize,
	)
	if err != nil {
		state.Logger.Error("Error inserting metadata into database", zap.Error(err))
	}

	// Respond to Request
	return uapi.HttpResponse{
		Status: http.StatusCreated,
		Json: Response{
			Key:      fileKey,
			UserID:   userID,
			Platform: platform,
			Type:     contentType,
			Size:     fileHeader.Size,
		},
	}
}
