"""Add OAuth identities for third-party sign-in.

Revision ID: 002_oauth_identities
Revises: 001_initial_schema
Create Date: 2026-09-10
"""

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

revision = "002_oauth_identities"
down_revision = "001_initial_schema"
branch_labels = None
depends_on = None

def upgrade() -> None:
    op.create_table(
        "user_identities",
        sa.Column("id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("provider", sa.Text(), nullable=False),
        sa.Column("provider_id", sa.Text(), nullable=False),
        sa.Column("email", sa.String(length=254), nullable=False),
        sa.Column("username", sa.Text(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.text("now()")),
        sa.ForeignKeyConstraint(["user_id"], ["public.users.id"], ondelete="CASCADE"),
        sa.PrimaryKeyConstraint("id"),
        sa.UniqueConstraint("provider", "provider_id", name="user_identities_provider_subject_key"),
        schema="public",
    )
    op.create_index("user_identities_user_id_idx", "user_identities", ["user_id"], schema="public")

def downgrade() -> None:
    op.drop_table("user_identities", schema="public")
