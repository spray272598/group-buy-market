-- ============================================================================
-- 拼团营销中台 - 测试数据初始化脚本
-- 用于业务集成测试和单元测试的测试数据准备
-- ============================================================================

-- 清空测试数据（可选，测试前执行）
DELETE FROM group_buy_order_list;
DELETE FROM group_buy_order;
DELETE FROM notify_task;
DELETE FROM group_buy_activity;
DELETE FROM group_buy_discount;
DELETE FROM sc_sku_activity;
DELETE FROM sku;
DELETE FROM crowd_tags_detail;
DELETE FROM crowd_tags_job;
DELETE FROM crowd_tags;

-- ============================================================================
-- 1. 折扣配置表
-- ============================================================================
INSERT INTO group_buy_discount (id, discount_id, discount_name, discount_desc, discount_type, market_plan, market_expr, tag_id, create_time, update_time)
VALUES 
(1, 'DIS_ZJ_001', '直减20元', '原价100直减20', 1, 'ZJ', '20', NULL, NOW(), NOW()),
(2, 'DIS_MJ_001', '满减优惠', '满100减30', 2, 'MJ', '100,30', NULL, NOW(), NOW()),
(3, 'DIS_N_001', 'N元购', '29.9元购买', 3, 'N', '29.9', NULL, NOW(), NOW()),
(4, 'DIS_ZK_001', '折扣优惠', '5折优惠', 4, 'ZK', '0.5', NULL, NOW(), NOW());

-- ============================================================================
-- 2. 商品表
-- ============================================================================
INSERT INTO sku (id, source, channel, goods_id, goods_name, original_price, create_time, update_time)
VALUES 
(1, 's01', 'c01', '9890001', '手写MyBatis：渐进式源码实践（全彩）', 100.00, NOW(), NOW()),
(2, 's01', 'c01', '9890002', 'Spring Boot 实战', 200.00, NOW(), NOW()),
(3, 's01', 'c02', '9890003', '分布式系统设计', 150.00, NOW(), NOW());

-- ============================================================================
-- 3. 渠道商品活动关联表
-- ============================================================================
INSERT INTO sc_sku_activity (id, source, channel, activity_id, goods_id, create_time, update_time)
VALUES 
(1, 's01', 'c01', 100123, '9890001', NOW(), NOW()),
(2, 's01', 'c01', 100124, '9890002', NOW(), NOW());

-- ============================================================================
-- 4. 拼团活动表
-- ============================================================================
INSERT INTO group_buy_activity (id, activity_id, activity_name, discount_id, group_type, take_limit_count, target, valid_time, status, start_time, end_time, tag_id, tag_scope, create_time, update_time)
VALUES 
-- 活动1：直减20元，3人团，每人限2次，有效期15分钟
(1, 100123, '春季图书大促', 'DIS_ZJ_001', 1, 2, 3, 15, 1, DATE_SUB(NOW(), INTERVAL 1 HOUR), DATE_ADD(NOW(), INTERVAL 24 HOUR), NULL, NULL, NOW(), NOW()),
-- 活动2：满减30元，5人团，每人限1次，有效期30分钟
(2, 100124, '技术书籍优惠', 'DIS_MJ_001', 1, 1, 5, 30, 1, DATE_SUB(NOW(), INTERVAL 1 HOUR), DATE_ADD(NOW(), INTERVAL 24 HOUR), NULL, NULL, NOW(), NOW()),
-- 活动3：已结束的活动（status=0）
(3, 100125, '已结束活动', 'DIS_ZJ_001', 1, 2, 3, 15, 0, DATE_SUB(NOW(), INTERVAL 48 HOUR), DATE_SUB(NOW(), INTERVAL 24 HOUR), NULL, NULL, NOW(), NOW()),
-- 活动4：带人群标签的活动
(4, 100126, 'VIP专属活动', 'DIS_N_001', 1, 1, 2, 15, 1, DATE_SUB(NOW(), INTERVAL 1 HOUR), DATE_ADD(NOW(), INTERVAL 24 HOUR), 'TAG_VIP', '1,2', NOW(), NOW());

-- ============================================================================
-- 5. 拼团订单表（模拟已存在的订单）
-- ============================================================================
INSERT INTO group_buy_order (id, team_id, activity_id, source, channel, original_price, deduction_price, pay_price, target_count, complete_count, lock_count, status, valid_start_time, valid_end_time, notify_type, notify_url, create_time, update_time)
VALUES 
-- 已成团的订单（status=2）
(1, 'TEAM_001', 100123, 's01', 'c01', 100.00, 20.00, 80.00, 3, 3, 3, 2, DATE_SUB(NOW(), INTERVAL 30 MINUTE), DATE_ADD(NOW(), INTERVAL 15 MINUTE), 'MQ', NULL, NOW(), NOW()),
-- 未成团的订单（status=1）
(2, 'TEAM_002', 100123, 's01', 'c01', 100.00, 20.00, 80.00, 3, 1, 2, 1, DATE_SUB(NOW(), INTERVAL 10 MINUTE), DATE_ADD(NOW(), INTERVAL 15 MINUTE), 'MQ', NULL, NOW(), NOW()),
-- 已过期未支付的订单（status=0）
(3, 'TEAM_003', 100123, 's01', 'c01', 100.00, 20.00, 80.00, 3, 0, 1, 0, DATE_SUB(NOW(), INTERVAL 30 MINUTE), DATE_SUB(NOW(), INTERVAL 15 MINUTE), 'MQ', NULL, NOW(), NOW());

-- ============================================================================
-- 6. 拼团订单明细表
-- ============================================================================
INSERT INTO group_buy_order_list (id, user_id, team_id, order_id, activity_id, start_time, end_time, goods_id, source, channel, original_price, deduction_price, pay_price, status, out_trade_no, out_trade_time, biz_id, create_time, update_time)
VALUES 
-- TEAM_001 成员（已成团）
(1, 'xfg01', 'TEAM_001', 'ORDER_001', 100123, DATE_SUB(NOW(), INTERVAL 30 MINUTE), DATE_ADD(NOW(), INTERVAL 15 MINUTE), '9890001', 's01', 'c01', 100.00, 20.00, 80.00, 2, 'OUT_TRADE_001', NOW(), NULL, NOW(), NOW()),
(2, 'xfg02', 'TEAM_001', 'ORDER_002', 100123, DATE_SUB(NOW(), INTERVAL 25 MINUTE), DATE_ADD(NOW(), INTERVAL 15 MINUTE), '9890001', 's01', 'c01', 100.00, 20.00, 80.00, 2, 'OUT_TRADE_002', NOW(), NULL, NOW(), NOW()),
(3, 'xfg03', 'TEAM_001', 'ORDER_003', 100123, DATE_SUB(NOW(), INTERVAL 20 MINUTE), DATE_ADD(NOW(), INTERVAL 15 MINUTE), '9890001', 's01', 'c01', 100.00, 20.00, 80.00, 2, 'OUT_TRADE_003', NOW(), NULL, NOW(), NOW()),
-- TEAM_002 成员（未成团，仅1人已支付）
(4, 'xfg01', 'TEAM_002', 'ORDER_004', 100123, DATE_SUB(NOW(), INTERVAL 10 MINUTE), DATE_ADD(NOW(), INTERVAL 15 MINUTE), '9890001', 's01', 'c01', 100.00, 20.00, 80.00, 2, 'OUT_TRADE_004', NOW(), NULL, NOW(), NOW()),
-- TEAM_003 成员（已过期未支付）
(5, 'xfg01', 'TEAM_003', 'ORDER_005', 100123, DATE_SUB(NOW(), INTERVAL 30 MINUTE), DATE_SUB(NOW(), INTERVAL 15 MINUTE), '9890001', 's01', 'c01', 100.00, 20.00, 80.00, 0, 'OUT_TRADE_005', NULL, NULL, NOW(), NOW());

-- ============================================================================
-- 7. 回调任务表
-- ============================================================================
INSERT INTO notify_task (id, activity_id, team_id, notify_category, notify_type, notify_mq, notify_url, notify_count, notify_status, parameter_json, uuid, create_time, update_time)
VALUES 
-- 待执行的通知任务
(1, 100123, 'TEAM_001', 'team_success', 'MQ', 'topic.team_success', NULL, 0, 0, '{"teamId":"TEAM_001","outTradeNoList":["OUT_TRADE_001","OUT_TRADE_002","OUT_TRADE_003"]}', 'UUID_001', NOW(), NOW()),
-- 已失败的通知任务（需重试）
(2, 100123, 'TEAM_002', 'team_success', 'HTTP', NULL, 'http://localhost:8080/callback', 1, 2, '{"teamId":"TEAM_002"}', 'UUID_002', NOW(), NOW()),
-- 已成功的通知任务
(3, 100123, 'TEAM_000', 'team_success', 'MQ', 'topic.team_success', NULL, 2, 1, '{"teamId":"TEAM_000"}', 'UUID_003', NOW(), NOW());

-- ============================================================================
-- 8. 人群标签相关表
-- ============================================================================
INSERT INTO crowd_tags (id, tag_id, tag_name, tag_desc, statistics, create_time, update_time)
VALUES 
(1, 'TAG_VIP', 'VIP用户', 'VIP专属用户标签', 100, NOW(), NOW()),
(2, 'TAG_NEW', '新用户', '新注册用户标签', 50, NOW(), NOW());

INSERT INTO crowd_tags_detail (id, tag_id, user_id, create_time, update_time)
VALUES 
(1, 'TAG_VIP', 'xfg01', NOW(), NOW()),
(2, 'TAG_VIP', 'xfg02', NOW(), NOW()),
(3, 'TAG_NEW', 'xfg03', NOW(), NOW());

INSERT INTO crowd_tags_job (id, tag_id, batch_id, tag_type, tag_rule, stat_start_time, stat_end_time, status, create_time, update_time)
VALUES 
(1, 'TAG_VIP', 'BATCH_001', 1, '{"condition":"purchase_amount>=1000"}', DATE_SUB(NOW(), INTERVAL 7 DAY), NOW(), 1, NOW(), NOW());
