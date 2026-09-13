package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"InfraMap/utils"

	"github.com/gin-gonic/gin"
)

const vercelAPIURL = "https://api.vercel.com"

var errVercelConnectionNotFound = errors.New("vercel connection not found")

func ConnectVercelIntegration(ctx context.Context, db *sql.DB, accountID, userID string, token utils.OAuthTokenBundle, teamID, configurationID string) error {
	if configurationID == "" {
		return fmt.Errorf("vercel callback did not include a configuration id")
	}
	accountName := FetchVercelAccountName(ctx, token.AccessToken, teamID)
	name := "Vercel"
	if accountName != "" {
		name += " · " + accountName
	}
	return SaveOAuthConnection(ctx, db, OAuthConnection{
		AccountID: accountID, UserID: userID, Provider: "vercel",
		Name: name, ExternalID: configurationID, Token: token,
		Metadata: map[string]any{
			"configuration_id": configurationID, "team_id": teamID,
			"account_name": accountName, "scope": token.Scope,
		},
	})
}

type VercelService struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Framework string `json:"framework"`
	AccountID string `json:"accountId"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	Selected  bool   `json:"selected"`
}

type VercelDeployment struct {
	UID          string                 `json:"uid"`
	Name         string                 `json:"name"`
	URL          string                 `json:"url"`
	State        string                 `json:"state"`
	ReadyState   string                 `json:"readyState"`
	Target       string                 `json:"target"`
	Created      int64                  `json:"created"`
	BuildingAt   int64                  `json:"buildingAt"`
	Ready        int64                  `json:"ready"`
	Creator      map[string]interface{} `json:"creator"`
	Meta         map[string]interface{} `json:"meta"`
	InspectorURL string                 `json:"inspectorUrl"`
}

type updateVercelServicesRequest struct {
	ServiceIDs []string `json:"service_ids"`
}

func GetVercelServices(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")
		if !validConnectionIDs(accountID, connectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or connection ID"})
			return
		}

		token, teamID, selected, _, err := LoadVercelConnection(c, db, userID, accountID, connectionID)
		if errors.Is(err, errVercelConnectionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vercel connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open the Vercel connection"})
			return
		}
		services, err := FetchVercelServices(c, token, teamID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Vercel services could not be loaded"})
			return
		}
		for index := range services {
			services[index].Selected = selected[services[index].ID]
		}
		c.JSON(http.StatusOK, gin.H{"services": services})
	}
}

func UpdateVercelServices(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")
		if !validConnectionIDs(accountID, connectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or connection ID"})
			return
		}

		var request updateVercelServicesRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Choose the Vercel services to attach"})
			return
		}
		if len(request.ServiceIDs) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Too many services selected"})
			return
		}

		token, teamID, _, role, err := LoadVercelConnection(c, db, userID, accountID, connectionID)
		if errors.Is(err, errVercelConnectionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vercel connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open the Vercel connection"})
			return
		}
		if role != "owner" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only workspace owners and admins can change Vercel project access"})
			return
		}

		services, err := FetchVercelServices(c, token, teamID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Vercel services could not be verified"})
			return
		}
		selected, err := selectServices(services, request.ServiceIDs)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A selected Vercel service is no longer accessible"})
			return
		}
		err = saveServiceSelection(c, db, accountID, connectionID, "vercel", selected)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vercel connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save Vercel service selection"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"services": selected})
	}
}

func LoadVercelConnection(ctx context.Context, db *sql.DB, userID, accountID, connectionID string) (string, string, map[string]bool, string, error) {
	bundle, metadata, role, err := loadProviderConnection(ctx, db, userID, accountID, connectionID, "vercel")
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil, "", errVercelConnectionNotFound
	}
	if err != nil {
		return "", "", nil, "", err
	}
	var stored struct {
		TeamID string `json:"team_id"`
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &stored); err != nil {
			return "", "", nil, "", err
		}
	}
	selected, err := selectedServiceIDs[VercelService](metadata)
	if err != nil {
		return "", "", nil, "", err
	}
	return bundle.AccessToken, stored.TeamID, selected, role, nil
}

func FetchVercelServices(ctx context.Context, accessToken, teamID string) ([]VercelService, error) {
	services := make([]VercelService, 0)
	var until int64
	for page := 0; page < 10; page++ {
		query := url.Values{"limit": {"100"}}
		if teamID != "" {
			query.Set("teamId", teamID)
		}
		if until > 0 {
			query.Set("until", strconv.FormatInt(until, 10))
		}
		var result struct {
			Projects   []VercelService `json:"projects"`
			Pagination struct {
				Next *int64 `json:"next"`
			} `json:"pagination"`
		}
		if err := providerRequest(ctx, http.MethodGet, vercelAPIURL+"/v9/projects?"+query.Encode(), accessToken, nil, &result, 15*time.Second); err != nil {
			return nil, err
		}
		if result.Projects == nil {
			return nil, fmt.Errorf("vercel returned an invalid project list")
		}
		services = append(services, result.Projects...)
		if result.Pagination.Next == nil || *result.Pagination.Next == 0 {
			return services, nil
		}
		until = *result.Pagination.Next
	}
	return nil, fmt.Errorf("vercel service pagination limit reached")
}

func FetchVercelDeployments(ctx context.Context, accessToken, teamID, projectID string) ([]VercelDeployment, error) {
	query := url.Values{"limit": {"20"}, "projectId": {projectID}}
	if teamID != "" {
		query.Set("teamId", teamID)
	}
	var result struct {
		Deployments []VercelDeployment `json:"deployments"`
	}
	err := providerRequest(ctx, http.MethodGet, vercelAPIURL+"/v6/deployments?"+query.Encode(), accessToken, nil, &result, 15*time.Second)
	return result.Deployments, err
}

func (service VercelService) serviceID() string { return service.ID }

func (service VercelService) withSelection() VercelService {
	service.Selected = true
	return service
}

func FetchVercelAccountName(ctx context.Context, accessToken, teamID string) string {
	endpoint := vercelAPIURL + "/v2/user"
	if teamID != "" {
		endpoint = vercelAPIURL + "/v2/teams/" + url.PathEscape(teamID)
	}
	type account struct{ Name, Username, Slug string }
	var result struct {
		account
		User *account `json:"user"`
	}
	if err := providerRequest(ctx, http.MethodGet, endpoint, accessToken, nil, &result, 10*time.Second); err != nil {
		return ""
	}
	if result.User != nil {
		if result.User.Name != "" {
			return result.User.Name
		}
		return result.User.Username
	}
	if result.Name != "" {
		return result.Name
	}
	return result.Slug
}
