// Package assembler 入站防腐层（Inbound Anti-Corruption Layer）。
//
// 将外部契约（api/dto）翻译为领域可理解的模型，避免领域被 HTTP JSON 形状污染。
// 反向：领域结果 → 响应 DTO。
package assembler
