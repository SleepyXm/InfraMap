import { addFavLLM } from "@/app/components/handlers/fav";
import { Message, createTemporaryConversation } from "@/app/components/handlers/chat";
import { fetchConversationMessages } from "@/app/conversation";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE;

export type ModelSettings = {
  temperature: number;
  top_p: number;
  max_tokens: number;
  presence_penalty: number;
  frequency_penalty: number;
};

export const DEFAULT_MODEL_SETTINGS: ModelSettings = {
  temperature: 0.7,
  top_p: 0.95,
  max_tokens: 1024,
  presence_penalty: 0,
  frequency_penalty: 0,
};

export const handleFavClick = async (
  modelId: string | undefined,
  setIsFav: React.Dispatch<React.SetStateAction<boolean>>
) => {
  if (!modelId) {
    console.error("Missing modelId");
    return;
  }

  try {
    const data = await addFavLLM(modelId);
    console.log(data.message);
    setIsFav(true);
  } catch (err: unknown) {
    const error = err instanceof Error ? err : new Error("Unknown error");
    console.error("Failed to favorite LLM:", error.message);
  }
};

async function ensureConversation(
  currentConversationId: string | null,
  setCurrentConversationId: React.Dispatch<React.SetStateAction<string | null>>,
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>,
  modelId: string
) {
  if (currentConversationId) return currentConversationId;

  const conv = await createTemporaryConversation(modelId);
  setCurrentConversationId(conv.id);

  const messages = await fetchConversationMessages(conv.id);
  setMessages(messages);

  return conv.id;
}

function appendUserMessage(
  input: string,
  setInput: React.Dispatch<React.SetStateAction<string>>,
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>,
) {
  const userMessage: Message = { role: "user", content: input };
  setMessages(prev => [...prev, userMessage]);
  setInput("");
  return userMessage;
}

async function streamAssistantResponse(
  conversationId: string,
  newMessages: Message[],
  modelId: string,
  hfTokenName: string,
  settings: ModelSettings,
  customisationId: string,
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>,
) {
  const response = await fetch(`${API_BASE}/api/llm/chat/stream?conversation_id=${conversationId}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    // The backend owns conversation memory. Only submit messages created by
    // this interaction so stored history is not duplicated on every request.
    body: JSON.stringify({ modelId, hfTokenName, conversation: newMessages, settings, customisationId }),
  });

  if (!response.ok) {
    const data = await response.json().catch(() => null);
    throw new Error(data?.error || `Request failed with status ${response.status}`);
  }

  if (!response.body) throw new Error("No response body");

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let done = false;
  let partial = "";

  while (!done) {
    const { value, done: readerDone } = await reader.read();
    done = readerDone;
    if (value) {
      partial += decoder.decode(value, { stream: true });

      // catch mid-stream JSON errors before rendering them
      try {
        const parsed = JSON.parse(partial);
        if (parsed?.error) throw new Error(parsed.error);
      } catch (e) {
        if (e instanceof SyntaxError) {
          // not JSON yet, keep streaming
        } else {
          throw e;
        }
      }

      setMessages(prev => {
        const newMessages = [...prev];
        const lastMsg = newMessages[newMessages.length - 1];
        if (lastMsg?.role === "assistant") {
          lastMsg.content = partial;
        } else {
          newMessages.push({ role: "assistant", content: partial });
        }
        return newMessages;
      });
    }
  }
}

export const sendMessage = async (args: {
  input: string;
  setInput: React.Dispatch<React.SetStateAction<string>>;
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>;
  currentConversationId: string | null;
  setCurrentConversationId: React.Dispatch<React.SetStateAction<string | null>>;
  modelId: string;
  hfTokenName: string;
  settings: ModelSettings;
  customisationId: string;
}) => {
  const { input, setInput, setMessages, currentConversationId, setCurrentConversationId, modelId, hfTokenName, settings, customisationId } = args;
  if (!input.trim()) return;

  const conversationId = await ensureConversation(currentConversationId, setCurrentConversationId, setMessages, modelId);

  const userMessage = appendUserMessage(input, setInput, setMessages);


  await streamAssistantResponse(conversationId, [userMessage], modelId, hfTokenName, settings, customisationId, setMessages);
};
