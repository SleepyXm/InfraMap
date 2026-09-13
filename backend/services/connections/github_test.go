package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWriteGitHubRepositoryErrorRequiresWorkspaceReconnect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	WriteGitHubRepositoryError(context, &GitHubAPIError{Status: http.StatusUnauthorized, Message: "Bad credentials"}, "loaded")
	if recorder.Code != http.StatusFailedDependency {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusFailedDependency)
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["code"] != "github_reauthentication_required" {
		t.Fatalf("code = %v", payload["code"])
	}
}
