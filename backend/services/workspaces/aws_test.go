package workspaces

import (
	"strings"
	"testing"

	"InfraMap/utils"
)

func TestAWSExternalIDIsWorkspaceScopedAndStable(t *testing.T) {
	original := utils.Cfg.SecretKey
	utils.Cfg.SecretKey = "test-secret"
	t.Cleanup(func() { utils.Cfg.SecretKey = original })

	first := awsExternalID("account-one", "workspace-one")
	if first != awsExternalID("account-one", "workspace-one") {
		t.Fatal("expected the same workspace to receive a stable external ID")
	}
	if first == awsExternalID("account-one", "workspace-two") {
		t.Fatal("expected different workspaces to receive different external IDs")
	}
	if !strings.HasPrefix(first, "inframap-") || len(first) != len("inframap-")+32 {
		t.Fatalf("unexpected external ID format: %q", first)
	}
}

func TestAWSControlSourceRequiresOneIPv4Address(t *testing.T) {
	for _, value := range []string{"203.0.113.10/32", "10.0.0.1/32"} {
		if !validIPv4HostCIDR(value) {
			t.Fatalf("expected %q to be accepted", value)
		}
	}
	for _, value := range []string{"203.0.113.0/24", "0.0.0.0/0", "2001:db8::1/128", "203.0.113.10"} {
		if validIPv4HostCIDR(value) {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}

func TestCloudFormationTemplateKeepsManagementPortsClosed(t *testing.T) {
	for _, required := range []string{"HttpTokens: required", "sts:ExternalId", "aws:SourceIp", "ssm:resourceTag/InfraMapWorkspaceId"} {
		if !strings.Contains(ezdeployHostTemplate, required) {
			t.Fatalf("expected CloudFormation template to contain %q", required)
		}
	}
	if strings.Contains(ezdeployHostTemplate, "FromPort: 22") || strings.Contains(ezdeployHostTemplate, "ToPort: 22") {
		t.Fatal("CloudFormation template must not open SSH")
	}
}
