import { useEffect, useRef } from "react";
import { useChat } from "../../context/ChatContext";
import { initials, avatarColor } from "../../utils/avatar";
import MessageBubble from "./MessageBubble";
import MessageInput from "./MessageInput";

export default function ChatWindow() {
  const { currentChat, messages, loadingMessages, chatLabel, connected } =
    useChat();
  const bottomRef = useRef(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  if (!currentChat) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center gap-2 text-slate-400 bg-slate-50">
        <div className="h-14 w-14 rounded-full bg-white shadow-sm flex items-center justify-center text-2xl">
          👋
        </div>
        <p className="text-sm">Select a chat or start a new one.</p>
      </div>
    );
  }

  const label = chatLabel(currentChat);

  return (
    <div className="flex-1 flex flex-col h-full bg-slate-50">
      <div className="flex items-center gap-3 px-4 py-2.5 border-b border-slate-200 bg-white shadow-sm z-10">
        <div
          className={`h-9 w-9 rounded-full ${avatarColor(
            currentChat.id
          )} flex items-center justify-center text-white text-xs font-semibold`}
        >
          {initials(label)}
        </div>
        <div className="min-w-0 flex-1">
          <h2 className="text-sm font-semibold text-slate-800 truncate">
            {label}
          </h2>
          <p className="flex items-center gap-1 text-[11px] text-slate-400">
            <span
              className={`h-1.5 w-1.5 rounded-full ${
                connected ? "bg-emerald-500" : "bg-slate-300"
              }`}
            />
            {connected ? "Online" : "Connecting..."}
          </p>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto px-4 py-4">
        {loadingMessages ? (
          <p className="text-xs text-slate-400 text-center mt-4">
            Loading messages...
          </p>
        ) : messages.length === 0 ? (
          <p className="text-xs text-slate-400 text-center mt-4">
            No messages yet. Say hi!
          </p>
        ) : (
          messages.map((m) => <MessageBubble key={m.id} message={m} />)
        )}
        <div ref={bottomRef} />
      </div>

      <MessageInput />
    </div>
  );
}
