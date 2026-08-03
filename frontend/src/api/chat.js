import { request } from "./client";

export function listUsers(token) {
  return request("/users", { token });
}

export function listChats(token) {
  return request("/chats", { token });
}

export function createChat(token, userIds) {
  return request("/chats", { method: "POST", body: { userIds }, token });
}

export function listMessages(token, chatId) {
  return request(`/chats/${chatId}/messages`, { token });
}
