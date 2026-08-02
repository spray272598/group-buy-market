# 快速启动

## 方式 A：Docker 中间件 + 本机应用（推荐）

### 1. 启动中间件

```bash
cd docs/dev-ops
docker compose -f docker-compose-environment.yml up -d
docker compose -f docker-compose-environment.yml ps
```

等待 MySQL healthy（约 20～40s）。库表自动初始化。

### 2. 启动应用

```bash
cd ../..  # group-buy-market 根目录
go run ./cmd/server -config configs/config.yaml
```

### 3. 验证中间件连通

| 检查 | 命令/地址 |
|------|-----------|
| 健康 | `curl http://127.0.0.1:8091/health` |
| Metrics | `curl http://127.0.0.1:8091/metrics` |
| RabbitMQ | http://127.0.0.1:15672 |
| phpMyAdmin | http://127.0.0.1:8899 |

## 方式 B：全 Docker（应用也容器化）

```bash
cd docs/dev-ops
docker compose -f docker-compose-app.yml up -d --build
```

应用配置：`configs/config-docker.yaml`。

## 完整业务冒烟

```bash
# 0 打标
curl "http://127.0.0.1:8091/api/v1/gbm/tag/exec_tag_batch_job?tagId=RQ_KJHKL98UU78H66554GFDV&batchId=10001"

# 1 试算
curl -X POST http://127.0.0.1:8091/api/v1/gbm/index/query_group_buy_market_config \
  -H "Content-Type: application/json" \
  -d "{\"userId\":\"xfg01\",\"source\":\"s01\",\"channel\":\"c01\",\"goodsId\":\"9890001\"}"

# 2 三人成团：锁单（每人不同 outTradeNo）
# user xfg01 开团
curl -X POST http://127.0.0.1:8091/api/v1/gbm/trade/lock_market_pay_order \
  -H "Content-Type: application/json" \
  -d "{\"userId\":\"xfg01\",\"activityId\":100123,\"goodsId\":\"9890001\",\"source\":\"s01\",\"channel\":\"c01\",\"outTradeNo\":\"200000000001\",\"notifyConfigVO\":{\"notifyType\":\"MQ\"}}"
# 记下返回 teamId，xfg02/xfg03 参团时带上 teamId

# 3 结算（每人支付后调一次）
curl -X POST http://127.0.0.1:8091/api/v1/gbm/trade/settlement_market_pay_order \
  -H "Content-Type: application/json" \
  -d "{\"userId\":\"xfg01\",\"source\":\"s01\",\"channel\":\"c01\",\"outTradeNo\":\"200000000001\",\"outTradeTime\":\"2026-08-02T12:00:00+08:00\"}"

# 4 退单
curl -X POST http://127.0.0.1:8091/api/v1/gbm/trade/refund_market_pay_order \
  -H "Content-Type: application/json" \
  -d "{\"userId\":\"xfg01\",\"outTradeNo\":\"200000000001\",\"source\":\"s01\",\"channel\":\"c01\"}"

# 5 DCC 降级开关
curl "http://127.0.0.1:8091/api/v1/gbm/dcc/update_config?key=downgradeSwitch&value=0"
```

## 单测

```bash
go test ./internal/domain/... ./internal/design/...
```

## 停中间件

```bash
cd docs/dev-ops
docker compose -f docker-compose-environment.yml down
# 清数据卷
docker compose -f docker-compose-environment.yml down -v
```
