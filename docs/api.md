# HTTP API 文档

Base URL: `http://127.0.0.1:8091`

统一响应：

```json
{
  "code": "0000",
  "info": "成功",
  "data": {}
}
```

## 1. 健康检查

`GET /health`

## 2. 查询拼团营销配置

`POST /api/v1/gbm/index/query_group_buy_market_config`

```json
{
  "userId": "xfg01",
  "source": "s01",
  "channel": "c01",
  "goodsId": "9890001"
}
```

返回：活动 ID、试算价格、在拼团队、统计数据。

## 3. 锁单

`POST /api/v1/gbm/trade/lock_market_pay_order`

```json
{
  "userId": "xfg01",
  "teamId": "",
  "activityId": 100123,
  "goodsId": "9890001",
  "source": "s01",
  "channel": "c01",
  "outTradeNo": "100000000001",
  "notifyConfigVO": {
    "notifyType": "HTTP",
    "notifyUrl": "http://127.0.0.1:8091/api/v1/test/group_buy_notify"
  }
}
```

- `teamId` 空：开新团  
- `teamId` 有值：加入已有团  

## 4. 结算（支付成功回调拼团）

`POST /api/v1/gbm/trade/settlement_market_pay_order`

```json
{
  "userId": "xfg01",
  "source": "s01",
  "channel": "c01",
  "outTradeNo": "100000000001",
  "outTradeTime": "2026-08-02T12:00:00+08:00"
}
```

## 5. 退单

`POST /api/v1/gbm/trade/refund_market_pay_order`

```json
{
  "userId": "xfg01",
  "outTradeNo": "100000000001",
  "source": "s01",
  "channel": "c01"
}
```

## 6. 动态配置 DCC

`GET /api/v1/gbm/dcc/query`

`GET /api/v1/gbm/dcc/update_config?key=downgradeSwitch&value=0`（对齐 Java）

`POST /api/v1/gbm/dcc/update`

```json
{
  "key": "downgradeSwitch",
  "value": "0"
}
```

支持 key：`downgradeSwitch`、`cutRange`、`scBlacklist`、`cacheSwitch`。

## 7. 人群标签批次

`GET /api/v1/gbm/tag/exec_tag_batch_job?tagId=RQ_KJHKL98UU78H66554GFDV&batchId=10001`

写入 `crowd_tags_detail` + Redis BitSet。

## 8. 测试回调 / 健康 / Metrics

- `POST /api/v1/test/group_buy_notify` — HTTP 回调接收
- `GET /health`
- `GET /metrics` — Prometheus

## 常见业务码

| code | 含义 |
|------|------|
| 0000 | 成功 |
| 0002 | 非法参数 |
| E0002 | 无拼团营销配置 |
| E0003 | 降级拦截 |
| E0006 | 拼团目标已达成 |
| E0007 | 人群限定不可参与 |
| E0008 | 缓存库存不足 |
| E0101 | 活动未生效 |
| E0103 | 参与次数达上限 |
| E0104 | 外部单号不存在或已退 |
| E0106 | 支付时间超拼团有效期 |
