package auth

import (
	"context"
	"fmt"
	"krikati/src/api"
	"krikati/src/db"
	"net/http"
)

func (h *Handler) emailInIuseMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := r.Context().Value("body").(admin)

		exists := db.Database.First(&db.Admin{}, "email = ?", body.Email)

		if exists.Error == nil {
			http.Error(w, "{\"message\":\"Email already in use\"}", http.StatusBadRequest)
			return
		}

		next(w, r)
	})
}

func (h *Handler) accountExistsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("a")
		body := r.Context().Value("body").(admin)

		var admin db.Admin
		fmt.Println("oxi", body.Email)
		exists := db.Database.First(&admin, "email = ?", body.Email)
		fmt.Println("hm", exists.Error)

		if exists.Error != nil {
			http.Error(w, "{\"message\":\"Account registered with this email does not exist\"}", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "adm", admin)
		r = r.WithContext(ctx)

		next(w, r)
	})
}

func (h *Handler) isPasswordValidMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := r.Context().Value("body").(admin)
		adm := r.Context().Value("adm").(db.Admin)

		isEqual := api.VerifyPassword(body.Password, adm.Password)

		if !isEqual {
			http.Error(w, "{\"message\":\"Invalid password\"}", http.StatusBadRequest)
			return
		}

		next(w, r)
	})
}
