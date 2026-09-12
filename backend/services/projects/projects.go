package projects

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"InfraMap/structs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var invalidSlugCharacters = regexp.MustCompile(`[^a-z0-9]+`)
var errInvalidProjectID = errors.New("invalid workspace or project ID")

func GetProjects(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		if _, err := uuid.Parse(accountID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
			return
		}

		if _, err := membershipRole(c, db, userID, accountID); err != nil {
			writeMembershipError(c, err)
			return
		}

		rows, err := db.QueryContext(c, `
			SELECT id, account_id, name, slug, description, status,
			       COALESCE(created_by::text, ''), created_at, updated_at
			FROM projects
			WHERE account_id = $1
			ORDER BY updated_at DESC`, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load projects"})
			return
		}
		defer rows.Close()

		available := make([]structs.Project, 0)
		for rows.Next() {
			var project structs.Project
			if err := rows.Scan(&project.ID, &project.AccountID, &project.Name, &project.Slug, &project.Description, &project.Status, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load projects"})
				return
			}
			available = append(available, project)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load projects"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"projects": available})
	}
}

func GetProject(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		projectID := c.Param("projectID")
		project, role, err := loadProject(c, db, userID, accountID, projectID)
		if err != nil {
			writeProjectAccessError(c, err)
			return
		}

		resources, err := loadProjectResources(c, db, projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load project resources"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"project": project, "resources": resources, "role": role})
	}
}

func CreateProject(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		if _, err := uuid.Parse(accountID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
			return
		}

		role, err := membershipRole(c, db, userID, accountID)
		if err != nil {
			writeMembershipError(c, err)
			return
		}
		if role == "viewer" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Viewers cannot create projects"})
			return
		}

		var request structs.CreateProjectRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Enter a project name"})
			return
		}
		request.Name = strings.TrimSpace(request.Name)
		request.Description = strings.TrimSpace(request.Description)
		if request.Name == "" || utf8.RuneCountInString(request.Name) > 80 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project names must be between 1 and 80 characters"})
			return
		}
		if utf8.RuneCountInString(request.Description) > 280 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project descriptions cannot exceed 280 characters"})
			return
		}

		projectID := uuid.New()
		slug := projectSlug(request.Name)
		project, err := insertProject(c, db, projectID, accountID, userID, request, slug)
		if isUniqueViolation(err) {
			slug += "-" + projectID.String()[:8]
			project, err = insertProject(c, db, projectID, accountID, userID, request, slug)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create project"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"project": project})
	}
}

func insertProject(c *gin.Context, db *sql.DB, projectID uuid.UUID, accountID, userID string, request structs.CreateProjectRequest, slug string) (structs.Project, error) {
	var project structs.Project
	err := db.QueryRowContext(c, `
		INSERT INTO projects (id, account_id, name, slug, description, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'draft', $6, NOW(), NOW())
		RETURNING id, account_id, name, slug, description, status,
		          COALESCE(created_by::text, ''), created_at, updated_at`,
		projectID, accountID, request.Name, slug, request.Description, userID,
	).Scan(&project.ID, &project.AccountID, &project.Name, &project.Slug, &project.Description, &project.Status, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt)
	return project, err
}

func membershipRole(c *gin.Context, db *sql.DB, userID, accountID string) (string, error) {
	var role string
	err := db.QueryRowContext(c, `SELECT role FROM account_memberships WHERE account_id = $1 AND user_id = $2`, accountID, userID).Scan(&role)
	return role, err
}

func loadProject(c *gin.Context, db *sql.DB, userID, accountID, projectID string) (structs.Project, string, error) {
	if _, err := uuid.Parse(accountID); err != nil {
		return structs.Project{}, "", errInvalidProjectID
	}
	if _, err := uuid.Parse(projectID); err != nil {
		return structs.Project{}, "", errInvalidProjectID
	}

	var project structs.Project
	var role string
	err := db.QueryRowContext(c, `
		SELECT p.id, p.account_id, p.name, p.slug, p.description, p.status,
		       COALESCE(p.created_by::text, ''), p.created_at, p.updated_at, am.role
		FROM projects p
		JOIN account_memberships am ON am.account_id = p.account_id
		WHERE p.id = $1 AND p.account_id = $2 AND am.user_id = $3`, projectID, accountID, userID).Scan(
		&project.ID, &project.AccountID, &project.Name, &project.Slug, &project.Description,
		&project.Status, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt, &role,
	)
	return project, role, err
}

func loadProjectResources(c *gin.Context, db *sql.DB, projectID string) ([]structs.ProjectResource, error) {
	rows, err := db.QueryContext(c, `
		SELECT id, project_id, connection_id, provider, resource_type, external_id,
		       name, status, metadata, COALESCE(created_by::text, ''), created_at, updated_at
		FROM project_resources
		WHERE project_id = $1
		ORDER BY created_at ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resources := make([]structs.ProjectResource, 0)
	for rows.Next() {
		var resource structs.ProjectResource
		var metadata []byte
		if err := rows.Scan(&resource.ID, &resource.ProjectID, &resource.ConnectionID, &resource.Provider, &resource.ResourceType, &resource.ExternalID, &resource.Name, &resource.Status, &metadata, &resource.CreatedBy, &resource.CreatedAt, &resource.UpdatedAt); err != nil {
			return nil, err
		}
		resource.Metadata = make(map[string]interface{})
		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &resource.Metadata); err != nil {
				return nil, err
			}
		}
		resources = append(resources, resource)
	}
	return resources, rows.Err()
}

func writeMembershipError(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this workspace"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify workspace access"})
}

func writeProjectAccessError(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if errors.Is(err, errInvalidProjectID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or project ID"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load project"})
}

func projectSlug(name string) string {
	slug := strings.Trim(invalidSlugCharacters.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if slug == "" {
		return "project"
	}
	if len(slug) > 64 {
		slug = strings.Trim(slug[:64], "-")
	}
	return slug
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
