package projects

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	connections "InfraMap/services/connections"
	"InfraMap/structs"

	"github.com/gin-gonic/gin"
)

type projectControlResponse struct {
	ObservedAt       time.Time                           `json:"observed_at"`
	Status           string                              `json:"status"`
	Providers        []projectProviderStatus             `json:"providers"`
	Deployments      []connections.VercelDeployment      `json:"deployments"`
	SupabaseServices []connections.SupabaseServiceHealth `json:"supabase_services"`
	SupabaseUsage    []connections.SupabaseUsagePoint    `json:"supabase_usage"`
	Warnings         []string                            `json:"warnings"`
}

type projectProviderStatus struct {
	Provider  string `json:"provider"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Detail    string `json:"detail"`
	Connected bool   `json:"connected"`
}

func GetProjectControl(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		projectID := c.Param("projectID")
		if _, _, err := loadProject(c, db, userID, accountID, projectID); err != nil {
			writeProjectAccessError(c, err)
			return
		}
		resources, err := loadProjectResources(c, db, projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load the project system"})
			return
		}

		response := projectControlResponse{
			ObservedAt:       time.Now().UTC(),
			Status:           "operational",
			Providers:        make([]projectProviderStatus, 0, 4),
			Deployments:      make([]connections.VercelDeployment, 0),
			SupabaseServices: make([]connections.SupabaseServiceHealth, 0),
			SupabaseUsage:    make([]connections.SupabaseUsagePoint, 0),
			Warnings:         make([]string, 0),
		}

		for _, provider := range []string{"github", "vercel", "supabase", "aws"} {
			resource := findProjectResource(resources, provider)
			if resource == nil {
				response.Providers = append(response.Providers, projectProviderStatus{Provider: provider, Name: providerName(provider), Status: "not_connected", Detail: "No resource attached", Connected: false})
				continue
			}
			response.Providers = append(response.Providers, projectProviderStatus{Provider: provider, Name: resource.Name, Status: normalizedControlStatus(resource.Status), Detail: resourceDetail(*resource), Connected: true})
		}

		if resource := findProjectResource(resources, "vercel"); resource != nil {
			token, teamID, _, _, err := connections.LoadVercelConnection(c, db, userID, accountID, resource.ConnectionID)
			if err == nil {
				response.Deployments, err = connections.FetchVercelDeployments(c, token, teamID, resource.ExternalID)
			}
			if err != nil {
				response.Warnings = append(response.Warnings, "Vercel deployment telemetry is temporarily unavailable")
			}
		}

		if resource := findProjectResource(resources, "supabase"); resource != nil {
			token, _, _, err := connections.LoadSupabaseConnection(c, db, userID, accountID, resource.ConnectionID)
			if err == nil {
				projects, projectErr := connections.FetchSupabaseProjects(c, token)
				if projectErr == nil {
					for _, project := range projects {
						if project.Ref == resource.ExternalID {
							updateProviderStatus(response.Providers, "supabase", normalizedControlStatus(project.Status), projectStatusDetail(project.Status))
							break
						}
					}
				}
				response.SupabaseServices, err = connections.FetchSupabaseProjectHealth(c, token, resource.ExternalID)
				if usage, usageErr := connections.FetchSupabaseRequestActivity(c, token, resource.ExternalID); usageErr == nil {
					response.SupabaseUsage = usage
				} else {
					response.Warnings = append(response.Warnings, "Supabase Analytics log telemetry is temporarily unavailable")
				}
			}
			if err != nil {
				response.Warnings = append(response.Warnings, "Supabase service health is temporarily unavailable")
			}
		}

		response.Status = overallControlStatus(response.Providers, response.SupabaseServices)
		if response.Status == "operational" && len(response.Warnings) > 0 {
			response.Status = "partial"
		}
		c.JSON(http.StatusOK, response)
	}
}

func findProjectResource(resources []structs.ProjectResource, provider string) *structs.ProjectResource {
	for index := range resources {
		if resources[index].Provider == provider {
			return &resources[index]
		}
	}
	return nil
}

func updateProviderStatus(providers []projectProviderStatus, provider, status, detail string) {
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

func overallControlStatus(providers []projectProviderStatus, services []connections.SupabaseServiceHealth) string {
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

func resourceDetail(resource structs.ProjectResource) string {
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

func projectStatusDetail(status string) string {
	if strings.EqualFold(status, "inactive") {
		return "Paused or unavailable"
	}
	return strings.ReplaceAll(strings.ToLower(status), "_", " ")
}
