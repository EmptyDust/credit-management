import path from "path";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8082",
        changeOrigin: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes("node_modules")) {
            return undefined;
          }
          if (
            id.includes("react") ||
            id.includes("react-router") ||
            id.includes("@tanstack/react-query") ||
            id.includes("@radix-ui") ||
            id.includes("cmdk") ||
            id.includes("lucide-react") ||
            id.includes("zustand")
          ) {
            return "react-vendor";
          }
          return "vendor";
        },
      },
    },
  },
});
