package category

import (
	"context"
	"krikati/src/db"
	"net/http"
	"strings"
)

func (h *Handler) nameUniqueMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := r.Context().Value("body").(category)

		exists := db.Database.First(&db.Category{}, "name = ?", body.Name)

		if exists.Error == nil {
			http.Error(w, "{\"message\":\"Category with this name already exists\"}", http.StatusConflict)
			return
		}

		next(w, r)
	})
}

func (h *Handler) categoryExistsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/")

		if id == "" {
			http.Error(w, "{\"message\":\"ID is required\"}", http.StatusBadRequest)
			return
		}

		exists := db.Database.First(&db.Category{}, "id = ?", id)

		if exists.Error != nil {
			http.Error(w, "{\"message\":\"Category with id "+id+" does not exist\"}", http.StatusNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "category_id", id)
		r = r.WithContext(ctx)

		next(w, r)
	})
}
