package workspaces

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"

	"InfraMap/structs"
	"InfraMap/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//go:embed templates/ezdeploy-host.yaml
var ezdeployHostTemplate string

var (
	awsRolePattern      = regexp.MustCompile(`^arn:aws:iam::([0-9]{12}):role/[A-Za-z0-9+=,.@_/-]+$`)
	awsPrincipalPattern = regexp.MustCompile(`^arn:aws:iam::[0-9]{12}:(?:role|user)/[A-Za-z0-9+=,.@_/-]+$`)
	awsRegionPattern    = regexp.MustCompile(`^[a-z]{2}(?:-gov)?-[a-z]+-[0-9]$`)
	ec2InstancePattern  = regexp.MustCompile(`^i-[0-9a-f]{8,17}$`)
	ssmDocumentPattern  = regexp.MustCompile(`^[A-Za-z0-9_.-]{3,128}$`)
)

type awsConnectionSecret struct {
	RoleARN    string `json:"role_arn"`
	ExternalID string `json:"external_id"`
}

type AWSRuntime struct {
	ConnectionID string
	InstanceID   string
	Region       string
	DocumentName string
	RoleARN      string
	ExternalID   string
}

func GetAWSCloudFormation(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID, workspaceID := c.Param("accountID"), c.Param("workspaceID")
		workspace, role, err := LoadWorkspace(c, db, userID, accountID, workspaceID)
		if err != nil {
			WriteWorkspaceAccessError(c, err)
			return
		}
		if role != "owner" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only workspace owners and admins can provision AWS hosts"})
			return
		}
		if !awsPrincipalPattern.MatchString(utils.Cfg.AWSControlPrincipalARN) || !validIPv4HostCIDR(utils.Cfg.AWSControlSourceCIDR) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AWS control-plane identity and fixed egress IP are not configured"})
			return
		}

		externalID := awsExternalID(accountID, workspaceID)
		template := strings.NewReplacer(
			"{{INFRAMAP_PRINCIPAL_ARN}}", utils.Cfg.AWSControlPrincipalARN,
			"{{INFRAMAP_EXTERNAL_ID}}", externalID,
			"{{INFRAMAP_SOURCE_CIDR}}", utils.Cfg.AWSControlSourceCIDR,
			"{{WORKSPACE_ID}}", workspaceID,
		).Replace(ezdeployHostTemplate)
		filename := fmt.Sprintf("inframap-%s-ezdeploy.yaml", workspace.Slug)
		c.JSON(http.StatusOK, gin.H{"template": template, "filename": filename, "external_id": externalID, "stack_name": "inframap-" + workspace.Slug})
	}
}

func AttachAWSHost(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID, workspaceID := c.Param("accountID"), c.Param("workspaceID")
		_, role, err := LoadWorkspace(c, db, userID, accountID, workspaceID)
		if err != nil {
			WriteWorkspaceAccessError(c, err)
			return
		}
		if role != "owner" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only workspace owners and admins can attach AWS hosts"})
			return
		}

		var request structs.AttachAWSHostRequest
		if err := c.ShouldBindJSON(&request); err != nil || !validAWSHostRequest(request) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Enter the exact outputs from the InfraMap CloudFormation stack"})
			return
		}
		expectedExternalID := awsExternalID(accountID, workspaceID)
		if subtle.ConstantTimeCompare([]byte(request.ExternalID), []byte(expectedExternalID)) != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The CloudFormation stack belongs to a different InfraMap workspace"})
			return
		}
		awsAccountID := awsRolePattern.FindStringSubmatch(request.RoleARN)[1]
		if err := verifyAWSHost(c, request.RoleARN, request.ExternalID, request.Region, request.InstanceID, request.DocumentName, awsAccountID); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "AWS could not verify the workspace-scoped role, SSM document, and online host", "detail": err.Error()})
			return
		}

		secret, err := encryptAWSConnection(awsConnectionSecret{RoleARN: request.RoleARN, ExternalID: request.ExternalID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AWS connection could not be protected"})
			return
		}
		connectionID, err := saveAWSConnection(c, db, userID, accountID, workspaceID, request, awsAccountID, secret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AWS connection could not be saved"})
			return
		}
		resource, err := SaveResource(c, db, userID, workspaceID, ResourceInput{
			ConnectionID: connectionID,
			Provider:     "aws",
			ResourceType: "ec2_instance",
			ExternalID:   request.InstanceID,
			Name:         strings.TrimSpace(request.Name),
			Status:       "active",
			Metadata: map[string]interface{}{
				"account_id": awsAccountID, "region": request.Region, "document_name": request.DocumentName,
				"stack_id": request.StackID, "public_ip": request.PublicIP, "control": "ssm",
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AWS host could not be attached to the workspace"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"service": resource})
	}
}

func awsExternalID(accountID, workspaceID string) string {
	mac := hmac.New(sha256.New, []byte(utils.Cfg.SecretKey))
	_, _ = mac.Write([]byte("aws-control:" + accountID + ":" + workspaceID))
	return "inframap-" + hex.EncodeToString(mac.Sum(nil)[:16])
}

func validIPv4HostCIDR(value string) bool {
	ip, network, err := net.ParseCIDR(value)
	if err != nil || ip.To4() == nil {
		return false
	}
	ones, bits := network.Mask.Size()
	return bits == 32 && ones == 32
}

func validAWSHostRequest(request structs.AttachAWSHostRequest) bool {
	return strings.TrimSpace(request.Name) != "" && awsRolePattern.MatchString(request.RoleARN) && awsRegionPattern.MatchString(request.Region) &&
		ec2InstancePattern.MatchString(request.InstanceID) && ssmDocumentPattern.MatchString(request.DocumentName) && request.ExternalID != ""
}

func metadataString(metadata map[string]interface{}, key string) string {
	value, _ := metadata[key].(string)
	return value
}

func encryptAWSConnection(secret awsConnectionSecret) (string, error) {
	payload, err := json.Marshal(secret)
	if err != nil {
		return "", err
	}
	return utils.Encrypt(string(payload))
}

func saveAWSConnection(ctx context.Context, db *sql.DB, userID, accountID, workspaceID string, request structs.AttachAWSHostRequest, awsAccountID, secret string) (string, error) {
	resources, err := LoadResources(ctx, db, workspaceID)
	if err != nil {
		return "", err
	}
	if existing := FindResource(resources, "aws"); existing != nil {
		_, err = db.ExecContext(ctx, `UPDATE connections SET name = $1, external_account_id = $2, secret_ref = $3,
			metadata = $4::jsonb, updated_at = NOW() WHERE id = $5 AND account_id = $6 AND provider = 'aws'`,
			request.Name, awsAccountID, secret, awsConnectionMetadata(request, awsAccountID), existing.ConnectionID, accountID)
		return existing.ConnectionID, err
	}
	connectionID := uuid.NewString()
	_, err = db.ExecContext(ctx, `INSERT INTO connections (
		id, account_id, name, provider, connection_type, external_account_id, secret_ref, metadata, created_by, created_at, updated_at
	) VALUES ($1, $2, $3, 'aws', 'cross_account_role', $4, $5, $6::jsonb, $7, NOW(), NOW())`,
		connectionID, accountID, request.Name, awsAccountID, secret, awsConnectionMetadata(request, awsAccountID), userID)
	return connectionID, err
}

func awsConnectionMetadata(request structs.AttachAWSHostRequest, awsAccountID string) string {
	payload, _ := json.Marshal(map[string]interface{}{
		"account_id": awsAccountID, "region": request.Region, "instance_id": request.InstanceID,
		"document_name": request.DocumentName, "stack_id": request.StackID, "public_ip": request.PublicIP,
	})
	return string(payload)
}

func verifyAWSHost(ctx context.Context, roleARN, externalID, region, instanceID, documentName, accountID string) error {
	client, err := NewSSMClient(ctx, region, roleARN, externalID)
	if err != nil {
		return err
	}
	result, err := client.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{
		Filters:    []ssmtypes.InstanceInformationStringFilter{{Key: aws.String("InstanceIds"), Values: []string{instanceID}}},
		MaxResults: aws.Int32(5),
	})
	if err != nil {
		return err
	}
	if len(result.InstanceInformationList) != 1 || result.InstanceInformationList[0].PingStatus != ssmtypes.PingStatusOnline {
		return errors.New("the EC2 instance is not online in Systems Manager")
	}
	_, err = client.DescribeDocument(ctx, &ssm.DescribeDocumentInput{Name: aws.String(documentName)})
	if err != nil {
		return fmt.Errorf("deployment document is not accessible in AWS account %s: %w", accountID, err)
	}
	return nil
}

func NewSSMClient(ctx context.Context, region, roleARN, externalID string) (*ssm.Client, error) {
	configuration, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	provider := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(configuration), roleARN, func(options *stscreds.AssumeRoleOptions) {
		options.ExternalID = aws.String(externalID)
		options.RoleSessionName = "inframap-workspace-control"
	})
	configuration.Credentials = aws.NewCredentialsCache(provider)
	if _, err := configuration.Credentials.Retrieve(ctx); err != nil {
		return nil, err
	}
	return ssm.NewFromConfig(configuration), nil
}

func LoadAWSRuntime(ctx context.Context, db *sql.DB, accountID string, resource structs.WorkspaceResource) (AWSRuntime, error) {
	var encrypted string
	if err := db.QueryRowContext(ctx, `SELECT secret_ref FROM connections WHERE id = $1 AND account_id = $2 AND provider = 'aws'`, resource.ConnectionID, accountID).Scan(&encrypted); err != nil {
		return AWSRuntime{}, err
	}
	payload, err := utils.Decrypt(encrypted)
	if err != nil {
		return AWSRuntime{}, err
	}
	var secret awsConnectionSecret
	if err := json.Unmarshal([]byte(payload), &secret); err != nil {
		return AWSRuntime{}, err
	}
	runtime := AWSRuntime{
		ConnectionID: resource.ConnectionID, InstanceID: resource.ExternalID, Region: metadataString(resource.Metadata, "region"),
		DocumentName: metadataString(resource.Metadata, "document_name"), RoleARN: secret.RoleARN, ExternalID: secret.ExternalID,
	}
	if !validAWSHostRequest(structs.AttachAWSHostRequest{Name: resource.Name, RoleARN: runtime.RoleARN, ExternalID: runtime.ExternalID, Region: runtime.Region, InstanceID: runtime.InstanceID, DocumentName: runtime.DocumentName}) {
		return AWSRuntime{}, errors.New("stored AWS runtime is incomplete")
	}
	return runtime, nil
}
