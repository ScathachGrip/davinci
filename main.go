// Package main provides the entry point, webhook handlers, and utility functions for the DaVinci GitHub Bot.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppVersion defines the current version of the application.
var AppVersion = "1.0.2-alpha"

func main() {
	// Load .env file if it exists (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("[Main] Info: No .env file found, relying on system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appID, webhookSecret, privateKeyBytes := loadConfig()

	// Initialize WebhookServer
	webhookServer, err := NewWebhookServer(appID, webhookSecret, privateKeyBytes)
	if err != nil {
		log.Fatalf("[Main] Failed to initialize webhook server: %v", err)
	}

	// Register route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			resp := fmt.Sprintf(`{"status":"online","version":%q,"message":"DaVinci GitHub Bot is active and running! Send webhook payloads via POST."}`, AppVersion)
			if _, err := w.Write([]byte(resp)); err != nil {
				log.Printf("[Main] Error writing GET response: %v", err)
			}
			return
		}
		if r.Method == http.MethodPost {
			webhookServer.ServeHTTP(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	log.Printf("[Main] Starting DaVinci Bot v%s...", AppVersion)
	log.Printf("[Main] Listening on port %s...", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatalf("[Main] Server failed: %v", err)
	}
}

func loadConfig() (int64, string, []byte) {
	appIDStr := os.Getenv("APP_ID")
	if appIDStr == "" {
		appIDStr = os.Getenv("GITHUB_APP_ID")
	}
	if appIDStr == "" {
		log.Fatal("[Main] APP_ID environment variable is required")
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		log.Fatalf("[Main] APP_ID must be a valid 64-bit integer: %v", err)
	}

	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	if webhookSecret == "" {
		webhookSecret = os.Getenv("GITHUB_WEBHOOK_SECRET")
	}
	if webhookSecret == "" {
		log.Fatal("[Main] WEBHOOK_SECRET environment variable is required")
	}

	privateKeyBytes, err := findPrivateKey()
	if err != nil {
		log.Fatalf("[Main] %v", err)
	}

	return appID, webhookSecret, privateKeyBytes
}

// findPrivateKey searches for the GitHub App private key.
// It checks GITHUB_PRIVATE_KEY_PATH, checks GITHUB_PRIVATE_KEY/PRIVATE_KEY, and then
// searches ".", "/keys", and "/config" for any .pem files (allowing auto-detection in Docker).
func findPrivateKey() ([]byte, error) {
	// 1. Check file path env
	path := os.Getenv("GITHUB_PRIVATE_KEY_PATH")
	if path != "" {
		return os.ReadFile(path)
	}

	// 2. Check raw PEM env
	if key := os.Getenv("PRIVATE_KEY"); key != "" {
		return []byte(key), nil
	}
	if key := os.Getenv("GITHUB_PRIVATE_KEY"); key != "" {
		return []byte(key), nil
	}

	// 3. Search directories for any .pem files (auto-detection style)
	searchDirs := []string{".", "/keys", "/config"}
	for _, dir := range searchDirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip directory if it doesn't exist or has permissions issues
		}
		for _, f := range files {
			if !f.IsDir() && len(f.Name()) > 4 && f.Name()[len(f.Name())-4:] == ".pem" {
				keyPath := dir + "/" + f.Name()
				log.Printf("[Main] Auto-detected private key file: %s", keyPath)
				return os.ReadFile(keyPath)
			}
		}
	}

	return nil, fmt.Errorf("no private key found (place a .pem file in this directory or mount to /keys)")
}
