"use client";

import { useCallback, useEffect, useState } from "react";
import { createWorkspace, deleteWorkspace, getWorkspaces } from "@/app/components/handlers/workspaces";
import type { CreateWorkspace, Workspace } from "@/app/components/types/workspaces";

export function useWorkspaces(accountID?: string) {
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [loading, setLoading] = useState(Boolean(accountID));
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    if (!accountID) return;
    setLoading(true);
    setError(null);
    try {
      setWorkspaces(await getWorkspaces(accountID));
    } catch (cause) {
      setWorkspaces([]);
      setError(cause instanceof Error ? cause.message : "Workspaces could not be loaded.");
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
        const available = await getWorkspaces(accountID);
        if (active) setWorkspaces(available);
      } catch (cause) {
        if (!active) return;
        setWorkspaces([]);
        setError(cause instanceof Error ? cause.message : "Workspaces could not be loaded.");
      } finally {
        if (active) setLoading(false);
      }
    });
    return () => { active = false; };
  }, [accountID]);

  const create = useCallback(async (input: CreateWorkspace) => {
    if (!accountID) throw new Error("Choose a workspace first.");
    const workspace = await createWorkspace(accountID, input);
    setWorkspaces((current) => [workspace, ...current]);
    return workspace;
  }, [accountID]);

  const remove = useCallback(async (workspaceID: string) => {
    if (!accountID) throw new Error("Choose an account first.");
    await deleteWorkspace(accountID, workspaceID);
    setWorkspaces((current) => current.filter((workspace) => workspace.id !== workspaceID));
  }, [accountID]);

  return { workspaces, loading, error, refresh, createWorkspace: create, deleteWorkspace: remove };
}
