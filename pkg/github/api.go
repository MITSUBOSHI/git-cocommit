package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

// GetUserEmail gets an email address from a GitHub username
// Uses authenticated API if GITHUB_TOKEN is in the environment variables,
// otherwise uses unauthenticated API (be careful of rate limits)
func GetUserEmail(username string) (string, error) {
	var client *github.Client

	// Get GitHub Personal Access Token from environment variable
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		// Client with authentication
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client = github.NewClient(tc)
	} else {
		// Client without authentication
		client = github.NewClient(nil)
	}

	// Set timeout for context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user information via GitHub API
	user, resp, err := client.Users.Get(ctx, username)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("GitHub user '%s' not found", username)
		}
		return "", fmt.Errorf("failed to get GitHub user info: %w", err)
	}

	// Get and validate email address
	email := user.GetEmail()
	if email == "" {
		// If public email address is not set
		// Return GitHub's no-reply format email address (ID+USERNAME@users.noreply.github.com)
		userID := user.GetID()
		email = fmt.Sprintf("%d+%s@users.noreply.github.com", userID, username)
	}

	return email, nil
}

// FormatCoAuthor creates a Co-Authored-By format string
// from username and email address
func FormatCoAuthor(username, email string) string {
	return fmt.Sprintf("%s <%s>", username, email)
}
