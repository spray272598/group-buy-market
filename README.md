# group-buy-market（Go · DDD · 拼团营销）

Java「小傅哥拼团」完整 Go 重写：**严格 DDD** + 责任链/策略树/策略模式 + **MySQL / Redis / RabbitMQ** 中间件 + Docker 本地一键环境。

> 目录：`D:\project_go\group-buy-market`（`project_go` 为多项目父目录）

---

## 为什么够秋招深度

| 维度 | 体现 |
|------|------|
| DDD | 限界上下文 activity / trade / tag；聚合事务；仓储端口倒置 |
| 设计模式 | 试算策略树、锁单/结算/退单责任链、折扣/退单策略 |
| 高并发 | Redis 组队库存 occupy + recovery；回调/Job 分布式锁 |
| 最终一致 | 本地消息表 `notify_task` + 定时补偿 + MQ 重试 |
| 可运维 | 限流 + Prometheus/Grafana 全链路看板 + 告警、DCC 热更新 |
| 完整链路 | 试算→锁单→结算成团→回调→退单逆向→超时退单 |

---

## 工程结构

```
group-buy-market/
├── cmd/server/                 # 启动入口
├── configs/                    # 本地 / Docker 配置
├── docs/
│   ├── ddd.md                  # DDD 分层与领域模型
│   ├── design-patterns.md      # 设计模式
│   ├── api.md                  # HTTP API
│   ├── middleware.md           # Docker 中间件
│   ├── quickstart.md
│   └── dev-ops/                # ★ Docker Compose（对齐 Java）
│       ├── docker-compose-environment.yml
│       ├── docker-compose-app.yml
│       ├── mysql/ redis/ rabbitmq/ prometheus/
├── scripts/sql/
├── Dockerfile
└── internal/
    ├── api/                    # 对外契约 DTO/接口
    ├── application/            # 用例编排 + 入站防腐 assembler
    ├── design/                 # 责任链/策略树骨架（无业务）
    ├── domain/                 # 领域 + 出站 port
    ├── infrastructure/         # repository + **acl 出站防腐** + 中间件
    ├── trigger/                # 瘦 HTTP/Job/Listener
    └── app/                    # 组装根
```

**依赖方向：**

```
trigger → application(实现 api) → domain → design
                                    ▲
                         infrastructure（含出站 ACL）
```

| 层 | 一句话 |
|----|--------|
| **api** | 对外合同 |
| **application** | 用例 + **入站防腐** |
| **domain** | 业务规则 + 出站端口定义 |
| **infrastructure/acl** | **出站防腐**（HTTP/MQ） |
| **design** | 模式骨架 |
| **trigger** | 协议适配（越瘦越好） |

防腐层详解：[docs/acl.md](docs/acl.md)

> 合规审计：[docs/ddd-compliance-audit.md](docs/ddd-compliance-audit.md) · design 说明：[docs/design-layer.md](docs/design-layer.md)

---

## 5 分钟启动（推荐 Docker 中间件）

### 1）起中间件

```bash
cd docs/dev-ops
docker compose -f docker-compose-environment.yml up -d
```

| 服务 | 地址 |
|------|------|
| MySQL | 127.0.0.1:13306 root/123456 |
| Redis | 127.0.0.1:16379 |
| RabbitMQ 管理台 | http://127.0.0.1:15672 admin/admin |
| phpMyAdmin | http://127.0.0.1:8899 |
| Redis Admin | http://127.0.0.1:8081 admin/admin |

### 2）起应用

```bash
cd ../..   # 回到 group-buy-market 根
go mod tidy
go run ./cmd/server -config configs/config.yaml
```

### 3）测试台 / Swagger / 观测

| 入口 | 地址 |
|------|------|
| **API 测试页面** | http://127.0.0.1:8091/test/ |
| **Swagger UI** | http://127.0.0.1:8091/swagger/ |
| Grafana 全链路 | http://127.0.0.1:3000 → **Full Stack** + **Dependencies** |
| Prometheus | http://127.0.0.1:9090/targets （app/mysql/redis/rabbitmq） |
| Grafana + ELK | `docker compose -f docker-compose-environment.yml -f docker-compose-observability.yml up -d` |

### 4）冒烟

```bash
# 健康检查
curl http://127.0.0.1:8091/health

# 人群打标（写入 DB + Redis BitSet）
curl "http://127.0.0.1:8091/api/v1/gbm/tag/exec_tag_batch_job?tagId=RQ_KJHKL98UU78H66554GFDV&batchId=10001"

# 营销试算
curl -X POST http://127.0.0.1:8091/api/v1/gbm/index/query_group_buy_market_config \
  -H "Content-Type: application/json" \
  -d "{\"userId\":\"xfg01\",\"source\":\"s01\",\"channel\":\"c01\",\"goodsId\":\"9890001\"}"

# 锁单（开团）
curl -X POST http://127.0.0.1:8091/api/v1/gbm/trade/lock_market_pay_order \
  -H "Content-Type: application/json" \
  -d "{\"userId\":\"xfg01\",\"activityId\":100123,\"goodsId\":\"9890001\",\"source\":\"s01\",\"channel\":\"c01\",\"outTradeNo\":\"100000000001\",\"notifyConfigVO\":{\"notifyType\":\"MQ\"}}"

# 结算
curl -X POST http://127.0.0.1:8091/api/v1/gbm/trade/settlement_market_pay_order \
  -H "Content-Type: application/json" \
  -d "{\"userId\":\"xfg01\",\"source\":\"s01\",\"channel\":\"c01\",\"outTradeNo\":\"100000000001\",\"outTradeTime\":\"2026-08-02T12:00:00+08:00\"}"
```

完整接口见 [docs/api.md](docs/api.md)，中间件细节见 [docs/middleware.md](docs/middleware.md)。

---

## 核心业务与中间件映射

| 业务 | 技术点 |
|------|--------|
| 营销试算 | 策略树 Root→Switch→Market→Tag→End；折扣 ZJ/MJ/N/ZK |
| 降级/切量/黑名单 | DCC + **Redis Pub/Sub 跨实例广播** |
| 锁单 | 责任链 + Redis 库存 + MySQL 事务写团/明细 |
| 结算成团 | 责任链 + complete_count + 达标写 notify_task |
| 回调 | TradePort：Redis 抢锁 → HTTP 或 MQ 持久化消息 |
| 退单 | 三策略 + MQ 通知 + Consumer 恢复 recovery 库存 |
| 超时未支付 | Job + 分布式锁 + 走退单链路 |
| 人群 | Tag 批次任务；Redis BitSet + crowd_tags_detail |

---

## 与 Java 模块映射

| Java | Go |
|------|-----|
| **group-buy-market-api** | **internal/api** |
| group-buy-market-domain | internal/domain |
| group-buy-market-infrastructure | internal/infrastructure |
| group-buy-market-trigger | internal/trigger |
| group-buy-market-types | internal/types |
| group-buy-market-app | internal/app + cmd/server |
| docs/dev-ops | docs/dev-ops |
| wrench design chain/tree | internal/design |

---

## 文档索引

- [docs/ddd.md](docs/ddd.md) — DDD 与分层依赖
- [docs/acl.md](docs/acl.md) — **防腐层（有/在哪/怎么讲）**
- [docs/ddd-compliance-audit.md](docs/ddd-compliance-audit.md) — DDD 合规审计
- [docs/design-layer.md](docs/design-layer.md) — 为何有 design 层
- [docs/architecture-layers.md](docs/architecture-layers.md) — 为何需要 api 层
- [docs/design-patterns.md](docs/design-patterns.md) — 业务中的模式用法
- [docs/api.md](docs/api.md) — HTTP API + 契约接口
- [docs/middleware.md](docs/middleware.md) — Docker / 中间件
- [docs/quickstart.md](docs/quickstart.md) — 快速上手
- [docs/interview.md](docs/interview.md) — 秋招面试讲解提纲

## 测试

```bash
# 领域 + 设计模式 + DCC 跨实例 mock + 限流
go test ./internal/... -count=1
```

覆盖：折扣策略、试算责任链、锁单/结算规则链、退单类型、DCC 广播、限流等。

## License

学习与秋招展示用途，业务对齐原 Java 教学项目。
