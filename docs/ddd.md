# DDD 架构说明（group-buy-market Go 版）

本项目严格按 **领域驱动设计（DDD）** 分层，结构对齐 Java 原版（小傅哥拼团）工程模型。

## 1. 分层与依赖方向

```
┌─────────────────────────────────────────────────────────┐
│  trigger（用户接口层 / 触发器）                            │
│  HTTP Controller、定时 Job、消息 Listener                  │
└───────────────────────────┬─────────────────────────────┘
                            │ 调用
                            ▼
┌─────────────────────────────────────────────────────────┐
│  domain（领域层） ★ 核心，不依赖外设                        │
│  model(entity/valobj/aggregate) + service + adapter端口   │
└───────────────────────────┬─────────────────────────────┘
                            │ 端口由基础设施实现
                            ▼
┌─────────────────────────────────────────────────────────┐
│  infrastructure（基础设施层）                              │
│  Repository 实现、Redis、DCC、HTTP/MQ 通知、GORM/PO        │
└─────────────────────────────────────────────────────────┘
                            ▲
┌───────────────────────────┴─────────────────────────────┐
│  app / cmd（组装根 Composition Root）                     │
│  读取配置、注入依赖、启动 HTTP & Job                        │
└─────────────────────────────────────────────────────────┘
```

**依赖铁律：**

| 层 | 可以依赖 | 禁止依赖 |
|----|---------|---------|
| domain | types、design（纯抽象） | infrastructure、trigger、gin、gorm、redis 客户端 |
| infrastructure | domain 端口接口、外部中间件 | 不调用 trigger |
| trigger | domain 服务接口、DTO | 不直接访问 DAO/SQL |
| app | 所有层（仅组装） | 不含业务规则 |

## 2. 限界上下文（Bounded Context）

| 上下文 | 路径 | 职责 |
|--------|------|------|
| **activity** | `internal/domain/activity` | 拼团活动、营销试算、折扣策略、人群可见/可参与 |
| **trade** | `internal/domain/trade` | 锁单、支付结算成团、退单逆向、回调任务 |
| **tag** | `internal/domain/tag`（可扩展） | 人群标签任务与打标 |

各上下文通过 **领域服务编排** 协作；跨上下文不直接引用对方仓储实现，只在 app 组装时注入。

## 3. 战术设计

### 3.1 实体（Entity）

- `MarketProductEntity`：试算请求
- `TrialBalanceEntity`：试算结果
- `MarketPayOrderEntity`：营销支付订单
- `GroupBuyTeamEntity`：拼团组队
- `NotifyTaskEntity`：回调任务

### 3.2 值对象（Value Object）

- `GroupBuyActivityDiscountVO`、`SkuVO`、`NotifyConfigVO`
- `TradeOrderStatus`、`RefundType`、`GroupBuyProgressVO`

### 3.3 聚合（Aggregate）

- `GroupBuyOrderAggregate`：锁单（用户 + 活动 + 优惠）
- `GroupBuyTeamSettlementAggregate`：结算
- `GroupBuyRefundAggregate`：退单

仓储方法以 **聚合** 为事务边界写入。

### 3.4 领域服务

| 服务 | 设计模式 | 说明 |
|------|---------|------|
| 试算 `IndexGroupBuyMarketService` | 策略树/责任链 | Root→Switch→Market→Tag→End |
| 折扣 `ZJ/MJ/N/ZK` | 策略模式 | market_plan 路由 |
| 锁单 `TradeLockOrderService` | 责任链 | 活动可用→参与次数→库存占用 |
| 结算 `TradeSettlementOrderService` | 责任链 | SC黑名单→外部单号→有效时间→结束 |
| 退单 `TradeRefundOrderService` | 责任链+策略 | 数据加载→幂等→策略执行 |

### 3.5 仓储端口（Repository Port）

定义在领域层：

- `domain/activity/adapter/repository.IActivityRepository`
- `domain/trade/adapter/repository.ITradeRepository`
- `domain/trade/adapter/port.ITradePort`

实现在基础设施层：

- `infrastructure/repository.ActivityRepository`
- `infrastructure/repository.TradeRepository`
- `infrastructure/notify.TradePort`

编译期通过 `var _ IActivityRepository = (*ActivityRepository)(nil)` 校验。

## 4. 核心业务流程（领域语言）

### 4.1 营销试算

1. 参数校验  
2. 降级开关 / 流量切量  
3. 并行加载活动配置 + SKU  
4. 按 market_plan 折扣试算  
5. 人群标签控制 visible/enable  

### 4.2 锁单

1. 外部单号幂等  
2. 团队目标是否已满  
3. 试算 + 人群校验  
4. 规则链（活动有效、参与上限、Redis 库存）  
5. 开团或入团，写 `group_buy_order` + `group_buy_order_list`  

### 4.3 结算成团

1. SC 黑名单  
2. 订单存在且未退  
3. 支付时间在拼团有效期内  
4. 明细完成 + complete_count++  
5. 达标则写 notify_task 并异步回调  

### 4.4 退单

按「组队状态 × 订单状态」选择策略：

| 组队 | 订单 | 策略 |
|------|------|------|
| 拼单中 | 锁定 | unpaid2Refund |
| 拼单中 | 已支付 | paid2Refund |
| 已完成/完成含退 | 已支付 | paidTeam2Refund |

## 5. 与 Java 工程映射

| Java 模块 | Go 路径 |
|-----------|---------|
| group-buy-market-types | `internal/types` |
| group-buy-market-domain | `internal/domain` |
| group-buy-market-infrastructure | `internal/infrastructure` |
| group-buy-market-trigger | `internal/trigger` |
| group-buy-market-api | `internal/trigger/http` DTO + response |
| group-buy-market-app | `internal/app` + `cmd/server` |
| wrench design chain/tree | `internal/design/chain`、`tree` |

## 6. 秋招面试可讲点

1. **为何 DDD**：拼团规则复杂（试算、库存、成团、逆向），用领域模型+责任链把规则显式化，便于扩展新折扣/新退单类型。  
2. **依赖倒置**：领域定义接口，基础设施实现，测试可用 mock 仓储。  
3. **事务边界**：锁单/结算/退单以聚合为单位，避免贫血 CRUD。  
4. **缓存库存**：Redis incr 降压 DB，失败 recovery 补偿。  
5. **本地消息表**：notify_task + 定时补偿，保证成团回调最终一致。  
