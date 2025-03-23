package word

import (
	"context"
	"krikati/src/db"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) categoryExistsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := r.Context().Value("body").(word)

		exists := db.Database.First(&db.Category{}, "id = ?", body.CategoryID)

		if exists.Error != nil {
			http.Error(w, "{\"message\":\"Category with id "+strconv.Itoa(int(body.CategoryID))+" does not exist\"}", http.StatusNotFound)
			return
		}

		next(w, r)
	})
}

func (h *Handler) wordExistMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := strings.Split(r.URL.Path, "/")
		id := url[len(url)-1]

		exists := db.Database.First(&db.Word{}, "id = ?", id)

		if exists.Error != nil {
			http.Error(w, "{\"message\":\"Word with id "+id+" does not exist\"}", http.StatusNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "word_id", id)
		r = r.WithContext(ctx)

		next(w, r)
	})
}

func (h *Handler) attachmentExistsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := strings.Split(r.URL.Path, "/")
		id := url[len(url)-1]

		exists := db.Database.First(&db.Attachment{}, "id = ?", id)

		if exists.Error != nil {
			http.Error(w, "{\"message\":\"Attachment with id "+id+" does not exist\"}", http.StatusNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "attachment_id", id)
		r = r.WithContext(ctx)

		next(w, r)
	})
}
