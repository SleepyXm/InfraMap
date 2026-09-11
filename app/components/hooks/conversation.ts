import { ConversationItem, Message } from "@/app/components/handlers/chat";
import { useState, useEffect } from "react";
import { request } from "@/app/components/handlers/auth";

export const useConversations = () => {
  const [conversations, setConversations] = useState<ConversationItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [reload, setReload] = useState(0);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setLoadError("");
    request<{ conversations: ConversationItem[] }>("/api/conversation/list", { method: "GET" })
      .then((data) => {
        if (!Array.isArray(data.conversations)) throw new Error("The server returned an invalid conversation list.");
        if (active) setConversations(data.conversations);
      })
      .catch((error) => {
        if (active) setLoadError(error instanceof Error ? error.message : "Could not load conversations");
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [reload]);

  useEffect(() => {
    return onConversationCreated((conversation) => {
      setConversations((current) => [
        conversation,
        ...current.filter((item) => item.id !== conversation.id),
      ]);
    });
  }, []);

  const renameConversation = (id: string, title: string) => {
    setConversations(prev => prev.map(c => c.id === id ? { ...c, title } : c));
  };

  const removeConversation = (id: string) => {
    setConversations(prev => prev.filter(c => c.id !== id));
  };

  return {
    conversations,
    loading,
    loadError,
    retry: () => setReload((value) => value + 1),
    renameConversation,
    removeConversation,
  };
};

export const updateConversationTitle = async (conversationId: string, title: string) => {
  return request(`/api/conversation/${conversationId}`, {
    method: "PATCH",
    body: JSON.stringify({ title }),
  });
};

export const deleteConversation = async (conversationId: string) => {
  return request(`/api/conversation/${conversationId}`, {
    method: "DELETE",
  });
};

export const fetchConversationMessages = async (conversationId: string) => {
  const data = await request<{ messages: Message[] }>(
    `/api/conversation/${conversationId}/chunk`,
    { method: "GET" },
  );
  if (!Array.isArray(data.messages)) throw new Error("The conversation exists, but its messages did not load.");
  return data.messages;
};

const conversationBus = new EventTarget();

export function emitConversationSelected(id: string) {
  conversationBus.dispatchEvent(new CustomEvent("conversationSelected", { detail: id }));
}

export function onConversationSelected(callback: (id: string) => void) {
  const handler = (e: Event) => {
    const event = e as CustomEvent<string>;
    callback(event.detail);
  };
  conversationBus.addEventListener("conversationSelected", handler);
  return () => conversationBus.removeEventListener("conversationSelected", handler);
}

export function emitNewConversationRequested() {
  conversationBus.dispatchEvent(new Event("newConversationRequested"));
}

export function onNewConversationRequested(callback: () => void) {
  conversationBus.addEventListener("newConversationRequested", callback);
  return () => conversationBus.removeEventListener("newConversationRequested", callback);
}

export function emitConversationCreated(conversation: ConversationItem) {
  conversationBus.dispatchEvent(new CustomEvent("conversationCreated", { detail: conversation }));
}

export function onConversationCreated(callback: (conversation: ConversationItem) => void) {
  const handler = (event: Event) => {
    callback((event as CustomEvent<ConversationItem>).detail);
  };
  conversationBus.addEventListener("conversationCreated", handler);
  return () => conversationBus.removeEventListener("conversationCreated", handler);
}
