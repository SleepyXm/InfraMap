"use client";

import { useMemo, useState } from "react";
import { runEZDeployAnalysis, saveEZDeployDecisions } from "@/app/components/handlers/scans";
import type { WorkspaceLayer } from "../control/WorkspaceSetupGuide";
import { normalizeEZDeployAnalysis } from "@/app/components/handlers/scans";
import { type EZDeployAnalysis, type EZDeployComponent, type EZDeployDecision, type EZDeployDependency } from "@/app/components/types/scans";
import { type WorkspaceDetail, type WorkspaceResource } from "@/app/components/types/workspaces";
import styles from "@/app/UI/WorkspaceControls.module.css";

type AnalysisItem = (EZDeployComponent & { itemType: "component"; }) | (EZDeployDependency & { itemType: "dependency"; });

type Props = {
  accountID: string;
  workspaceID: string;
  role: WorkspaceDetail["access"];
  source: WorkspaceResource;
  resources: WorkspaceResource[];
  onUpdated: (resource: WorkspaceResource) => void;
  onNavigate: (layer: WorkspaceLayer) => void;
};

export function EZDeployAnalysisStage({ accountID, workspaceID, role, source, resources, onUpdated, onNavigate }: Props) {
  const initialAnalysis = normalizeEZDeployAnalysis(source.metadata.ezdeploy_analysis);
  const initialDecisions = decisionsFromMetadata(source);
  const [analysis, setAnalysis] = useState<EZDeployAnalysis | null>(initialAnalysis);
  const [decisions, setDecisions] = useState<Record<string, EZDeployDecision>>(() => initialAnalysis ? decisionMap(initialAnalysis, initialDecisions) : {});
  const [scanning, setScanning] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(initialDecisions.length > 0);
  const [error, setError] = useState("");
  const canChange = role !== "viewer";
  const items = useMemo<AnalysisItem[]>(() => analysis ? [
    ...analysis.components.map((component) => ({ ...component, itemType: "component" as const })),
    ...analysis.dependencies.map((dependency) => ({ ...dependency, itemType: "dependency" as const })),
  ] : [], [analysis]);

  async function scan() {
    setScanning(true);
    setSaved(false);
    setError("");
    try {
      const response = await runEZDeployAnalysis(accountID, workspaceID);
      setAnalysis(response.analysis);
      setDecisions(decisionMap(response.analysis, response.decisions));
      if (response.source) onUpdated(response.source);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "EZDeploy could not scan this repository.");
    } finally {
      setScanning(false);
    }
  }

  async function savePlan() {
    if (!analysis || !canChange) return;
    setSaving(true);
    setError("");
    try {
      const response = await saveEZDeployDecisions(accountID, workspaceID, items.map((item) => decisions[item.id]));
      setDecisions(decisionMap(response.analysis, response.decisions));
      setSaved(true);
      if (response.source) onUpdated(response.source);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "The component plan could not be saved.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <section id="analysis-layer" className={styles.section}>
      <header className={styles.heading}>
        <div><p className={styles.eyebrow}>Architecture discovery</p><h2>Let EZDeploy read the repository.</h2><p>InfraMap sends this exact GitHub revision through EZDeploy&apos;s source scanner, then turns what it finds into deployment decisions. The analysis reads files but never builds or runs workspace code.</p></div>
        <span className={styles.providerMark}>EZ</span>
      </header>

      {!analysis ? (
        <div className={styles.empty}>
          <div><span>Repository ready</span><h3>{source.name}</h3><p>Scan the default branch to identify deployable backends, containers, routes, environment requirements and infrastructure dependencies.</p></div>
          <button type="button" disabled={scanning || !canChange} onClick={() => void scan()}>{scanning ? "Scanning repository…" : "Scan with EZDeploy"}</button>
        </div>
      ) : (
        <div className={styles.analysis}>
          <div className={styles.summary}>
            <div><span>Revision</span><strong>{analysis.revision.slice(0, 8)}</strong></div>
            <div><span>Files scanned</span><strong>{analysis.files_scanned.toLocaleString()}</strong></div>
            <div><span>Components</span><strong>{analysis.components.length}</strong></div>
            <div><span>Dependencies</span><strong>{analysis.dependencies.length}</strong></div>
            <button type="button" disabled={scanning || !canChange} onClick={() => void scan()}>{scanning ? "Scanning…" : "Scan latest revision"}</button>
          </div>

          <div className={styles.findingsHeader}><div><p className={styles.eyebrow}>What EZDeploy found</p><h3>Decide what happens to each part.</h3></div><span>{languageSummary(analysis.languages)}</span></div>
          {items.length === 0 ? <div className={styles.noFindings}><strong>No deployable components were classified.</strong><span>The scan is still saved. Update EZDeploy&apos;s walker when this repository needs another framework or project type.</span></div> : (
            <div className={styles.findings}>
              {items.map((item, index) => (
                <article key={item.id} className={styles.finding}>
                  <div className={styles.findingIdentity}><span>{String(index + 1).padStart(2, "0")}</span><div><small>{item.itemType === "component" ? item.kind : "Required service"}</small><h4>{item.name}</h4><p>{item.itemType === "component" ? componentDetail(item) : dependencyDetail(item)}</p></div></div>
                  <div className={styles.evidence}>{evidenceFor(item).slice(0, 4).map((value) => <span key={value}>{value}</span>)}</div>
                  <label><span>What should InfraMap do?</span><select value={decisionValue(decisions[item.id])} disabled={!canChange} onChange={(event) => { setDecisions((current) => ({ ...current, [item.id]: parseDecision(item.id, event.target.value) })); setSaved(false); }}>{decisionOptions(item).map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}</select></label>
                </article>
              ))}
            </div>
          )}
          <div className={styles.planActions}>
            <div>{analysis.warnings.map((warning) => <span key={warning}>{warning}</span>)}</div>
            {items.length > 0 && <button type="button" disabled={saving || !canChange} onClick={() => void savePlan()}>{saving ? "Saving plan…" : saved ? "Plan saved" : "Save component plan"}</button>}
          </div>
          {saved && <ProviderNextSteps decisions={Object.values(decisions)} resources={resources} onNavigate={onNavigate} />}
        </div>
      )}
      {error && <p className={styles.error}>{error}</p>}
    </section>
  );
}

function decisionMap(analysis: EZDeployAnalysis, saved: EZDeployDecision[]): Record<string, EZDeployDecision> {
  const stored = new Map(saved.map((decision) => [decision.component_id, decision]));
  const items: AnalysisItem[] = [
    ...analysis.components.map((component) => ({ ...component, itemType: "component" } as AnalysisItem)),
    ...analysis.dependencies.map((dependency) => ({ ...dependency, itemType: "dependency" } as AnalysisItem)),
  ];
  return Object.fromEntries(items.map((item) => [item.id, stored.get(item.id) ?? defaultDecision(item)]));
}

function defaultDecision(item: AnalysisItem): EZDeployDecision {
  if (item.itemType === "dependency" && item.kind === "database") return { component_id: item.id, action: "connect", provider: "supabase" };
  if (item.itemType === "dependency") return { component_id: item.id, action: "existing", provider: "external" };
  return { component_id: item.id, action: "deploy", provider: item.kind === "frontend" ? "vercel" : "aws" };
}

function decisionOptions(item: AnalysisItem): Array<{ value: string; label: string; }> {
  if (item.itemType === "dependency" && item.kind === "database") return [
    { value: "connect:supabase", label: "Connect an existing Supabase project" },
    { value: "deploy:supabase", label: "Create a new Supabase project" },
    { value: "existing:external", label: "It already exists elsewhere" },
    { value: "later:", label: "Decide later" },
  ];
  if (item.itemType === "dependency") return [
    { value: "connect:aws", label: "Connect it through AWS" },
    { value: "existing:external", label: "It already exists elsewhere" },
    { value: "later:", label: "Decide later" },
  ];
  return [
    { value: "deploy:aws", label: "Deploy with EZDeploy on AWS" },
    { value: "connect:aws", label: "Connect an existing AWS service" },
    ...(item.kind === "frontend" ? [{ value: "deploy:vercel", label: "Deploy this frontend on Vercel" }] : []),
    { value: "existing:external", label: "It is already hosted elsewhere" },
    { value: "later:", label: "Decide later" },
  ];
}

function decisionValue(decision?: EZDeployDecision): string {
  return decision ? `${decision.action}:${decision.provider}` : "later:";
}

function parseDecision(componentID: string, value: string): EZDeployDecision {
  const [action, provider = ""] = value.split(":") as [EZDeployDecision["action"], EZDeployDecision["provider"]];
  return { component_id: componentID, action, provider };
}

function componentDetail(component: EZDeployComponent): string {
  const location = component.root && component.root !== "." ? component.root : "Repository root";
  return `${location} · ${component.runtime || "Runtime unresolved"} · ${component.confidence || "Unknown"} confidence`;
}

function dependencyDetail(dependency: EZDeployDependency): string {
  return `Inferred from ${dependency.evidence.length} environment requirement${dependency.evidence.length === 1 ? "" : "s"}`;
}

function evidenceFor(item: AnalysisItem): string[] {
  if (item.itemType === "dependency") return item.evidence;
  return [...item.routes.map((route) => `Route ${route}`), ...item.environment.map((name) => `ENV ${name}`), ...item.dockerfiles.map((file) => `Docker ${file}`), ...item.evidence];
}

function languageSummary(languages: Record<string, number>): string {
  const entries = Object.entries(languages).sort(([, left], [, right]) => right - left).slice(0, 4);
  return entries.length ? entries.map(([name]) => name).join(" · ") : "No language signal";
}

function ProviderNextSteps({ decisions, resources, onNavigate }: { decisions: EZDeployDecision[]; resources: WorkspaceResource[]; onNavigate: (layer: WorkspaceLayer) => void; }) {
  const providers = [...new Set(decisions.map((decision) => decision.provider).filter((provider) => provider && provider !== "external"))];
  if (!providers.length) return null;
  const layers: Record<string, WorkspaceLayer> = { github: "source", vercel: "deployment", supabase: "data", aws: "runtime" };
  return <nav className={styles.nextSteps} aria-label="Required project providers"><span>Continue the project plan</span>{providers.map((provider) => <button type="button" key={provider} onClick={() => onNavigate(layers[provider])}>{resources.some((resource) => resource.provider === provider) ? `${provider} attached` : `Configure ${provider}`} <b>→</b></button>)}</nav>;
}

function decisionsFromMetadata(source: WorkspaceResource): EZDeployDecision[] {
  return Array.isArray(source.metadata.ezdeploy_decisions) ? source.metadata.ezdeploy_decisions as EZDeployDecision[] : [];
}
