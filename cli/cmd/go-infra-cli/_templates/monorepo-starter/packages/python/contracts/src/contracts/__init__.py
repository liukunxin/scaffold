"""跨 Project 契约的 Python 绑定。

定义在 contracts/，本包只是 Python 侧的绑定，不要在这里改契约语义。
"""

from .events import (
    EVENT_DEMO_PING_COMPLETED,
    EVENT_TYPE_PATTERN,
    Envelope,
    new_envelope,
)

__all__ = [
    "EVENT_DEMO_PING_COMPLETED",
    "EVENT_TYPE_PATTERN",
    "Envelope",
    "new_envelope",
]
