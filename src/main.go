package main

import (
	"fmt"
	"krikati/src/api/auth"
	"krikati/src/api/category"
	"krikati/src/api/text"
	"krikati/src/api/word"
	"krikati/src/db"
	"krikati/src/env"
	"net/http"
)

func main() {
	db.Connect()
	db.InitializeStorage()

	authHandler := auth.NewHandler()
	categoryHandler := category.NewHandler()
	wordHandler := word.NewHandler()
	textHandler := text.NewHandler()

	mux := http.NewServeMux()
	mux.Handle("/auth/", http.StripPrefix("/auth", authHandler.RegisterRoutes()))
	mux.Handle("/category/", http.StripPrefix("/category", categoryHandler.RegisterRoutes()))
	mux.Handle("/word/", http.StripPrefix("/word", wordHandler.RegisterRoutes()))
	mux.Handle("/text/", http.StripPrefix("/text", textHandler.RegisterRoutes()))

	port := env.Get("PORT", "8080")

	if port == "" {
		port = "8080"
	}

	fmt.Println("Server is running on port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		panic(err)
	}
}
