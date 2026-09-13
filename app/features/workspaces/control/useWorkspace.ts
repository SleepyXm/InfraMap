"use client";

import { useCallback, useEffect, useState } from "react";
import { getWorkspace } from "@/app/components/handlers/workspaces";
import { normalizeEZDeployAnalysis } from "@/app/components/handlers/scans";
import { type WorkspaceDetail, type WorkspaceResource } from "@/app/components/types/workspaces";

export function useWorkspace(accountID: string, workspaceID: string) {
  const [detail, setDetail] = useState<WorkspaceDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setDetail(await getWorkspace(accountID, workspaceID));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Workspace could not be loaded.");
    } finally {
      setLoading(false);
    }
  }, [accountID, workspaceID]);

  useEffect(() => {
    let active = true;
    void Promise.resolve().then(async () => {
      if (!active) return;
      setLoading(true);
      setError(null);
      try {
        const loaded = await getWorkspace(accountID, workspaceID);
        if (active) setDetail(loaded);
      } catch (cause) {
        if (active) setError(cause instanceof Error ? cause.message : "Workspace could not be loaded.");
      } finally {
        if (active) setLoading(false);
      }
    });
    return () => { active = false; };
  }, [accountID, workspaceID]);

  const setResource = useCallback((resource: WorkspaceResource) => {
    setDetail((current) => current ? {
      ...current,
      workspace: { ...current.workspace, status: "active" },
      sources: resource.provider === "github" ? [...(current.sources ?? []).filter((item) => item.id !== resource.id), resource] : current.sources ?? [],
      components: resource.provider === "github" ? [
        ...current.components.filter((component) => component.source_id !== resource.id),
        ...(normalizeEZDeployAnalysis(resource.metadata.ezdeploy_analysis)?.components ?? []).map((component) => ({ ...component, source_id: resource.id })),
      ] : current.components,
      services: resource.provider === "github" ? current.services : [...current.services.filter((item) => !(item.provider === resource.provider && item.resource_type === resource.resource_type)), resource],
    } : current);
  }, []);

  return { detail, loading, error, refresh, setResource };
}
