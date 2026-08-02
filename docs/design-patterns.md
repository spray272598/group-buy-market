# 设计模式说明

## 1. 责任链模式（Chain of Responsibility）

### 1.1 框架位置

`internal/design/chain`

- `LogicHandler`：节点接口  
- `BaseHandler`：持有 next，提供 `Next()`  
- `LinkedList`：按序装配链  

### 1.2 业务使用

| 链路 | 节点顺序 | 文件 |
|------|---------|------|
| 锁单规则 | 活动可用 → 参与次数 → 组队库存 | `trade/service/lock/filter` |
| 结算规则 | SC黑名单 → 外部单号 → 有效时间 → End | `trade/service/settlement/filter` |
| 退单规则 | 数据加载 → 幂等 → 策略执行 | `trade/service/refund/filter` |

**好处：** 新增规则只需加节点，不改核心服务代码（开闭原则）。

## 2. 策略树 / 路由树（Strategy Router Tree）

试算流程使用 **带路由的多节点树**（对齐 Java AbstractMultiThreadStrategyRouter）：

```
RootNode（参数校验）
   └── SwitchNode（降级 / 切量）
          └── MarketNode（并行查活动+SKU，折扣试算）
                 ├── TagNode → EndNode
                 └── ErrorNode（无配置）
```

实现：`domain/activity/service/trial/node`

MarketNode 内用 `errgroup`/WaitGroup **并行加载** 活动与 SKU，缩短 RT。

## 3. 策略模式（Strategy）

### 3.1 折扣计算

| 编码 | 含义 | 表达式示例 |
|------|------|-----------|
| ZJ | 直减 | `20` → 减 20 元 |
| MJ | 满减 | `100,10` → 满 100 减 10 |
| N | N 元购 | `9.9` → 直接 9.9 |
| ZK | 折扣 | `0.8` → 打 8 折 |

注册表：`discount.Registry`，按 `market_plan` 路由。

### 3.2 退单策略

| Bean 名 | 场景 |
|---------|------|
| unpaid2RefundStrategy | 未支付未成团 |
| paid2RefundStrategy | 已支付未成团 |
| paidTeam2RefundStrategy | 已支付已成团 |

由 `RefundTypeEnum` 根据组队状态×订单状态选择。

## 4. 仓储模式（Repository）

- 领域只依赖接口  
- 基础设施用 GORM 实现  
- 聚合内事务保证一致性  

## 5. 端口适配器（Ports & Adapters / 六边形）

- **入站适配器**：HTTP Controller、Job  
- **出站适配器**：MySQL Repository、Redis、HTTP Notify  

领域在中心，外设可替换（例如 MQ 可从日志模拟换成真实 RabbitMQ）。

## 6. 工厂 / 组装

- 规则链在 `NewTradeLockRuleFilter` 等工厂方法中装配  
- `app.NewApplication` 作为 Composition Root 统一注入  
