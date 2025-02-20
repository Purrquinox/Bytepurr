package files

import (
	"context"
	"errors"
	"net/http"
	"time"

	"bytepurr/state"
	"bytepurr/types"

	"bytepurr/uapi"

	docs "bytepurr/doclib"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func GetMetadataDocs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get Metadata",
		Description: "Get File Metadata",
		Params: []docs.Parameter{{
			Name:        "file",
			In:          "path",
			Description: "File Key",
			Required:    true,
			Schema:      docs.IdSchema,
		}},
		Resp: types.Metadata{},
	}
}

func GetMetadataRoute(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	ctx := context.Background()
	key := chi.URLParam(r, "file")
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
	var metadata types.Metadata
	err = state.Pool.QueryRow(
		context.Background(),
		`SELECT key, "userID", platform, "fileType", "fileSize" FROM "Metadata" WHERE key = $1`,
		key,
	).Scan(&metadata.Key, &metadata.UserID, &metadata.Platform, &metadata.FileType, &metadata.FileSize)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			state.Logger.Warn("No metadata found for file", zap.String("key", key))
		} else {
			state.Logger.Error("Error fetching metadata", zap.String("key", key), zap.Error(err))
		}
	}

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
