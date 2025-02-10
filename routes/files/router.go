package files

import (
	"github.com/go-chi/chi/v5"
	"github.com/infinitybotlist/eureka/uapi"
)

const tagName = "Objects"

type Router struct{}

func (b Router) Tag() (string, string) {
	return tagName, "These API endpoints are related to Popkat files"
}

func (b Router) Routes(r *chi.Mux) {
	uapi.Route{
		Pattern: "/objects/list",
		OpId:    "listObjects",
		Method:  uapi.GET,
		Docs:    ListDocs,
		Handler: ListRoute,
	}.Route(r)

	uapi.Route{
		Pattern: "/{file}",
		OpId:    "getFile",
		Method:  uapi.GET,
		Docs:    GetFileDocs,
		Handler: GetRoute,
	}.Route(r)
}
