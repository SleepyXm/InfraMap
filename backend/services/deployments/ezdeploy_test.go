package deployments

import (
	"testing"

	"InfraMap/structs"
)

func TestRepositoryURLValidationRejectsCredentials(t *testing.T) {
	if !githubClonePattern.MatchString("https://github.com/example/service.git") {
		t.Fatal("expected a public GitHub clone URL to be accepted")
	}
	for _, value := range []string{"https://token@github.com/example/service.git", "https://github.com/example/service.git?token=secret", "git@github.com:example/service.git"} {
		if githubClonePattern.MatchString(value) {
			t.Fatalf("expected credential-bearing or SSH URL %q to be rejected", value)
		}
	}
}

func TestDeploymentRequestValidation(t *testing.T) {
	if !validRequest(structs.CreateEZDeployDeploymentRequest{Domain: "api.example.com", Email: "ops@example.com", Runtime: "native"}) {
		t.Fatal("expected a valid native deployment request")
	}
	for _, request := range []structs.CreateEZDeployDeploymentRequest{
		{Domain: "api.example.com;reboot", Email: "ops@example.com", Runtime: "native"},
		{Domain: "api.example.com", Email: "Ops <ops@example.com>", Runtime: "native"},
		{Domain: "api.example.com", Email: "ops@example.com", Runtime: "privileged"},
	} {
		if validRequest(request) {
			t.Fatalf("expected deployment request to be rejected: %#v", request)
		}
	}
}
