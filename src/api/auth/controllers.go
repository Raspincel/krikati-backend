package auth

import (
	"krikati/src/env"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	adm := r.Context().Value("body").(adminRegister)

	err := createAdmin(adm.Name, adm.Email, adm.Password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"account created"}`))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	adm := r.Context().Value("body").(adminLogin)

	token := jwt.New(jwt.SigningMethodHS256)
	token.EncodeSegment([]byte(adm.Email))
	secret := env.Get("JWT_SECRET", "")

	str, err := token.SignedString([]byte(secret))

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\":\"Logged in\", \"token\":\"" + str + "\"}"))
}
