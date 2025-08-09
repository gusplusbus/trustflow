import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
    proxy: {
      "/auth": {
        target: "http://auth_server:4000",
        changeOrigin: true,
      },
      "/api": {
        target: "http://api:8080",
        changeOrigin: true,
      },
    },
  },
});
