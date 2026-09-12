"use client";

import { useCallback, useEffect, useState } from "react";
import { getProjectControl } from "./api";
import type { ProjectControl } from "./types";

export function useProjectControl(accountID: string, projectID: string) {
  const [control, setControl] = useState<ProjectControl | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setControl(await getProjectControl(accountID, projectID));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Project telemetry could not be loaded.");
    } finally {
      setLoading(false);
    }
  }, [accountID, projectID]);

  useEffect(() => {
    let active = true;
    void getProjectControl(accountID, projectID).then((next) => {
      if (active) setControl(next);
    }).catch((cause) => {
      if (active) setError(cause instanceof Error ? cause.message : "Project telemetry could not be loaded.");
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [accountID, projectID]);

  return { control, loading, error, refresh };
}
