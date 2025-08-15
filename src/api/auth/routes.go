package auth

import (
	"krikati/src/api"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type adminRegister struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type adminLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h adminRegister) IsEmpty() bool {
	return h.Email == "" && h.Password == ""
}

func (h adminLogin) IsEmpty() bool {
	return h.Email == "" && h.Password == ""
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("POST /register", api.ValidateSchemaMiddleware[adminRegister](api.MultipleMiddleware(
		h.register,
		api.RecoveryMiddleware,
		h.emailInIuseMiddleware,
		api.AuthMiddleware,
		api.CORSMiddleware,
	)))

	r.HandleFunc("OPTIONS /register", api.MultipleMiddleware(
		func(w http.ResponseWriter, r *http.Request) {},
		api.CORSMiddleware,
	))

	r.HandleFunc("POST /login", api.ValidateSchemaMiddleware[adminLogin](api.MultipleMiddleware(
		h.login,
		api.RecoveryMiddleware,
		h.accountExistsMiddleware,
		h.isPasswordValidMiddleware,
		api.CORSMiddleware,
	)))

	r.HandleFunc("OPTIONS /login", api.MultipleMiddleware(
		func(w http.ResponseWriter, r *http.Request) {},
		api.CORSMiddleware,
	))

	return r
}
