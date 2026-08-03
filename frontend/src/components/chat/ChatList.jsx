import { useChat } from "../../context/ChatContext";
import { initials, avatarColor } from "../../utils/avatar";

function formatPreviewTime(iso) {
  if (!iso) return "";
  const d = new Date(iso);
  const today = new Date();
  const sameDay = d.toDateString() === today.toDateString();
  return sameDay
    ? d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    : d.toLocaleDateString([], { month: "short", day: "numeric" });
}

export default function ChatList() {
  const { chats, users, currentChat, chatLabel, openChat, startChatWith } =
    useChat();

  return (
    <div className="w-72 shrink-0 border-r border-slate-200 flex flex-col h-full bg-white">
      <div className="px-4 py-3 border-b border-slate-100">
        <h2 className="text-xs font-semibold text-slate-400 uppercase tracking-wide">
          Chats
        </h2>
      </div>

      <div className="flex-1 overflow-y-auto">
        {chats.length === 0 && (
          <div className="flex flex-col items-center justify-center gap-2 text-center px-6 py-10">
            <div className="h-10 w-10 rounded-full bg-slate-100 flex items-center justify-center text-lg">
              💭
            </div>
            <p className="text-xs text-slate-400">
              No conversations yet. Pick someone below to say hi.
            </p>
          </div>
        )}
        {chats.map((chat) => {
          const active = currentChat?.id === chat.id;
          const label = chatLabel(chat);
          return (
            <button
              key={chat.id}
              onClick={() => openChat(chat)}
              className={`w-full flex items-center gap-3 text-left px-4 py-2.5 border-l-2 transition ${
                active
                  ? "bg-indigo-50 border-indigo-600"
                  : "border-transparent hover:bg-slate-50"
              }`}
            >
              <div
                className={`h-9 w-9 shrink-0 rounded-full ${avatarColor(
                  chat.id
                )} flex items-center justify-center text-white text-xs font-semibold`}
              >
                {initials(label)}
              </div>
              <div className="min-w-0 flex-1">
                <div className="flex items-center justify-between gap-2">
                  <p
                    className={`text-sm truncate ${
                      active
                        ? "font-semibold text-indigo-900"
                        : "font-medium text-slate-800"
                    }`}
                  >
                    {label}
                  </p>
                  {chat.latestMessage?.createdAt && (
                    <span className="text-[10px] text-slate-400 shrink-0">
                      {formatPreviewTime(chat.latestMessage.createdAt)}
                    </span>
                  )}
                </div>
                {chat.latestMessage && (
                  <p className="text-xs text-slate-500 truncate">
                    {chat.latestMessage.text}
                  </p>
                )}
              </div>
            </button>
          );
        })}
      </div>

      {users.length > 0 && (
        <div className="border-t border-slate-100">
          <div className="px-4 py-2 text-xs font-semibold text-slate-400 uppercase tracking-wide">
            Start a chat
          </div>
          <div className="max-h-48 overflow-y-auto pb-2">
            {users.map((u) => (
              <button
                key={u.id}
                onClick={() => startChatWith(u.id)}
                className="w-full flex items-center gap-3 text-left px-4 py-1.5 hover:bg-slate-50 transition"
              >
                <div
                  className={`h-7 w-7 shrink-0 rounded-full ${avatarColor(
                    u.id
                  )} flex items-center justify-center text-white text-[10px] font-semibold`}
                >
                  {initials(u.name)}
                </div>
                <span className="text-sm text-slate-700 truncate">
                  {u.name}
                </span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
