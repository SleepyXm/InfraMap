package workspaces

import (
	"database/sql"
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
var errInvalidWorkspaceID = errors.New("invalid account or workspace ID")

func GetWorkspaces(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		if _, err := uuid.Parse(accountID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
			return
		}

		if _, err := membershipRole(c, db, userID, accountID); err != nil {
			writeMembershipError(c, err)
			return
		}

		rows, err := db.QueryContext(c, `
			SELECT id, account_id, name, slug, description, status,
			       COALESCE(created_by::text, ''), created_at, updated_at
			FROM workspaces
			WHERE account_id = $1
			ORDER BY updated_at DESC`, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load workspaces"})
			return
		}
		defer rows.Close()

		available := make([]structs.Workspace, 0)
		for rows.Next() {
			var workspace structs.Workspace
			if err := rows.Scan(&workspace.ID, &workspace.AccountID, &workspace.Name, &workspace.Slug, &workspace.Description, &workspace.Status, &workspace.CreatedBy, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load workspaces"})
				return
			}
			available = append(available, workspace)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load workspaces"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"workspaces": available})
	}
}

func GetWorkspace(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		workspaceID := c.Param("workspaceID")
		workspace, role, err := LoadWorkspace(c, db, userID, accountID, workspaceID)
		if err != nil {
			WriteWorkspaceAccessError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"workspace": workspace, "access": role})
	}
}

func DeleteWorkspace(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID, workspaceID := c.Param("accountID"), c.Param("workspaceID")
		_, role, err := LoadWorkspace(c, db, userID, accountID, workspaceID)
		if err != nil {
			WriteWorkspaceAccessError(c, err)
			return
		}
		if role != "owner" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only account owners and admins can delete workspaces"})
			return
		}
		result, err := db.ExecContext(c, `DELETE FROM workspaces
			WHERE id = $1 AND account_id = $2 AND EXISTS (
				SELECT 1 FROM account_memberships
				WHERE account_id = $2 AND user_id = $3 AND role IN ('owner', 'admin')
			)`, workspaceID, accountID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Workspace could not be deleted"})
			return
		}
		rows, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Workspace deletion could not be confirmed"})
			return
		}
		if rows != 1 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workspace no longer exists or access has changed"})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func CreateWorkspace(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		if _, err := uuid.Parse(accountID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
			return
		}

		role, err := membershipRole(c, db, userID, accountID)
		if err != nil {
			writeMembershipError(c, err)
			return
		}
		if role == "viewer" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Viewers cannot create workspaces"})
			return
		}

		var request structs.CreateWorkspaceRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Enter a workspace name"})
			return
		}
		request.Name = strings.TrimSpace(request.Name)
		request.Description = strings.TrimSpace(request.Description)
		if request.Name == "" || utf8.RuneCountInString(request.Name) > 80 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Workspace names must be between 1 and 80 characters"})
			return
		}
		if utf8.RuneCountInString(request.Description) > 280 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Workspace descriptions cannot exceed 280 characters"})
			return
		}

		workspaceID := uuid.New()
		slug := workspaceSlug(request.Name)
		workspace, err := insertWorkspace(c, db, workspaceID, accountID, userID, request, slug)
		if isUniqueViolation(err) {
			slug += "-" + workspaceID.String()[:8]
			workspace, err = insertWorkspace(c, db, workspaceID, accountID, userID, request, slug)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create workspace"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"workspace": workspace})
	}
}

func insertWorkspace(c *gin.Context, db *sql.DB, workspaceID uuid.UUID, accountID, userID string, request structs.CreateWorkspaceRequest, slug string) (structs.Workspace, error) {
	var workspace structs.Workspace
	err := db.QueryRowContext(c, `
		INSERT INTO workspaces (id, account_id, name, slug, description, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'draft', $6, NOW(), NOW())
		RETURNING id, account_id, name, slug, description, status,
		          COALESCE(created_by::text, ''), created_at, updated_at`,
		workspaceID, accountID, request.Name, slug, request.Description, userID,
	).Scan(&workspace.ID, &workspace.AccountID, &workspace.Name, &workspace.Slug, &workspace.Description, &workspace.Status, &workspace.CreatedBy, &workspace.CreatedAt, &workspace.UpdatedAt)
	return workspace, err
}

func membershipRole(c *gin.Context, db *sql.DB, userID, accountID string) (string, error) {
	var role string
	err := db.QueryRowContext(c, `SELECT role FROM account_memberships WHERE account_id = $1 AND user_id = $2`, accountID, userID).Scan(&role)
	return role, err
}

func LoadWorkspace(c *gin.Context, db *sql.DB, userID, accountID, workspaceID string) (structs.Workspace, string, error) {
	if _, err := uuid.Parse(accountID); err != nil {
		return structs.Workspace{}, "", errInvalidWorkspaceID
	}
	if _, err := uuid.Parse(workspaceID); err != nil {
		return structs.Workspace{}, "", errInvalidWorkspaceID
	}

	var workspace structs.Workspace
	var role string
	err := db.QueryRowContext(c, `
		SELECT p.id, p.account_id, p.name, p.slug, p.description, p.status,
		       COALESCE(p.created_by::text, ''), p.created_at, p.updated_at, am.role
		FROM workspaces p
		JOIN account_memberships am ON am.account_id = p.account_id
		WHERE p.id = $1 AND p.account_id = $2 AND am.user_id = $3`, workspaceID, accountID, userID).Scan(
		&workspace.ID, &workspace.AccountID, &workspace.Name, &workspace.Slug, &workspace.Description,
		&workspace.Status, &workspace.CreatedBy, &workspace.CreatedAt, &workspace.UpdatedAt, &role,
	)
	return workspace, role, err
}

func writeMembershipError(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this workspace"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify workspace access"})
}

func WriteWorkspaceAccessError(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
		return
	}
	if errors.Is(err, errInvalidWorkspaceID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account or workspace ID"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load workspace"})
}

func workspaceSlug(name string) string {
	slug := strings.Trim(invalidSlugCharacters.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if slug == "" {
		return "workspace"
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
