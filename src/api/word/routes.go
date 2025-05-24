package word

import (
	"krikati/src/api"
	"mime/multipart"
	"net/http"
	"strconv"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type attachment struct {
	Name   string                `json:"name"`
	Source string                `json:"source"`
	file   *multipart.FileHeader `json:"-"`
}
type word struct {
	Name        string       `json:"name" validate:"required"`
	Meaning     string       `json:"meaning" validate:"required"`
	CategoryID  uint         `json:"category_id" validate:"required"`
	Attachments []attachment `json:"attachments"`
	api.Constructable
}

type updateWord struct {
	Name    string `json:"name,omitempty"`
	Meaning string `json:"meaning,omitempty"`
}

type addAttachment struct {
	attachments []attachment
}

type updateAttachment struct {
	Source string `json:"source" validate:"required"`
}

func (u updateAttachment) IsEmpty() bool {
	return u.Source == ""
}

func (w *addAttachment) SetValue(field, value string) {}

func (w *addAttachment) SetFile(field, value string, file *multipart.FileHeader) {
	w.attachments = append(w.attachments, attachment{
		file:   file,
		Name:   field,
		Source: value,
	})
}

func (h updateWord) IsEmpty() bool {
	return h.Name == "" && h.Meaning == ""
}

func (w *word) SetValue(field, value string) {
	if field == "name" {
		w.Name = value
	}

	if field == "meaning" {
		w.Meaning = value
	}

	if field == "category_id" {
		val, err := strconv.Atoi(value)

		if err != nil {
			return
		}

		w.CategoryID = uint(val)
	}
}

func (w *word) SetFile(field, value string, file *multipart.FileHeader) {
	w.Attachments = append(w.Attachments, attachment{
		file:   file,
		Name:   field,
		Source: value,
	})
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("POST /new", api.ValidateMultipartSchemaMiddleware[word](
		[]string{"name", "meaning", "category_id"},
		api.MultipleMiddleware(
			h.create,
			api.RecoveryMiddleware,
			api.AuthMiddleware,
			h.categoryExistsMiddleware,
			api.CORSMiddleware,
		)))

	r.HandleFunc("OPTIONS /new", api.MultipleMiddleware(
		func(w http.ResponseWriter, r *http.Request) {},
		api.CORSMiddleware,
	))

	r.HandleFunc("GET /all", api.MultipleMiddleware(
		h.get,
		api.RecoveryMiddleware,
		api.CORSMiddleware,
	))

	r.HandleFunc("OPTIONS /all", api.MultipleMiddleware(
		func(w http.ResponseWriter, r *http.Request) {},
		api.CORSMiddleware,
	))

	r.HandleFunc("POST /attachment/{id}", api.ValidateMultipartSchemaMiddleware[addAttachment](
		[]string{},
		api.MultipleMiddleware(
			h.addAttachment,
			api.RecoveryMiddleware,
			api.AuthMiddleware,
			h.wordExistMiddleware,
			api.CORSMiddleware,
		)))

	r.HandleFunc("OPTIONS /attachment/{id}", api.MultipleMiddleware(
		func(w http.ResponseWriter, r *http.Request) {},
		api.CORSMiddleware,
	))

	r.HandleFunc("PUT /details/{id}", api.ValidateSchemaMiddleware[updateWord](api.MultipleMiddleware(
		h.update,
		api.RecoveryMiddleware,
		api.AuthMiddleware,
		h.wordExistMiddleware,
		api.CORSMiddleware,
	)))

	r.HandleFunc("OPTIONS /details/{id}", api.MultipleMiddleware(
		func(w http.ResponseWriter, r *http.Request) {},
		api.CORSMiddleware,
	))

	r.HandleFunc("DELETE /attachment/{id}", api.MultipleMiddleware(
		h.deleteAttachment,
		api.RecoveryMiddleware,
		api.AuthMiddleware,
		h.attachmentExistsMiddleware,
		api.CORSMiddleware,
	))

	r.HandleFunc("OPTIONS /attachment/{id}", api.MultipleMiddleware(
		func(w http.ResponseWriter, r *http.Request) {},
		api.CORSMiddleware,
	))

	r.HandleFunc("PUT /attachment/{id}", api.ValidateSchemaMiddleware[updateAttachment](
		api.MultipleMiddleware(
			h.updateAttachment,
			api.RecoveryMiddleware,
			api.AuthMiddleware,
			h.attachmentExistsMiddleware,
			api.CORSMiddleware,
		),
	))

	r.HandleFunc("OPTIONS /attachment/{id}", api.MultipleMiddleware(
		func(w http.ResponseWriter, r *http.Request) {},
		api.CORSMiddleware,
	))

	return r
}
