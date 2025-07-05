package text

import (
	"context"
	"krikati/src/db"
	"net/http"
	"strings"
)

func (h *Handler) textExistsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := strings.Split(r.URL.Path, "/")
		id := url[len(url)-1]

		exists := db.Database.First(&db.Text{}, "id = ?", id)

		if exists.Error != nil {
			http.Error(w, "{\"message\":\"Text with id "+id+" does not exist\"}", http.StatusNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "text_id", id)
		r = r.WithContext(ctx)

		next(w, r)
	})
}
