package projects

import (
	"database/sql"
	"net/http"
	"strings"
	"unicode/utf8"

	connections "InfraMap/services/connections"
	"InfraMap/structs"
	"InfraMap/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var allowedSupabaseRegionGroups = map[string]bool{"americas": true, "emea": true, "apac": true}

func AttachSupabaseProject(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		projectID := c.Param("projectID")
		_, role, err := loadProject(c, db, userID, accountID, projectID)
		if err != nil {
			writeProjectAccessError(c, err)
			return
		}
		if role == "viewer" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Viewers cannot attach project resources"})
			return
		}

		var request structs.AttachSupabaseProjectRequest
		if err := c.ShouldBindJSON(&request); err != nil || !validUUID(request.ConnectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Choose a valid Supabase project"})
			return
		}
		accessToken, _, _, err := connections.LoadSupabaseConnection(c, db, userID, accountID, request.ConnectionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Supabase connection could not be opened"})
			return
		}
		available, err := connections.FetchSupabaseProjects(c, accessToken)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Supabase projects could not be verified"})
			return
		}

		var selected *connections.SupabaseProject
		for index := range available {
			if available[index].Ref == request.ProjectRef {
				selected = &available[index]
				break
			}
		}
		if selected == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "That Supabase project is no longer accessible"})
			return
		}

		resource, err := saveSupabaseResource(c, db, userID, projectID, request.ConnectionID, *selected, "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Supabase project could not be attached"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"resource": resource})
	}
}

func CreateSupabaseProject(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		projectID := c.Param("projectID")
		_, role, err := loadProject(c, db, userID, accountID, projectID)
		if err != nil {
			writeProjectAccessError(c, err)
			return
		}
		if role != "owner" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only workspace owners and admins can create Supabase projects"})
			return
		}

		var request structs.CreateSupabaseProjectRequest
		if err := c.ShouldBindJSON(&request); err != nil || !validUUID(request.ConnectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Complete the Supabase project details"})
			return
		}
		request.Name = strings.TrimSpace(request.Name)
		request.OrganizationSlug = strings.TrimSpace(request.OrganizationSlug)
		request.RegionGroup = strings.ToLower(strings.TrimSpace(request.RegionGroup))
		if request.Name == "" || utf8.RuneCountInString(request.Name) > 80 || !allowedSupabaseRegionGroups[request.RegionGroup] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Supabase project details are invalid"})
			return
		}
		if utf8.RuneCountInString(request.DatabasePassword) < 16 || utf8.RuneCountInString(request.DatabasePassword) > 128 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Use a database password between 16 and 128 characters"})
			return
		}

		accessToken, _, _, err := connections.LoadSupabaseConnection(c, db, userID, accountID, request.ConnectionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Supabase connection could not be opened"})
			return
		}
		organizations, err := connections.FetchSupabaseOrganizations(c, accessToken)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Supabase organizations could not be verified"})
			return
		}
		authorizedOrganization := false
		for _, organization := range organizations {
			if organization.Slug == request.OrganizationSlug {
				authorizedOrganization = true
				break
			}
		}
		if !authorizedOrganization {
			c.JSON(http.StatusBadRequest, gin.H{"error": "That Supabase organization is no longer accessible"})
			return
		}

		created, err := connections.CreateSupabaseProject(c, accessToken, connections.CreateSupabaseProjectInput{
			DatabasePassword: request.DatabasePassword,
			Name:             request.Name,
			OrganizationSlug: request.OrganizationSlug,
			RegionSelection:  connections.SupabaseRegionSelection{Type: "smartGroup", Code: request.RegionGroup},
		})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		encryptedPassword, err := utils.Encrypt(request.DatabasePassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Supabase project was created, but its database credential could not be secured. Attach the project manually."})
			return
		}
		resource, err := saveSupabaseResource(c, db, userID, projectID, request.ConnectionID, created, encryptedPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Supabase project was created, but could not be attached. Attach it from the existing project list."})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"resource": resource})
	}
}

func saveSupabaseResource(c *gin.Context, db *sql.DB, userID, projectID, connectionID string, project connections.SupabaseProject, encryptedPassword string) (structs.ProjectResource, error) {
	return saveProjectResource(c, db, userID, projectID, projectResourceInput{
		ConnectionID: connectionID,
		Provider:     "supabase",
		ResourceType: "project",
		ExternalID:   project.Ref,
		Name:         project.Name,
		Status:       normalizedSupabaseStatus(project.Status),
		Metadata: map[string]interface{}{
			"organization_id":   project.OrganizationID,
			"organization_slug": project.OrganizationSlug,
			"region":            project.Region,
			"created_at":        project.CreatedAt,
			"database":          project.Database,
		},
		SecretRef: encryptedPassword,
	})
}

func normalizedSupabaseStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "provisioning"
	}
	return status
}

func validUUID(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}
