package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var AuthClient *auth.Client

func InitFirebase() {
	// 1. Check for encoded credentials in Env Var (Best for Railway)
	cwd, _ := os.Getwd()
	fmt.Printf("DATA DEBUG: CWD: %s\n", cwd)
	files, _ := os.ReadDir(".")
	for _, f := range files {
		fmt.Printf("DATA DEBUG: Found File: %s\n", f.Name())
	}
	credsBase64 := os.Getenv("FIREBASE_CREDENTIALS_BASE64")
	if credsBase64 == "" {
		fmt.Println("DATA DEBUG: Env Var FIREBASE_CREDENTIALS_BASE64 is EMPTY/MISSING")
	} else {
		fmt.Printf("DATA DEBUG: Env Var FIREBASE_CREDENTIALS_BASE64 is FOUND (Len: %d)\n", len(credsBase64))
	}

	var opt option.ClientOption

	if credsBase64 != "" {
		// Decode base64 string
		credsJSON, err := base64.StdEncoding.DecodeString(credsBase64)
		if err != nil {
			log.Fatalf("Failed to decode FIREBASE_CREDENTIALS_BASE64: %v", err)
		}
		opt = option.WithCredentialsJSON(credsJSON)
		fmt.Println("🔥 Firebase initialized using Environment Variable")
	} else {
		// 2. Fallback to local file (for local dev)
		opt = option.WithCredentialsFile("serviceAccountKey.json")
		fmt.Println("🔥 Firebase initialized using local file")
	}

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	AuthClient = client
}
