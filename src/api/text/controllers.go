package text

import (
	"encoding/json"
	"krikati/src/db"
	"net/http"
)

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	body := r.Context().Value("body").(text)

	coverName, bucketID, err := h.uploadFilesService(body.Cover)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	err = h.createTextService(body, coverName, bucketID)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("{\"message\":\"Text created\"}"))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	texts, err := h.getTextsService()

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	type response struct {
		Message string    `json:"message"`
		Data    []db.Text `json:"data"`
	}

	res := response{
		Message: "Texts retrieved",
		Data:    texts,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(res)
}
