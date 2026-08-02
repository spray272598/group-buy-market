// Package application 应用层（用例编排）。
//
// 职责：
//   - 编排一个或多个领域服务完成用户用例；
//   - 通过 assembler 做入站防腐（api DTO ↔ 领域模型）；
//   - 被 trigger 调用，不依赖 gin。
//
// 不把业务规则写在这里（规则在 domain）；这里只做「先试算再锁单」类流程胶水。
package application
