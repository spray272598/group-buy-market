package node

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/adapter/repository"
	"group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/activity/model/valobj"
	"group-buy-market/internal/domain/activity/service/discount"
	"group-buy-market/internal/domain/activity/service/trial/factory"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
)

// TrialHandler 试算节点函数类型
type TrialHandler func(ctx context.Context, req *entity.MarketProductEntity, dc *factory.DynamicContext) (*entity.TrialBalanceEntity, error)

// Chain 试算策略树：Root -> Switch -> Market -> Tag -> End（或 Error）
type Chain struct {
	repo     repository.IActivityRepository
	discount *discount.Registry
	timeout  time.Duration
	// sem 控制并发度，防止 goroutine 泄露（对齐 Java ThreadPoolConfig）
	sem *semaphore.Weighted
}

func NewChain(repo repository.IActivityRepository, discountReg *discount.Registry) *Chain {
	return &Chain{
		repo:     repo,
		discount: discountReg,
		timeout:  5 * time.Second,
		// 并发度上限 = 4，防止高并发下 goroutine 飙升
		sem: semaphore.NewWeighted(4),
	}
}

// Apply 从 Root 节点开始执行
func (c *Chain) Apply(ctx context.Context, req *entity.MarketProductEntity) (*entity.TrialBalanceEntity, error) {
	dc := &factory.DynamicContext{}
	return c.root(ctx, req, dc)
}

// Root: 参数校验
func (c *Chain) root(ctx context.Context, req *entity.MarketProductEntity, dc *factory.DynamicContext) (*entity.TrialBalanceEntity, error) {
	slog.Info("拼团商品查询试算服务-RootNode", "userId", req.UserID)
	if req.UserID == "" || req.GoodsID == "" || req.Source == "" || req.Channel == "" {
		return nil, exception.NewAppException(enums.ILLEGAL_PARAMETER)
	}
	return c.switchNode(ctx, req, dc)
}

// Switch: 降级 + 切量
func (c *Chain) switchNode(ctx context.Context, req *entity.MarketProductEntity, dc *factory.DynamicContext) (*entity.TrialBalanceEntity, error) {
	slog.Info("拼团商品查询试算服务-SwitchNode", "userId", req.UserID)
	if c.repo.DowngradeSwitch() {
		slog.Info("拼团活动降级拦截", "userId", req.UserID)
		return nil, exception.NewAppException(enums.E0003)
	}
	if !c.repo.CutRange(req.UserID) {
		slog.Info("拼团活动切量拦截", "userId", req.UserID)
		return nil, exception.NewAppException(enums.E0004)
	}
	return c.marketNode(ctx, req, dc)
}

// Market: 并行加载活动/SKU + 折扣试算
func (c *Chain) marketNode(ctx context.Context, req *entity.MarketProductEntity, dc *factory.DynamicContext) (*entity.TrialBalanceEntity, error) {
	slog.Info("拼团商品查询试算服务-MarketNode", "userId", req.UserID)

	// 多线程加载
	if err := c.multiThreadLoad(ctx, req, dc); err != nil {
		return nil, err
	}
	slog.Info("拼团商品查询试算服务-MarketNode 异步加载完成", "userId", req.UserID)

	if dc.GroupBuyActivityDiscountVO == nil || dc.SkuVO == nil {
		return c.errorNode(ctx, req, dc)
	}
	gbd := dc.GroupBuyActivityDiscountVO.GroupBuyDiscount
	if gbd == nil {
		return c.errorNode(ctx, req, dc)
	}

	svc := c.discount.Get(gbd.MarketPlan)
	if svc == nil {
		slog.Info("不存在对应折扣计算服务", "plan", gbd.MarketPlan, "support", c.discount.Keys())
		return nil, exception.NewAppException(enums.E0001)
	}

	payPrice, err := svc.Calculate(ctx, req.UserID, dc.SkuVO.OriginalPrice, gbd)
	if err != nil {
		return nil, err
	}
	deduction := dc.SkuVO.OriginalPrice.Sub(payPrice)
	dc.DeductionPrice = &deduction
	dc.PayPrice = &payPrice

	if dc.DeductionPrice == nil {
		return c.errorNode(ctx, req, dc)
	}
	return c.tagNode(ctx, req, dc)
}

// multiThreadLoad 使用 errgroup 并行加载活动配置与 SKU 数据，
// 并通过信号量控制并发度，防止高流量下 goroutine 飙升。
// 对应 Java 端 CompletableFuture + ThreadPoolExecutor 语义。
func (c *Chain) multiThreadLoad(ctx context.Context, req *entity.MarketProductEntity, dc *factory.DynamicContext) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	// 信号量控制：最多允许 4 个试算请求并行（防止 goroutine 泄露）
	if err := c.sem.Acquire(gCtx, 1); err != nil {
		return err
	}
	defer c.sem.Release(1)

	var (
		act *valobj.GroupBuyActivityDiscountVO
		sku  *valobj.SkuVO
	)

	// 并行加载拼团活动折扣信息
	g.Go(func() error {
		var err error
		act, err = c.loadActivity(gCtx, req)
		return err
	})

	// 并行加载 SKU 商品信息
	g.Go(func() error {
		var err error
		sku, err = c.repo.QuerySkuByGoodsID(gCtx, req.GoodsID)
		return err
	})

	// 任一失败立即取消另一个（errgroup.WithContext 取消传播）
	if err := g.Wait(); err != nil {
		return err
	}

	dc.GroupBuyActivityDiscountVO = act
	dc.SkuVO = sku
	return nil
}

func (c *Chain) loadActivity(ctx context.Context, req *entity.MarketProductEntity) (*valobj.GroupBuyActivityDiscountVO, error) {
	activityID := req.ActivityID
	if activityID == nil {
		sc, err := c.repo.QuerySCSkuActivityBySCGoodsID(ctx, req.Source, req.Channel, req.GoodsID)
		if err != nil {
			return nil, err
		}
		if sc == nil {
			return nil, nil
		}
		id := sc.ActivityID
		activityID = &id
	}
	return c.repo.QueryGroupBuyActivityDiscountVO(ctx, *activityID)
}

// Tag: 人群标签可见/可参与
func (c *Chain) tagNode(ctx context.Context, req *entity.MarketProductEntity, dc *factory.DynamicContext) (*entity.TrialBalanceEntity, error) {
	g := dc.GroupBuyActivityDiscountVO
	tagID := g.TagID
	visible := g.IsVisible()
	enable := g.IsEnable()

	if tagID == "" {
		dc.Visible = true
		dc.Enable = true
		return c.endNode(ctx, req, dc)
	}

	within, err := c.repo.IsTagCrowdRange(ctx, tagID, req.UserID)
	if err != nil {
		return nil, err
	}
	// visible/enable 为 true 表示无限制；为 false 表示需要在人群内
	dc.Visible = visible || within
	dc.Enable = enable || within
	return c.endNode(ctx, req, dc)
}

func (c *Chain) endNode(ctx context.Context, req *entity.MarketProductEntity, dc *factory.DynamicContext) (*entity.TrialBalanceEntity, error) {
	slog.Info("拼团商品查询试算服务-EndNode", "userId", req.UserID)
	g := dc.GroupBuyActivityDiscountVO
	sku := dc.SkuVO
	var deduction, pay decimal.Decimal
	if dc.DeductionPrice != nil {
		deduction = *dc.DeductionPrice
	}
	if dc.PayPrice != nil {
		pay = *dc.PayPrice
	}
	return &entity.TrialBalanceEntity{
		GoodsID:                    sku.GoodsID,
		GoodsName:                  sku.GoodsName,
		OriginalPrice:              sku.OriginalPrice,
		DeductionPrice:             deduction,
		PayPrice:                   pay,
		TargetCount:                g.Target,
		StartTime:                  g.StartTime,
		EndTime:                    g.EndTime,
		IsVisible:                  dc.Visible,
		IsEnable:                   dc.Enable,
		GroupBuyActivityDiscountVO: g,
	}, nil
}

func (c *Chain) errorNode(ctx context.Context, req *entity.MarketProductEntity, dc *factory.DynamicContext) (*entity.TrialBalanceEntity, error) {
	slog.Info("拼团商品查询试算服务-ErrorNode", "userId", req.UserID, "goodsId", req.GoodsID)
	if dc.GroupBuyActivityDiscountVO == nil || dc.SkuVO == nil {
		return nil, exception.NewAppException(enums.E0002)
	}
	return &entity.TrialBalanceEntity{}, nil
}
