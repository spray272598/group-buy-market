package tree

import "context"

// StrategyHandler 策略树节点接口（对齐 Java StrategyHandler）
// TRequest 请求，TContext 动态上下文，TResult 结果
type StrategyHandler[TRequest any, TContext any, TResult any] interface {
	Apply(ctx context.Context, request TRequest, dynamicCtx *TContext) (TResult, error)
}

// StrategyMapper 路由到下一节点
type StrategyMapper[TRequest any, TContext any, TResult any] interface {
	Get(ctx context.Context, request TRequest, dynamicCtx *TContext) (StrategyHandler[TRequest, TContext, TResult], error)
}

// AbstractStrategyRouter 策略路由抽象：doApply -> router(get next)
type AbstractStrategyRouter[TRequest any, TContext any, TResult any] struct {
	DoApply func(ctx context.Context, request TRequest, dynamicCtx *TContext) (TResult, error)
	GetNext func(ctx context.Context, request TRequest, dynamicCtx *TContext) (StrategyHandler[TRequest, TContext, TResult], error)
	// MultiThread 可选：节点执行前异步加载数据
	MultiThread func(ctx context.Context, request TRequest, dynamicCtx *TContext) error
}

func (r *AbstractStrategyRouter[TRequest, TContext, TResult]) Apply(ctx context.Context, request TRequest, dynamicCtx *TContext) (TResult, error) {
	var zero TResult
	if r.MultiThread != nil {
		if err := r.MultiThread(ctx, request, dynamicCtx); err != nil {
			return zero, err
		}
	}
	return r.DoApply(ctx, request, dynamicCtx)
}

func (r *AbstractStrategyRouter[TRequest, TContext, TResult]) Router(ctx context.Context, request TRequest, dynamicCtx *TContext) (TResult, error) {
	var zero TResult
	next, err := r.GetNext(ctx, request, dynamicCtx)
	if err != nil {
		return zero, err
	}
	if next == nil {
		return zero, nil
	}
	return next.Apply(ctx, request, dynamicCtx)
}

// DefaultNilHandler 空处理器，链终止
type DefaultNilHandler[TRequest any, TContext any, TResult any] struct{}

func (d *DefaultNilHandler[TRequest, TContext, TResult]) Apply(ctx context.Context, request TRequest, dynamicCtx *TContext) (TResult, error) {
	var zero TResult
	return zero, nil
}
