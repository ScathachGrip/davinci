// Package main provides the entry point, webhook handlers, and utility functions for the DaVinci GitHub Bot.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// NekosBestResponse represents the JSON response from nekos.best API
type NekosBestResponse struct {
	Results []struct {
		URL string `json:"url"`
	} `json:"results"`
}

// WaifuPicsResponse represents the JSON response from waifu.pics API
type WaifuPicsResponse struct {
	URL string `json:"url"`
}

// GetReactionURL returns a random anime reaction GIF URL for the given category.
// It maps aliases if necessary and supports fallback logic across APIs.
func GetReactionURL(ctx context.Context, category string) (string, error) {
	mappedCategory := category
	if category == "uwu" {
		mappedCategory = "smile"
	}

	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	var url string
	var err error

	if mappedCategory == "pout" {
		// nekos.best supports pout, waifu.pics does not
		url, err = fetchNekosBest(ctx, client, "pout")
	} else {
		// For categories like smug, pat, happy, nom, smile:
		// We try waifu.pics first and fall back to nekos.best.
		url, err = fetchWaifuPics(ctx, client, mappedCategory)
		if err != nil {
			url, err = fetchNekosBest(ctx, client, mappedCategory)
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to fetch reaction for category '%s' (mapped to '%s'): %w", category, mappedCategory, err)
	}

	return url, nil
}

func fetchNekosBest(ctx context.Context, client *http.Client, category string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://nekos.best/api/v2/%s", category), http.NoBody)
	if err != nil {
		return "", err
	}
	// nekos.best requires a descriptive User-Agent
	req.Header.Set("User-Agent", "davinci-bot/1.0 (https://github.com/ScathachGrip/da-vinci)")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Printf("[HTTP] Error closing response body: %v", errClose)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nekos.best returned status %d", resp.StatusCode)
	}

	var apiResp NekosBestResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	if len(apiResp.Results) == 0 {
		return "", fmt.Errorf("nekos.best returned empty results")
	}

	return apiResp.Results[0].URL, nil
}

func fetchWaifuPics(ctx context.Context, client *http.Client, category string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.waifu.pics/sfw/%s", category), http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "davinci-bot/1.0 (https://github.com/ScathachGrip/da-vinci)")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Printf("[HTTP] Error closing response body: %v", errClose)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("waifu.pics returned status %d", resp.StatusCode)
	}

	var apiResp WaifuPicsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	if apiResp.URL == "" {
		return "", fmt.Errorf("waifu.pics returned empty URL")
	}

	return apiResp.URL, nil
}
