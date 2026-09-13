"use client";

import type { WorkspaceResource } from "@/app/components/types/workspaces";
import styles from "@/app/UI/WorkspaceControls.module.css";

export type WorkspaceLayer = "source" | "analysis" | "deployment" | "data" | "runtime";

type Layer = {
  id: WorkspaceLayer;
  label: string;
  provider?: WorkspaceResource["provider"];
};

const layers: Layer[] = [
  { id: "source", label: "GitHub source", provider: "github" },
  { id: "analysis", label: "EZDeploy scan" },
  { id: "deployment", label: "Deployment", provider: "vercel" },
  { id: "data", label: "Data", provider: "supabase" },
  { id: "runtime", label: "Runtime", provider: "aws" },
];

export function WorkspaceSetupGuide({ workspaceName, resources, activeLayer, onSelect }: { workspaceName: string; resources: WorkspaceResource[]; activeLayer: WorkspaceLayer; onSelect: (layer: WorkspaceLayer) => void; }) {
  const source = resources.find((resource) => resource.provider === "github");
  const scanned = Boolean(source?.metadata.ezdeploy_analysis);
  const attached = new Set(resources.map((resource) => resource.provider));
  const next = !source ? "Connect the repository first." : !scanned ? "Scan the repository before choosing infrastructure." : "Map each detected component to its service.";

  return (
    <section className={styles.resourceNavigator} aria-label="Workspace resource workflow">
      <header><div><p className={styles.eyebrow}>Workspace resources</p><h2>{workspaceName}</h2></div><p>{next}</p></header>
      <ol>
        {layers.map((layer, index) => {
          const complete = layer.id === "analysis" ? scanned : layer.provider ? attached.has(layer.provider) : false;
          const blocked = layer.id === "analysis" && !source;
          return (
            <li key={layer.id}>
              <button type="button" className={activeLayer === layer.id ? styles.layerActive : undefined} aria-current={activeLayer === layer.id ? "step" : undefined} disabled={blocked} onClick={() => onSelect(layer.id)}>
                <span>{String(index + 1).padStart(2, "0")}</span><strong>{layer.label}</strong><em>{complete ? "Connected" : blocked ? "Needs source" : layer.id === "source" ? "Required" : "Available"}</em>
              </button>
            </li>
          );
        })}
      </ol>
    </section>
  );
}
