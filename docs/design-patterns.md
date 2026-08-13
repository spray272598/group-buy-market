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

## 7. 状态机模式（State Machine / FSM）

### 7.1 建模对象：回调任务 `NotifyTask`

回调补偿链路上，`notify_task` 的状态原本是散落的魔法数字（`0/1/2/3`），流转逻辑手写在 `TradeTaskService` 的 switch 里。为消除魔法数字、约束非法流转，引入了**轻量状态机**。

### 7.2 实现位置

`domain/trade/model/valobj/notify_task_status.go`（纯 Go 手写，零第三方依赖）

### 7.3 状态与迁移表

```
        ┌──────── 成功(1) ────────┐
初始(0) ─┤                        ├─ 终态
        └─ 重试(2) ─ 成功(1)/失败(3)
        └─ 失败(3) ───────────────┘
```

| 当前状态 | 允许迁移到 | 说明 |
|---------|-----------|------|
| 初始 0 | 成功 1 / 重试 2 / 失败 3 | 首次投递 |
| 重试 2 | 成功 1 / 失败 3 | 补偿重投 |
| 成功 1 | —（终态） | 不再处理 |
| 失败 3 | —（终态） | 人工介入 |

### 7.4 核心 API

- `CanTransition(to)`：判断当前状态能否迁到目标状态
- `MoveTo(to)`：执行迁移，非法迁移返回 error（不修改原状态）
- `IsTerminal()`：是否为终态（成功/失败），用于跳过无需再补偿的任务

### 7.5 为什么不用第三方 FSM 框架（如 looplab/fsm）

- 只有 4 个状态、迁移规则简单，手写迁移表足够清晰
- 更贴合 DDD **充血模型**：状态机逻辑内聚在领域值对象上
- 少一个外部依赖，避免"两套并行抽象"（退单已用策略模式覆盖）

### 7.6 迁移守卫

"重试 ≤ 4 次才允许进入失败态"这一业务规则保留在 `TradeTaskService` 中（`NotifyCount > 4` 才调 `MoveTo(Error)`），作为状态机的**外部守卫条件**，与迁移表内聚的"非法流转拦截"形成两层防护。

