-- 拼团营销 group_buy_market 初始化脚本（对齐 Java 项目表结构）
CREATE DATABASE IF NOT EXISTS `group_buy_market` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
USE `group_buy_market`;

DROP TABLE IF EXISTS `crowd_tags`;
CREATE TABLE `crowd_tags` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `tag_id` varchar(32) NOT NULL COMMENT '人群ID',
  `tag_name` varchar(64) NOT NULL COMMENT '人群名称',
  `tag_desc` varchar(256) NOT NULL COMMENT '人群描述',
  `statistics` int NOT NULL COMMENT '人群标签统计量',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_tag_id` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='人群标签';

INSERT INTO `crowd_tags` (`tag_id`, `tag_name`, `tag_desc`, `statistics`)
VALUES ('RQ_KJHKL98UU78H66554GFDV', '潜在消费用户', '潜在消费用户', 11);

DROP TABLE IF EXISTS `crowd_tags_detail`;
CREATE TABLE `crowd_tags_detail` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `tag_id` varchar(32) NOT NULL,
  `user_id` varchar(16) NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_tag_user` (`tag_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='人群标签明细';

INSERT INTO `crowd_tags_detail` (`tag_id`, `user_id`) VALUES
('RQ_KJHKL98UU78H66554GFDV','xiaofuge'),
('RQ_KJHKL98UU78H66554GFDV','liergou'),
('RQ_KJHKL98UU78H66554GFDV','xfg01'),
('RQ_KJHKL98UU78H66554GFDV','xfg02'),
('RQ_KJHKL98UU78H66554GFDV','xfg03');

DROP TABLE IF EXISTS `crowd_tags_job`;
CREATE TABLE `crowd_tags_job` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `tag_id` varchar(32) NOT NULL,
  `batch_id` varchar(8) NOT NULL,
  `tag_type` tinyint(1) NOT NULL DEFAULT '1',
  `tag_rule` varchar(8) NOT NULL,
  `stat_start_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `stat_end_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` tinyint(1) NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_batch_id` (`batch_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='人群标签任务';

DROP TABLE IF EXISTS `group_buy_activity`;
CREATE TABLE `group_buy_activity` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `activity_id` bigint NOT NULL,
  `activity_name` varchar(128) NOT NULL,
  `discount_id` varchar(8) NOT NULL,
  `group_type` tinyint(1) NOT NULL DEFAULT '0',
  `take_limit_count` int NOT NULL DEFAULT '1',
  `target` int NOT NULL DEFAULT '1',
  `valid_time` int NOT NULL DEFAULT '15',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0创建、1生效、2过期、3废弃',
  `start_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `end_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `tag_id` varchar(64) DEFAULT NULL,
  `tag_scope` varchar(4) DEFAULT NULL COMMENT '1可见限制、2参与限制',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_activity_id` (`activity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='拼团活动';

INSERT INTO `group_buy_activity`
(`activity_id`,`activity_name`,`discount_id`,`group_type`,`take_limit_count`,`target`,`valid_time`,`status`,`start_time`,`end_time`,`tag_id`,`tag_scope`)
VALUES
(100123,'测试活动','25120207',0,3,3,15,1,'2024-12-07 10:19:40','2029-12-07 10:19:40','RQ_KJHKL98UU78H66554GFDV','1');

DROP TABLE IF EXISTS `group_buy_discount`;
CREATE TABLE `group_buy_discount` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `discount_id` varchar(8) NOT NULL,
  `discount_name` varchar(64) NOT NULL,
  `discount_desc` varchar(256) NOT NULL,
  `discount_type` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0:base、1:tag',
  `market_plan` varchar(4) NOT NULL DEFAULT 'ZJ' COMMENT 'ZJ直减 MJ满减 N元购 ZK折扣',
  `market_expr` varchar(32) NOT NULL,
  `tag_id` varchar(8) DEFAULT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_discount_id` (`discount_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `group_buy_discount`
(`discount_id`,`discount_name`,`discount_desc`,`discount_type`,`market_plan`,`market_expr`,`tag_id`)
VALUES
('25120207','测试优惠','直减20元',0,'ZJ','20',NULL);

DROP TABLE IF EXISTS `group_buy_order`;
CREATE TABLE `group_buy_order` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `team_id` varchar(8) NOT NULL,
  `activity_id` bigint NOT NULL,
  `source` varchar(8) NOT NULL,
  `channel` varchar(8) NOT NULL,
  `original_price` decimal(8,2) NOT NULL,
  `deduction_price` decimal(8,2) NOT NULL,
  `pay_price` decimal(8,2) NOT NULL,
  `target_count` int NOT NULL,
  `complete_count` int NOT NULL,
  `lock_count` int NOT NULL,
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0拼单中 1完成 2失败 3完成含退单',
  `valid_start_time` datetime NOT NULL,
  `valid_end_time` datetime NOT NULL,
  `notify_type` varchar(8) NOT NULL DEFAULT 'HTTP',
  `notify_url` varchar(512) DEFAULT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_team_id` (`team_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `group_buy_order_list`;
CREATE TABLE `group_buy_order_list` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user_id` varchar(64) NOT NULL,
  `team_id` varchar(8) NOT NULL,
  `order_id` varchar(12) NOT NULL,
  `activity_id` bigint NOT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `goods_id` varchar(16) NOT NULL,
  `source` varchar(8) NOT NULL,
  `channel` varchar(8) NOT NULL,
  `original_price` decimal(8,2) NOT NULL,
  `deduction_price` decimal(8,2) NOT NULL,
  `pay_price` decimal(8,2) NOT NULL,
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0锁定 1完成 2退单',
  `out_trade_no` varchar(12) NOT NULL,
  `out_trade_time` datetime DEFAULT NULL,
  `biz_id` varchar(64) NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_order_id` (`order_id`),
  UNIQUE KEY `uq_biz_id` (`biz_id`),
  KEY `idx_user_id_activity_id` (`user_id`,`activity_id`),
  KEY `idx_out_trade_no` (`user_id`,`out_trade_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `notify_task`;
CREATE TABLE `notify_task` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `activity_id` bigint NOT NULL,
  `team_id` varchar(8) NOT NULL,
  `notify_category` varchar(64) DEFAULT NULL,
  `notify_type` varchar(8) NOT NULL DEFAULT 'HTTP',
  `notify_mq` varchar(32) DEFAULT NULL,
  `notify_url` varchar(128) DEFAULT NULL,
  `notify_count` int NOT NULL,
  `notify_status` tinyint(1) NOT NULL COMMENT '0初始 1完成 2重试 3失败',
  `parameter_json` varchar(512) NOT NULL,
  `uuid` varchar(128) NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uuid` (`uuid`),
  KEY `idx_status` (`notify_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 人群批次任务种子（表在上方 CREATE）
INSERT INTO `crowd_tags_job` (`tag_id`, `batch_id`, `tag_type`, `tag_rule`, `stat_start_time`, `stat_end_time`, `status`)
VALUES ('RQ_KJHKL98UU78H66554GFDV', '10001', 0, '100', NOW(), NOW(), 0);

DROP TABLE IF EXISTS `sc_sku_activity`;
CREATE TABLE `sc_sku_activity` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `source` varchar(8) NOT NULL,
  `channel` varchar(8) NOT NULL,
  `activity_id` bigint NOT NULL,
  `goods_id` varchar(16) NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_sc_goodsid` (`source`,`channel`,`goods_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='渠道商品活动配置关联表';

INSERT INTO `sc_sku_activity` (`source`,`channel`,`activity_id`,`goods_id`)
VALUES ('s01','c01',100123,'9890001');

DROP TABLE IF EXISTS `sku`;
CREATE TABLE `sku` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `source` varchar(8) NOT NULL,
  `channel` varchar(8) NOT NULL,
  `goods_id` varchar(16) NOT NULL,
  `goods_name` varchar(128) NOT NULL,
  `original_price` decimal(10,2) NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_goods_id` (`goods_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品信息';

INSERT INTO `sku` (`source`,`channel`,`goods_id`,`goods_name`,`original_price`)
VALUES ('s01','c01','9890001','《手写MyBatis：渐进式源码实践》',100.00);
