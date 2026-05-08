import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: "../web/public/vite",
    emptyOutDir: true,
    rollupOptions: {
      input: "./src/main.ts",
      output: {
        entryFileNames: "myblog.js",
        chunkFileNames: "chunks/[name]-[hash].js",
        assetFileNames: "assets/[name]-[hash][extname]",
      },
    },
  },
});
