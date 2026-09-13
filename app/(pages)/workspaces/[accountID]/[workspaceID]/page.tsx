import { WorkspaceControl } from "@/app/features/workspaces/control";

export default async function WorkspacePage({ params, searchParams }: PageProps<"/workspaces/[accountID]/[workspaceID]">) {
  const { accountID, workspaceID } = await params;
  const setup = (await searchParams).setup === "1";
  return <WorkspaceControl accountID={accountID} workspaceID={workspaceID} initialSetup={setup} />;
}
