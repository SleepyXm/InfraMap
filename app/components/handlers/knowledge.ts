import { request } from "@/app/auth";

export type KnowledgeBase = {
  id: string;
  name: string;
  description: string;
  embedding_model_id: string;
  hf_token_name: string;
  chunk_size_runes: number;
  chunk_overlap_runes: number;
  embedding_dimension: number | null;
};

export type KnowledgeDocument = {
  id: string;
  knowledge_base_id: string;
  filename: string;
  mime_type: string;
  size_bytes: number;
  status: "processing" | "ready" | "failed";
  failure_reason: string | null;
};

export type KnowledgeResult = {
  citation_id: string;
  document_id: string;
  filename: string;
  page: number | null;
  chunk_index: number;
  content: string;
  score: number;
};

export async function listKnowledgeBases(): Promise<KnowledgeBase[]> {
  const response = await request<{ data: KnowledgeBase[] }>("/api/knowledge-bases", { method: "GET" });
  return response.data;
}

export async function createKnowledgeBase(input: {
  name: string;
  description: string;
  embedding_model_id: string;
  hf_token_name: string;
  chunk_size_runes: number;
  chunk_overlap_runes: number;
}): Promise<KnowledgeBase> {
  const response = await request<{ data: KnowledgeBase }>("/api/knowledge-bases", {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.data;
}

export async function listKnowledgeDocuments(knowledgeBaseID: string): Promise<KnowledgeDocument[]> {
  const response = await request<{ data: KnowledgeDocument[] }>(
    `/api/knowledge-bases/${knowledgeBaseID}/documents`,
    { method: "GET" },
  );
  return response.data;
}

export async function uploadKnowledgeDocument(knowledgeBaseID: string, file: File): Promise<KnowledgeDocument> {
  const form = new FormData();
  form.append("file", file);
  const response = await request<{ data: KnowledgeDocument }>(
    `/api/knowledge-bases/${knowledgeBaseID}/documents`,
    { method: "POST", body: form },
  );
  return response.data;
}

export async function searchKnowledge(knowledgeBaseID: string, query: string): Promise<KnowledgeResult[]> {
  const response = await request<{ data: { results: KnowledgeResult[] } }>(
    `/api/knowledge-bases/${knowledgeBaseID}/search`,
    { method: "POST", body: JSON.stringify({ query, limit: 6 }) },
  );
  return response.data.results;
}
