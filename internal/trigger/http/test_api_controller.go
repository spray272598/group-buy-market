package http

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/api/response"
)

// TestAPIController 测试回调接收（非核心业务契约）
type TestAPIController struct{}

func NewTestAPIController() *TestAPIController { return &TestAPIController{} }

func (c *TestAPIController) Register(r *gin.Engine) {
	r.POST("/api/v1/test/group_buy_notify", c.GroupBuyNotify)
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "up"})
	})
}

func (c *TestAPIController) GroupBuyNotify(ctx *gin.Context) {
	body, _ := io.ReadAll(ctx.Request.Body)
	slog.Info("收到拼团回调", "body", string(body))
	ctx.JSON(http.StatusOK, response.Success("success"))
}
