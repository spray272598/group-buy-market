// Package api 对外契约层（对齐 Java group-buy-market-api）。
//
// 职责：
//  1. 定义对外 HTTP/RPC 的请求/响应 DTO（与领域实体解耦）
//  2. 定义对外服务接口（IMarketIndexService / IMarketTradeService 等）
//  3. 统一响应包装 Response
//
// 依赖方向：
//
//	trigger（适配器）──实现──► api 接口
//	api  ──不依赖──► domain / infrastructure
//	domain ──不依赖──► api
//
// 为何需要独立 api 层：
//   - 契约稳定：前端/调用方依赖 DTO，领域模型可自由演进
//   - 多入口复用：HTTP、未来 gRPC、定时回调测试可共用同一契约
//   - 与原 Java 工程模块映射清晰，便于秋招讲解
package api
