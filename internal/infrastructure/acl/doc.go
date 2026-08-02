// Package acl 基础设施侧的防腐层实现（Anti-Corruption Layer）。
//
// 入站防腐：HTTP JSON ↔ 领域命令 在 application/assembler 完成。
// 出站防腐：领域通知意图 ↔ HTTP/MQ 协议 在本包完成（如 TradeNotifyACL）。
//
// 领域只依赖 domain/*/adapter/port 接口，永不 import 本包。
package acl
