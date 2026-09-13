"use client";

import { useCallback, useEffect, useState } from "react";
import { getWorkspaceMetrics } from "@/app/components/handlers/metrics";
import type { WorkspaceMetrics } from "@/app/components/types/metrics";

export function useWorkspaceMetrics(accountID: string, workspaceID: string) {
  const [metrics, setMetrics] = useState<WorkspaceMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setMetrics(await getWorkspaceMetrics(accountID, workspaceID));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Workspace metrics could not be loaded.");
    } finally {
      setLoading(false);
    }
  }, [accountID, workspaceID]);

  useEffect(() => {
    let active = true;
    void getWorkspaceMetrics(accountID, workspaceID).then((next) => {
      if (active) setMetrics(next);
    }).catch((cause) => {
      if (active) setError(cause instanceof Error ? cause.message : "Workspace metrics could not be loaded.");
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [accountID, workspaceID]);

  return { metrics, loading, error, refresh };
}
