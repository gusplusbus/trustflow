import { api } from "./api";
export async function updateUser(email: string, password: string) {
  return api("/auth/users", {
    method: "PUT",
    body: JSON.stringify({ email, password }),
  });
}
