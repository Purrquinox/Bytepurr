package files

import (
	"context"
	"net/http"
	"time"

	"bytepurr/state"

	"bytepurr/uapi"

	docs "bytepurr/doclib"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

func GetMetadataDocs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get Metadata",
		Description: "Get File Metadata",
		Params: []docs.Parameter{},
		Resp: minio.ObjectInfo{},
	}
}

func GetMetadataRoute(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	ctx := context.Background()
	key := chi.URLParam(r, "*")
	cacheKey := "file_metadata_cache:" + key

	// Check Redis Cache
	cachedData, err := state.Redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		return uapi.HttpResponse{
			Status: http.StatusOK,
			Json:   cachedData,
		}
	}

	// Fetch metadata from the database
	var metadata minio.ObjectInfo
	metadata, err = state.S3.StatObject(
		ctx,
		"bytepurr",
		key,
		minio.StatObjectOptions{},
	)
	if err != nil {
		state.Logger.Error("Failed to fetch metadata from S3", zap.Error(err))
	}

	state.Logger.Info("User metadata retrieved", zap.Any("userMetadata", metadata.UserMetadata))

	// Cache the file in Redis with an expiration time (e.g., 1 hour)
	err = state.Redis.Set(ctx, cacheKey, metadata, time.Hour).Err()
	if err != nil {
		state.Logger.Warn("Failed to cache file in Redis", zap.Error(err))
	}

	// Prepare response with headers
	return uapi.HttpResponse{
		Status: http.StatusOK,
		Json:   metadata,
	}
}
