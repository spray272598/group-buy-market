package chain

import "context"

// LogicHandler 责任链节点接口（对齐 Java ILogicHandler）
type LogicHandler[TRequest any, TContext any, TResult any] interface {
	Apply(ctx context.Context, request TRequest, dynamicCtx *TContext) (TResult, error)
	// SetNext 由链路装配时注入
	SetNext(next LogicHandler[TRequest, TContext, TResult])
	Next(ctx context.Context, request TRequest, dynamicCtx *TContext) (TResult, error)
}

// BaseHandler 责任链基类，提供 next 跳转
type BaseHandler[TRequest any, TContext any, TResult any] struct {
	next LogicHandler[TRequest, TContext, TResult]
}

func (b *BaseHandler[TRequest, TContext, TResult]) SetNext(next LogicHandler[TRequest, TContext, TResult]) {
	b.next = next
}

func (b *BaseHandler[TRequest, TContext, TResult]) Next(ctx context.Context, request TRequest, dynamicCtx *TContext) (TResult, error) {
	var zero TResult
	if b.next == nil {
		return zero, nil
	}
	return b.next.Apply(ctx, request, dynamicCtx)
}

// LinkedList 责任链容器
type LinkedList[TRequest any, TContext any, TResult any] struct {
	head LogicHandler[TRequest, TContext, TResult]
	name string
}

func NewLinkedList[TRequest any, TContext any, TResult any](name string, handlers ...LogicHandler[TRequest, TContext, TResult]) *LinkedList[TRequest, TContext, TResult] {
	if len(handlers) == 0 {
		return &LinkedList[TRequest, TContext, TResult]{name: name}
	}
	for i := 0; i < len(handlers)-1; i++ {
		handlers[i].SetNext(handlers[i+1])
	}
	return &LinkedList[TRequest, TContext, TResult]{
		head: handlers[0],
		name: name,
	}
}

func (l *LinkedList[TRequest, TContext, TResult]) Apply(ctx context.Context, request TRequest, dynamicCtx *TContext) (TResult, error) {
	var zero TResult
	if l.head == nil {
		return zero, nil
	}
	return l.head.Apply(ctx, request, dynamicCtx)
}

func (l *LinkedList[TRequest, TContext, TResult]) Name() string {
	return l.name
}
