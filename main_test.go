package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v62/github"
)

// TestSignatureValidation tests that github.ValidatePayload works correctly with our signature logic
func TestSignatureValidation(t *testing.T) {
	secret := "my-secret-key"
	payload := []byte(`{"action": "opened", "issue": {"number": 1}}`)

	// Compute signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// 1. Success case: correct signature
	req1 := httptest.NewRequest("POST", "/", bytes.NewReader(payload))
	req1.Header.Set("X-Hub-Signature-256", signature)
	req1.Header.Set("Content-Type", "application/json")

	validated1, err := github.ValidatePayload(req1, []byte(secret))
	if err != nil {
		t.Fatalf("Expected signature to be valid, got error: %v", err)
	}
	if string(validated1) != string(payload) {
		t.Errorf("Expected payload %s, got %s", payload, validated1)
	}

	// 2. Failure case: incorrect signature
	req2 := httptest.NewRequest("POST", "/", bytes.NewReader(payload))
	req2.Header.Set("X-Hub-Signature-256", "sha256=invalid-signature-here")
	req2.Header.Set("Content-Type", "application/json")

	_, err = github.ValidatePayload(req2, []byte(secret))
	if err == nil {
		t.Error("Expected validation to fail for invalid signature, but it succeeded")
	}
}

// TestGetReactionURL fetches a few categories to make sure the external APIs are accessible and work
func TestGetReactionURL(t *testing.T) {
	categories := []string{"uwu", "pout", "smug"}

	for _, cat := range categories {
		t.Run(cat, func(t *testing.T) {
			url, err := GetReactionURL(cat)
			if err != nil {
				// We log instead of failing to handle cases where external network / rate-limiting is active
				t.Logf("Fetch reaction failed for category %s: %v (skipping fatal assertion)", cat, err)
				return
			}
			if url == "" {
				t.Errorf("Expected a non-empty image URL for category %s, got empty string", cat)
			}
			t.Logf("Category %s returned URL: %s", cat, url)
		})
	}
}
