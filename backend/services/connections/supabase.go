package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"InfraMap/utils"

	"github.com/gin-gonic/gin"
)

const supabaseManagementAPI = "https://api.supabase.com"

var errSupabaseConnectionNotFound = errors.New("supabase connection not found")

func ConnectSupabaseIntegration(ctx context.Context, db *sql.DB, accountID, userID string, token utils.OAuthTokenBundle) error {
	organizations, err := FetchSupabaseOrganizations(ctx, token.AccessToken)
	if err != nil {
		return err
	}
	if len(organizations) == 0 {
		return fmt.Errorf("supabase returned no authorised organization")
	}
	organization := organizations[0]
	return SaveOAuthConnection(ctx, db, OAuthConnection{
		AccountID: accountID, UserID: userID, Provider: "supabase",
		Name: "Supabase · " + organization.Name, ExternalID: organization.ID,
		Token:    token,
		Metadata: map[string]any{"organization": organization, "organizations": organizations, "scope": token.Scope},
	})
}

type SupabaseService struct {
	ID               string                   `json:"id"`
	Ref              string                   `json:"ref"`
	OrganizationID   string                   `json:"organization_id"`
	OrganizationSlug string                   `json:"organization_slug"`
	Name             string                   `json:"name"`
	Region           string                   `json:"region"`
	CreatedAt        string                   `json:"created_at"`
	Status           string                   `json:"status"`
	Database         *SupabaseServiceDatabase `json:"database,omitempty"`
	Selected         bool                     `json:"selected"`
}

type SupabaseServiceDatabase struct {
	Host           string `json:"host"`
	Version        string `json:"version"`
	PostgresEngine string `json:"postgres_engine"`
	ReleaseChannel string `json:"release_channel"`
}

type SupabaseServiceHealth struct {
	Name    string                 `json:"name"`
	Healthy bool                   `json:"healthy"`
	Status  string                 `json:"status"`
	Info    map[string]interface{} `json:"info"`
	Error   string                 `json:"error"`
}

type SupabaseUsagePoint struct {
	Timestamp             string `json:"timestamp"`
	TotalAuthRequests     int64  `json:"total_auth_requests"`
	TotalRealtimeRequests int64  `json:"total_realtime_requests"`
	TotalRestRequests     int64  `json:"total_rest_requests"`
	TotalStorageRequests  int64  `json:"total_storage_requests"`
}

type SupabaseOrganization struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type SupabaseRegionSelection struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

type CreateSupabaseServiceInput struct {
	DatabasePassword string                  `json:"db_pass"`
	Name             string                  `json:"name"`
	OrganizationSlug string                  `json:"organization_slug"`
	RegionSelection  SupabaseRegionSelection `json:"region_selection"`
}

type updateSupabaseServicesRequest struct {
	ServiceRefs []string `json:"service_refs"`
}

func GetSupabaseServices(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")

		if !validConnectionIDs(accountID, connectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or connection ID"})
			return
		}

		token, selected, _, err := LoadSupabaseConnection(c, db, userID, accountID, connectionID)
		if errors.Is(err, errSupabaseConnectionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supabase connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open the Supabase connection"})
			return
		}

		services, err := FetchSupabaseServices(c, token)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Supabase services could not be loaded"})
			return
		}
		for index := range services {
			services[index].Selected = selected[services[index].Ref]
		}
		c.JSON(http.StatusOK, gin.H{"services": services})
	}
}

func UpdateSupabaseServices(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")

		if !validConnectionIDs(accountID, connectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or connection ID"})
			return
		}

		var request updateSupabaseServicesRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Choose the Supabase services to attach"})
			return
		}
		if len(request.ServiceRefs) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Too many services selected"})
			return
		}

		token, _, role, err := LoadSupabaseConnection(c, db, userID, accountID, connectionID)
		if errors.Is(err, errSupabaseConnectionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supabase connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open the Supabase connection"})
			return
		}
		if role != "owner" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only workspace owners and admins can change project access"})
			return
		}

		services, err := FetchSupabaseServices(c, token)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Supabase services could not be verified"})
			return
		}
		selected, err := selectServices(services, request.ServiceRefs)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A selected Supabase service is no longer accessible"})
			return
		}
		err = saveServiceSelection(c, db, accountID, connectionID, "supabase", selected)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supabase connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save Supabase service selection"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"services": selected})
	}
}

func LoadSupabaseConnection(ctx context.Context, db *sql.DB, userID, accountID, connectionID string) (string, map[string]bool, string, error) {
	bundle, metadata, role, err := loadProviderConnection(ctx, db, userID, accountID, connectionID, "supabase")
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil, "", errSupabaseConnectionNotFound
	}
	if err != nil {
		return "", nil, "", err
	}
	if !bundle.ExpiresAt.IsZero() && time.Now().Add(time.Minute).After(bundle.ExpiresAt) {
		bundle, err = refreshSupabaseToken(ctx, bundle)
		if err != nil {
			return "", nil, "", err
		}
		if err := saveSupabaseToken(ctx, db, connectionID, bundle); err != nil {
			return "", nil, "", err
		}
	}
	selected, err := selectedServiceIDs[SupabaseService](metadata)
	if err != nil {
		return "", nil, "", err
	}
	return bundle.AccessToken, selected, role, nil
}

func FetchSupabaseServices(ctx context.Context, accessToken string) ([]SupabaseService, error) {
	var services []SupabaseService
	err := providerRequest(ctx, http.MethodGet, supabaseManagementAPI+"/v1/projects", accessToken, nil, &services, 15*time.Second)
	return services, err
}

func FetchSupabaseServiceHealth(ctx context.Context, accessToken, projectRef string) ([]SupabaseServiceHealth, error) {
	query := url.Values{"services": {"auth,db,realtime,rest,storage"}}
	endpoint := supabaseManagementAPI + "/v1/projects/" + url.PathEscape(projectRef) + "/health?" + query.Encode()
	var services []SupabaseServiceHealth
	err := providerRequest(ctx, http.MethodGet, endpoint, accessToken, nil, &services, 15*time.Second)
	return services, err
}

func FetchSupabaseRequestActivity(ctx context.Context, accessToken, projectRef string) ([]SupabaseUsagePoint, error) {
	end := time.Now().UTC().Truncate(time.Minute)
	query := url.Values{
		"iso_timestamp_start": {end.Add(-24 * time.Hour).Format(time.RFC3339)},
		"iso_timestamp_end":   {end.Format(time.RFC3339)},
		"sql": {`SELECT
formatDateTime(toStartOfHour(timestamp), '%Y-%m-%dT%H:%i:%SZ', 'UTC') AS timestamp,
countIf(source = 'auth_logs') AS total_auth_requests,
countIf(source = 'realtime_logs') AS total_realtime_requests,
countIf(source = 'edge_logs') AS total_rest_requests,
countIf(source = 'storage_logs') AS total_storage_requests
FROM logs
WHERE source IN ('auth_logs', 'realtime_logs', 'edge_logs', 'storage_logs')
GROUP BY timestamp
ORDER BY timestamp ASC`},
	}
	endpoint := supabaseManagementAPI + "/v1/projects/" + url.PathEscape(projectRef) + "/analytics/endpoints/logs?" + query.Encode()
	var result struct {
		Result []struct {
			Timestamp             string          `json:"timestamp"`
			TotalAuthRequests     json.RawMessage `json:"total_auth_requests"`
			TotalRealtimeRequests json.RawMessage `json:"total_realtime_requests"`
			TotalRestRequests     json.RawMessage `json:"total_rest_requests"`
			TotalStorageRequests  json.RawMessage `json:"total_storage_requests"`
		} `json:"result"`
		Error json.RawMessage `json:"error"`
	}
	if err := providerRequest(ctx, http.MethodGet, endpoint, accessToken, nil, &result, 15*time.Second); err != nil {
		return nil, err
	}
	if len(result.Error) > 0 && string(result.Error) != "null" && string(result.Error) != `""` {
		return nil, fmt.Errorf("supabase analytics query failed: %s", strings.TrimSpace(string(result.Error)))
	}
	points := make([]SupabaseUsagePoint, 0, len(result.Result))
	for _, item := range result.Result {
		points = append(points, SupabaseUsagePoint{
			Timestamp:             item.Timestamp,
			TotalAuthRequests:     analyticsCount(item.TotalAuthRequests),
			TotalRealtimeRequests: analyticsCount(item.TotalRealtimeRequests),
			TotalRestRequests:     analyticsCount(item.TotalRestRequests),
			TotalStorageRequests:  analyticsCount(item.TotalStorageRequests),
		})
	}
	return points, nil
}

func analyticsCount(value json.RawMessage) int64 {
	text := strings.Trim(string(value), `"`)
	var count int64
	_, _ = fmt.Sscan(text, &count)
	return count
}

func FetchSupabaseOrganizations(ctx context.Context, accessToken string) ([]SupabaseOrganization, error) {
	var organizations []SupabaseOrganization
	err := providerRequest(ctx, http.MethodGet, supabaseManagementAPI+"/v1/organizations", accessToken, nil, &organizations, 15*time.Second)
	return organizations, err
}

func CreateSupabaseService(ctx context.Context, accessToken string, input CreateSupabaseServiceInput) (SupabaseService, error) {
	var service SupabaseService
	err := providerRequest(ctx, http.MethodPost, supabaseManagementAPI+"/v1/projects", accessToken, input, &service, 30*time.Second)
	return service, err
}

func refreshSupabaseToken(ctx context.Context, current utils.OAuthTokenBundle) (utils.OAuthTokenBundle, error) {
	if current.RefreshToken == "" {
		return utils.OAuthTokenBundle{}, fmt.Errorf("supabase refresh token is missing")
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {current.RefreshToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, utils.Cfg.SupabaseOAuth.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return utils.OAuthTokenBundle{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(utils.Cfg.SupabaseOAuth.ClientID, utils.Cfg.SupabaseOAuth.ClientSecret)

	response, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return utils.OAuthTokenBundle{}, err
	}
	defer response.Body.Close()

	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return utils.OAuthTokenBundle{}, err
	}
	if response.StatusCode >= 300 || token.AccessToken == "" {
		return utils.OAuthTokenBundle{}, fmt.Errorf("supabase token refresh failed: %s", token.Error)
	}
	if token.RefreshToken == "" {
		token.RefreshToken = current.RefreshToken
	}
	return utils.OAuthTokenBundle{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Scope:        token.Scope,
		ExpiresAt:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	}, nil
}

func saveSupabaseToken(ctx context.Context, db *sql.DB, connectionID string, bundle utils.OAuthTokenBundle) error {
	encrypted, err := encryptOAuthToken(bundle)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `UPDATE connections SET secret_ref = $1, updated_at = NOW() WHERE id = $2 AND provider = 'supabase'`, encrypted, connectionID)
	return err
}

func (service SupabaseService) serviceID() string { return service.Ref }

func (service SupabaseService) withSelection() SupabaseService {
	service.Selected = true
	return service
}
