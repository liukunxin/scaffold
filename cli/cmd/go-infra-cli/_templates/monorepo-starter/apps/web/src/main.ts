import "./style.css";

import { ping } from "./api";

// 这是一个最小可跑的 Project：只做一件事——调 services/gateway，把原始响应打出来。
// 它存在的意义是示范「一个非 Go Project 在 Monorepo 里怎么放、怎么引用 packages/、怎么调别的 Project」。
const app = document.querySelector<HTMLElement>("#app");
if (!app) {
  throw new Error("找不到 #app 挂载点");
}

app.innerHTML = `
  <h1>web · monorepo demo</h1>
  <p class="muted">
    这是 <code>apps/web</code>：一个普通的前端 Project。它通过
    <code>@repo/contracts</code> 复用跨 Project 契约类型，通过 HTTP 调
    <code>services/gateway</code>，不 import 对方的任何内部代码。
  </p>
  <p>
    <input id="name" value="go-infra" />
    <button id="send">调用 /api/demo/ping</button>
  </p>
  <pre id="out">点上面的按钮发起请求（需要先启动 services/gateway）</pre>
`;

const nameInput = app.querySelector<HTMLInputElement>("#name");
const out = app.querySelector<HTMLPreElement>("#out");

app.querySelector<HTMLButtonElement>("#send")?.addEventListener("click", async () => {
  if (!out) {
    return;
  }
  out.classList.remove("error");
  out.textContent = "请求中…";
  try {
    const resp = await ping(nameInput?.value ?? "");
    out.textContent = JSON.stringify(resp, null, 2);
  } catch (err) {
    out.classList.add("error");
    out.textContent = err instanceof Error ? err.message : String(err);
  }
});
