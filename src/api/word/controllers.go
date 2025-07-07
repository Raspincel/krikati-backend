package word

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	body := r.Context().Value("body").(word)

	bucketID, files := h.uploadFilesService(body.Attachments)
	word, err := h.createWordService(bucketID, files, body)

	if err != nil {
		h.deleteFilesService(files)
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	wordJson, err := json.Marshal(word)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(wordJson)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	d := h.getWordsService()

	type response struct {
		Message string                 `json:"message"`
		Data    map[string][]full_data `json:"data"`
	}

	res := response{
		Message: "Words retrieved",
		Data:    d,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) addAttachment(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("word_id").(string)
	body := r.Context().Value("body").(addAttachment)

	bucketID, files := h.uploadFilesService(body.attachments)

	attachments, err := h.addAttachmentService(id, bucketID, files)

	if err != nil {
		h.deleteFilesService(files)
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	attachmentsJson, err := json.Marshal(attachments)
	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(attachmentsJson)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("word_id").(string)
	body := r.Context().Value("body").(updateWord)

	word, err := h.updateWordService(id, body)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	wordJson, err := json.Marshal(word)
	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(wordJson)
}

func (h *Handler) deleteAttachment(w http.ResponseWriter, r *http.Request) {
	err := h.deleteAttachmentService(r.Context().Value("attachment_id").(string))

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\":\"Attachment deleted\"}"))
}

func (h *Handler) updateAttachment(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("attachment_id").(string)
	body := r.Context().Value("body").(updateAttachment)

	attachment, err := h.updateAttachmentService(id, body.Source)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	attachmentJson, err := json.Marshal(attachment)
	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(attachmentJson)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("word_id").(string)

	err := h.deleteWordService(id)

	if err != nil {
		http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\":\"Word deleted\"}"))
}
