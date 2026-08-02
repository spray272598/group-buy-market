package http

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// RegisterStatic 注册 API 测试台、Swagger UI、OpenAPI 文档
func RegisterStatic(r *gin.Engine, webRoot string) {
	if webRoot == "" {
		webRoot = "web/static"
	}
	// OpenAPI 规范（优先 docs/openapi.yaml）
	r.GET("/openapi.yaml", func(c *gin.Context) {
		for _, p := range []string{"docs/openapi.yaml", filepath.Join("..", "docs", "openapi.yaml")} {
			if _, err := os.Stat(p); err == nil {
				c.File(p)
				return
			}
		}
		c.String(http.StatusNotFound, "openapi.yaml not found")
	})

	// 测试台首页
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/test/")
	})
	r.GET("/test", func(c *gin.Context) { c.Redirect(http.StatusFound, "/test/") })
	r.Static("/test", webRoot)
	// swagger 目录
	swaggerDir := filepath.Join(webRoot, "swagger")
	r.Static("/swagger", swaggerDir)
	r.GET("/swagger/", func(c *gin.Context) {
		c.File(filepath.Join(swaggerDir, "index.html"))
	})
}
