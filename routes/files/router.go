package files

import (
	"popkat/uapi"

	"github.com/go-chi/chi/v5"
)

type Router struct{}

func (b Router) Tag() (string, string) {
	return "Files", "Hello, there. This category of endpoints are to allow users to add new content and view currently uploaded files."
}

func (b Router) Routes(r *chi.Mux) {
	uapi.Route{
		Pattern: "/upload",
		OpId:    "upload",
		Method:  uapi.POST,
		Docs:    UploadDocs,
		Handler: UploadRoute,
	}.Route(r)

	uapi.Route{
		Pattern: "/{file}",
		OpId:    "getFile",
		Method:  uapi.GET,
		Docs:    GetFileDocs,
		Handler: GetRoute,
	}.Route(r)

	uapi.Route{
		Pattern: "/{file}/meta",
		OpId:    "getFileMetadata",
		Method:  uapi.GET,
		Docs:    GetMetadataDocs,
		Handler: GetMetadataRoute,
	}.Route(r)
}
