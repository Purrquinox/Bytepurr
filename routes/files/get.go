package files

import (
	"context"
	"io"
	"net/http"
	"time"

	"bytepurr/constants"
	"bytepurr/state"

	"bytepurr/uapi"

	docs "bytepurr/doclib"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

func GetFileDocs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get File",
		Description: "Render File",
		Params: []docs.Parameter{{
			Name:        "file",
			In:          "path",
			Description: "File Key",
			Required:    true,
			Schema:      docs.IdSchema,
		}},
		Resp: []byte{},
	}
}

func GetRoute(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	ctx := context.Background()
	key := chi.URLParam(r, "file")
	cacheKey := "file_cache:" + key

	// Check Redis Cache
	cachedData, err := state.Redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		return uapi.HttpResponse{
			Status: http.StatusOK,
			Bytes:  cachedData,
			Headers: map[string]string{
				"Content-Type":        http.DetectContentType(cachedData),
				"Content-Disposition": "inline",
			},
		}
	}

	// Get the object from MinIO
	obj, err := state.S3.GetObject(ctx, "bytepurr", key, minio.GetObjectOptions{})
	if err != nil {
		state.Logger.Error("Error fetching object", zap.Error(err))
		return uapi.HttpResponse{
			Status: http.StatusInternalServerError,
			Data:   constants.InternalServerError,
		}
	}
	defer obj.Close()

	// Read the file content
	fileData, err := io.ReadAll(obj)
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			return uapi.HttpResponse{
				Status: http.StatusNotFound,
				Data:   constants.FileNotFound,
			}
		}

		state.Logger.Error("Error reading content", zap.Error(err))
		return uapi.HttpResponse{
			Status: http.StatusInternalServerError,
			Data:   constants.InternalServerError,
		}
	}

	// Cache the file in Redis with an expiration time (e.g., 1 hour)
	err = state.Redis.Set(ctx, cacheKey, fileData, time.Hour).Err()
	if err != nil {
		state.Logger.Warn("Failed to cache file in Redis", zap.Error(err))
	}

	// Prepare response with headers
	return uapi.HttpResponse{
		Status: http.StatusOK,
		Bytes:  fileData,
		Headers: map[string]string{
			"Content-Type":        http.DetectContentType(fileData),
			"Content-Disposition": "inline",
		},
	}
}
