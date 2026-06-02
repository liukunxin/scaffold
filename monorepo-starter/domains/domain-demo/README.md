# domain-demo

示例领域模块，仅处理业务逻辑，通过事件与 runtime 交互。

约束：

- 不依赖 runtime 实现细节
- 不直接耦合 transport（http/grpc/ws）
- 通过 `api` 定义输入输出契约

