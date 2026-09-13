package services

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	connections "InfraMap/services/connections"
	"InfraMap/structs"
)

func TestNormalizeSignup(t *testing.T) {
	request := structs.UserCreate{
		Username: "  Dave_123 ",
		Email:    " DAVE123@EXAMPLE.COM ",
		Password: "safe password",
	}
	if err := normalizeSignup(&request); err != nil {
		t.Fatal(err)
	}
	if request.Username != "Dave_123" || request.Email != "dave123@example.com" {
		t.Fatalf("signup was not normalized: %#v", request)
	}
}

func TestNormalizeSignupRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		request structs.UserCreate
	}{
		{name: "short username", request: structs.UserCreate{Username: "ab", Email: "a@example.com", Password: "password"}},
		{name: "punctuation in username", request: structs.UserCreate{Username: "dave!", Email: "a@example.com", Password: "password"}},
		{name: "display-name email", request: structs.UserCreate{Username: "dave", Email: "Dave <a@example.com>", Password: "password"}},
		{name: "blank password", request: structs.UserCreate{Username: "dave", Email: "a@example.com", Password: "        "}},
		{name: "bcrypt overflow", request: structs.UserCreate{Username: "dave", Email: "a@example.com", Password: strings.Repeat("x", 73)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := normalizeSignup(&test.request); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestNormalizeLogin(t *testing.T) {
	request := structs.UserLogin{Email: " DAVE@EXAMPLE.COM ", Password: "password"}
	if err := normalizeLogin(&request); err != nil {
		t.Fatal(err)
	}
	if request.Email != "dave@example.com" {
		t.Fatalf("email = %q", request.Email)
	}
}

func TestVercelCompletionURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		allowed bool
	}{
		{name: "vercel", value: "https://vercel.com/dashboard/integrations", allowed: true},
		{name: "vercel subdomain", value: "https://app.vercel.com/integrations", allowed: true},
		{name: "lookalike host", value: "https://vercel.com.example.com/steal", allowed: false},
		{name: "insecure scheme", value: "http://vercel.com/integrations", allowed: false},
		{name: "external host", value: "https://example.com", allowed: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, allowed := vercelCompletionURL(test.value)
			if allowed != test.allowed {
				t.Fatalf("allowed = %v, want %v", allowed, test.allowed)
			}
		})
	}
}

func TestSafeOAuthReturnTo(t *testing.T) {
	allowed := "/workspaces/account/workspace?setup=1"
	if actual := safeOAuthReturnTo(allowed); actual != allowed {
		t.Fatalf("return path = %q, want %q", actual, allowed)
	}
	for _, value := range []string{"https://example.com/steal", "//example.com/steal", "workspaces/workspace"} {
		if actual := safeOAuthReturnTo(value); actual != "" {
			t.Fatalf("unsafe return path %q was accepted as %q", value, actual)
		}
	}
}

func TestIntegrationOAuthProblemPreservesGitHubFailure(t *testing.T) {
	err := fmt.Errorf("verify repositories: %w", &connections.GitHubAPIError{Status: http.StatusForbidden})
	if problem := integrationOAuthProblem("github", err); problem != "github_repository_access_forbidden" {
		t.Fatalf("problem = %q", problem)
	}
}
