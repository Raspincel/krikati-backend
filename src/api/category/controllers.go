package category

import (
	"encoding/json"
	"krikati/src/db"
	"net/http"
)

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	body := r.Context().Value("body").(category)

	category, err := h.createCategoryService(body.Name)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonCategory, err := json.Marshal(category)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonCategory)
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

	category, err := h.updateCategoryService(catID, body.Name)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonCategory, err := json.Marshal(category)
	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonCategory)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	catID := r.Context().Value("category_id").(string)

	err := h.deleteCategoryService(catID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\":\"Category deleted\"}"))
}
