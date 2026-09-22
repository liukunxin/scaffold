package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/liukunxin/go-infra/pkg/base/uuid"
	contractevents "monorepo-starter/packages/go/contracts/events"
	"monorepo-starter/services/gateway/internal/app/demo/dto"
	"monorepo-starter/services/gateway/internal/app/demo/vo"
)

// EventDemoPingCompleted 是本服务产生的事件类型名，形如 <domain>.<entity>.<action>.v<major>。
const EventDemoPingCompleted = "demo.ping.completed.v1"

// DemoService 演示 Project 内部的三层调用：controller -> service（无 logic）。
type DemoService interface {
	// Ping 返回一条演示用 pong 消息，包在跨 Project 事件信封里。
	Ping(ctx context.Context, in dto.PingInput) (*vo.PingResp, error)
}

type demoService struct{}

func NewDemoService() DemoService {
	return &demoService{}
}

func (s *demoService) Ping(_ context.Context, in dto.PingInput) (*vo.PingResp, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = "world"
	}
	// 事件 ID 用 go-infra 的 snowflake 服务生成，不自己拼时间戳。
	eventID := strconv.FormatInt(uuid.GetIDService().GenerateUserID(), 10)
	return &vo.PingResp{
		Envelope: contractevents.New(eventID, EventDemoPingCompleted, map[string]any{
			"message": fmt.Sprintf("pong, %s", name),
		}),
	}, nil
}
