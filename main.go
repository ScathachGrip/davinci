// Package main provides the entry point, webhook handlers, and utility functions for the DaVinci GitHub Bot.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v62/github"
	"github.com/joho/godotenv"
)

// AppVersion defines the current version of the application.
var AppVersion = "1.1.3-alpha"

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

	// Log the active repositories being watched in the background at startup
	go logActiveRepositories(webhookServer)

	// Run periodic scanning every 1 hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			log.Println("[Main] Starting periodic active repository scan...")
			logActiveRepositories(webhookServer)
		}
	}()

	// Register route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			name, avatar, htmlURL := webhookServer.GetAppMetadata(r.Context())
			activeRepos := webhookServer.GetActiveRepos()
			html := GetDashboardHTML(AppVersion, name, avatar, htmlURL, activeRepos)
			if _, err := w.Write([]byte(html)); err != nil {
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

// logActiveRepositories lists all repositories accessible to all installations of the GitHub App
// and prints the total count to the log.
func logActiveRepositories(server *WebhookServer) {
	ctx := context.Background()
	appClient := github.NewClient(&http.Client{Transport: server.appsTransport})

	opt := &github.ListOptions{PerPage: 100}
	var totalRepos int
	for {
		installations, resp, err := appClient.Apps.ListInstallations(ctx, opt)
		if err != nil {
			log.Printf("[Main] Error listing installations: %v", err)
			return
		}

		for _, inst := range installations {
			count, err := countInstallationRepos(ctx, server, inst.GetID())
			if err != nil {
				log.Printf("[Main] Error counting repositories for installation %d: %v", inst.GetID(), err)
				continue
			}
			totalRepos += count
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	log.Printf("[Main] Watching %d repository active", totalRepos)
	server.SetActiveRepos(totalRepos)
}

// countInstallationRepos retrieves the count of active repositories for a specific installation.
func countInstallationRepos(ctx context.Context, server *WebhookServer, instID int64) (int, error) {
	itr := ghinstallation.NewFromAppsTransport(server.appsTransport, instID)
	instClient := github.NewClient(&http.Client{Transport: itr})

	var count int
	opt := &github.ListOptions{PerPage: 100}
	for {
		repos, resp, err := instClient.Apps.ListRepos(ctx, opt)
		if err != nil {
			return 0, err
		}
		count += len(repos.Repositories)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return count, nil
}
