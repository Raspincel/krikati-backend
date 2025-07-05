package text

import (
	"encoding/json"
	"fmt"
	"krikati/src/db"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n=== CREATE START === %s\n", time.Now().Format("15:04:05.000"))
	body := r.Context().Value("body").(text)

	// Log the incoming request
	fmt.Printf("Creating text with title: %s\n", body.Title)

	coverName, bucketID, err := h.uploadFilesService(body.Cover)
	fmt.Printf("Upload complete - coverName: %s, bucketID: %s, err: %v\n", coverName, bucketID, err)

	if err != nil {
		fmt.Printf("Upload failed: %v\n", err)
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	text, err := h.createTextService(body, coverName, bucketID)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	jsonText, err := json.Marshal(text)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(jsonText)
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

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n=== DELETE START === %s\n", time.Now().Format("15:04:05.000"))
	id := r.Context().Value("text_id").(string)
	fmt.Printf("Deleting text with ID: %s\n", id)

	// First, let's see what we're trying to delete
	var text db.Text
	err := db.Database.First(&text, "id = ?", id).Error
	if err != nil {
		fmt.Printf("Failed to find text: %v\n", err)
		http.Error(w, "{\"message\":\"Text not found: "+err.Error()+"\"}", http.StatusNotFound)
		return
	}

	fmt.Printf("Found text to delete: ID=%v, Title=%s, CoverURL=%s\n", text.ID, text.Title, text.CoverURL)

	textID, err := strconv.Atoi(id)
	err = h.deleteTextService(uint(textID))
	fmt.Printf("Delete service complete - err: %v\n", err)

	if err != nil {
		fmt.Printf("Delete service failed: %v\n", err)
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	fmt.Printf("=== DELETE END === %s\n\n", time.Now().Format("15:04:05.000"))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\":\"Text deleted\"}"))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	body := r.Context().Value("body").(updateText)
	id := r.Context().Value("text_id").(string)

	err := h.replaceFileService(id, body.Cover)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	text, err := h.updateTextService(id, body)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	jsonText, err := json.Marshal(text)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonText)
}
