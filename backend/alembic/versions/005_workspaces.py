"""Rename the InfraMap project aggregate to workspace.

Revision ID: 005_workspaces
Revises: 004_project_resources
Create Date: 2026-09-13
"""

from alembic import op

revision = "005_workspaces"
down_revision = "004_project_resources"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.rename_table("projects", "workspaces", schema="public")
    op.execute("ALTER TABLE public.workspaces RENAME CONSTRAINT projects_pkey TO workspaces_pkey")
    op.execute("ALTER TABLE public.workspaces RENAME CONSTRAINT projects_account_slug_key TO workspaces_account_slug_key")
    op.execute("ALTER TABLE public.workspaces RENAME CONSTRAINT projects_status_check TO workspaces_status_check")
    op.execute("ALTER INDEX public.projects_account_id_idx RENAME TO workspaces_account_id_idx")

    op.rename_table("project_resources", "workspace_resources", schema="public")
    op.alter_column("workspace_resources", "project_id", new_column_name="workspace_id", schema="public")
    op.execute("ALTER TABLE public.workspace_resources RENAME CONSTRAINT project_resources_pkey TO workspace_resources_pkey")
    op.execute("ALTER TABLE public.workspace_resources RENAME CONSTRAINT project_resources_project_id_fkey TO workspace_resources_workspace_id_fkey")
    op.execute("ALTER TABLE public.workspace_resources RENAME CONSTRAINT project_resources_connection_id_fkey TO workspace_resources_connection_id_fkey")
    op.execute("ALTER TABLE public.workspace_resources RENAME CONSTRAINT project_resources_created_by_fkey TO workspace_resources_created_by_fkey")
    op.execute("ALTER TABLE public.workspace_resources RENAME CONSTRAINT project_resources_provider_type_key TO workspace_resources_provider_type_key")
    op.execute("ALTER INDEX public.project_resources_project_id_idx RENAME TO workspace_resources_workspace_id_idx")


def downgrade() -> None:
    op.execute("ALTER INDEX public.workspace_resources_workspace_id_idx RENAME TO project_resources_project_id_idx")
    op.execute("ALTER TABLE public.workspace_resources RENAME CONSTRAINT workspace_resources_provider_type_key TO project_resources_provider_type_key")
    op.execute("ALTER TABLE public.workspace_resources RENAME CONSTRAINT workspace_resources_created_by_fkey TO project_resources_created_by_fkey")
    op.execute("ALTER TABLE public.workspace_resources RENAME CONSTRAINT workspace_resources_connection_id_fkey TO project_resources_connection_id_fkey")
    op.execute("ALTER TABLE public.workspace_resources RENAME CONSTRAINT workspace_resources_workspace_id_fkey TO project_resources_project_id_fkey")
    op.execute("ALTER TABLE public.workspace_resources RENAME CONSTRAINT workspace_resources_pkey TO project_resources_pkey")
    op.alter_column("workspace_resources", "workspace_id", new_column_name="project_id", schema="public")
    op.rename_table("workspace_resources", "project_resources", schema="public")

    op.execute("ALTER INDEX public.workspaces_account_id_idx RENAME TO projects_account_id_idx")
    op.execute("ALTER TABLE public.workspaces RENAME CONSTRAINT workspaces_status_check TO projects_status_check")
    op.execute("ALTER TABLE public.workspaces RENAME CONSTRAINT workspaces_account_slug_key TO projects_account_slug_key")
    op.execute("ALTER TABLE public.workspaces RENAME CONSTRAINT workspaces_pkey TO projects_pkey")
    op.rename_table("workspaces", "projects", schema="public")
