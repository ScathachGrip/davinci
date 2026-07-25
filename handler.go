package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v62/github"
)

// WebhookServer handles incoming GitHub webhook requests.
type WebhookServer struct {
	appID         int64
	webhookSecret string
	appsTransport *ghinstallation.AppsTransport
	rng           *rand.Rand
}

// NewWebhookServer initializes a WebhookServer using the App ID, Webhook Secret, and Private Key PEM bytes.
func NewWebhookServer(appID int64, webhookSecret string, privateKey []byte) (*WebhookServer, error) {
	// Initialize AppsTransport once to cache JWT signing setup
	tr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create apps transport: %w", err)
	}

	return &WebhookServer{
		appID:         appID,
		webhookSecret: webhookSecret,
		appsTransport: tr,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// ServeHTTP satisfies the http.Handler interface.
func (s *WebhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(s.webhookSecret))
	if err != nil {
		log.Printf("[Webhook] Payload verification failed: %v", err)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("[Webhook] Parsing payload failed: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Extract Installation ID to authenticate as the target GitHub App Installation
	var installationID int64
	switch e := event.(type) {
	case *github.IssuesEvent:
		if e.Installation != nil {
			installationID = e.Installation.GetID()
		}
	case *github.PullRequestEvent:
		if e.Installation != nil {
			installationID = e.Installation.GetID()
		}
	}

	if installationID == 0 {
		log.Printf("[Webhook] Event has no installation context; skipping")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Authenticate client specifically for this installation
	itr := ghinstallation.NewFromAppsTransport(s.appsTransport, installationID)
	client := github.NewClient(&http.Client{
		Transport: itr,
		Timeout:   15 * time.Second,
	})

	ctx := r.Context()

	switch e := event.(type) {
	case *github.IssuesEvent:
		if e.GetAction() == "opened" {
			go s.handleIssueOpened(ctx, client, e)
		}
	case *github.PullRequestEvent:
		if e.GetAction() == "opened" {
			go s.handlePullRequestOpened(ctx, client, e)
		} else if e.GetAction() == "closed" {
			go s.handlePullRequestClosed(ctx, client, e)
		}
	default:
		log.Printf("[Webhook] Unsupported event type: %s", github.WebHookType(r))
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Event accepted for processing"))
}

func (s *WebhookServer) handleIssueOpened(ctx context.Context, client *github.Client, e *github.IssuesEvent) {
	owner := e.Repo.Owner.GetLogin()
	repo := e.Repo.GetName()
	issueNum := e.Issue.GetNumber()
	sender := e.Sender.GetLogin()

	log.Printf("[Issues] Issue opened in %s/%s #%d by %s", owner, repo, issueNum, sender)

	// Determine reaction category randomly
	categories := []string{"uwu", "uwu", "pout", "smug"}
	cat := categories[s.rng.Intn(len(categories))]

	imgURL, err := GetReactionURL(cat)
	if err != nil {
		log.Printf("[Issues] Error fetching reaction: %v. Using fallback.", err)
		imgURL = "https://i.waifu.pics/k5K7t5t.gif" // A cute static anime image/gif fallback
	}

	thanks := []string{
		"You'll get a response soon! UwU",
		"Thanks for opening this issue, we will get back to you soon! OwO",
		"I see you opened an issue, an admins or collaborators will get back to you soon! UmU",
	}
	thanksMsg := thanks[s.rng.Intn(len(thanks))]

	commentBody := fmt.Sprintf(
		"Hey! @%s %s\n<details>\n<summary>Click here to make your day UmU</summary>\n\n![abc](%s \"UmU\")\n</details>",
		sender, thanksMsg, imgURL,
	)

	// Create issue comment
	_, _, err = client.Issues.CreateComment(ctx, owner, repo, issueNum, &github.IssueComment{
		Body: github.String(commentBody),
	})
	if err != nil {
		log.Printf("[Issues] Error creating comment: %v", err)
	}

	// Add triage label
	_, _, err = client.Issues.AddLabelsToIssue(ctx, owner, repo, issueNum, []string{"triage"})
	if err != nil {
		log.Printf("[Issues] Error adding label: %v", err)
	}
}

func (s *WebhookServer) handlePullRequestOpened(ctx context.Context, client *github.Client, e *github.PullRequestEvent) {
	owner := e.Repo.Owner.GetLogin()
	repo := e.Repo.GetName()
	prNum := e.PullRequest.GetNumber()
	sender := e.Sender.GetLogin()

	log.Printf("[PR] Pull request opened in %s/%s #%d by %s", owner, repo, prNum, sender)

	categories := []string{"pat", "happy", "pat", "nom"}
	cat := categories[s.rng.Intn(len(categories))]

	imgURL, err := GetReactionURL(cat)
	if err != nil {
		log.Printf("[PR] Error fetching reaction: %v. Using fallback.", err)
		imgURL = "https://i.waifu.pics/k5K7t5t.gif"
	}

	thanks := []string{
		"Thanks for making a pull request, an collaborators will get back to you soon! UwU",
		"We will get back to you soon! OwO",
		"What a great improvements, we will consider it! UmU",
	}
	thanksMsg := thanks[s.rng.Intn(len(thanks))]

	commentBody := fmt.Sprintf(
		"Hey?! @%s %s\n<details>\n<summary>Click here to make your day UmU</summary>\n\n![abc](%s \"UmU\")\n</details>",
		sender, thanksMsg, imgURL,
	)

	// Create comment
	_, _, err = client.Issues.CreateComment(ctx, owner, repo, prNum, &github.IssueComment{
		Body: github.String(commentBody),
	})
	if err != nil {
		log.Printf("[PR] Error creating comment: %v", err)
	}

	// Add pr:pending label
	_, _, err = client.Issues.AddLabelsToIssue(ctx, owner, repo, prNum, []string{"pr:pending"})
	if err != nil {
		log.Printf("[PR] Error adding label: %v", err)
	}
}

func (s *WebhookServer) handlePullRequestClosed(ctx context.Context, client *github.Client, e *github.PullRequestEvent) {
	owner := e.Repo.Owner.GetLogin()
	repo := e.Repo.GetName()
	prNum := e.PullRequest.GetNumber()

	log.Printf("[PR] Pull request closed in %s/%s #%d", owner, repo, prNum)

	// Remove pr:pending label
	_, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, prNum, "pr:pending")
	if err != nil {
		log.Printf("[PR] Error removing label 'pr:pending' (might not exist): %v", err)
	}
}
