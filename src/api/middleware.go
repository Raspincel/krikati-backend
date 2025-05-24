package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"krikati/src/env"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type Body interface {
	IsEmpty() bool
}

func ValidateSchemaMiddleware[T Body](next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "{\"message\":{\"Invalid request. Body is missing\"}", http.StatusBadRequest)
			return
		}

		var validation T
		err = json.Unmarshal(data, &validation)

		if err != nil {
			http.Error(w, "{\"message\":\"Invalid request. Could not unmarshal body\"}", http.StatusBadRequest)
			return
		}

		if validation.IsEmpty() {
			http.Error(w, "{\"message\":\"Invalid request. Body is empty\"}", http.StatusBadRequest)
			return
		}

		validate := validator.New()
		err = validate.Struct(validation)

		if err != nil {
			fmt.Println(err)
			http.Error(w, "{\"message\":\""+err.Error()+"\"}", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "body", validation)
		r = r.WithContext(ctx)

		restoreBody(r, data)
		next(w, r)
	})
}

type Constructable interface {
	SetValue(field string, value string)
	SetFile(field, value string, file *multipart.FileHeader)
}

func ValidateMultipartSchemaMiddleware[T any, U interface {
	*T
	Constructable
}](expectedFields []string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)

		if r.MultipartForm == nil {
			http.Error(w, "{\"message\":\"Invalid request. Body is missing\"}", http.StatusBadRequest)
			return
		}

		var validation U = new(T)

		data := map[string]string{}

		for k, v := range r.MultipartForm.Value {
			if slices.Contains(expectedFields, k) {
				validation.SetValue(k, v[0])
				continue
			}

			data[k] = v[0]
		}

		for k, v := range r.MultipartForm.File {
			validation.SetFile(k, data[k], v[0])
		}

		validate := validator.New()
		err := validate.Struct(validation)

		if err != nil {
			http.Error(w, "{\"message\":\""+strings.ReplaceAll(err.Error(), "\n", ". ")+"\"}", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "body", *validation)
		r = r.WithContext(ctx)

		next(w, r)
	})
}

func restoreBody(r *http.Request, body []byte) {
	r.Body = io.NopCloser(bytes.NewBuffer(body))
}

func MultipleMiddleware(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	if len(m) < 1 {
		return h
	}

	wrapped := h

	// loop in reverse to preserve middleware order
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}

	return wrapped
}

func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Internal server error:", err)
				http.Error(w, "{\"message\":\"Internal server error\"}", http.StatusInternalServerError)
			}
		}()

		next(w, r)
	})
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if token == "" {
			http.Error(w, "{\"message\":\"Unauthorized. Token is required\"}", http.StatusUnauthorized)
			return
		}

		if len(token) < 7 || token[:7] != "Bearer " {
			http.Error(w, "{\"message\":\"Unauthorized. Token should have \"Bearer\" prefix\"}", http.StatusUnauthorized)
			return
		}

		token = token[7:]

		t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			secret := env.Get("JWT_SECRET", "")
			return []byte(secret), nil
		})

		if !t.Valid || err != nil {
			http.Error(w, "{\"message\":\"Unauthorized\"}", http.StatusUnauthorized)
			return
		}

		next(w, r)
	})
}

func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	})
}
