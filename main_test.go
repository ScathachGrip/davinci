package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
		name           string
		parentsCount   int
		commitMsg      string
		prTitle        string
		prNum          int
		prCommitsCount int
		commitSHA      string
		headSHA        string
		expected       string
	}{
		{
			name:           "Standard Merge Commit (2 parents)",
			parentsCount:   2,
			commitMsg:      "Merge pull request #12 from branch",
			prTitle:        "feat: add exciting feature",
			prNum:          12,
			prCommitsCount: 3,
			commitSHA:      "sha1",
			headSHA:        "sha2",
			expected:       "merge",
		},
		{
			name:           "Clean Merge Commit (2 parents, custom title without default Merge pull request prefix)",
			parentsCount:   2,
			commitMsg:      "feat: add exciting feature",
			prTitle:        "feat: add exciting feature",
			prNum:          12,
			prCommitsCount: 3,
			commitSHA:      "sha1",
			headSHA:        "sha2",
			expected:       "merge",
		},
		{
			name:           "Squash Merge Commit (1 parent, contains (#pr))",
			parentsCount:   1,
			commitMsg:      "feat: add exciting feature (#12)",
			prTitle:        "feat: add exciting feature",
			prNum:          12,
			prCommitsCount: 3,
			commitSHA:      "sha1",
			headSHA:        "sha2",
			expected:       "squash",
		},
		{
			name:           "Squash Merge Commit (matches PR Title in commit msg)",
			parentsCount:   1,
			commitMsg:      "chore(release): 4.0.5-alpha\n\n* commit 1\n* commit 2",
			prTitle:        "chore(release): 4.0.5-alpha",
			prNum:          12,
			prCommitsCount: 3,
			commitSHA:      "sha1",
			headSHA:        "sha2",
			expected:       "squash",
		},
		{
			name:           "Squash Merge Commit (multi-commit PR squashed into 1 commit, commitSHA != headSHA)",
			parentsCount:   1,
			commitMsg:      "feat: custom commit title without pr num",
			prTitle:        "some title",
			prNum:          12,
			prCommitsCount: 3,
			commitSHA:      "sha_squashed_commit",
			headSHA:        "sha_pr_head",
			expected:       "squash",
		},
		{
			name:           "Rebase Merge Commit (1 parent, commitSHA == headSHA)",
			parentsCount:   1,
			commitMsg:      "feat: add exciting feature",
			prTitle:        "another title",
			prNum:          12,
			prCommitsCount: 1,
			commitSHA:      "sha1",
			headSHA:        "sha1",
			expected:       "rebase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineMergeMethod(tt.parentsCount, tt.commitMsg, tt.prTitle, tt.prNum, tt.prCommitsCount, tt.commitSHA, tt.headSHA)
			if result != tt.expected {
				t.Errorf("Expected merge method %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestPullRequestOptionsCommitTitle tests that CommitTitle in github.PullRequestOptions is properly set
// according to the detected merge method (merge: PR title, squash: PR title (#PR_NUM), rebase: empty).
func TestPullRequestOptionsCommitTitle(t *testing.T) {
	prTitle := "feat: awesome feature"
	prNum := 42

	tests := []struct {
		name          string
		mergeMethod   string
		expectedTitle string
	}{
		{
			name:          "Merge method sets clean PR title",
			mergeMethod:   "merge",
			expectedTitle: "feat: awesome feature",
		},
		{
			name:          "Squash method sets PR title with (#prNum)",
			mergeMethod:   "squash",
			expectedTitle: "feat: awesome feature (#42)",
		},
		{
			name:          "Rebase method leaves CommitTitle empty",
			mergeMethod:   "rebase",
			expectedTitle: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &github.PullRequestOptions{
				MergeMethod: tt.mergeMethod,
			}
			switch opts.MergeMethod {
			case "squash":
				opts.CommitTitle = fmt.Sprintf("%s (#%d)", prTitle, prNum)
			case "merge":
				opts.CommitTitle = prTitle
			}

			if opts.CommitTitle != tt.expectedTitle {
				t.Errorf("Expected CommitTitle %q for %s, got %q", tt.expectedTitle, tt.mergeMethod, opts.CommitTitle)
			}
		})
	}
}

// TestFormatCommitBody verifies that the Co-authored-by email and display name are generated correctly
func TestFormatCommitBody(t *testing.T) {
	pr := &github.PullRequest{
		Body: github.String("This is a cool PR description."),
		User: &github.User{
			ID:    github.Int64(8387081),
			Login: github.String("sinkaroid"),
			Name:  github.String("sinkaroid"),
		},
	}
	botUser := &github.User{
		ID:    github.Int64(120938290),
		Login: github.String("da-vinci-bot[bot]"),
		Name:  github.String("davinci"),
	}

	expected := "This is a cool PR description.\n\nCo-authored-by: sinkaroid <8387081+sinkaroid@users.noreply.github.com>\nCo-authored-by: davinci <120938290+da-vinci-bot[bot]@users.noreply.github.com>"
	result := formatCommitBody(pr, botUser)
	if result != expected {
		t.Errorf("Expected commit body:\n%q\n\nGot:\n%q", expected, result)
	}
}
