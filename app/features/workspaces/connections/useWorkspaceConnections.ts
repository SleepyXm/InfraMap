"use client";

import { useCallback, useEffect, useState } from "react";
import { getWorkspaceConnections } from "@/app/components/handlers/connections";
import type { Connection } from "@/app/components/types/connections";

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
    if (!accountID) return;
    let active = true;
    void Promise.resolve().then(async () => {
      if (!active) return;
      setLoading(true);
      setError(null);
      try {
        const available = await getWorkspaceConnections(accountID);
        if (active) setConnections(available);
      } catch (cause) {
        if (!active) return;
        setConnections([]);
        setError(cause instanceof Error ? cause.message : "Connections could not be loaded.");
      } finally {
        if (active) setLoading(false);
      }
    });
    return () => {
      active = false;
    };
  }, [accountID]);

  return { connections, loading, error, refresh };
}
