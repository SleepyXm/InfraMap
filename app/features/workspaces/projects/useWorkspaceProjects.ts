"use client";

import { useCallback, useEffect, useState } from "react";
import { createWorkspaceProject, getWorkspaceProjects } from "./api";
import type { CreateWorkspaceProject, WorkspaceProject } from "./types";

export function useWorkspaceProjects(accountID?: string) {
  const [projects, setProjects] = useState<WorkspaceProject[]>([]);
  const [loading, setLoading] = useState(Boolean(accountID));
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    if (!accountID) return;
    setLoading(true);
    setError(null);
    try {
      setProjects(await getWorkspaceProjects(accountID));
    } catch (cause) {
      setProjects([]);
      setError(cause instanceof Error ? cause.message : "Projects could not be loaded.");
    } finally {
      setLoading(false);
    }
  }, [accountID]);

  useEffect(() => {
    if (!accountID) return;
    let active = true;
    void Promise.resolve().then(async () => {
      if (!active) return;
      setLoading(true);
      setError(null);
      try {
        const available = await getWorkspaceProjects(accountID);
        if (active) setProjects(available);
      } catch (cause) {
        if (!active) return;
        setProjects([]);
        setError(cause instanceof Error ? cause.message : "Projects could not be loaded.");
      } finally {
        if (active) setLoading(false);
      }
    });
    return () => { active = false; };
  }, [accountID]);

  const createProject = useCallback(async (input: CreateWorkspaceProject) => {
    if (!accountID) throw new Error("Choose a workspace first.");
    const project = await createWorkspaceProject(accountID, input);
    setProjects((current) => [project, ...current]);
    return project;
  }, [accountID]);

  return { projects, loading, error, refresh, createProject };
}
