import { request } from "@/app/components/handlers/auth";

export interface FavLLMResponse {
  status: "success" | "error";
  message: string;
}

export async function addFavLLM(llmId: string): Promise<FavLLMResponse> {
  return request<FavLLMResponse>("/api/user/add_fav", { method: "POST", body: JSON.stringify({ hf_id: llmId }) });
}

export async function removeFavLLM(hfId: string): Promise<FavLLMResponse> {
  return request<FavLLMResponse>("/api/user/remove_fav", { method: "POST", body: JSON.stringify({ hf_id: hfId }) });
}
