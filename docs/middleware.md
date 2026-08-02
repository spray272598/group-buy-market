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
| RabbitMQ | 5672 / 15672 | admin / admin |
| phpMyAdmin | 8899 | - |
| Redis Commander | 8081 | admin / admin |
| Prometheus | 9090 | - |

## 2. 观测栈（Grafana + ELK，推荐秋招演示）

```bash
cd docs/dev-ops
docker compose -f docker-compose-environment.yml -f docker-compose-observability.yml up -d
```

也可拆开：

```bash
# 仅 Grafana（依赖已启动的 prometheus）
docker compose -f docker-compose-environment.yml -f docker-compose-grafana.yml up -d

# 仅 ELK
docker compose -f docker-compose-environment.yml up -d
docker compose -f docker-compose-elk.yml up -d
```

| 组件 | 端口 | 说明 |
|------|------|------|
| Grafana | 3000 | admin/admin，自动导入 Prometheus 与 Runtime 看板 |
| Elasticsearch | 9200 | 日志索引 |
| Logstash | 4560 | TCP json_lines 接入 |
| Kibana | 5601 | 日志检索 |

### 应用日志进 ELK

```bash
# Windows PowerShell
$env:GBM_LOGSTASH_ADDR="127.0.0.1:4560"
go run ./cmd/server -config configs/config.yaml
```

在 Kibana 中创建 Index Pattern：`group-buy-market-log-*`。

### Grafana

浏览器打开 http://127.0.0.1:3000 ，看板 **Group Buy Market · Runtime**（goroutine / 内存 / CPU）。

应用需先监听 8091，Prometheus 通过 `host.docker.internal:8091/metrics` 抓取。

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
指标      ── /metrics → Prometheus → Grafana
```

## 5. 停止

```bash
docker compose -f docker-compose-environment.yml -f docker-compose-observability.yml down
# 清数据
docker compose -f docker-compose-environment.yml -f docker-compose-observability.yml down -v
```
