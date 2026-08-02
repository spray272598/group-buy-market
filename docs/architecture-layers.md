# 分层说明：为什么需要 api 层？

> 文档随架构变更同步更新。当前版本：引入独立 `internal/api` 契约层。

## 一句话

**api 层 = 对外「合同」；domain = 对内「业务」；trigger = 把 HTTP/MQ 接到合同上。**

## 和「没有 api 层」时的对比

| | 无 api（早期写法） | 有 api（当前） |
|--|-------------------|---------------|
| DTO 放哪 | `trigger/http/dto.go` | `api/dto` |
| 接口定义 | 无 | `IMarketIndexService` 等 |
| 换 gRPC | 再抄一遍 DTO | 复用 api 契约 |
| 领域改字段 | 容易误伤 HTTP JSON | 领域与 DTO 可分别演进 |
| 对标 Java | 缺模块 | 对齐 `group-buy-market-api` |

## 依赖图

```
        ┌──────────── api ────────────┐
        │  DTO / Response / Interface │  ← 不依赖 domain
        └──────────────▲──────────────┘
                       │ 实现
┌────────── trigger ───┴────┐
│ Controller / Job / MQ     │
└──────────┬────────────────┘
           │ 调用领域服务
           ▼
        domain  ◄── infrastructure
```

## 各层一句话职责

| 层 | 路径 | 职责 |
|----|------|------|
| api | `internal/api` | 定义「系统对外承诺什么」 |
| domain | `internal/domain` | 定义「拼团业务规则是什么」 |
| infrastructure | `internal/infrastructure` | 定义「数据与中间件怎么落地」 |
| trigger | `internal/trigger` | 定义「请求从哪进来、怎么适配」 |
| app | `internal/app` | 定义「怎么组装启动」 |

## 实现约定

1. Controller 必须 `var _ api.IXxxService = (*Controller)(nil)` 编译期校验。  
2. Handler 只做 bind/限流/JSON，核心逻辑进契约方法再调 domain。  
3. **禁止** domain import api。  
4. **禁止** api import domain / gin / gorm。  

## 面试怎么讲

> 「我们有独立 api 模块，和 Java 的 api 包一样，只放 DTO 和对外接口。HTTP 适配器实现这些接口，领域层完全不知道 HTTP 的存在，这样契约和业务模型分离，符合 DDD 防腐思想。」
