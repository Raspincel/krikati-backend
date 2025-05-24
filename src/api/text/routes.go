package text

import (
	"krikati/src/api"
	"mime/multipart"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type text struct {
	Title    string                `json:"title" validate:"required"`
	Subtitle string                `json:"subtitle" validate:"required"`
	Content  string                `json:"content" validate:"required"`
	Cover    *multipart.FileHeader `json:"cover" validate:"required"`
	api.Constructable
}

func (t text) IsEmpty() bool {
	return t.Title == "" && t.Subtitle == "" && t.Content == ""
}

func (t *text) SetValue(field, value string) {
	switch field {
	case "title":
		t.Title = value
	case "subtitle":
		t.Subtitle = value
	case "content":
		t.Content = value
	}
}

func (t *text) SetFile(field, value string, file *multipart.FileHeader) {
	if field == "cover" {
		t.Cover = file
	}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("POST /", api.ValidateMultipartSchemaMiddleware[text](
		[]string{"title", "subtitle", "content", "cover"},
		api.MultipleMiddleware(
			h.create,
			api.RecoveryMiddleware,
			api.AuthMiddleware,
			api.CORSMiddleware,
		)))

	r.HandleFunc("OPTIONS /", api.MultipleMiddleware(
		func(w http.ResponseWriter, r *http.Request) {},
		api.CORSMiddleware,
	))

	r.HandleFunc("GET /", api.MultipleMiddleware(
		h.get,
		api.RecoveryMiddleware,
		api.CORSMiddleware,
	))

	return r
}
