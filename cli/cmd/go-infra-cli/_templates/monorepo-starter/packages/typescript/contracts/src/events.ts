/**
 * contracts/events/envelope.schema.json 的 TypeScript 绑定。
 *
 * 与 Go 绑定（packages/go/contracts/events/envelope.go）字段一一对应。
 * 三份绑定（Go / TS / Python）必须同时改，且只允许向后兼容的变更：
 * 加字段可以，改语义、删字段不行。
 */

/** 跨 Project 事件的统一信封。 */
export interface Envelope {
  /** 全局唯一事件 ID。 */
  eventId: string;
  /** 稳定的事件类型名，形如 `<domain>.<entity>.<action>.v<major>`。 */
  eventType: string;
  /** 可选的会话 / 聚合根标识。 */
  sessionId?: string;
  /** 同一 sessionId 内单调递增。 */
  seq?: number;
  /** Unix 毫秒。 */
  timestamp: number;
  /** 业务负载。消费方必须忽略未知字段，以兼容新增字段。 */
  payload?: Record<string, unknown>;
}

/** 事件类型名格式，与 JSON Schema 的 `pattern` 保持一致。 */
export const EVENT_TYPE_PATTERN = /^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)*\.v[0-9]+$/;

/** 构造一个信封。`eventId` 需由调用方保证全局唯一。 */
export function newEnvelope(
  eventId: string,
  eventType: string,
  payload?: Record<string, unknown>,
): Envelope {
  if (!EVENT_TYPE_PATTERN.test(eventType)) {
    throw new Error(`invalid eventType: ${eventType}`);
  }
  return { eventId, eventType, timestamp: Date.now(), ...(payload ? { payload } : {}) };
}

/** 校验任意值是否符合信封约定；消费外部消息时先过这一层。 */
export function isEnvelope(value: unknown): value is Envelope {
  if (typeof value !== "object" || value === null) {
    return false;
  }
  const env = value as Partial<Envelope>;
  return (
    typeof env.eventId === "string" &&
    env.eventId.length > 0 &&
    typeof env.eventType === "string" &&
    EVENT_TYPE_PATTERN.test(env.eventType) &&
    typeof env.timestamp === "number"
  );
}
