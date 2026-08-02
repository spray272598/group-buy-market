# 秋招面试讲解提纲（建议 8～12 分钟）

## 1. 项目一句话

Go 实现的 **拼团营销中台**：DDD 拆分活动试算、交易锁单/结算/退单、人群标签；用责任链把规则显式化；用 Redis 库存 + 本地消息表 + RabbitMQ 保证高并发与最终一致。

## 2. 为什么 DDD + 独立 api 层

- 规则多：降级、切量、人群、参与次数、库存、成团条件、三种退单。
- 若写成 Controller 堆 if-else，扩展一个折扣/退单类型就要改核心流。
- **领域服务 + 责任链/策略**：开闭原则；仓储端口方便单测 mock。
- **api 层**：对外 DTO/接口与领域实体分离（对齐 Java `group-buy-market-api`），防腐 + 多入口复用。  
  详见 [architecture-layers.md](./architecture-layers.md)。

## 3. 核心链路怎么讲

### 试算（读路径）

参数校验 → 降级/切量 → **并行**查活动+SKU → 折扣策略 → 人群 visible/enable。

### 锁单（写路径）

幂等 outTradeNo → 试算+人群 → 规则链（活动有效、次数、**Redis occupy**）→ 事务写团/明细；失败 **recovery**。

### 结算

SC 黑名单 → 单号存在 → 支付时间在拼团有效期 → complete++ → 达标写 **notify_task** → 异步回调（HTTP/MQ）。

### 退单

状态组合选策略 → 改订单/团状态 → MQ 退单事件 → 消费端 **orderId 锁防重** 恢复库存。

## 4. 必问点准备

**Q: Redis 库存和 DB lock_count 不一致怎么办？**  
A: 以 DB 事务为准；Redis 做前置过滤降压；失败/退单走 recovery；最终可对账 Job。

**Q: 为什么本地消息表还要 MQ？**  
A: 先落库保证「一定会尝试通知」；MQ 解耦下游；Job 补偿未成功任务；多实例回调用 Redis 锁防重。

**Q: 多实例 Job 如何不重复执行？**  
A: `tryLock(wait, lease)` 抢 `group_buy_market_*_job_exec`。

**Q: DDD 依赖方向？**  
A: domain 不依赖 gorm/gin；infrastructure 实现端口；app 组装。

## 5. 可展示的扩展

- 限流（userId QPS）
- Prometheus metrics
- Docker 一键中间件
- Tag BitSet 人群（百万级判断 O(1)）

## 6. 诚实边界

- MQ 消费侧示例业务（成团成功）以日志/库存恢复为主，可接商城履约。
- 未做完整 ELK 链路（可讲日志规范与扩展点）。
