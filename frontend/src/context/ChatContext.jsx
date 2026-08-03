import { createContext, useContext, useEffect, useState, useCallback } from "react";
import { useAuth } from "./AuthContext";
import { useGrpcStream } from "../hooks/useGrpcStream";
import * as chatApi from "../api/chat";
import * as grpc from "../lib/grpc";

const ChatContext = createContext(null);

export function ChatProvider({ children }) {
  const { token, user } = useAuth();

  const [users, setUsers] = useState([]);
  const [chats, setChats] = useState([]);
  const [currentChat, setCurrentChat] = useState(null);
  const [messages, setMessages] = useState([]);
  const [loadingChats, setLoadingChats] = useState(false);
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [error, setError] = useState(null);

  const userById = useCallback(
    (id) => users.find((u) => u.id === id),
    [users]
  );

  const chatLabel = useCallback(
    (chat) => {
      const otherIds = (chat?.userIds || []).filter((id) => id !== user?.id);
      const names = otherIds.map((id) => userById(id)?.name || "Unknown");
      return names.join(", ") || "Just you";
    },
    [user?.id, userById]
  );

  const refresh = useCallback(async () => {
    if (!token) return;
    setLoadingChats(true);
    setError(null);
    try {
      const [chatList, userList] = await Promise.all([
        chatApi.listChats(token),
        chatApi.listUsers(token),
      ]);
      setChats(chatList || []);
      setUsers((userList || []).filter((u) => u.id !== user?.id));
    } catch (err) {
      setError(err.message);
    } finally {
      setLoadingChats(false);
    }
  }, [token, user?.id]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const openChat = useCallback(
    (chat) => {
      setCurrentChat(chat);
      setMessages([]);
      setLoadingMessages(true);
      chatApi
        .listMessages(token, chat.id)
        .then((msgs) => setMessages(msgs || []))
        .catch((err) => setError(err.message))
        .finally(() => setLoadingMessages(false));
    },
    [token]
  );

  const startChatWith = useCallback(
    async (otherUserId) => {
      const existing = chats.find(
        (c) => c.userIds?.length === 2 && c.userIds.includes(otherUserId)
      );
      if (existing) {
        openChat(existing);
        return;
      }
      try {
        const chat = await chatApi.createChat(token, [otherUserId]);
        setChats((prev) => [chat, ...prev]);
        openChat(chat);
      } catch (err) {
        setError(err.message);
      }
    },
    [chats, token, openChat]
  );

  const sendMessage = useCallback(
    async (text) => {
      const trimmed = text.trim();
      if (!trimmed || !currentChat) return;
      try {
        await grpc.sendMessage(token, currentChat.id, trimmed);
      } catch (err) {
        setError(err.message);
      }
    },
    [token, currentChat]
  );

  // Fed by the gRPC stream. Deliberately not wrapped in useCallback: passing
  // a fresh closure every render means useGrpcStream's internal ref always
  // sees the current `chats`/`currentChat`, with no manual ref-mirroring
  // needed (that's what bit the old socket.io version of this handler).
  const handleIncomingMessage = (message) => {
    if (message?.chatId === currentChat?.id) {
      setMessages((prev) => [...prev, message]);
    }

    const knowsChat = chats.some((c) => c.id === message.chatId);
    if (!knowsChat) {
      refresh();
      return;
    }

    setChats((prev) => {
      const idx = prev.findIndex((c) => c.id === message.chatId);
      if (idx === -1) return prev;
      const updated = { ...prev[idx], latestMessage: message };
      const rest = prev.filter((c) => c.id !== message.chatId);
      return [updated, ...rest];
    });
  };

  const { connected } = useGrpcStream(token, handleIncomingMessage);

  return (
    <ChatContext.Provider
      value={{
        connected,
        users,
        chats,
        currentChat,
        messages,
        loadingChats,
        loadingMessages,
        error,
        chatLabel,
        userById,
        openChat,
        startChatWith,
        sendMessage,
        refresh,
      }}
    >
      {children}
    </ChatContext.Provider>
  );
}

export function useChat() {
  return useContext(ChatContext);
}
