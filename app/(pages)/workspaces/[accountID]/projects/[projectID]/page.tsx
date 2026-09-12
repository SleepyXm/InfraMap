import { ProjectOperations } from "@/app/features/workspaces/projects";

export default async function ProjectPage({ params }: PageProps<"/workspaces/[accountID]/projects/[projectID]">) {
  const { accountID, projectID } = await params;
  return <ProjectOperations accountID={accountID} projectID={projectID} />;
}
