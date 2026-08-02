# 中间件与 Docker 编排

## 需要吗？Swagger？

**需要。** 秋招项目里 Swagger/OpenAPI 是加分项：接口可自测、可演示、可写简历。  
本项目已提供：

| 入口 | 地址 |
|------|------|
| API 测试台 | http://127.0.0.1:8091/test/ |
| Swagger UI | http://127.0.0.1:8091/swagger/ |
| OpenAPI 规范 | http://127.0.0.1:8091/openapi.yaml |

使用 **手写 OpenAPI 3 + Swagger UI CDN**，无需 `swag init` 代码生成，维护成本低。

---

## 1. 核心中间件（必起）

```bash
cd docs/dev-ops
docker compose -f docker-compose-environment.yml up -d
```

| 组件 | 端口 | 账号 |
|------|------|------|
| MySQL | 13306 | root / 123456 |
| Redis | 16379 | - |
| RabbitMQ | 5672 / 15672 / **15692** | admin / admin（15692=Prometheus） |
| phpMyAdmin | 8899 | - |
| Redis Commander | 8081 | admin / admin |
| Prometheus | 9090 | - |
| mysqld-exporter | 9104 | - |
| redis-exporter | 9121 | - |

## 2. 全链路监控（Prometheus + Grafana + Exporter）

```text
                    ┌─────────────────────────────────────────┐
  试算/锁单/结算/退单 │  Gin MetricsMiddleware + 业务码解析     │
  限流 / DCC / 库存  │  gbm_* 应用指标  →  :8091/metrics        │
  Job / Notify / MQ │  Redis 客户端 / DB 池 / 依赖探针          │
                    └──────────────────┬──────────────────────┘
                                       │ scrape
         ┌─────────────────────────────┼─────────────────────────────┐
         ▼                             ▼                             ▼
   mysqld-exporter              redis-exporter              rabbitmq:15692
   (MySQL 全局状态)              (命令/内存/客户端)           (队列堆积/吞吐)
         └─────────────────────────────┬─────────────────────────────┘
                                       ▼
                               Prometheus :9090
                               (+ rules 告警)
                                       ▼
                               Grafana :3000
                     Full Stack 看板  +  Dependencies 看板
```

```bash
cd docs/dev-ops
# 核心中间件 + Prometheus + Exporter
docker compose -f docker-compose-environment.yml up -d
# Grafana（及可选 ELK）
docker compose -f docker-compose-environment.yml -f docker-compose-observability.yml up -d
```

| 组件 | 端口 | 说明 |
|------|------|------|
| Grafana | 3000 | admin/admin，自动导入双看板 |
| Prometheus | 9090 | 抓取 app + mysql + redis + rabbitmq |
| Elasticsearch | 9200 | 日志索引（可选） |
| Logstash | 4560 | TCP json_lines |
| Kibana | 5601 | 日志检索 |

### Grafana 看板

| 看板 | UID | 内容 |
|------|-----|------|
| **Group Buy Market · Full Stack** | `gbm-fullstack` | 入口 RED → 业务链路/库存/限流/DCC → Job/Notify/MQ → Redis 客户端/DB 池 → Runtime |
| **Group Buy Market · Dependencies** | `gbm-deps` | MySQL / Redis / RabbitMQ 中间件指标 + 应用探针 |

应用需先监听 **8091**，Prometheus 通过 `host.docker.internal:8091/metrics` 抓取。

### 应用侧指标一览（`gbm_*`）

| 分层 | 指标 |
|------|------|
| 入口 | `gbm_http_requests_total` / `gbm_http_request_duration_seconds` / `gbm_http_requests_in_flight` |
| 限流 | `gbm_rate_limit_total{allow\|deny}` |
| 业务码 | `gbm_biz_requests_total{path,code}` |
| 库存 | `gbm_stock_ops_total{op,result}` occupy/recovery/refund_recovery |
| DCC | `gbm_dcc_updates_total{key,source}` |
| Job | `gbm_job_runs_total` / `gbm_job_processed_total` / `gbm_job_duration_seconds` |
| Notify | `gbm_notify_total` / `gbm_notify_duration_seconds` |
| MQ | `gbm_mq_publish_total` / `gbm_mq_consume_total` / `gbm_mq_consume_duration_seconds` |
| Redis 客户端 | `gbm_redis_ops_total` / `gbm_redis_op_duration_seconds` |
| DB 池 | `gbm_db_open_connections` / `in_use` / `idle` / `max_open` / `wait_count` |
| 依赖健康 | `gbm_dependency_up{name=mysql\|redis\|rabbitmq}` |

### 演示限流 + 看板

```bash
# 同一 userId 连续打试算（config rate_limit.index_qps=1）
for i in 1 2 3 4 5; do
  curl -s -X POST http://127.0.0.1:8091/api/v1/gbm/index/query_group_buy_market_config \
    -H "Content-Type: application/json" \
    -d '{"userId":"xfg01","source":"s01","channel":"c01","goodsId":"9890001"}'
  echo
done
# 观察：第二次起 code=0006；Grafana Full Stack「限流」deny 上升
```

### 自检 Targets

1. 应用：http://127.0.0.1:8091/metrics 搜 `gbm_`
2. Prometheus：http://127.0.0.1:9090/targets 四类 job 应为 UP  
   `group-buy-market` / `mysql` / `redis` / `rabbitmq`
3. Alerts：http://127.0.0.1:9090/alerts  
   覆盖 TargetDown、依赖 Down、5xx、P95、限流、业务成功率、库存失败、Redis/MQ 错误、DB 池饱和等

> RabbitMQ 若已有旧数据卷，改 `enabled_plugins` 后需重建容器：  
> `docker compose -f docker-compose-environment.yml up -d --force-recreate rabbitmq prometheus`

### 应用日志进 ELK

```bash
# Windows PowerShell
$env:GBM_LOGSTASH_ADDR="127.0.0.1:4560"
go run ./cmd/server -config configs/config.yaml
```

在 Kibana 中创建 Index Pattern：`group-buy-market-log-*`。
---

## 3. DCC 跨实例通知

对齐 Java `RTopic`：

1. HTTP `update_config` 更新本机内存  
2. Redis **Pub/Sub** 频道 `group_buy_market_dcc_topic` 广播 `{key,value,origin}`  
3. 其他实例 Subscribe 后 `applyLocal`  

多实例验证：起两个进程不同端口，改 DCC 后两边 `query` 一致。

---

## 4. 中间件在业务中的位置

```
试算/锁单 ── MySQL + Redis(库存/BitSet/DCC)
结算      ── MySQL notify_task + TradePort(Redis锁) + RabbitMQ/HTTP
退单      ── MySQL + MQ topic.team_refund + Consumer 恢复 recovery
Job       ── Redis 分布式锁
日志      ── Logstash → ES → Kibana
指标      ── App gbm_* + Exporters → Prometheus → Grafana（Full Stack / Dependencies）
```

## 5. 停止

```bash
docker compose -f docker-compose-environment.yml -f docker-compose-observability.yml down
# 清数据
docker compose -f docker-compose-environment.yml -f docker-compose-observability.yml down -v
```
