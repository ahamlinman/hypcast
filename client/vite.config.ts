import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    // Defaults to "dist" if undefined.
    outDir: process.env.BUILD_PATH,
  },
});
