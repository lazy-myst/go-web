import { request } from "./client";

export function login(email, password) {
  return request("/auth/login", { method: "POST", body: { email, password } });
}

export function register(name, email, password) {
  return request("/auth/register", {
    method: "POST",
    body: { name, email, password },
  });
}

export function getProfile(token) {
  return request("/users/profile", { token });
}
