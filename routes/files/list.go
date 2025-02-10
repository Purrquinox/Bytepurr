package files

import (
	"context"
	"net/http"
	"time"

	"popkat/state"

	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/minio/minio-go/v7"
)

type Object struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastmodified"`
	ETag         string    `json:"etag"`
}

func ListDocs() *docs.Doc {
	return &docs.Doc{
		Summary:     "List Objects",
		Description: "List All Objects",
		Params:      []docs.Parameter{},
		Resp:        []Object{},
	}
}

func ListRoute(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	objects := []Object{}
	objectCh := state.S3.ListObjects(context.Background(), "popkat", minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return uapi.DefaultResponse(http.StatusInternalServerError)
		}
		objects = append(objects, Object{
			Key:          object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			ETag:         object.ETag,
		})
	}

	return uapi.HttpResponse{
		Json: objects,
	}
}
