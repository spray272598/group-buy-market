package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	activityservice "group-buy-market/internal/domain/activity/service"
	"group-buy-market/internal/domain/activity/service/discount"
	"group-buy-market/internal/domain/activity/service/trial/node"
	tagservice "group-buy-market/internal/domain/tag/service"
	"group-buy-market/internal/domain/trade/service/lock"
	"group-buy-market/internal/domain/trade/service/refund"
	"group-buy-market/internal/domain/trade/service/refund/business"
	"group-buy-market/internal/domain/trade/service/settlement"
	"group-buy-market/internal/domain/trade/service/task"
	"group-buy-market/internal/infrastructure/dcc"
	"group-buy-market/internal/infrastructure/gateway"
	"group-buy-market/internal/infrastructure/mq"
	"group-buy-market/internal/infrastructure/notify"
	"group-buy-market/internal/infrastructure/redis"
	infrarepo "group-buy-market/internal/infrastructure/repository"
	triggerhttp "group-buy-market/internal/trigger/http"
	"group-buy-market/internal/trigger/job"
	"group-buy-market/internal/trigger/listener"
)

// Application 组装根：trigger -> domain <- infrastructure
type Application struct {
	Config *Config
	DB     *gorm.DB
	Redis  *redis.Service
	MQ     *mq.Client
	DCC    *dcc.Service
	Engine *gin.Engine

	notifyJob  *job.NotifyJob
	timeoutJob *job.TimeoutRefundJob
}

func NewApplication(cfg *Config) (*Application, error) {
	// ----- MySQL -----
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("mysql connect: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MySQL.MaxIdle > 0 {
		sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdle)
	}
	if cfg.MySQL.MaxOpen > 0 {
		sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpen)
	}
	sqlDB.SetConnMaxLifetime(time.Hour)

	// ----- Redis -----
	rdb := redis.New(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	{
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := rdb.Ping(ctx); err != nil {
			slog.Warn("Redis 暂不可用，库存/限流/锁能力会受影响", "err", err)
		}
		cancel()
	}

	// ----- DCC -----
	dccSvc := dcc.New(cfg.DCC.DowngradeSwitch, cfg.DCC.CutRange, cfg.DCC.SCBlacklist, cfg.DCC.CacheSwitch)

	// ----- RabbitMQ -----
	var mqClient *mq.Client
	mqClient, err = mq.NewClient(mq.Config{
		URL:      cfg.RabbitMQ.URL,
		Exchange: cfg.RabbitMQ.Exchange,
		TeamSuccess: mq.TopicConfig{
			RoutingKey: firstNonEmpty(cfg.RabbitMQ.TeamSuccess.RoutingKey, cfg.Notify.TopicTeamSuccess, "topic.team_success"),
			Queue:      firstNonEmpty(cfg.RabbitMQ.TeamSuccess.Queue, "group_buy_market_queue_2_topic_team_success"),
		},
		TeamRefund: mq.TopicConfig{
			RoutingKey: firstNonEmpty(cfg.RabbitMQ.TeamRefund.RoutingKey, cfg.Notify.TopicTeamRefund, "topic.team_refund"),
			Queue:      firstNonEmpty(cfg.RabbitMQ.TeamRefund.Queue, "group_buy_market_queue_2_topic_team_refund"),
		},
	})
	if err != nil {
		slog.Warn("RabbitMQ 连接失败，MQ 回调降级为日志（本地消息表仍会写入）", "err", err)
		mqClient = nil
	}

	publisher := mq.NewEventPublisher(mqClient)
	httpGW := gateway.NewGroupBuyNotifyService(cfg.Notify.HTTPTimeoutSec)
	tradePort := notify.NewTradePort(rdb, httpGW, publisher)

	// ----- Repositories（实现领域端口）-----
	activityRepo := infrarepo.NewActivityRepository(db, rdb, dccSvc)
	tradeRepo := infrarepo.NewTradeRepository(db, rdb, dccSvc, cfg.Notify.TopicTeamSuccess, cfg.Notify.TopicTeamRefund)
	tagRepo := infrarepo.NewTagRepository(db, rdb)

	// ----- Domain services -----
	discountReg := discount.NewRegistry(
		discount.NewZJ(activityRepo),
		discount.NewMJ(activityRepo),
		discount.NewN(activityRepo),
		discount.NewZK(activityRepo),
	)
	trialChain := node.NewChain(activityRepo, discountReg)
	indexSvc := activityservice.NewIndexGroupBuyMarketService(activityRepo, trialChain)
	tagSvc := tagservice.NewTagService(tagRepo)

	taskSvc := task.NewTradeTaskService(tradeRepo, tradePort)
	lockSvc := lock.NewTradeLockOrderService(tradeRepo)
	settlementSvc := settlement.NewTradeSettlementOrderService(tradeRepo, taskSvc)
	refundStrategies := business.NewStrategyMap(
		business.NewUnpaid2Refund(tradeRepo, tradePort, taskSvc),
		business.NewPaid2Refund(tradeRepo, tradePort, taskSvc),
		business.NewPaidTeam2Refund(tradeRepo, tradePort, taskSvc),
	)
	refundSvc := refund.NewTradeRefundOrderService(tradeRepo, refundStrategies)

	// ----- MQ consumers -----
	if err := listener.StartConsumers(mqClient, refundSvc); err != nil {
		slog.Error("启动 MQ 消费者失败", "err", err)
	}

	// ----- HTTP Trigger -----
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger())
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	rlStore := triggerhttp.NewRateLimitStore(cfg.RateLimit.IndexQPS, cfg.RateLimit.Burst)
	triggerhttp.NewMarketIndexController(indexSvc, rlStore).Register(engine)
	triggerhttp.NewMarketTradeController(indexSvc, lockSvc, settlementSvc, refundSvc).Register(engine)
	triggerhttp.NewDCCController(dccSvc).Register(engine)
	triggerhttp.NewTagController(tagSvc).Register(engine)
	triggerhttp.NewTestAPIController().Register(engine)

	// ----- Jobs -----
	notifyJob := job.NewNotifyJob(taskSvc, rdb, cfg.Job.NotifyIntervalSec)
	timeoutJob := job.NewTimeoutRefundJob(refundSvc, rdb, cfg.Job.TimeoutRefundIntervalSec)

	return &Application{
		Config:     cfg,
		DB:         db,
		Redis:      rdb,
		MQ:         mqClient,
		DCC:        dccSvc,
		Engine:     engine,
		notifyJob:  notifyJob,
		timeoutJob: timeoutJob,
	}, nil
}

func (a *Application) Start() error {
	a.notifyJob.Start()
	a.timeoutJob.Start()
	addr := fmt.Sprintf(":%d", a.Config.Server.Port)
	slog.Info("group-buy-market 启动",
		"addr", addr,
		"redis", a.Config.Redis.Addr,
		"rabbitmq_ready", a.MQ != nil,
	)
	return a.Engine.Run(addr)
}

func (a *Application) Stop() {
	a.notifyJob.Stop()
	a.timeoutJob.Stop()
	if a.MQ != nil {
		_ = a.MQ.Close()
	}
	_ = a.Redis.Close()
	if sqlDB, err := a.DB.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
