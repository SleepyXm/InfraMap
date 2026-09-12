"use client";

import { useCallback, useEffect, useState } from "react";
import { getWorkspaceProject } from "./api";
import type { ProjectResource, WorkspaceProjectDetail } from "./types";

export function useWorkspaceProject(accountID: string, projectID: string) {
  const [detail, setDetail] = useState<WorkspaceProjectDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setDetail(await getWorkspaceProject(accountID, projectID));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Project could not be loaded.");
    } finally {
      setLoading(false);
    }
  }, [accountID, projectID]);

  useEffect(() => {
    let active = true;
    void Promise.resolve().then(async () => {
      if (!active) return;
      setLoading(true);
      setError(null);
      try {
        const loaded = await getWorkspaceProject(accountID, projectID);
        if (active) setDetail(loaded);
      } catch (cause) {
        if (active) setError(cause instanceof Error ? cause.message : "Project could not be loaded.");
      } finally {
        if (active) setLoading(false);
      }
    });
    return () => { active = false; };
  }, [accountID, projectID]);

  const setResource = useCallback((resource: ProjectResource) => {
    setDetail((current) => current ? {
      ...current,
      project: { ...current.project, status: "active" },
      resources: [...current.resources.filter((item) => !(item.provider === resource.provider && item.resource_type === resource.resource_type)), resource],
    } : current);
  }, []);

  return { detail, loading, error, refresh, setResource };
}
