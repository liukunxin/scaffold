package route

import (
	"github.com/gin-gonic/gin"
	llmcontroller "single-starter/internal/app/llm/controller"
	llmservice "single-starter/internal/app/llm/service"
)

// setupLLM 由 `init --features llm` 注入到 Setup 的 FEATURE_ROUTES 锚点。
// LLM 业务代码不在模板主线里，只有勾选该能力时才会被拷进项目。
func setupLLM(api *gin.RouterGroup) {
	llmController := llmcontroller.NewLLMController(llmservice.NewLLMService())
	api.POST("/llm/ping", llmController.Ping)
}
