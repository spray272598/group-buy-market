package node

import (
	"context"
	"os"
	"testing"
	"time"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/activity/service/trial/factory"
	infracrepo "group-buy-market/internal/infrastructure/repository"
	"group-buy-market/internal/infrastructure/dcc"
	redisx "group-buy-market/internal/infrastructure/redis"
)

// BenchmarkTrialLoadReal 连真实 Docker MySQL 的基准测试（默认跳过）。
//
// 运行方式（需先 docker compose up -d mysql redis）：
//   GBM_REAL_DB=1 go test -bench=BenchmarkTrialLoadReal -benchmem -benchtime=500x \
//     ./internal/domain/activity/service/trial/node/
//
// 环境变量 GBM_REAL_DB=1 时启用；否则 t.Skip 跳过，不影响普通单测。
func BenchmarkTrialLoadReal(b *testing.B) {
	if os.Getenv("GBM_REAL_DB") != "1" {
		b.Skip("set GBM_REAL_DB=1 to run real DB benchmark (requires docker mysql)")
	}

	dsn := "root:123456@tcp(127.0.0.1:13306)/group_buy_market?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{Logger: loggerSilent{}})
	if err != nil {
		b.Fatalf("connect mysql: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	rdb := redisx.New("127.0.0.1:16379", "", 0)
	dccSvc := dcc.New("false", "false", "false", "false")
	realRepo := infracrepo.NewActivityRepository(db, rdb, dccSvc)

	req := &entity.MarketProductEntity{
		UserID:  "xfg01",
		Source:  "s01",
		Channel: "c01",
		GoodsID: "9890001",
	}

	b.Run("parallel", func(b *testing.B) {
		c := NewChain(realRepo, nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dc := &factory.DynamicContext{}
			_ = c.multiThreadLoad(context.Background(), req, dc)
		}
	})

	b.Run("serial", func(b *testing.B) {
		c := NewChain(realRepo, nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dc := &factory.DynamicContext{}
			_ = serialLoad(context.Background(), c, req, dc)
		}
	})
}

// loggerSilent 关闭 gorm 日志噪声，避免 benchmark 输出刷屏
type loggerSilent struct{}

func (loggerSilent) LogMode(gormlogger.LogLevel) gormlogger.Interface { return loggerSilent{} }
func (loggerSilent) Info(context.Context, string, ...interface{})     {}
func (loggerSilent) Warn(context.Context, string, ...interface{})     {}
func (loggerSilent) Error(context.Context, string, ...interface{})    {}
func (loggerSilent) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
}
