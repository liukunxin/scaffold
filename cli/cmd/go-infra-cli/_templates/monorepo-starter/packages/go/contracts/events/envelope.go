// Package events 是 contracts/events/envelope.schema.json 的 Go 绑定。
//
// 跨 Project 的事件统一用这个信封承载，任何语言看到的是同一份定义。
package events

import "time"

// Envelope 是跨 Project 事件的统一信封。
//
// 字段与 contracts/events/envelope.schema.json 一一对应；
// 变更该结构时必须同步改契约，且只允许向后兼容的改动。
type Envelope struct {
	// EventID 全局唯一事件 ID。
	EventID string `json:"eventId"`
	// EventType 稳定的事件类型名，形如 <domain>.<entity>.<action>.v<major>。
	// 版本表达在 eventType 里，不要靠 URL 路径版本。
	EventType string `json:"eventType"`
	// SessionID 可选的会话 / 聚合根标识。
	SessionID string `json:"sessionId,omitempty"`
	// Seq 同一 SessionID 内单调递增。
	Seq int64 `json:"seq,omitempty"`
	// Timestamp 事件发生时间，Unix 毫秒。
	Timestamp int64 `json:"timestamp"`
	// Payload 业务负载。消费方必须忽略未知字段，以兼容新增字段。
	Payload map[string]any `json:"payload,omitempty"`
}

// New 按当前时间构造一个信封。EventID 由调用方保证全局唯一
// （通常用 go-infra 的 pkg/base/uuid 生成）。
func New(eventID, eventType string, payload map[string]any) Envelope {
	return Envelope{
		EventID:   eventID,
		EventType: eventType,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	}
}
