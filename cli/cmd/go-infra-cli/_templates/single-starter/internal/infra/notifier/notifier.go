package notifier

import (
	"context"

	"github.com/liukunxin/go-infra/pkg/base/log"
)

// Notifier 是通知能力的端口，业务层只依赖该接口，不关心具体通道。
type Notifier interface {
	// NotifyUserCreated 在用户创建成功后发出通知。
	NotifyUserCreated(ctx context.Context, userID string) error
}

// LogNotifier 是默认实现：只写一条结构化日志，便于开箱即跑。
// 替换为消息队列 / 邮件 / 短信实现时，只需改 internal/route/init.go 里的注入点。
type LogNotifier struct{}

func NewLogNotifier() Notifier {
	return &LogNotifier{}
}

func (n *LogNotifier) NotifyUserCreated(ctx context.Context, userID string) error {
	log.WithContext(ctx).Infof("user created: uid=%s", userID)
	return nil
}
