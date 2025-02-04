package list_objects

import (
	"context"
	"net/http"

	"popkat/state"

	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/minio/minio-go/v7"
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "List Objects",
		Description: "List All Objects",
		Params:      []docs.Parameter{},
		Resp:        []map[string]interface{}{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	objects := []map[string]interface{}{}
	objectCh := state.S3.ListObjects(context.Background(), "popkat", minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return uapi.DefaultResponse(http.StatusInternalServerError)
		}
		objects = append(objects, map[string]interface{}{
			"Key":          object.Key,
			"Size":         object.Size,
			"LastModified": object.LastModified,
			"ETag":         object.ETag,
		})
	}

	return uapi.HttpResponse{
		Json: objects,
	}
}
