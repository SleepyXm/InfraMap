package workspaces

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

const deleteTestAccount = "11111111-1111-4111-8111-111111111111"
const deleteTestWorkspace = "22222222-2222-4222-8222-222222222222"

type deletionDB struct {
	role                string
	missing             bool
	execError           error
	affected            int64
	queryArgs, execArgs []driver.NamedValue
	statement           string
}

func (db *deletionDB) Connect(context.Context) (driver.Conn, error) { return &deletionConn{db}, nil }
func (db *deletionDB) Driver() driver.Driver                        { return deletionDriver{} }

type deletionDriver struct{}

func (deletionDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type deletionConn struct{ db *deletionDB }

func (*deletionConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (*deletionConn) Close() error              { return nil }
func (*deletionConn) Begin() (driver.Tx, error) { return nil, errors.New("unexpected transaction") }
func (c *deletionConn) QueryContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Rows, error) {
	c.db.queryArgs = args
	return &deletionRows{db: c.db}, nil
}
func (c *deletionConn) ExecContext(_ context.Context, statement string, args []driver.NamedValue) (driver.Result, error) {
	c.db.statement, c.db.execArgs = statement, args
	return driver.RowsAffected(c.db.affected), c.db.execError
}

type deletionRows struct {
	db   *deletionDB
	read bool
}

func (*deletionRows) Columns() []string {
	return []string{"id", "account_id", "name", "slug", "description", "status", "created_by", "created_at", "updated_at", "role"}
}
func (*deletionRows) Close() error { return nil }
func (r *deletionRows) Next(values []driver.Value) error {
	if r.read || r.db.missing {
		return io.EOF
	}
	r.read = true
	copy(values, []driver.Value{deleteTestWorkspace, deleteTestAccount, "Test", "test", "", "active", "user", time.Now(), time.Now(), r.db.role})
	return nil
}

func TestDeleteWorkspaceAccessAndFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name, role string
		missing    bool
		execError  error
		affected   int64
		status     int
		exec       bool
	}{
		{"owner", "owner", false, nil, 1, 204, true},
		{"admin", "admin", false, nil, 1, 204, true},
		{"member", "member", false, nil, 1, 403, false},
		{"viewer", "viewer", false, nil, 1, 403, false},
		{"missing or foreign workspace", "", true, nil, 0, 404, false},
		{"database failure", "owner", false, errors.New("database unavailable"), 0, 500, true},
		{"already deleted or access revoked", "owner", false, nil, 0, 404, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			connector := &deletionDB{role: test.role, missing: test.missing, execError: test.execError, affected: test.affected}
			db := sql.OpenDB(connector)
			defer db.Close()
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Set("userID", "user")
			c.Params = gin.Params{{Key: "accountID", Value: deleteTestAccount}, {Key: "workspaceID", Value: deleteTestWorkspace}}
			DeleteWorkspace(db)(c)
			c.Writer.WriteHeaderNow()
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d: %s", recorder.Code, test.status, recorder.Body.String())
			}
			if (connector.statement != "") != test.exec {
				t.Fatalf("unexpected delete execution: %q", connector.statement)
			}
			if connector.queryArgs[0].Value != deleteTestWorkspace || connector.queryArgs[1].Value != deleteTestAccount || connector.queryArgs[2].Value != "user" {
				t.Fatal("access lookup was not scoped to the user, account and workspace")
			}
			if test.exec {
				if !strings.Contains(connector.statement, "DELETE FROM workspaces") || !strings.Contains(connector.statement, "role IN ('owner', 'admin')") {
					t.Fatal("delete must retain its permission check")
				}
				if connector.execArgs[0].Value != deleteTestWorkspace || connector.execArgs[1].Value != deleteTestAccount || connector.execArgs[2].Value != "user" {
					t.Fatal("delete was not scoped to the user, account and workspace")
				}
			}
		})
	}
}

func TestDeleteWorkspaceRejectsInvalidIDsBeforeDatabaseAccess(t *testing.T) {
	for _, ids := range [][2]string{{"invalid", deleteTestWorkspace}, {deleteTestAccount, "invalid"}} {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Set("userID", "user")
		c.Params = gin.Params{{Key: "accountID", Value: ids[0]}, {Key: "workspaceID", Value: ids[1]}}
		DeleteWorkspace(nil)(c)
		if recorder.Code != 400 {
			t.Fatalf("status = %d, want 400", recorder.Code)
		}
	}
}
