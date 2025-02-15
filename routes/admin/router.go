package admin

import (
	"github.com/go-chi/chi/v5"
	"github.com/infinitybotlist/eureka/uapi"
)

type Router struct{}

func (b Router) Tag() (string, string) {
	return "Admin", "Hello, there. This category of endpoints are to allow platforms or admins to fully access the content, allowing access for deleting content, editing metadata and more."
}

func (b Router) Routes(r *chi.Mux) {
	uapi.Route{
		Pattern: "/admin/{file}/delete",
		OpId:    "adminDelete",
		Method:  uapi.GET,
		Docs:    AdminDelete,
		Handler: AdminDeleteRoute,
	}.Route(r)
}
