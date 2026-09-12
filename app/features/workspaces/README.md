# Workspace features

Everything beneath this directory belongs to the workspace control plane. Route files should compose these features; provider logic, workspace state, and operational views should live here rather than in global UI or generic component folders.

- `overview`: workspace summary and project-level entry points.
- `projects`: project composition across source, deployment, data, and infrastructure providers.
- `connections`: provider authorization, connected resources, and workspace attachment.
- `map-view`: infrastructure relationships and topology interaction.
- `deployments`: deployment history, promotion, rollback, and release state.
- `environments`: environment variables, scopes, branches, and provider synchronization.
- `observability`: health, events, logs, metrics, and incidents.

Reusable visual primitives stay in `app/UI`. Backend provider clients remain in `backend/services`.
