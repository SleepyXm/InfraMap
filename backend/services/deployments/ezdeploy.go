package deployments

import (
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"sort"
	"strings"

	scan "InfraMap/services/scan"
	workspaces "InfraMap/services/workspaces"
	"InfraMap/structs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/gin-gonic/gin"
)

var (
	commandIDPattern   = regexp.MustCompile(`^[0-9a-f-]{36}$`)
	domainPattern      = regexp.MustCompile(`^(?:\*\.)?[A-Za-z0-9](?:[A-Za-z0-9.-]{0,251}[A-Za-z0-9])?$`)
	gitBranchPattern   = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)
	githubClonePattern = regexp.MustCompile(`^https://github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+\.git$`)
)

type commandResponse struct {
	CommandID      string `json:"command_id"`
	Status         string `json:"status"`
	StatusDetail   string `json:"status_detail,omitempty"`
	ResponseCode   int32  `json:"response_code"`
	StandardOutput string `json:"standard_output,omitempty"`
	StandardError  string `json:"standard_error,omitempty"`
}

func Create(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, accountID, workspaceID := c.MustGet("userID").(string), c.Param("accountID"), c.Param("workspaceID")
		_, role, err := workspaces.LoadWorkspace(c, db, userID, accountID, workspaceID)
		if err != nil {
			workspaces.WriteWorkspaceAccessError(c, err)
			return
		}
		if role != "owner" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only workspace owners and admins can deploy to AWS"})
			return
		}
		var request structs.CreateEZDeployDeploymentRequest
		if err := c.ShouldBindJSON(&request); err != nil || !validRequest(request) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Enter a valid domain, certificate email, and runtime"})
			return
		}
		resources, err := workspaces.LoadResources(c, db, workspaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Workspace attachments could not be loaded"})
			return
		}
		source, host := workspaces.FindResource(resources, "github"), workspaces.FindResource(resources, "aws")
		if source == nil || host == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Attach both a GitHub source and AWS service before deploying"})
			return
		}
		components, err := plannedAWSComponents(*source)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if environment := componentEnvironment(components); len(environment) > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Configure the environment required by the selected components before deploying", "environment": environment})
			return
		}
		if request.Runtime == "docker" && len(components) != 1 {
			c.JSON(http.StatusConflict, gin.H{"error": "Docker deployment currently requires exactly one component in the saved AWS plan"})
			return
		}
		if private, _ := source.Metadata["private"].(bool); private {
			c.JSON(http.StatusConflict, gin.H{"error": "Private repository deployment needs a repository-scoped GitHub App token and is not enabled yet"})
			return
		}
		repositoryURL, branch := metadataString(source.Metadata, "clone_url"), metadataString(source.Metadata, "default_branch")
		if repositoryURL == "" {
			repositoryURL = strings.TrimSuffix(metadataString(source.Metadata, "html_url"), ".git") + ".git"
		}
		if !githubClonePattern.MatchString(repositoryURL) || !gitBranchPattern.MatchString(branch) {
			c.JSON(http.StatusConflict, gin.H{"error": "The attached GitHub source does not have a deployable URL and branch"})
			return
		}
		runtime, err := workspaces.LoadAWSRuntime(c, db, accountID, *host)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "AWS runtime credentials could not be opened"})
			return
		}
		client, err := workspaces.NewSSMClient(c, runtime.Region, runtime.RoleARN, runtime.ExternalID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "AWS runtime could not be reached"})
			return
		}
		result, err := client.SendCommand(c, &ssm.SendCommandInput{
			Comment: aws.String("InfraMap EZDeploy operation for workspace " + workspaceID), DocumentName: aws.String(runtime.DocumentName), InstanceIds: []string{runtime.InstanceID},
			Parameters:     map[string][]string{"Operation": {"deploy"}, "RepositoryURL": {repositoryURL}, "Branch": {branch}, "Domain": {request.Domain}, "Email": {request.Email}, "Runtime": {request.Runtime}, "Service": {componentSelector(components, request.Runtime)}},
			TimeoutSeconds: aws.Int32(600),
		})
		if err != nil || result.Command == nil || result.Command.CommandId == nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "EZDeploy command could not be started"})
			return
		}
		commandID := aws.ToString(result.Command.CommandId)
		_, _ = db.ExecContext(c, `UPDATE workspace_resources SET metadata = metadata || jsonb_build_object(
			'last_command_id', $1::text, 'deployment_status', 'pending', 'domain', $2::text, 'runtime', $3::text
		), updated_at = NOW() WHERE id = $4`, commandID, request.Domain, request.Runtime, host.ID)
		c.JSON(http.StatusAccepted, commandResponse{CommandID: commandID, Status: "Pending"})
	}
}

func Get(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, accountID, workspaceID, commandID := c.MustGet("userID").(string), c.Param("accountID"), c.Param("workspaceID"), c.Param("commandID")
		if !commandIDPattern.MatchString(commandID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment command ID"})
			return
		}
		if _, _, err := workspaces.LoadWorkspace(c, db, userID, accountID, workspaceID); err != nil {
			workspaces.WriteWorkspaceAccessError(c, err)
			return
		}
		resources, err := workspaces.LoadResources(c, db, workspaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Workspace attachments could not be loaded"})
			return
		}
		host := workspaces.FindResource(resources, "aws")
		if host == nil || metadataString(host.Metadata, "last_command_id") != commandID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deployment command not found"})
			return
		}
		runtime, err := workspaces.LoadAWSRuntime(c, db, accountID, *host)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "AWS runtime credentials could not be opened"})
			return
		}
		client, err := workspaces.NewSSMClient(c, runtime.Region, runtime.RoleARN, runtime.ExternalID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "AWS runtime could not be reached"})
			return
		}
		result, err := client.GetCommandInvocation(c, &ssm.GetCommandInvocationInput{CommandId: aws.String(commandID), InstanceId: aws.String(runtime.InstanceID)})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Deployment status could not be read"})
			return
		}
		status := string(result.Status)
		_, _ = db.ExecContext(c, `UPDATE workspace_resources SET metadata = metadata || jsonb_build_object('deployment_status', $1::text), updated_at = NOW() WHERE id = $2`, status, host.ID)
		c.JSON(http.StatusOK, commandResponse{CommandID: commandID, Status: status, StatusDetail: aws.ToString(result.StatusDetails), ResponseCode: result.ResponseCode, StandardOutput: aws.ToString(result.StandardOutputContent), StandardError: aws.ToString(result.StandardErrorContent)})
	}
}

func validRequest(request structs.CreateEZDeployDeploymentRequest) bool {
	address, err := mail.ParseAddress(request.Email)
	return domainPattern.MatchString(request.Domain) && err == nil && address.Address == request.Email && (request.Runtime == "native" || request.Runtime == "docker")
}

func plannedAWSComponents(source structs.WorkspaceResource) ([]scan.Component, error) {
	analysis, ok := scan.AnalysisFromSource(source)
	if !ok {
		return nil, errors.New("scan the GitHub source with EZDeploy before deploying")
	}
	decisions := scan.DecisionsFromSource(source)
	if len(decisions) == 0 {
		return nil, errors.New("save the EZDeploy component plan before deploying")
	}
	selected := map[string]bool{}
	for _, decision := range decisions {
		if decision.Action == "deploy" && decision.Provider == "aws" {
			selected[decision.ComponentID] = true
		}
	}
	components := make([]scan.Component, 0, len(selected))
	for _, component := range analysis.Components {
		if selected[component.ID] {
			components = append(components, component)
		}
	}
	if len(components) == 0 {
		return nil, errors.New("the saved component plan does not assign a deployable component to AWS")
	}
	return components, nil
}

func componentEnvironment(components []scan.Component) []string {
	values := map[string]bool{}
	for _, component := range components {
		for _, name := range component.Environment {
			values[name] = true
		}
	}
	keys := make([]string, 0, len(values))
	for value := range values {
		keys = append(keys, value)
	}
	sort.Strings(keys)
	return keys
}

func componentSelector(components []scan.Component, runtime string) string {
	if runtime == "docker" {
		return ""
	}
	selectors := make([]string, 0, len(components))
	for _, component := range components {
		selector := strings.TrimSpace(component.Root)
		if selector == "" || selector == "." {
			selector = strings.TrimSpace(component.Entry)
		}
		if selector != "" {
			selectors = append(selectors, selector)
		}
	}
	return strings.Join(selectors, ",")
}

func metadataString(metadata map[string]interface{}, key string) string {
	value, _ := metadata[key].(string)
	return value
}
