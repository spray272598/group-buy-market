# DDD 合规审计报告

> 审计日期：随仓库维护更新。方法：依赖扫描 + 职责对照 + 与 Java 原项目映射。

## 1. 结论摘要

| 维度 | 结论 | 说明 |
|------|------|------|
| 依赖方向 | **基本合规** | domain 不依赖 infra/api/gin/gorm/redis |
| 契约层 api | **合规** | api 不依赖 domain |
| design 层 | **合规** | 纯模式框架，domain 可依赖 |
| 仓储端口 | **合规** | 接口在 domain，实现在 infrastructure |
| 限界上下文 | **可接受** | 有少量跨上下文引用（与原项目类似） |
| 应用层 | **已具备** | `internal/application` 用例 + assembler 入站防腐 |
| 防腐层 | **已具备** | 入站 assembler + 出站 `infrastructure/acl` |
| 待改进 | 见第 4 节 | 非阻断 |

**总体：符合本项目目标下的 DDD 工程实践（对齐小傅哥四层/六边形），不是贫血 CRUD。**

---

## 2. 依赖铁律检查

### 2.1 domain → 禁止依赖

| 禁止项 | 结果 |
|--------|------|
| infrastructure | ✅ 无 import |
| trigger | ✅ 无 |
| api | ✅ 无 |
| gin / gorm / go-redis | ✅ 无 |

domain 允许依赖：

- `internal/types`（枚举、业务异常）
- `internal/design`（责任链/策略树骨架）
- 同领域或共享模型（见 3.2）

### 2.2 api → 禁止依赖

| 禁止项 | 结果 |
|--------|------|
| domain | ✅ 无（仅注释提及） |
| infrastructure | ✅ 无 |
| trigger | ✅ 无 |
| gin | ✅ 无 |

### 2.3 design → 禁止依赖

| 禁止项 | 结果 |
|--------|------|
| 任何业务包 | ✅ 仅 `context` |

### 2.4 infrastructure

- ✅ 实现 `IActivityRepository` / `ITradeRepository` / `ITagRepository` / `ITradePort`
- ✅ 编译期 `var _ Interface = (*Impl)(nil)`

### 2.5 trigger

- ✅ 实现 `api.IMarketIndexService` 等契约
- ✅ 调用 domain 服务完成业务
- ⚠️ 直接依赖部分 infrastructure（DCC、Redis 锁、MQ Client）— **适配器层可接受**，更纯的做法是再包一层 domain 端口

---

## 3. 战术设计检查

### 3.1 已具备

| 元素 | 位置 | 状态 |
|------|------|------|
| 实体 Entity | `domain/*/model/entity` | ✅ |
| 值对象 VO | `domain/*/model/valobj` | ✅ |
| 聚合 Aggregate | 锁单/结算/退单聚合 | ✅ |
| 领域服务 | 试算、锁单、结算、退单、任务、标签 | ✅ |
| 仓储端口 | `adapter/repository` | ✅ |
| 防腐/出站端口 | `ITradePort` | ✅ |
| 策略 | 折扣 ZJ/MJ/N/ZK；退单三策略 | ✅ |
| 责任链 | 锁单/结算/退单规则链 | ✅ |
| 限界上下文 | activity / trade / tag | ✅ |

### 3.2 跨上下文引用

```
trade → activity.model.entity.UserGroupBuyOrderDetailEntity
```

用途：超时未支付扫描返回列表结构。  
评价：**弱耦合、可接受**；更严可抽 shared kernel 或 trade 内自有 DTO。  
原 Java 亦有 trade 使用 activity 侧明细类型的类似情况。

### 3.3 无独立 Application 层

编排主要在：

- `domain/*/service`（领域服务）
- `trigger/http`（协议适配 + 少量用例串联，如锁单前试算）

评价：与原 Java「Controller → Domain Service」一致；若追求教科书式 DDD，可把「试算+锁单」收成 application 用例服务。当前 **深度够秋招**，非硬伤。

---

## 4. 待改进清单

| 状态 | 项 | 说明 |
|------|----|------|
| ✅ 已做 | application 层 | 用例从 Controller 下沉 |
| ✅ 已做 | 出站 ACL 包 | `infrastructure/acl.TradeNotifyACL` |
| ✅ 已做 | 入站 assembler | `application/assembler` |
| ✅ 已做 | trade 超时模型 | `TimeoutUnpaidOrderEntity`，不再依赖 activity entity |
| 可选 | 试算接入 `design/tree` | 现 hand-written Chain 语义等价 |
| 可选 | DCC 再包一层端口 | trigger 仍可直连 dcc（配置中心适配器） |

---

## 5. design 层在 DDD 中的位置（答疑）

**问：写了 design 是否破坏 DDD？**  
**答：不破坏。** design = 与业务无关的模式库，不是第四业务层。

```
[api] 契约
[trigger] 适配器
[domain] 业务 + 使用 design 组装责任链/策略
[design] 纯模式（类似工具库）
[infrastructure] 技术实现
```

详见 [design-layer.md](./design-layer.md)。

---

## 6. 快速自检命令

```bash
# domain 不得出现这些依赖（应无输出）
rg "infrastructure|gin-gonic|gorm.io|go-redis|internal/api" internal/domain --glob "*.go"

# api 不得依赖 domain（应无代码 import）
rg "internal/domain|gin-gonic|gorm" internal/api --glob "*.go"

# design 应只有标准库
rg "group-buy-market" internal/design --glob "*.go"
```

---

## 7. 审计结论（面试可用）

> 项目按 DDD 分包：api 契约、domain 核心、infrastructure 实现端口、trigger 入站适配。domain 不依赖框架与中间件。责任链/策略通过 design 通用组件落地。与 Java 四层模型一致；跨上下文与 application 层做了工程化取舍，已记录改进项。
