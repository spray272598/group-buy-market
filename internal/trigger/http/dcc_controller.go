package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/infrastructure/dcc"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/response"
)

// DCCController 动态配置（对齐 Java DCCController）
type DCCController struct {
	dcc *dcc.Service
}

func NewDCCController(svc *dcc.Service) *DCCController {
	return &DCCController{dcc: svc}
}

func (c *DCCController) Register(r *gin.Engine) {
	g := r.Group("/api/v1/gbm/dcc")
	g.GET("/query", c.Query)
	g.GET("/update_config", c.UpdateConfig) // 对齐 Java GET update_config?key=&value=
	g.POST("/update", c.Update)
}

func (c *DCCController) Query(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, response.Success(c.dcc.Snapshot()))
}

// UpdateConfig curl "http://127.0.0.1:8091/api/v1/gbm/dcc/update_config?key=downgradeSwitch&value=1"
func (c *DCCController) UpdateConfig(ctx *gin.Context) {
	key := ctx.Query("key")
	value := ctx.Query("value")
	if key == "" {
		ctx.JSON(http.StatusOK, response.Fail[bool](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	slog.Info("DCC 动态配置值变更", "key", key, "value", value)
	c.dcc.Update(key, value)
	ctx.JSON(http.StatusOK, response.Success(true))
}

type dccUpdateReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (c *DCCController) Update(ctx *gin.Context) {
	var req dccUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Key == "" {
		ctx.JSON(http.StatusOK, response.Fail[any](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	c.dcc.Update(req.Key, req.Value)
	ctx.JSON(http.StatusOK, response.Success(c.dcc.Snapshot()))
}
