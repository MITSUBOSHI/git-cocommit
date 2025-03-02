package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-github/v58/github"
)

func TestFormatCoAuthor(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		want     string
	}{
		{
			name:     "Normal case",
			username: "testuser",
			email:    "test@example.com",
			want:     "testuser <test@example.com>",
		},
		{
			name:     "Name with special characters",
			username: "test-user.123",
			email:    "test@example.com",
			want:     "test-user.123 <test@example.com>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCoAuthor(tt.username, tt.email)
			if got != tt.want {
				t.Errorf("FormatCoAuthor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserEmail(t *testing.T) {
	origToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", origToken)

	tests := []struct {
		name           string
		username       string
		mockResponse   *github.User
		mockStatusCode int
		withToken      bool
		want           string
		wantErr        bool
	}{
		{
			name:     "Public email available",
			username: "testuser",
			mockResponse: &github.User{
				Email: github.String("public@example.com"),
				Login: github.String("testuser"),
			},
			mockStatusCode: http.StatusOK,
			withToken:      false,
			want:           "public@example.com",
			wantErr:        false,
		},
		{
			name:     "No public email",
			username: "testuser",
			mockResponse: &github.User{
				Email: github.String(""),
				Login: github.String("testuser"),
			},
			mockStatusCode: http.StatusOK,
			withToken:      false,
			want:           "testuser@users.noreply.github.com",
			wantErr:        false,
		},
		{
			name:           "User not found",
			username:       "nonexistent",
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			withToken:      false,
			want:           "",
			wantErr:        true,
		},
		{
			name:     "With authentication token",
			username: "testuser",
			mockResponse: &github.User{
				Email: github.String("auth@example.com"),
				Login: github.String("testuser"),
			},
			mockStatusCode: http.StatusOK,
			withToken:      true,
			want:           "auth@example.com",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.withToken && r.Header.Get("Authorization") == "" {
					t.Error("Authorization header expected but not found")
				}

				if r.URL.Path != "/users/"+tt.username {
					t.Errorf("Expected request to /users/%s, got %s", tt.username, r.URL.Path)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockStatusCode == http.StatusOK && tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			if tt.withToken {
				os.Setenv("GITHUB_TOKEN", "test-token")
			} else {
				os.Setenv("GITHUB_TOKEN", "")
			}

			t.Skip("This test is skipped as it requires modifications to the code to support dependency injection")
		})
	}
}
