import { api } from "./api";
export async function updateUser(email: string, password: string) {
  return api("/api/users", {
    method: "PUT",
    body: JSON.stringify({ email, password }),
  });
}
