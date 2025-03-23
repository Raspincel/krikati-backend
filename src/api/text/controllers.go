package text

import (
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
