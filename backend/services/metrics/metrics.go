package metrics

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	connections "InfraMap/services/connections"
	scan "InfraMap/services/scan"
	workspaces "InfraMap/services/workspaces"
	"InfraMap/structs"

	"github.com/gin-gonic/gin"
)

type workspaceMetricsResponse struct {
	ObservedAt       time.Time                           `json:"observed_at"`
	Status           string                              `json:"status"`
	Providers        []workspaceProviderStatus           `json:"providers"`
	Resources        []workspaceResourceTelemetry        `json:"resources"`
	Deployments      []connections.VercelDeployment      `json:"deployments"`
	SupabaseServices []connections.SupabaseServiceHealth `json:"supabase_services"`
	SupabaseUsage    []connections.SupabaseUsagePoint    `json:"supabase_usage"`
	Warnings         []string                            `json:"warnings"`
}

type workspaceResourceTelemetry struct {
	ResourceID       string                              `json:"resource_id"`
	Provider         string                              `json:"provider"`
	Name             string                              `json:"name"`
	ResourceType     string                              `json:"resource_type"`
	Status           string                              `json:"status"`
	ServiceIDs       []string                            `json:"service_ids"`
	Deployments      []connections.VercelDeployment      `json:"deployments"`
	SupabaseServices []connections.SupabaseServiceHealth `json:"supabase_services"`
	SupabaseUsage    []connections.SupabaseUsagePoint    `json:"supabase_usage"`
	Warnings         []string                            `json:"warnings"`
}

type workspaceProviderStatus struct {
	Provider  string `json:"provider"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Detail    string `json:"detail"`
	Connected bool   `json:"connected"`
}

func GetMetrics(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		workspaceID := c.Param("workspaceID")
		if _, _, err := workspaces.LoadWorkspace(c, db, userID, accountID, workspaceID); err != nil {
			workspaces.WriteWorkspaceAccessError(c, err)
			return
		}
		resources, err := workspaces.LoadResources(c, db, workspaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load the workspace system"})
			return
		}

		response := workspaceMetricsResponse{
			ObservedAt:       time.Now().UTC(),
			Status:           "operational",
			Providers:        make([]workspaceProviderStatus, 0, 4),
			Resources:        make([]workspaceResourceTelemetry, 0, len(resources)),
			Deployments:      make([]connections.VercelDeployment, 0),
			SupabaseServices: make([]connections.SupabaseServiceHealth, 0),
			SupabaseUsage:    make([]connections.SupabaseUsagePoint, 0),
			Warnings:         make([]string, 0),
		}

		for _, provider := range []string{"github", "vercel", "supabase", "aws"} {
			resource := workspaces.FindResource(resources, provider)
			if resource == nil {
				response.Providers = append(response.Providers, workspaceProviderStatus{Provider: provider, Name: providerName(provider), Status: "not_connected", Detail: "No resource attached", Connected: false})
				continue
			}
			response.Providers = append(response.Providers, workspaceProviderStatus{Provider: provider, Name: resource.Name, Status: normalizedControlStatus(resource.Status), Detail: resourceDetail(*resource), Connected: true})
		}

		serviceIDs := serviceIDsByProvider(resources)
		for _, resource := range resources {
			telemetry := workspaceResourceTelemetry{
				ResourceID: resource.ID, Provider: resource.Provider, Name: resource.Name,
				ResourceType: resource.ResourceType, Status: normalizedControlStatus(resource.Status),
				ServiceIDs: append([]string{}, serviceIDs[resource.Provider]...), Deployments: make([]connections.VercelDeployment, 0),
				SupabaseServices: make([]connections.SupabaseServiceHealth, 0), SupabaseUsage: make([]connections.SupabaseUsagePoint, 0), Warnings: make([]string, 0),
			}
			if resource.Provider == "vercel" {
				token, teamID, _, _, loadErr := connections.LoadVercelConnection(c, db, userID, accountID, resource.ConnectionID)
				if loadErr == nil {
					telemetry.Deployments, loadErr = connections.FetchVercelDeployments(c, token, teamID, resource.ExternalID)
				}
				if loadErr != nil {
					telemetry.Warnings = append(telemetry.Warnings, "Vercel deployment telemetry is temporarily unavailable")
				}
				response.Deployments = append(response.Deployments, telemetry.Deployments...)
			}
			if resource.Provider == "supabase" {
				loadSupabaseTelemetry(c, db, userID, accountID, resource, &telemetry, response.Providers)
				response.SupabaseServices = append(response.SupabaseServices, telemetry.SupabaseServices...)
				response.SupabaseUsage = append(response.SupabaseUsage, telemetry.SupabaseUsage...)
			}
			response.Warnings = append(response.Warnings, telemetry.Warnings...)
			response.Resources = append(response.Resources, telemetry)
		}

		response.Status = overallControlStatus(response.Providers, response.SupabaseServices)
		if response.Status == "operational" && len(response.Warnings) > 0 {
			response.Status = "partial"
		}
		c.JSON(http.StatusOK, response)
	}
}

func loadSupabaseTelemetry(c *gin.Context, db *sql.DB, userID, accountID string, resource structs.WorkspaceResource, telemetry *workspaceResourceTelemetry, providers []workspaceProviderStatus) {
	token, _, _, err := connections.LoadSupabaseConnection(c, db, userID, accountID, resource.ConnectionID)
	if err == nil {
		services, serviceErr := connections.FetchSupabaseServices(c, token)
		if serviceErr == nil {
			for _, service := range services {
				if service.Ref == resource.ExternalID {
					telemetry.Status = normalizedControlStatus(service.Status)
					updateProviderStatus(providers, "supabase", telemetry.Status, workspaceStatusDetail(service.Status))
					break
				}
			}
		}
		telemetry.SupabaseServices, err = connections.FetchSupabaseServiceHealth(c, token, resource.ExternalID)
		if usage, usageErr := connections.FetchSupabaseRequestActivity(c, token, resource.ExternalID); usageErr == nil {
			telemetry.SupabaseUsage = usage
		} else {
			telemetry.Warnings = append(telemetry.Warnings, "Supabase Analytics log telemetry is temporarily unavailable")
		}
	}
	if err != nil {
		telemetry.Warnings = append(telemetry.Warnings, "Supabase service health is temporarily unavailable")
	}
}

func serviceIDsByProvider(resources []structs.WorkspaceResource) map[string][]string {
	serviceIDs := make(map[string][]string)
	for _, source := range resources {
		if source.Provider != "github" {
			continue
		}
		for _, decision := range scan.DecisionsFromSource(source) {
			if decision.Provider != "" && decision.ComponentID != "" {
				serviceIDs[decision.Provider] = append(serviceIDs[decision.Provider], decision.ComponentID)
			}
		}
	}
	return serviceIDs
}

func updateProviderStatus(providers []workspaceProviderStatus, provider, status, detail string) {
	for index := range providers {
		if providers[index].Provider == provider {
			providers[index].Status = status
			providers[index].Detail = detail
			return
		}
	}
}

func normalizedControlStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "inactive" {
		return "paused"
	}
	if status == "active_healthy" || status == "ready" || status == "active" {
		return "healthy"
	}
	if status == "error" || status == "failed" || status == "unhealthy" {
		return "error"
	}
	if status == "" {
		return "unknown"
	}
	return status
}

func overallControlStatus(providers []workspaceProviderStatus, services []connections.SupabaseServiceHealth) string {
	for _, provider := range providers {
		if provider.Connected && (provider.Status == "paused" || provider.Status == "error") {
			return "attention"
		}
	}
	for _, service := range services {
		if !service.Healthy {
			return "attention"
		}
	}
	return "operational"
}

func providerName(provider string) string {
	switch provider {
	case "github":
		return "GitHub"
	case "vercel":
		return "Vercel"
	case "supabase":
		return "Supabase"
	case "aws":
		return "AWS"
	default:
		return provider
	}
}

func resourceDetail(resource structs.WorkspaceResource) string {
	switch resource.Provider {
	case "github":
		if branch, ok := resource.Metadata["default_branch"].(string); ok && branch != "" {
			return "Branch " + branch
		}
	case "vercel":
		if framework, ok := resource.Metadata["framework"].(string); ok && framework != "" {
			return framework
		}
	case "supabase":
		if region, ok := resource.Metadata["region"].(string); ok && region != "" {
			return region
		}
	}
	return resource.ResourceType
}

func workspaceStatusDetail(status string) string {
	if strings.EqualFold(status, "inactive") {
		return "Paused or unavailable"
	}
	return strings.ReplaceAll(strings.ToLower(status), "_", " ")
}
