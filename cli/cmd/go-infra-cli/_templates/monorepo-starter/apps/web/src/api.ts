import { isEnvelope, type Envelope } from "@repo/contracts";

/**
 * go-infra 的统一响应壳（pkg/biz/controller.CommonResponse）。
 * 所有走 kcontroller.GinBase 的接口都返回这个结构。
 */
export interface ApiResponse<T> {
  code: number;
  msg: string;
  data?: T;
  trace_id: string;
}

/** /api/demo/ping 的 data：服务端直接返回跨 Project 的事件信封。 */
export type PingVO = Envelope;

/**
 * 调用 services/gateway 的演示接口。
 *
 * 注意这里没有 import 任何 gateway 的内部代码——前端与后端之间只有 HTTP + 契约，
 * 契约类型来自 packages/typescript/contracts（它是 contracts/ 的 TS 绑定）。
 */
export async function ping(name: string): Promise<ApiResponse<PingVO>> {
  const resp = await fetch(`/api/demo/ping?name=${encodeURIComponent(name)}`);
  if (!resp.ok) {
    throw new Error(`请求失败：HTTP ${resp.status}`);
  }

  const body = (await resp.json()) as ApiResponse<PingVO>;
  if (body.code !== 0) {
    throw new Error(`接口返回错误：code=${body.code} msg=${body.msg}`);
  }
  if (body.data && !isEnvelope(body.data)) {
    throw new Error("响应不符合契约：data 不是合法的 EventEnvelope");
  }
  return body;
}
