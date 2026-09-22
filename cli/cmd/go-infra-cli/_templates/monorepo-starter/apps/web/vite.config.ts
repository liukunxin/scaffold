import { defineConfig } from "vite";

// 开发期把 /api 反代到 services/gateway 的 HTTP 端口（configs/config.yml 里的 http.addr）。
// 这样前端只写相对路径，不必在浏览器里处理跨域，也不需要在代码里硬编码后端地址。
export default defineConfig({
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
      },
    },
  },
});
