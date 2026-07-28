package main

import (
	"bytes"
	"context"
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
	if !bytes.Equal(validated1, payload) {
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
			url, err := GetReactionURL(context.Background(), cat)
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

// TestDetermineMergeMethod verifies the decision logic for auto-detecting the merge strategy
func TestDetermineMergeMethod(t *testing.T) {
	tests := []struct {
		name         string
		parentsCount int
		commitMsg    string
		prNum        int
		expected     string
	}{
		{
			name:         "Standard Merge Commit (2 parents)",
			parentsCount: 2,
			commitMsg:    "Merge pull request #12 from branch",
			prNum:        12,
			expected:     "merge",
		},
		{
			name:         "Squash Merge Commit (1 parent, contains (#pr))",
			parentsCount: 1,
			commitMsg:    "feat: add exciting feature (#12)",
			prNum:        12,
			expected:     "squash",
		},
		{
			name:         "Rebase Merge Commit (1 parent, original message)",
			parentsCount: 1,
			commitMsg:    "feat: add exciting feature",
			prNum:        12,
			expected:     "rebase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineMergeMethod(tt.parentsCount, tt.commitMsg, tt.prNum)
			if result != tt.expected {
				t.Errorf("Expected merge method %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatCommitBody verifies that the Co-authored-by email and display name are generated correctly
func TestFormatCommitBody(t *testing.T) {
	pr := &github.PullRequest{
		Body: github.String("This is a cool PR description."),
	}
	botUser := &github.User{
		ID:    github.Int64(120938290),
		Login: github.String("da-vinci-bot[bot]"),
		Name:  github.String("davinci"),
	}

	expected := "This is a cool PR description.\n\nCo-authored-by: da-vinci-bot[bot] <120938290+da-vinci-bot[bot]@users.noreply.github.com>"
	result := formatCommitBody(pr, botUser)
	if result != expected {
		t.Errorf("Expected commit body:\n%q\n\nGot:\n%q", expected, result)
	}
}
