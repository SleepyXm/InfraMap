package services

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"InfraMap/structs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetConnections(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")

		if _, err := uuid.Parse(accountID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
			return
		}

		hasAccess, err := userHasAccountAccess(c, db, userID, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify account access"})
			return
		}
		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this account"})
			return
		}

		rows, err := db.QueryContext(c,
			`SELECT
				id,
				account_id,
				name,
				provider,
				connection_type,
				COALESCE(external_account_id, ''),
				metadata,
				COALESCE(created_by::text, ''),
				created_at,
				updated_at
			 FROM connections
			 WHERE account_id = $1
			 ORDER BY created_at DESC`,
			accountID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load connections"})
			return
		}
		defer rows.Close()

		connections := make([]structs.Connection, 0)

		for rows.Next() {
			var connection structs.Connection
			var metadata []byte

			if err := rows.Scan(
				&connection.ID,
				&connection.AccountID,
				&connection.Name,
				&connection.Provider,
				&connection.ConnectionType,
				&connection.ExternalAccountID,
				&metadata,
				&connection.CreatedBy,
				&connection.CreatedAt,
				&connection.UpdatedAt,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load connections"})
				return
			}

			if len(metadata) > 0 {
				if err := json.Unmarshal(metadata, &connection.Metadata); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not decode connection metadata"})
					return
				}
			}
			connection.Category = connectionCategory(connection.Provider)
			connection.Status = connectionStatus(connection.Metadata)

			connections = append(connections, connection)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load connections"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"connections": connections,
		})
	}
}

func GetConnection(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")

		if _, err := uuid.Parse(accountID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
			return
		}

		if _, err := uuid.Parse(connectionID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
			return
		}

		var connection structs.Connection
		var metadata []byte

		err := db.QueryRowContext(c,
			`SELECT
				c.id,
				c.account_id,
				c.name,
				c.provider,
				c.connection_type,
				COALESCE(c.external_account_id, ''),
				c.metadata,
				COALESCE(c.created_by::text, ''),
				c.created_at,
				c.updated_at
			 FROM connections c
			 JOIN account_memberships am
			   ON am.account_id = c.account_id
			 WHERE c.id = $1
			   AND c.account_id = $2
			   AND am.user_id = $3`,
			connectionID,
			accountID,
			userID,
		).Scan(
			&connection.ID,
			&connection.AccountID,
			&connection.Name,
			&connection.Provider,
			&connection.ConnectionType,
			&connection.ExternalAccountID,
			&metadata,
			&connection.CreatedBy,
			&connection.CreatedAt,
			&connection.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load connection"})
			return
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &connection.Metadata); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not decode connection metadata"})
				return
			}
		}
		connection.Category = connectionCategory(connection.Provider)
		connection.Status = connectionStatus(connection.Metadata)

		c.JSON(http.StatusOK, gin.H{
			"connection": connection,
		})
	}
}

func CreateConnection(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")

		if _, err := uuid.Parse(accountID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
			return
		}

		var req structs.CreateConnectionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection request"})
			return
		}

		hasAccess, err := userHasAccountAccess(c, db, userID, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify account access"})
			return
		}
		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this account"})
			return
		}

		metadata, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection metadata"})
			return
		}

		connectionID := uuid.New()

		var connection structs.Connection

		err = db.QueryRowContext(c,
			`INSERT INTO connections (
				id,
				account_id,
				name,
				provider,
				connection_type,
				external_account_id,
				metadata,
				created_by,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, NOW(), NOW())
			RETURNING
				id,
				account_id,
				name,
				provider,
				connection_type,
				COALESCE(external_account_id, ''),
				COALESCE(created_by::text, ''),
				created_at,
				updated_at`,
			connectionID,
			accountID,
			req.Name,
			req.Provider,
			req.ConnectionType,
			req.ExternalAccountID,
			metadata,
			userID,
		).Scan(
			&connection.ID,
			&connection.AccountID,
			&connection.Name,
			&connection.Provider,
			&connection.ConnectionType,
			&connection.ExternalAccountID,
			&connection.CreatedBy,
			&connection.CreatedAt,
			&connection.UpdatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create connection"})
			return
		}

		connection.Metadata = req.Metadata
		connection.Category = req.Category
		connection.Status = structs.ConnectionStatusActive

		c.JSON(http.StatusCreated, gin.H{
			"connection": connection,
		})
	}
}

func connectionCategory(provider string) structs.ConnectionCategory {
	switch provider {
	case "github", "gitlab", "bitbucket":
		return structs.ConnectionCategorySource
	case "aws", "gcp", "azure", "cloudflare", "digitalocean", "hetzner":
		return structs.ConnectionCategoryCloud
	case "vercel", "railway", "render", "netlify":
		return structs.ConnectionCategoryDeployment
	case "supabase", "neon", "planetscale":
		return structs.ConnectionCategoryData
	case "kubernetes", "docker", "ssh":
		return structs.ConnectionCategoryServer
	default:
		return structs.ConnectionCategoryAccount
	}
}

func connectionStatus(metadata map[string]any) structs.ConnectionStatus {
	status, _ := metadata["health_status"].(string)
	switch structs.ConnectionStatus(status) {
	case structs.ConnectionStatusPending, structs.ConnectionStatusInvalid, structs.ConnectionStatusDisabled:
		return structs.ConnectionStatus(status)
	default:
		return structs.ConnectionStatusActive
	}
}

func DeleteConnection(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")

		if _, err := uuid.Parse(accountID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
			return
		}

		if _, err := uuid.Parse(connectionID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
			return
		}

		hasAccess, err := userHasAccountAccess(c, db, userID, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify account access"})
			return
		}
		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this account"})
			return
		}

		result, err := db.ExecContext(c,
			`DELETE FROM connections
			 WHERE id = $1
			   AND account_id = $2`,
			connectionID,
			accountID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete connection"})
			return
		}

		rows, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete connection"})
			return
		}

		if rows == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Connection deleted successfully",
		})
	}
}

func userHasAccountAccess(
	c *gin.Context,
	db *sql.DB,
	userID string,
	accountID string,
) (bool, error) {
	var exists bool

	err := db.QueryRowContext(c,
		`SELECT EXISTS (
			SELECT 1
			FROM account_memberships
			WHERE account_id = $1
			  AND user_id = $2
		)`,
		accountID,
		userID,
	).Scan(&exists)

	return exists, err
}
