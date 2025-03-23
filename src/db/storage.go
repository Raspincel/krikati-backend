package db

import (
	"fmt"
	"krikati/src/env"

	"github.com/supabase-community/supabase-go"
)

var Storage *supabase.Client

func InitializeStorage() {
	apiURL := env.Get("SUPABASE_API_URL", "")
	apiKey := env.Get("SUPABASE_API_KEY", "")
	// supabaseEmail := env.Get("SUPABASE_EMAIL", "")
	// supabasePassword := env.Get("SUPABASE_PASSWORD", "")

	if apiURL == "" || apiKey == "" {
		panic("Missing storage env variables")
	}

	client, err := supabase.NewClient(apiURL, apiKey, &supabase.ClientOptions{})
	if err != nil {
		fmt.Println("cannot initalize client", err)
	}

	Storage = client

	// client.SignInWithEmailPassword(supabaseEmail, supabasePassword)
}
