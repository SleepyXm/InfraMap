import { request } from "./auth";
import type { EZDeployAnalysis, EZDeployAnalysisResponse, EZDeployComponent, EZDeployDecision, EZDeployDependency } from "../types/scans";

export function getEZDeployAnalysis(accountID: string, workspaceID: string): Promise<EZDeployAnalysisResponse> {
  return request<EZDeployAnalysisResponse>(`/api/accounts/${accountID}/scans/${workspaceID}`, { method: "GET" }).then(normalizeAnalysisResponse);
}

export function runEZDeployAnalysis(accountID: string, workspaceID: string): Promise<EZDeployAnalysisResponse> {
  return request<EZDeployAnalysisResponse>(`/api/accounts/${accountID}/scans/${workspaceID}`, { method: "POST" }).then(normalizeAnalysisResponse);
}

export function saveEZDeployDecisions(accountID: string, workspaceID: string, decisions: EZDeployDecision[]): Promise<EZDeployAnalysisResponse> {
  return request<EZDeployAnalysisResponse>(`/api/accounts/${accountID}/scans/${workspaceID}/plan`, { method: "PUT", body: JSON.stringify({ decisions }) }).then(normalizeAnalysisResponse);
}

export function normalizeEZDeployAnalysis(value: unknown): EZDeployAnalysis | null {
  if (!isRecord(value) || typeof value.id !== "string" || !value.id) return null;
  const components = Array.isArray(value.components) ? value.components.filter(isRecord).map((component) => ({
    ...component,
    evidence: stringArray(component.evidence),
    routes: stringArray(component.routes),
    environment: stringArray(component.environment),
    dockerfiles: stringArray(component.dockerfiles),
    suggested_providers: stringArray(component.suggested_providers),
  })) as EZDeployComponent[] : [];
  const dependencies = Array.isArray(value.dependencies) ? value.dependencies.filter(isRecord).map((dependency) => ({
    ...dependency,
    evidence: stringArray(dependency.evidence),
    suggested_providers: stringArray(dependency.suggested_providers),
  })) as EZDeployDependency[] : [];
  return {
    ...(value as EZDeployAnalysis),
    languages: numberRecord(value.languages),
    components,
    dependencies,
    dockerfiles: Array.isArray(value.dockerfiles) ? value.dockerfiles as EZDeployAnalysis["dockerfiles"] : [],
    warnings: stringArray(value.warnings),
  };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function stringArray(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string") : [];
}

function numberRecord(value: unknown): Record<string, number> {
  if (!isRecord(value)) return {};
  return Object.fromEntries(Object.entries(value).filter((entry): entry is [string, number] => typeof entry[1] === "number"));
}


function normalizeAnalysisResponse(response: EZDeployAnalysisResponse): EZDeployAnalysisResponse {
  const analysis = normalizeEZDeployAnalysis(response.analysis);
  if (!analysis) throw new Error("The server returned an invalid scan analysis.");
  return { ...response, analysis, decisions: response.decisions ?? [] };
}
