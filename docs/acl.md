# 防腐层（Anti-Corruption Layer）说明

## 我们有防腐层吗？

**有。** 分 **入站** 与 **出站** 两条，对应六边形架构两侧。

```
          外部调用方（HTTP/前端）
                    │
                    ▼
            ┌───────────────┐
            │  api（契约DTO） │
            └───────┬───────┘
                    │ 入站防腐
                    ▼
         application/assembler
         （DTO ↔ 领域模型）
                    │
                    ▼
            ┌───────────────┐
            │    domain     │
            └───────┬───────┘
                    │ 出站端口 ITradePort
                    ▼
         infrastructure/acl
         （领域通知 ↔ HTTP/MQ）
                    │
                    ▼
          外部系统（商城/消息总线）
```

## 1. 入站防腐（Inbound ACL）

| 组件 | 路径 | 作用 |
|------|------|------|
| 契约 | `internal/api/dto` | 外部 JSON 形状，领域不知道它 |
| 组装器 | `internal/application/assembler` | DTO → 领域 Entity/Command；领域结果 → 响应 DTO |
| 应用服务 | `internal/application` | 用例编排，不泄漏 HTTP |

**防止什么：** 领域被 `json:"userId"`、前端字段命名、notifyUrl 兼容字段绑架。

## 2. 出站防腐（Outbound ACL）

| 组件 | 路径 | 作用 |
|------|------|------|
| 端口（领域定义） | `domain/trade/adapter/port.ITradePort` | 领域只说「要通知」 |
| ACL 实现 | `infrastructure/acl.TradeNotifyACL` | 翻译成 HTTP 回调或 MQ 消息 + 分布式锁 |
| 网关 | `infrastructure/gateway` | 纯 HTTP 客户端细节 |

**防止什么：** 领域 service 里出现 `http.Client`、RabbitMQ SDK、锁 key 拼装。

## 3. 仓储适配也是一种防腐

`infrastructure/repository` 把 **PO/表结构** 转成 **领域对象**，隔离 MySQL 列名与领域模型。  
严格说这是 Repository Adapter；与对外系统的 ACL 同类思想。

## 4. 和「没有防腐」的对比

| | 无 ACL | 有 ACL（当前） |
|--|--------|----------------|
| Controller | 直接 new 领域对象、塞 JSON 字段 | 只 bind DTO，调 application |
| 换回调协议 | 改 domain | 只改 `TradeNotifyACL` |
| 领域测试 | 要 mock HTTP | mock `ITradePort` 即可 |

## 5. 面试一句话

> 「入站用 api DTO + assembler 把外部协议翻译成领域语言；出站用 ITradePort 端口，由 infrastructure/acl 翻译成 HTTP/MQ。领域核心不依赖任何外部协议，这就是防腐层。」
