// Package main provides the entry point, webhook handlers, and utility functions for the DaVinci GitHub Bot.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v62/github"
)

// WebhookServer handles incoming GH webhook requests.
type WebhookServer struct {
	appID         int64
	webhookSecret string
	appsTransport *ghinstallation.AppsTransport
	rng           *rand.Rand
	activeReposMu sync.RWMutex
	activeRepos   int
	botUserMu     sync.Mutex
	botUser       *github.User
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
	_, event, err := s.parsePayload(r)
	if err != nil {
		log.Printf("[Webhook] Payload verification/parsing failed: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	installationID := s.getInstallationID(event)
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

	s.routeEvent(client, event, r)

	w.WriteHeader(http.StatusAccepted)
	if _, err := w.Write([]byte("Event accepted for processing")); err != nil {
		log.Printf("[Webhook] Error writing response: %v", err)
	}
}

func (s *WebhookServer) parsePayload(r *http.Request) ([]byte, interface{}, error) {
	payload, err := github.ValidatePayload(r, []byte(s.webhookSecret))
	if err != nil {
		return nil, nil, err
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, nil, err
	}
	return payload, event, nil
}

func (s *WebhookServer) getInstallationID(event interface{}) int64 {
	switch e := event.(type) {
	case *github.IssuesEvent:
		if e.Installation != nil {
			return e.Installation.GetID()
		}
	case *github.PullRequestEvent:
		if e.Installation != nil {
			return e.Installation.GetID()
		}
	case *github.IssueCommentEvent:
		if e.Installation != nil {
			return e.Installation.GetID()
		}
	}
	return 0
}

func (s *WebhookServer) routeEvent(client *github.Client, event interface{}, r *http.Request) {
	switch e := event.(type) {
	case *github.IssuesEvent:
		if e.GetAction() == "opened" {
			go s.handleIssueOpened(context.Background(), client, e)
		} else if e.GetAction() == "closed" {
			go s.handleIssueClosed(context.Background(), client, e)
		}
	case *github.PullRequestEvent:
		if e.GetAction() == "opened" {
			go s.handlePullRequestOpened(context.Background(), client, e)
		} else if e.GetAction() == "closed" {
			go s.handlePullRequestClosed(context.Background(), client, e)
		}
	case *github.IssueCommentEvent:
		if e.GetAction() == "created" {
			go s.handleCommentCreated(context.Background(), client, e)
		}
	default:
		log.Printf("[Webhook] Unsupported event type: %s", github.WebHookType(r))
	}
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

	imgURL, err := GetReactionURL(ctx, cat)
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
	s.postComment(ctx, client, owner, repo, issueNum, commentBody)

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

	imgURL, err := GetReactionURL(ctx, cat)
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
	s.postComment(ctx, client, owner, repo, prNum, commentBody)

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

func (s *WebhookServer) handleIssueClosed(ctx context.Context, client *github.Client, e *github.IssuesEvent) {
	owner := e.Repo.Owner.GetLogin()
	repo := e.Repo.GetName()
	issueNum := e.Issue.GetNumber()

	log.Printf("[Issues] Issue closed in %s/%s #%d", owner, repo, issueNum)

	// Remove triage label
	_, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, issueNum, "triage")
	if err != nil {
		log.Printf("[Issues] Error removing label 'triage' (might not exist): %v", err)
	}
}

// postComment is a helper to post a comment to an issue or pull request.
func (s *WebhookServer) postComment(ctx context.Context, client *github.Client, owner, repo string, number int, body string) {
	_, _, err := client.Issues.CreateComment(ctx, owner, repo, number, &github.IssueComment{
		Body: github.String(body),
	})
	if err != nil {
		log.Printf("[Comment] Error creating comment in %s/%s #%d: %v", owner, repo, number, err)
	}
}

// detectLastMergeMethod queries the repository history to see how the most recent PR was merged,
// so the bot can reuse the same merge strategy (merge, squash, or rebase).
func (s *WebhookServer) detectLastMergeMethod(ctx context.Context, client *github.Client, owner, repo string) string {
	defaultMethod := "squash"

	opts := &github.PullRequestListOptions{
		State:       "closed",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 10},
	}
	prs, _, err := client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		log.Printf("[Merge] Failed to list PRs, falling back to '%s': %v", defaultMethod, err)
		return defaultMethod
	}

	var lastMergedPR *github.PullRequest
	for _, pr := range prs {
		if pr.MergedAt != nil && !pr.GetMergedAt().IsZero() {
			lastMergedPR = pr
			break
		}
	}

	if lastMergedPR == nil {
		log.Printf("[Merge] No merged PRs found in history, falling back to '%s'", defaultMethod)
		return defaultMethod
	}

	commitSHA := lastMergedPR.GetMergeCommitSHA()
	if commitSHA == "" {
		log.Printf("[Merge] Merged PR #%d has empty MergeCommitSHA, falling back to '%s'", lastMergedPR.GetNumber(), defaultMethod)
		return defaultMethod
	}

	prTitle := lastMergedPR.GetTitle()
	prCommitsCount := lastMergedPR.GetCommits()
	headSHA := lastMergedPR.GetHead().GetSHA()

	commit, _, err := client.Repositories.GetCommit(ctx, owner, repo, commitSHA, nil)
	if err != nil {
		log.Printf("[Merge] Failed to fetch commit %s, falling back to '%s': %v", commitSHA, defaultMethod, err)
		return defaultMethod
	}

	return determineMergeMethod(len(commit.Parents), commit.GetCommit().GetMessage(), prTitle, lastMergedPR.GetNumber(), prCommitsCount, commitSHA, headSHA)
}

// determineMergeMethod decides the merge strategy (merge, squash, or rebase) based on repo history:
// parent count, commit message, PR title, PR commit count, and commit SHA comparison.
func determineMergeMethod(parentsCount int, commitMsg, prTitle string, prNum, prCommitsCount int, commitSHA, headSHA string) string {
	if parentsCount == 2 {
		return "merge"
	}
	squashPattern := fmt.Sprintf("(#%d)", prNum)
	if strings.Contains(commitMsg, squashPattern) {
		return "squash"
	}
	if prTitle != "" && strings.Contains(commitMsg, prTitle) {
		return "squash"
	}
	// If a PR (single or multi-commit with prCommitsCount >= 1) was squashed by GitHub,
	// GitHub creates a new single commit on target branch (commitSHA != headSHA).
	if prCommitsCount >= 1 && commitSHA != "" && headSHA != "" && commitSHA != headSHA {
		return "squash"
	}
	return "rebase"
}

// checkPermission verifies if a commenter has write or admin permission to the repository.
func (s *WebhookServer) checkPermission(ctx context.Context, client *github.Client, owner, repo, commenter string) (bool, error) {
	permLevel, _, err := client.Repositories.GetPermissionLevel(ctx, owner, repo, commenter)
	if err != nil {
		return false, err
	}
	perm := permLevel.GetPermission()
	return perm == "admin" || perm == "write", nil
}

// formatCommitBody appends the Co-authored-by footer for the PR author and bot to the pull request description.
func formatCommitBody(pr *github.PullRequest, botUser *github.User) string {
	var coAuthors []string

	if line := formatUserCoAuthor(pr.GetUser()); line != "" {
		coAuthors = append(coAuthors, line)
	}
	if line := formatBotCoAuthor(pr.GetUser(), botUser); line != "" {
		coAuthors = append(coAuthors, line)
	}

	commitBody := strings.TrimSpace(pr.GetBody())
	if len(coAuthors) > 0 {
		footer := strings.Join(coAuthors, "\n")
		if commitBody != "" {
			return commitBody + "\n\n" + footer
		}
		return footer
	}
	return commitBody
}

func formatUserCoAuthor(prUser *github.User) string {
	if prUser == nil || prUser.GetLogin() == "" {
		return ""
	}
	prLogin := prUser.GetLogin()
	email := prUser.GetEmail()
	if email == "" {
		email = fmt.Sprintf("%d+%s@users.noreply.github.com", prUser.GetID(), prLogin)
	}
	name := prUser.GetName()
	if name == "" {
		name = prLogin
	}
	return fmt.Sprintf("Co-authored-by: %s <%s>", name, email)
}

func formatBotCoAuthor(prUser, botUser *github.User) string {
	if botUser == nil || botUser.GetLogin() == "" {
		return ""
	}
	botLogin := botUser.GetLogin()
	if prUser != nil && prUser.GetLogin() == botLogin {
		return ""
	}
	botName := botUser.GetName()
	if botName == "" {
		botName = strings.TrimSuffix(botLogin, "[bot]")
	}
	botEmail := fmt.Sprintf("%d+%s@users.noreply.github.com", botUser.GetID(), botLogin)
	return fmt.Sprintf("Co-authored-by: %s <%s>", botName, botEmail)
}

// handleCommentCreated processes comment creation events to detect "lgtm" and auto-merge the Pull Request.
func (s *WebhookServer) handleCommentCreated(ctx context.Context, client *github.Client, e *github.IssueCommentEvent) {
	// 1. Verify this comment is on a Pull Request (since PR comments are triggered via issue_comment)
	if e.Issue.PullRequestLinks == nil {
		return
	}

	// 2. Check if the comment body is exactly "lgtm" (case-insensitive)
	commentBody := strings.ToLower(strings.TrimSpace(e.Comment.GetBody()))
	if commentBody != "lgtm" {
		return
	}

	owner := e.Repo.Owner.GetLogin()
	repo := e.Repo.GetName()
	prNum := e.Issue.GetNumber()
	commenter := e.Comment.User.GetLogin()

	log.Printf("[Webhook] LGTM comment detected on PR %s/%s #%d by %s", owner, repo, prNum, commenter)

	// 3. Verify commenter permission level (must have write or admin access)
	hasPermission, err := s.checkPermission(ctx, client, owner, repo, commenter)
	if err != nil {
		log.Printf("[Webhook] Failed to get permission level for user %s: %v", commenter, err)
		return
	}
	if !hasPermission {
		log.Printf("[Webhook] User %s has insufficient permissions; ignoring LGTM", commenter)
		s.postComment(ctx, client, owner, repo, prNum, fmt.Sprintf("Sorry @%s, only collaborators with write or admin access can merge this Pull Request.", commenter))
		return
	}

	// 4. Fetch PR details
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNum)
	if err != nil {
		log.Printf("[Webhook] Failed to fetch PR details for #%d: %v", prNum, err)
		return
	}

	if pr.GetMerged() {
		log.Printf("[Webhook] PR #%d is already merged; skipping", prNum)
		return
	}

	// 5. Detect repository's active merge method from history
	mergeMethod := s.detectLastMergeMethod(ctx, client, owner, repo)
	log.Printf("[Webhook] Auto-detected last merge method: %s", mergeMethod)

	// 6. Build commit details with Co-authored-by footer
	commitBody := formatCommitBody(pr, s.getBotUser(ctx, client))

	// 7. Perform Merge
	opts := &github.PullRequestOptions{
		MergeMethod: mergeMethod,
		SHA:         pr.GetHead().GetSHA(),
	}
	switch mergeMethod {
	case "squash":
		opts.CommitTitle = fmt.Sprintf("%s (#%d)", pr.GetTitle(), prNum)
	case "merge":
		opts.CommitTitle = pr.GetTitle()
	}
	mergeResult, _, err := client.PullRequests.Merge(ctx, owner, repo, prNum, commitBody, opts)
	if err != nil {
		log.Printf("[Webhook] Merge failed for PR #%d: %v", prNum, err)
		s.postComment(ctx, client, owner, repo, prNum, fmt.Sprintf("Failed to merge this Pull Request: %v", err))
		return
	}

	log.Printf("[Webhook] Successfully merged PR #%d: %s", prNum, mergeResult.GetMessage())
	s.postComment(ctx, client, owner, repo, prNum, fmt.Sprintf("Merged successfully. (%s)", mergeResult.GetMessage()))
}

// getBotUser returns the GitHub App's bot user identity, fetching and caching it on first call.
func (s *WebhookServer) getBotUser(ctx context.Context, client *github.Client) *github.User {
	s.botUserMu.Lock()
	defer s.botUserMu.Unlock()
	if s.botUser != nil {
		return s.botUser
	}

	// Always use AppsTransport (App JWT) to fetch App info via GET /app
	appClient := github.NewClient(&http.Client{Transport: s.appsTransport})
	app, _, err := appClient.Apps.Get(ctx, "")

	var slug string
	var appName string
	if err != nil || app == nil {
		log.Printf("[Webhook] Failed to fetch app info for bot identity: %v", err)
		slug = "da-vinci-bot"
		appName = "davinci"
	} else {
		slug = app.GetSlug()
		if slug == "" {
			slug = "da-vinci-bot"
		}
		appName = app.GetName()
	}
	botLogin := slug + "[bot]"

	apiClient := client
	if apiClient == nil {
		apiClient = appClient
	}

	// Fetch the actual User object of the bot to get its correct database User ID
	botUser, _, err := apiClient.Users.Get(ctx, botLogin)
	if err != nil || botUser == nil {
		log.Printf("[Webhook] Failed to fetch bot user details for %s: %v. Falling back to App ID.", botLogin, err)
		s.botUser = &github.User{
			ID:    github.Int64(s.appID),
			Login: github.String(botLogin),
			Name:  github.String(appName),
		}
		return s.botUser
	}

	s.botUser = botUser
	return s.botUser
}

// GetAppMetadata returns the dynamic name, avatar URL, and HTML URL of the GitHub App.
func (s *WebhookServer) GetAppMetadata(ctx context.Context) (string, string, string) {
	defaultAvatar := fmt.Sprintf("https://avatars.githubusercontent.com/in/%d?v=4", s.appID)
	defaultHTMLURL := "https://github.com/apps/da-vinci-bot"
	defaultName := "davinci"

	client := github.NewClient(&http.Client{Transport: s.appsTransport})
	app, _, err := client.Apps.Get(ctx, "")
	if err != nil || app == nil {
		return defaultName, defaultAvatar, defaultHTMLURL
	}

	name := defaultName
	if app.Name != nil {
		name = *app.Name
	}
	htmlURL := defaultHTMLURL
	if app.HTMLURL != nil {
		htmlURL = *app.HTMLURL
	}

	return name, defaultAvatar, htmlURL
}

// GetActiveRepos returns the cached count of active repositories.
func (s *WebhookServer) GetActiveRepos() int {
	s.activeReposMu.RLock()
	defer s.activeReposMu.RUnlock()
	return s.activeRepos
}

// SetActiveRepos updates the cached count of active repositories.
func (s *WebhookServer) SetActiveRepos(count int) {
	s.activeReposMu.Lock()
	defer s.activeReposMu.Unlock()
	s.activeRepos = count
}
