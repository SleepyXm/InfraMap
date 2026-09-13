package workspaces

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

func AttachSupabaseService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		workspaceID := c.Param("workspaceID")
		if !requireServiceWriteAccess(c, db, userID, accountID, workspaceID, false) {
			return
		}

		var request structs.AttachSupabaseServiceRequest
		if err := c.ShouldBindJSON(&request); err != nil || !validUUID(request.ConnectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Choose a valid Supabase service"})
			return
		}
		accessToken, _, _, err := connections.LoadSupabaseConnection(c, db, userID, accountID, request.ConnectionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Supabase connection could not be opened"})
			return
		}
		available, err := connections.FetchSupabaseServices(c, accessToken)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Supabase services could not be verified"})
			return
		}

		var selected *connections.SupabaseService
		for index := range available {
			if available[index].Ref == request.ServiceRef {
				selected = &available[index]
				break
			}
		}
		if selected == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "That Supabase service is no longer accessible"})
			return
		}

		resource, err := saveSupabaseResource(c, db, userID, workspaceID, request.ConnectionID, *selected, "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Supabase service could not be attached"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"service": resource})
	}
}

func CreateSupabaseService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		workspaceID := c.Param("workspaceID")
		if !requireServiceWriteAccess(c, db, userID, accountID, workspaceID, true) {
			return
		}

		var request structs.CreateSupabaseServiceRequest
		if err := c.ShouldBindJSON(&request); err != nil || !validUUID(request.ConnectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Complete the Supabase service details"})
			return
		}
		request.Name = strings.TrimSpace(request.Name)
		request.OrganizationSlug = strings.TrimSpace(request.OrganizationSlug)
		request.RegionGroup = strings.ToLower(strings.TrimSpace(request.RegionGroup))
		if request.Name == "" || utf8.RuneCountInString(request.Name) > 80 || !allowedSupabaseRegionGroups[request.RegionGroup] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Supabase service details are invalid"})
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

		created, err := connections.CreateSupabaseService(c, accessToken, connections.CreateSupabaseServiceInput{
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Supabase service was created, but its database credential could not be secured. Attach the service manually."})
			return
		}
		resource, err := saveSupabaseResource(c, db, userID, workspaceID, request.ConnectionID, created, encryptedPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Supabase service was created, but could not be attached. Attach it from the existing service list."})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"service": resource})
	}
}

func saveSupabaseResource(c *gin.Context, db *sql.DB, userID, workspaceID, connectionID string, service connections.SupabaseService, encryptedPassword string) (structs.WorkspaceResource, error) {
	return SaveResource(c, db, userID, workspaceID, ResourceInput{
		ConnectionID: connectionID,
		Provider:     "supabase",
		ResourceType: "project",
		ExternalID:   service.Ref,
		Name:         service.Name,
		Status:       normalizedSupabaseStatus(service.Status),
		Metadata: map[string]interface{}{
			"organization_id":   service.OrganizationID,
			"organization_slug": service.OrganizationSlug,
			"region":            service.Region,
			"created_at":        service.CreatedAt,
			"database":          service.Database,
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
