package auth

import (
	"krikati/src/api"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type admin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h admin) IsEmpty() bool {
	return h.Email == "" && h.Password == ""
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("POST /register", api.ValidateSchemaMiddleware[admin](api.MultipleMiddleware(
		h.register,
		api.RecoveryMiddleware,
		h.emailInIuseMiddleware,
	)))

	r.HandleFunc("POST /login", api.ValidateSchemaMiddleware[admin](api.MultipleMiddleware(
		h.login,
		api.RecoveryMiddleware,
		h.accountExistsMiddleware,
		h.isPasswordValidMiddleware,
	)))

	return r
}
