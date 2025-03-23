package category

import (
	"krikati/src/api"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type category struct {
	Name string `json:"name" validate:"required"`
}

func (h category) IsEmpty() bool {
	return h.Name == ""
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("POST /new", api.ValidateSchemaMiddleware[category](api.MultipleMiddleware(
		h.create,
		api.RecoveryMiddleware,
		api.AuthMiddleware,
		h.nameUniqueMiddleware,
	)))

	r.HandleFunc("GET /all", api.MultipleMiddleware(
		h.get,
		api.RecoveryMiddleware,
	))

	r.HandleFunc("PUT /{id}", api.ValidateSchemaMiddleware[category](api.MultipleMiddleware(
		h.update,
		api.RecoveryMiddleware,
		api.AuthMiddleware,
		h.categoryExistsMiddleware,
	)))

	return r
}
