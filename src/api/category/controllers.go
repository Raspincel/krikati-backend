package category

import (
	"encoding/json"
	"krikati/src/db"
	"net/http"
)

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	body := r.Context().Value("body").(category)

	err := h.createCategoryService(body.Name)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("{\"message\":\"Category created\"}"))
}

func (h *Handler) get(w http.ResponseWriter, _ *http.Request) {
	categories := h.getCategoriesService()

	type response struct {
		Message    string        `json:"message"`
		Categories []db.Category `json:"categories"`
	}

	res := response{
		Message:    "Categories retrieved",
		Categories: categories,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	catID := r.Context().Value("category_id").(string)
	body := r.Context().Value("body").(category)

	err := h.updateCategoryService(catID, body.Name)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"message\":\"Category updated\"}"))
}
