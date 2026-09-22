"""contracts/events/envelope.schema.json 的 Python 绑定。

与 Go 绑定（packages/go/contracts/events/envelope.go）字段一一对应。
三份绑定（Go / TS / Python）必须同时改，且只允许向后兼容的变更：
加字段可以，改语义、删字段不行。
"""

from __future__ import annotations

import re
import time
from dataclasses import dataclass, field
from typing import Any, Mapping

# 事件类型名格式，与 JSON Schema 的 pattern 保持一致。
EVENT_TYPE_PATTERN = re.compile(r"^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)*\.v[0-9]+$")

# gateway 在 ping 用例里产生的事件类型（示例）。
EVENT_DEMO_PING_COMPLETED = "demo.ping.completed.v1"


@dataclass(slots=True)
class Envelope:
    """跨 Project 事件的统一信封。

    线上结构用 camelCase（与 Go / TS 绑定一致），Python 侧用 snake_case，
    转换收敛在 to_wire / from_wire 两个方法里。
    """

    event_id: str
    event_type: str
    timestamp: int
    session_id: str | None = None
    seq: int | None = None
    payload: dict[str, Any] = field(default_factory=dict)

    def to_wire(self) -> dict[str, Any]:
        """转成跨语言一致的 JSON 结构。"""
        raw: dict[str, Any] = {
            "eventId": self.event_id,
            "eventType": self.event_type,
            "timestamp": self.timestamp,
        }
        if self.session_id is not None:
            raw["sessionId"] = self.session_id
        if self.seq is not None:
            raw["seq"] = self.seq
        if self.payload:
            raw["payload"] = self.payload
        return raw

    @classmethod
    def from_wire(cls, raw: Mapping[str, Any]) -> "Envelope":
        """从外部消息解析。

        未知字段直接忽略——这是契约要求的向前兼容行为，不要在这里报错。
        """
        payload = raw.get("payload")
        return cls(
            event_id=str(raw["eventId"]),
            event_type=str(raw["eventType"]),
            timestamp=int(raw["timestamp"]),
            session_id=raw.get("sessionId"),
            seq=raw.get("seq"),
            payload=dict(payload) if isinstance(payload, Mapping) else {},
        )


def new_envelope(
    event_id: str,
    event_type: str,
    payload: Mapping[str, Any] | None = None,
) -> Envelope:
    """构造一个信封。`event_id` 需由调用方保证全局唯一。"""
    if not EVENT_TYPE_PATTERN.match(event_type):
        raise ValueError(f"invalid event_type: {event_type}")
    return Envelope(
        event_id=event_id,
        event_type=event_type,
        timestamp=int(time.time() * 1000),
        payload=dict(payload or {}),
    )
