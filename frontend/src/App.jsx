import { AuthProvider, useAuth } from "./context/AuthContext";
import { ChatProvider } from "./context/ChatContext";
import Login from "./components/Login";
import ChatList from "./components/chat/ChatList";
import ChatWindow from "./components/chat/ChatWindow";
import { initials, avatarColor } from "./utils/avatar";

function ChatApp() {
  const { user, logout } = useAuth();

  return (
    <ChatProvider>
      <div className="flex flex-col h-screen bg-slate-50">
        <header className="flex items-center justify-between px-4 py-2.5 border-b border-slate-200 bg-white shadow-sm z-10">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-xl bg-indigo-600 flex items-center justify-center text-white text-sm font-bold">
              💬
            </div>
            <h1 className="text-sm font-semibold text-slate-800">Chat</h1>
          </div>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <div
                className={`h-7 w-7 rounded-full ${avatarColor(
                  user?.id
                )} flex items-center justify-center text-white text-[11px] font-semibold`}
              >
                {initials(user?.name)}
              </div>
              <span className="text-sm text-slate-600 hidden sm:inline">
                {user?.name}
              </span>
            </div>
            <button
              onClick={logout}
              className="flex items-center gap-1 text-xs text-slate-500 hover:text-red-600 hover:bg-red-50 px-2 py-1.5 rounded-lg transition"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="h-3.5 w-3.5"
              >
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                <polyline points="16 17 21 12 16 7" />
                <line x1="21" y1="12" x2="9" y2="12" />
              </svg>
              Log out
            </button>
          </div>
        </header>
        <div className="flex flex-1 overflow-hidden">
          <ChatList />
          <ChatWindow />
        </div>
      </div>
    </ChatProvider>
  );
}

function AppContent() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen text-sm text-slate-400 bg-slate-50">
        Loading...
      </div>
    );
  }

  return user ? <ChatApp /> : <Login />;
}

export default function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
}
