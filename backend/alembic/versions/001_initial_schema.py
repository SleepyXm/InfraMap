"""Baseline the existing InfraMap schema. Covers Users, Accounts, Connections, and Instances tables.

Revision ID: 001_initial_schema
Revises:
Create Date: 2026-08-15
"""

from typing import Sequence, Union

from alembic import context, op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

revision: str = "001_initial_schema"
down_revision: Union[str, Sequence[str], None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None

def upgrade() -> None:
    op.execute("CREATE EXTENSION IF NOT EXISTS pgcrypto")
    inspector = None if context.is_offline_mode() else sa.inspect(op.get_bind())

    def table_missing(name: str) -> bool:
        return inspector is None or not inspector.has_table(name, schema="public")

    if table_missing("users"):
        op.create_table(
            "users",
            sa.Column("id", postgresql.UUID(as_uuid=True), nullable=False, server_default=sa.text("gen_random_uuid()")),
            sa.Column("username", sa.Text(), nullable=False),
            sa.Column("password", sa.Text(), nullable=False),
            sa.Column("email", sa.String(length=254), nullable=True),
            sa.Column("verified", sa.Boolean(), nullable=False, server_default=sa.text("false")),
            sa.Column("verification_token", sa.Text(), nullable=True),
            sa.Column("favourites", postgresql.ARRAY(sa.Text()), nullable=True, server_default=sa.text("'{}'::text[]")),
            sa.Column("created_at", sa.DateTime(timezone=True), nullable=True, server_default=sa.text("now()")),
            sa.Column("updated_at", sa.DateTime(timezone=True), nullable=True, server_default=sa.text("now()")),
            sa.PrimaryKeyConstraint("id", name="users_pkey"),
            sa.UniqueConstraint("username", name="users_username_key"),
            schema="public",
        )
    if table_missing("accounts"):
        op.create_table(
            "accounts",
            sa.Column("id", postgresql.UUID(as_uuid=True), nullable=False, server_default=sa.text("gen_random_uuid()")),
            sa.Column("name", sa.Text(), nullable=False),
            sa.Column("slug", sa.Text(), nullable=False),
            sa.Column("type", sa.Text(), nullable=False, server_default=sa.text("'organization'")),
            sa.Column("created_by", postgresql.UUID(as_uuid=True), nullable=True),
            sa.Column("created_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.text("now()")),
            sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.text("now()")),
            sa.ForeignKeyConstraint(["created_by"], ["public.users.id"], name="accounts_created_by_fkey", ondelete="SET NULL"),
            sa.PrimaryKeyConstraint("id", name="accounts_pkey"),
            sa.UniqueConstraint("slug", name="accounts_slug_key"),
            sa.CheckConstraint("type IN ('personal', 'organization')", name="accounts_type_check"),
            schema="public",
        )

    if table_missing("account_memberships"):
        op.create_table(
            "account_memberships",
            sa.Column("account_id", postgresql.UUID(as_uuid=True), nullable=False),
            sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
            sa.Column("role", sa.Text(), nullable=False, server_default=sa.text("'member'")),
            sa.Column("created_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.text("now()")),
            sa.ForeignKeyConstraint(["account_id"], ["public.accounts.id"], name="account_memberships_account_id_fkey", ondelete="CASCADE"),
            sa.ForeignKeyConstraint(["user_id"], ["public.users.id"], name="account_memberships_user_id_fkey", ondelete="CASCADE"),
            sa.PrimaryKeyConstraint("account_id", "user_id", name="account_memberships_pkey"),
            sa.CheckConstraint("role IN ('owner', 'admin', 'member', 'viewer')", name="account_memberships_role_check"),
            schema="public",
        )

    if table_missing("connections"):
        op.create_table(
            "connections",
            sa.Column("id", postgresql.UUID(as_uuid=True), nullable=False, server_default=sa.text("gen_random_uuid()")),
            sa.Column("account_id", postgresql.UUID(as_uuid=True), nullable=False),
            sa.Column("name", sa.Text(), nullable=False),
            sa.Column("provider", sa.Text(), nullable=False),
            sa.Column("connection_type", sa.Text(), nullable=False),
            sa.Column("external_account_id", sa.Text(), nullable=True),
            sa.Column("secret_ref", sa.Text(), nullable=True),
            sa.Column("metadata", postgresql.JSONB(astext_type=sa.Text()), nullable=False, server_default=sa.text("'{}'::jsonb")),
            sa.Column("created_by", postgresql.UUID(as_uuid=True), nullable=True),
            sa.Column("created_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.text("now()")),
            sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.text("now()")),
            sa.ForeignKeyConstraint(["account_id"], ["public.accounts.id"], name="connections_account_id_fkey", ondelete="CASCADE"),
            sa.ForeignKeyConstraint(["created_by"], ["public.users.id"], name="connections_created_by_fkey", ondelete="SET NULL"),
            sa.PrimaryKeyConstraint("id", name="connections_pkey"),
            schema="public",
        )

    op.create_index("connections_account_id_idx", "connections", ["account_id"], schema="public")
    op.create_index("connections_provider_idx", "connections", ["provider"], schema="public")

    op.execute("ALTER TABLE public.users ADD COLUMN IF NOT EXISTS email VARCHAR(254)")
    op.execute("UPDATE public.users SET email = NULL WHERE email IS NOT NULL AND btrim(email) = ''")
    op.execute("CREATE UNIQUE INDEX IF NOT EXISTS users_email_key ON public.users (email) WHERE email IS NOT NULL")

    op.execute("ALTER TABLE public.users ADD COLUMN IF NOT EXISTS verified BOOLEAN NOT NULL DEFAULT false")
    op.execute("ALTER TABLE public.users ADD COLUMN IF NOT EXISTS verification_token TEXT")
    op.execute("CREATE UNIQUE INDEX IF NOT EXISTS users_verification_token_key ON public.users (verification_token) WHERE verification_token IS NOT NULL")



def downgrade() -> None:
    op.drop_table("users", schema="public")