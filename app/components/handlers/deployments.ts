import { request } from "./auth";
import type { CreateEZDeployDeployment, EZDeployCommand } from "../types/deployments";

export function createEZDeployDeployment(accountID: string, workspaceID: string, input: CreateEZDeployDeployment): Promise<EZDeployCommand> {
  return request<EZDeployCommand>(`/api/accounts/${accountID}/deployments/${workspaceID}`, { method: "POST", body: JSON.stringify(input) });
}

export function getEZDeployDeployment(accountID: string, workspaceID: string, commandID: string): Promise<EZDeployCommand> {
  return request<EZDeployCommand>(`/api/accounts/${accountID}/deployments/${workspaceID}/${commandID}`, { method: "GET" });
}
