# 中间件与本地 Docker 部署

对齐 Java 原项目 `docs/dev-ops`，本地默认用 Docker 起全套中间件。

## 1. 中间件清单

| 组件 | 容器名 | 端口 | 用途 |
|------|--------|------|------|
| MySQL 8.0 | gbm-mysql | 13306 | 业务库 `group_buy_market` |
| phpMyAdmin | gbm-phpmyadmin | 8899 | 库表可视化 |
| Redis 6.2 | gbm-redis | 16379 | 组队库存、BitSet 人群、分布式锁、DCC 相关 |
| Redis Commander | gbm-redis-admin | 8081 | Redis 可视化 admin/admin |
| RabbitMQ 3.12 + 管理台 | gbm-rabbitmq | 5672 / 15672 | 成团/退单领域事件 admin/admin |
| Prometheus | gbm-prometheus | 9090 | 抓取 `/metrics` |

## 2. 一键启动中间件

```bash
cd docs/dev-ops
docker compose -f docker-compose-environment.yml up -d
```

查看状态：

```bash
docker compose -f docker-compose-environment.yml ps
```

MySQL 首次启动会自动执行：

`docs/dev-ops/mysql/sql/01-group_buy_market.sql`

## 3. 启动应用（本机 Go）

```bash
cd group-buy-market
go run ./cmd/server -config configs/config.yaml
```

配置已指向：

- MySQL `127.0.0.1:13306`
- Redis `127.0.0.1:16379`
- RabbitMQ `127.0.0.1:5672`

## 4. 全 Docker 启动应用 + 中间件

```bash
cd docs/dev-ops
docker compose -f docker-compose-app.yml up -d --build
```

应用使用 `configs/config-docker.yaml`（host 为容器服务名 mysql/redis/rabbitmq）。

## 5. 在拼团链路中的角色

```
试算/锁单 ──► MySQL（活动/SKU/订单）
         ──► Redis（切量无关；库存 occupy；人群 BitSet）
结算成团 ──► MySQL notify_task（本地消息表）
         ──► TradePort 抢 Redis 锁
         ──► HTTP 回调 或 MQ topic.team_success
退单    ──► MySQL 状态回滚 + notify_task
         ──► MQ topic.team_refund
         ──► Consumer 调 restoreTeamLockStock → Redis recovery
Job     ──► Redis 分布式锁防多实例重复
         ──► 补偿 notify_task / 超时未支付退单
```

## 6. RabbitMQ 拓扑（对齐 Java）

- Exchange: `group_buy_market_exchange`（topic）
- RoutingKey / Queue:
  - `topic.team_success` → `group_buy_market_queue_2_topic_team_success`
  - `topic.team_refund` → `group_buy_market_queue_2_topic_team_refund`

管理台：http://127.0.0.1:15672 （admin/admin）

## 7. 可观测

- 健康：`GET /health`
- Prometheus：`GET /metrics`
- 结构化日志：slog

## 8. 常见问题

1. **MySQL 连不上**：等 healthcheck 完成，或看 `docker logs gbm-mysql`
2. **RabbitMQ 启动慢**：应用会 Warn 并降级（仍写 notify_task，Job 会重试）
3. **SQL 已初始化但改了脚本**：需 `docker compose down -v` 清 volume 再 up
