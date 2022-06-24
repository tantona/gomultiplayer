import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // string shorthand
      // "/foo": "http://localhost:4567",
      // // with options
      // "/api": {
      //   target: "http://jsonplaceholder.typicode.com",
      //   changeOrigin: true,
      //   rewrite: (path) => path.replace(/^\/api/, ""),
      // },
      // // with RegEx
      // "^/fallback/.*": {
      //   target: "http://jsonplaceholder.typicode.com",
      //   changeOrigin: true,
      //   rewrite: (path) => path.replace(/^\/fallback/, ""),
      // },
      // // Using the proxy instance
      // "/api": {
      //   target: "http://jsonplaceholder.typicode.com",
      //   changeOrigin: true,
      //   configure: (proxy, options) => {
      //     // proxy will be an instance of 'http-proxy'
      //   },
      // },
      // // Proxying websockets or socket.io
      "/wsb": {
        target: "ws://localhost:8080",
        ws: true,
      },
    },
  },
});
