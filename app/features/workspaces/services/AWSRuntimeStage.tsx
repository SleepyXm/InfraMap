"use client";

import { FormEvent, useEffect, useState } from "react";
import { attachAWSService, getAWSCloudFormation } from "@/app/components/handlers/services";
import { createEZDeployDeployment, getEZDeployDeployment } from "@/app/components/handlers/deployments";
import type { WorkspaceLayer } from "../control/WorkspaceSetupGuide";
import { normalizeEZDeployAnalysis } from "@/app/components/handlers/scans";
import { type AWSCloudFormationBundle } from "@/app/components/types/services";
import { type EZDeployCommand } from "@/app/components/types/deployments";
import { type EZDeployDecision } from "@/app/components/types/scans";
import { type WorkspaceDetail, type WorkspaceResource } from "@/app/components/types/workspaces";
import styles from "@/app/UI/WorkspaceControls.module.css";

type AWSRuntimeStageProps = {
  accountID: string;
  workspaceID: string;
  role: WorkspaceDetail["access"];
  resource?: WorkspaceResource;
  source?: WorkspaceResource;
  onAttached: (resource: WorkspaceResource) => void;
  onNavigate: (layer: WorkspaceLayer) => void;
};

const terminalStatuses = new Set(["Success", "Cancelled", "Failed", "TimedOut", "Cancelling"]);

export function AWSRuntimeStage({ accountID, workspaceID, role, resource, source, onAttached, onNavigate }: AWSRuntimeStageProps) {
  const [bundle, setBundle] = useState<AWSCloudFormationBundle | null>(null);
  const [preparing, setPreparing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [outputs, setOutputs] = useState({ name: "EZDeploy host", role_arn: "", region: "eu-west-2", instance_id: "", document_name: "", stack_id: "", public_ip: "" });

  async function prepareStack() {
    setPreparing(true);
    setError("");
    try {
      const prepared = await getAWSCloudFormation(accountID, workspaceID);
      setBundle(prepared);
      downloadTemplate(prepared);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "The CloudFormation stack could not be prepared.");
    } finally {
      setPreparing(false);
    }
  }

  async function attach(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!bundle) return;
    setSaving(true);
    setError("");
    try {
      const attached = await attachAWSService(accountID, workspaceID, { ...outputs, external_id: bundle.external_id });
      onAttached(attached);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "The AWS host could not be verified.");
    } finally {
      setSaving(false);
    }
  }

  if (resource) return <ConnectedAWSRuntime accountID={accountID} workspaceID={workspaceID} resource={resource} source={source} role={role} onNavigate={onNavigate} />;

  return (
    <section id="runtime-layer" className={styles.section}>
      <header className={styles.heading}>
        <div><p className={styles.eyebrow}>Runtime service</p><h2>Create a guarded AWS host.</h2><p>CloudFormation creates one EC2 instance and one role limited to a workspace-specific Systems Manager command. InfraMap never receives permanent AWS keys.</p></div>
        <span className={styles.providerMark}>AWS</span>
      </header>
      <div className={styles.securityStrip}><span>No SSH ingress</span><span>Fixed InfraMap source IP</span><span>One tagged EC2 host</span><span>One validated EZDeploy document</span></div>
      <div className={styles.setupGrid}>
        <article className={styles.stackPanel}>
          <span className={styles.step}>Step 01</span><h3>Launch the stack in your AWS account</h3><p>Download the generated template, upload it in CloudFormation, choose a VPC and public subnet, then create the stack.</p>
          <button type="button" disabled={preparing || (role !== "owner" && role !== "admin")} onClick={() => void prepareStack()}>{preparing ? "Preparing…" : bundle ? "Download template again" : "Prepare CloudFormation template"}</button>
          {bundle && <div className={styles.stackReady}><strong>{bundle.stack_name}</strong><span>The external ID is already embedded in this workspace-specific template.</span></div>}
        </article>
        <form className={styles.outputsPanel} onSubmit={attach}>
          <div><span className={styles.step}>Step 02</span><h3>Attach the completed stack</h3><p>Copy these values from the CloudFormation Outputs tab. InfraMap verifies the role, SSM document, and online instance before saving anything.</p></div>
          <label><span>Connection name</span><input value={outputs.name} onChange={(event) => setOutputs({ ...outputs, name: event.target.value })} required /></label>
          <label className={styles.wide}><span>Control role ARN</span><input value={outputs.role_arn} onChange={(event) => setOutputs({ ...outputs, role_arn: event.target.value })} placeholder="arn:aws:iam::123456789012:role/..." required /></label>
          <label><span>AWS region</span><input value={outputs.region} onChange={(event) => setOutputs({ ...outputs, region: event.target.value })} placeholder="eu-west-2" required /></label>
          <label><span>EC2 instance ID</span><input value={outputs.instance_id} onChange={(event) => setOutputs({ ...outputs, instance_id: event.target.value })} placeholder="i-0123456789abcdef0" required /></label>
          <label className={styles.wide}><span>SSM document name</span><input value={outputs.document_name} onChange={(event) => setOutputs({ ...outputs, document_name: event.target.value })} required /></label>
          <label className={styles.wide}><span>Stack ID <small>Optional</small></span><input value={outputs.stack_id} onChange={(event) => setOutputs({ ...outputs, stack_id: event.target.value })} /></label>
          <label><span>Public IP <small>Optional</small></span><input value={outputs.public_ip} onChange={(event) => setOutputs({ ...outputs, public_ip: event.target.value })} /></label>
          <button className={styles.primaryAction} type="submit" disabled={saving || !bundle || (role !== "owner" && role !== "admin")}>{saving ? "Verifying AWS…" : "Verify and attach host"}</button>
        </form>
      </div>
      {error && <p className={styles.error}>{error}</p>}
    </section>
  );
}

function ConnectedAWSRuntime({ accountID, workspaceID, resource, source, role, onNavigate }: { accountID: string; workspaceID: string; resource: WorkspaceResource; source?: WorkspaceResource; role: WorkspaceDetail["access"]; onNavigate: (layer: WorkspaceLayer) => void; }) {
  const [domain, setDomain] = useState(metadata(resource, "domain"));
  const [email, setEmail] = useState("");
  const [runtime, setRuntime] = useState<"native" | "docker">(metadata(resource, "runtime") === "docker" ? "docker" : "native");
  const [command, setCommand] = useState<EZDeployCommand | null>(null);
  const [working, setWorking] = useState(false);
  const [error, setError] = useState("");
  const privateRepository = source?.metadata.private === true;
  const analysis = normalizeEZDeployAnalysis(source?.metadata.ezdeploy_analysis);
  const decisions = sourceDecisions(source);
  const plannedComponents = analysis?.components.filter((component) => decisions.some((decision) => decision.component_id === component.id && decision.action === "deploy" && decision.provider === "aws")) ?? [];
  const requiredEnvironment = [...new Set(plannedComponents.flatMap((component) => component.environment))];
  const planReady = plannedComponents.length > 0 && requiredEnvironment.length === 0;

  useEffect(() => {
    if (!command || terminalStatuses.has(command.status)) return;
    const timer = window.setTimeout(() => {
      void getEZDeployDeployment(accountID, workspaceID, command.command_id)
        .then((next) => { setCommand(next); setError(""); })
        .catch((cause) => setError(cause instanceof Error ? cause.message : "Deployment status could not be refreshed."));
    }, 2500);
    return () => window.clearTimeout(timer);
  }, [accountID, command, workspaceID]);

  async function deploy(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setWorking(true);
    setError("");
    try {
      setCommand(await createEZDeployDeployment(accountID, workspaceID, { domain, email, runtime }));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Deployment could not be started.");
    } finally {
      setWorking(false);
    }
  }

  return (
    <section id="runtime-layer" className={styles.section}>
      <header className={styles.heading}><div><p className={styles.eyebrow}>AWS runtime</p><h2>{resource.name}</h2><p>This host is controlled through the workspace-specific Systems Manager document created by CloudFormation.</p></div><span className={styles.providerMark}>AWS</span></header>
      <div className={styles.runtimeFacts}>
        <div><span>Instance</span><strong>{resource.external_id}</strong></div><div><span>Region</span><strong>{metadata(resource, "region")}</strong></div><div><span>Public IP</span><strong>{metadata(resource, "public_ip") || "Not reported"}</strong></div><div><span>Control</span><strong>SSM only</strong></div>
      </div>
      <form className={styles.deployPanel} onSubmit={deploy}>
        <div><span className={styles.step}>GitHub → EZDeploy → EC2</span><h3>Deploy the approved components</h3><p>{plannedComponents.length ? `${plannedComponents.map((component) => component.name).join(", ")} will be deployed from ${source?.name}.` : "Scan the repository and assign at least one component to this AWS runtime first."}</p></div>
        <label><span>Domain</span><input value={domain} onChange={(event) => setDomain(event.target.value)} placeholder="api.example.com" required /></label>
        <label><span>Certificate email</span><input type="email" value={email} onChange={(event) => setEmail(event.target.value)} placeholder="ops@example.com" required /></label>
        <label><span>Runtime</span><select value={runtime} onChange={(event) => setRuntime(event.target.value as "native" | "docker")}><option value="native">Native + systemd</option><option value="docker">Docker</option></select></label>
        <button className={styles.primaryAction} type="submit" disabled={working || !source || !planReady || privateRepository || (role !== "owner" && role !== "admin")}>{working ? "Starting…" : "Deploy approved plan"}</button>
      </form>
      {privateRepository && <p className={styles.notice}>Private repository deployment is deliberately disabled until InfraMap can issue a short-lived, repository-scoped GitHub App token.</p>}
      {!analysis && <p className={styles.notice}><button type="button" onClick={() => onNavigate("analysis")}>Run the EZDeploy architecture scan</button> before deploying this repository.</p>}
      {analysis && plannedComponents.length === 0 && <p className={styles.notice}><button type="button" onClick={() => onNavigate("analysis")}>Update the component plan</button> and assign a deployable component to AWS.</p>}
      {requiredEnvironment.length > 0 && <p className={styles.notice}>Environment configuration is still required for {requiredEnvironment.join(", ")}. These values must be resolved before InfraMap sends the deployment.</p>}
      {command && <div className={styles.command}><div><span>Command {command.command_id}</span><strong>{command.status_detail || command.status}</strong></div>{command.standard_output && <pre>{command.standard_output}</pre>}{command.standard_error && <pre className={styles.commandError}>{command.standard_error}</pre>}</div>}
      {error && <p className={styles.error}>{error}</p>}
    </section>
  );
}

function downloadTemplate(bundle: AWSCloudFormationBundle) {
  const url = URL.createObjectURL(new Blob([bundle.template], { type: "application/x-yaml" }));
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = bundle.filename;
  anchor.click();
  window.setTimeout(() => URL.revokeObjectURL(url), 0);
}

function metadata(resource: WorkspaceResource, key: string): string {
  return typeof resource.metadata[key] === "string" ? resource.metadata[key] as string : "";
}

function sourceDecisions(source?: WorkspaceResource): EZDeployDecision[] {
  const value = source?.metadata.ezdeploy_decisions;
  return Array.isArray(value) ? value as EZDeployDecision[] : [];
}
