package objects

import (
	list_objects "popkat/routes/objects/endpoints"

	"github.com/go-chi/chi/v5"
	"github.com/infinitybotlist/eureka/uapi"
)

const tagName = "Objects"

type Router struct{}

func (b Router) Tag() (string, string) {
	return tagName, "These API endpoints are related to Popkat Objects"
}

func (b Router) Routes(r *chi.Mux) {
	uapi.Route{
		Pattern: "/objects/list",
		OpId:    "listObjects",
		Method:  uapi.GET,
		Docs:    list_objects.Docs,
		Handler: list_objects.Route,
	}.Route(r)
}
