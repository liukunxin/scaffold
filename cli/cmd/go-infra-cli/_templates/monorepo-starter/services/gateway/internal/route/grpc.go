package route

import (
	ggrpc "google.golang.org/grpc"
	demogrpc "monorepo-starter/services/gateway/internal/app/demo/grpc"
	demoservice "monorepo-starter/services/gateway/internal/app/demo/service"
)

// RegisterGRPC 注册 gRPC 服务入口。Server 的创建与装配在 bootstrap，服务挂载统一收在这里，
// 与 HTTP 侧的 Setup 对称。只做传输层粘合，不写业务分支。
func RegisterGRPC(server *ggrpc.Server) {
	demogrpc.Register(server, demogrpc.NewService(demoservice.NewDemoService()))
}
