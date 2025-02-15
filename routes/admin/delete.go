package admin

import (
	"net/http"
	"popkat/types"

	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
)

func AdminDelete() *docs.Doc {
	return &docs.Doc{
		Summary:     "Delete File",
		Description: "Delete a file from the container.",
		Params: []docs.Parameter{{
			Name:        "file",
			In:          "path",
			Description: "File Key",
			Required:    true,
			Schema:      docs.IdSchema,
		}},
		Resp: types.Response{},
	}
}

func AdminDeleteRoute(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	//ctx := context.Background()
	//key := chi.URLParam(r, "file")

	// Prepare response with headers
	return uapi.HttpResponse{
		Status: http.StatusOK,
		Json: types.Response{
			Message: "WIP",
		},
	}
}
