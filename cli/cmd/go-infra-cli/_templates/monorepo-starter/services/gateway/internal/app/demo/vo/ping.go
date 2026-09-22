package vo

import contractevents "monorepo-starter/packages/go/contracts/events"

// PingResp 是 ping 的响应体。
//
// 它直接复用跨 Project 契约类型 events.Envelope，而不是就地定义一套字段：
// gateway 与其它语言 / Project 看到的是同一份事件信封定义
// （contracts/events/envelope.schema.json）。这就是「Project -> packages/contracts」
// 这条允许的依赖方向的最小示例。
type PingResp struct {
	contractevents.Envelope
}
