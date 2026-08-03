import { useAuth } from "../../context/AuthContext";
import { useChat } from "../../context/ChatContext";
import { initials, avatarColor } from "../../utils/avatar";

function formatTime(iso) {
  if (!iso) return "";
  const d = new Date(iso);
  return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

export default function MessageBubble({ message }) {
  const { user } = useAuth();
  const { userById } = useChat();
  const isMine = message.senderId === user?.id;
  const sender = userById(message.senderId);

  return (
    <div
      className={`flex items-end gap-2 mb-3 ${
        isMine ? "justify-end" : "justify-start"
      }`}
    >
      {!isMine && (
        <div
          className={`h-6 w-6 shrink-0 rounded-full ${avatarColor(
            message.senderId
          )} flex items-center justify-center text-white text-[9px] font-semibold`}
        >
          {initials(sender?.name)}
        </div>
      )}
      <div
        className={`max-w-[70%] px-3.5 py-2 text-sm shadow-sm ${
          isMine
            ? "bg-indigo-600 text-white rounded-2xl rounded-br-sm"
            : "bg-white text-slate-800 rounded-2xl rounded-bl-sm border border-slate-100"
        }`}
      >
        <p className="whitespace-pre-wrap break-words leading-relaxed">
          {message.text}
        </p>
        <p
          className={`text-[10px] mt-1 text-right ${
            isMine ? "text-indigo-200" : "text-slate-400"
          }`}
        >
          {formatTime(message.createdAt)}
        </p>
      </div>
    </div>
  );
}
