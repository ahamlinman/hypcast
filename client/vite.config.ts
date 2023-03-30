import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: process.env.BUILD_PATH, // Defaults to "dist" if the environment variable is unset.
  },
  server: {
    proxy: {
      "/api": {
        target: process.env.HYPCAST_SERVER ?? "http://localhost:9200",
        ws: true,
      },
    },
  },
});
