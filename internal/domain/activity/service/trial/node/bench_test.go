package node

import (
	"context"
	"testing"
	"time"

	"group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/activity/service/trial/factory"
)

// BenchmarkTrialLoad 对比 errgroup 并行加载 vs 串行加载的耗时。
//
// 模拟真实调用图（req.ActivityID == nil 分支）：
//   - loadActivity: QuerySCSkuActivityBySCGoodsID + QueryGroupBuyActivityDiscountVO = 2 次 DB 查询
//   - QuerySkuByGoodsID: 1 次 DB 查询
//   串行总查询 = 3 次；并行 = max(loadActivity的2次串行, sku的1次)
//
// 用法（先 go build）：
//   go test -bench=BenchmarkTrialLoad -benchmem -benchtime=1000x ./internal/domain/activity/service/trial/node/
//
// 说明：dbRT 模拟单次 SQL 在本地 Docker MySQL 的网络+磁盘 RT。
// 真实值请以本地 Docker 起中间件后连真实 DB 的基准数据为准。
func BenchmarkTrialLoad(b *testing.B) {
	req := &entity.MarketProductEntity{
		UserID:  "xfg01",
		Source:  "s01",
		Channel: "c01",
		GoodsID: "9890001",
	}

	cases := []struct {
		name string
		rt   time.Duration
	}{
		{"dbRT_3ms", 3 * time.Millisecond},
		{"dbRT_8ms", 8 * time.Millisecond},
	}

	for _, tc := range cases {
		b.Run(tc.name+"_parallel", func(b *testing.B) {
			c := NewChain(&mockRepo{dbRT: tc.rt}, nil)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dc := &factory.DynamicContext{}
				_ = c.multiThreadLoad(context.Background(), req, dc)
			}
		})

		b.Run(tc.name+"_serial", func(b *testing.B) {
			c := NewChain(&mockRepo{dbRT: tc.rt}, nil)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dc := &factory.DynamicContext{}
				_ = serialLoad(context.Background(), c, req, dc)
			}
		})
	}
}
