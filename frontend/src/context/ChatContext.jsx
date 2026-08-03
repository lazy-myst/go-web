import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  useCallback,
} from "react";
import { useAuth } from "./AuthContext";
import { useSocket } from "../hooks/useSocket";
import * as chatApi from "../api/chat";

const ChatContext = createContext(null);

export function ChatProvider({ children }) {
  const { token, user } = useAuth();
  const { socket, connected } = useSocket(token);

  const [users, setUsers] = useState([]);
  const [chats, setChats] = useState([]);
  const [currentChat, setCurrentChat] = useState(null);
  const [messages, setMessages] = useState([]);
  const [loadingChats, setLoadingChats] = useState(false);
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [error, setError] = useState(null);

  const currentChatRef = useRef(null);
  useEffect(() => {
    currentChatRef.current = currentChat;
  }, [currentChat]);

  const chatsRef = useRef([]);
  useEffect(() => {
    chatsRef.current = chats;
  }, [chats]);

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
        (c) =>
          c.userIds?.length === 2 && c.userIds.includes(otherUserId)
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
    (text) => {
      const trimmed = text.trim();
      if (!trimmed || !currentChatRef.current || !socket) return;
      socket.emit("newMessage", { chatId: currentChatRef.current.id, text: trimmed });
    },
    [socket]
  );

  // Registered once per socket instance; reads currentChatRef so it always
  // targets whichever chat is open right now without needing to resubscribe.
  useEffect(() => {
    if (!socket) return undefined;

    const handleNewMessage = (message) => {
      if (message?.chatId === currentChatRef.current?.id) {
        setMessages((prev) => [...prev, message]);
      }

      // First message of a chat this socket wasn't told about yet (created
      // by the other side just now) — refetch so it shows up instead of the
      // user unknowingly creating a duplicate via "start a chat". Checked
      // against a ref mirror of `chats`, since a setState updater's body
      // isn't guaranteed to run synchronously enough to read back here.
      const knowsChat = chatsRef.current.some((c) => c.id === message.chatId);
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

    socket.on("messageCreated", handleNewMessage);
    return () => socket.off("messageCreated", handleNewMessage);
  }, [socket, refresh]);

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
