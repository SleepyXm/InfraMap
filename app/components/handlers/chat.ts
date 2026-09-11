import { request } from "@/app/components/handlers/auth";

export interface Message {
  role: "user" | "assistant";
  content: string;       // required for internal messages
  id?: string;           // optional for API messages
  message?: {            // optional for API messages
    role: "user" | "assistant";
    content: string;
  };
  created_at?: string;   // optional for API messages
}

export interface ChatContextType {
  currentConversationId: string | null;
  setCurrentConversationId: (id: string | null) => void;
  messages: Message[];
  setMessages: (msgs: Message[]) => void;
}

export interface ConversationItem {
  id: string;
  title: string;
  llm_model: string;
}

export interface ConversationMessages {
  messages: Message[],
  user_id: string,
  llm_model: string,
}

export interface LLMCustomisation {
  id: string;
  name: string;
  system_prompt: string;
  builtin: boolean;
}

export function createTemporaryConversation(modelId: string): Promise<ConversationItem> {
  return request<ConversationItem>("/api/conversation/create", {
    method: "POST",
    body: JSON.stringify({ title: "New chat", llm_model: modelId }),
  });
}

export async function getLLMCustomisations(): Promise<LLMCustomisation[]> {
  const response = await request<{ customisations: LLMCustomisation[] }>(
    "/api/llm/customisations",
    { method: "GET" },
  );
  return response.customisations;
}

export function createLLMCustomisation(name: string, systemPrompt: string): Promise<LLMCustomisation> {
  return request<LLMCustomisation>(
    "/api/llm/customisations",
    { method: "POST", body: JSON.stringify({ name, system_prompt: systemPrompt }) },
  );
}

export function updateLLMCustomisation(id: string, name: string, systemPrompt: string): Promise<LLMCustomisation> {
  return request<LLMCustomisation>(
    `/api/llm/customisations/${id}`,
    { method: "PATCH", body: JSON.stringify({ name, system_prompt: systemPrompt }) },
  );
}

export function deleteLLMCustomisation(id: string): Promise<void> {
  return request<void>(`/api/llm/customisations/${id}`, { method: "DELETE" });
}
