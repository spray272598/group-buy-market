package main

import (
	"flag"
	"log/slog"
	"os"

	"group-buy-market/internal/app"
	"group-buy-market/internal/infrastructure/logx"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 日志：默认本地 Text；GBM_LOGSTASH_ADDR 设置时 JSON + 双写 Logstash（ELK）
	if os.Getenv("GBM_LOGSTASH_ADDR") != "" {
		logx.SetupDefault()
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	}

	cfg, err := app.LoadConfig(*cfgPath)
	if err != nil {
		slog.Error("加载配置失败", "err", err)
		os.Exit(1)
	}

	application, err := app.NewApplication(cfg)
	if err != nil {
		slog.Error("应用初始化失败", "err", err)
		os.Exit(1)
	}
	defer application.Stop()

	if err := application.Start(); err != nil {
		slog.Error("服务退出", "err", err)
		os.Exit(1)
	}
}
