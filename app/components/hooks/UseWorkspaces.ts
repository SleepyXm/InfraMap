"use client";

import { useCallback, useEffect, useState } from "react";
import { getWorkspaceConnections } from "@/app/components/handlers/workspaces";
import type { Connection } from "@/app/components/types/users";

export function useWorkspaceConnections(accountID?: string) {
  const [connections, setConnections] = useState<Connection[]>([]);
  const [loading, setLoading] = useState(Boolean(accountID));
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    if (!accountID) return;

    setLoading(true);
    setError(null);

    try {
      setConnections(await getWorkspaceConnections(accountID));
    } catch (cause) {
      setConnections([]);
      setError(cause instanceof Error ? cause.message : "Connections could not be loaded.");
    } finally {
      setLoading(false);
    }
  }, [accountID]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { connections, loading, error, refresh };
}
