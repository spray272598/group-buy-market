// Package port 交易出站端口（Outbound Port）= 领域对外的防腐接口。
//
// 防腐层（Anti-Corruption Layer）在此方向上的职责：
//   - 领域只依赖本包接口，不感知 HTTP Client / RabbitMQ / 第三方回调协议；
//   - infrastructure 中的 ACL 实现负责协议转换、重试、加锁与降级。
//
// 对应实现：internal/infrastructure/acl（出站防腐实现）。
package port
