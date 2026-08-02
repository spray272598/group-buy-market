package po

import (
	"time"

	"github.com/shopspring/decimal"
)

// GroupBuyActivity 拼团活动表
type GroupBuyActivity struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	ActivityID     int64     `gorm:"column:activity_id"`
	ActivityName   string    `gorm:"column:activity_name"`
	DiscountID     string    `gorm:"column:discount_id"`
	GroupType      int       `gorm:"column:group_type"`
	TakeLimitCount int       `gorm:"column:take_limit_count"`
	Target         int       `gorm:"column:target"`
	ValidTime      int       `gorm:"column:valid_time"`
	Status         int       `gorm:"column:status"`
	StartTime      time.Time `gorm:"column:start_time"`
	EndTime        time.Time `gorm:"column:end_time"`
	TagID          string    `gorm:"column:tag_id"`
	TagScope       string    `gorm:"column:tag_scope"`
	CreateTime     time.Time `gorm:"column:create_time"`
	UpdateTime     time.Time `gorm:"column:update_time"`
}

func (GroupBuyActivity) TableName() string { return "group_buy_activity" }

// GroupBuyDiscount 折扣配置表
type GroupBuyDiscount struct {
	ID           uint64    `gorm:"column:id;primaryKey"`
	DiscountID   string    `gorm:"column:discount_id"`
	DiscountName string    `gorm:"column:discount_name"`
	DiscountDesc string    `gorm:"column:discount_desc"`
	DiscountType int       `gorm:"column:discount_type"`
	MarketPlan   string    `gorm:"column:market_plan"`
	MarketExpr   string    `gorm:"column:market_expr"`
	TagID        string    `gorm:"column:tag_id"`
	CreateTime   time.Time `gorm:"column:create_time"`
	UpdateTime   time.Time `gorm:"column:update_time"`
}

func (GroupBuyDiscount) TableName() string { return "group_buy_discount" }

// Sku 商品表
type Sku struct {
	ID            uint64          `gorm:"column:id;primaryKey"`
	Source        string          `gorm:"column:source"`
	Channel       string          `gorm:"column:channel"`
	GoodsID       string          `gorm:"column:goods_id"`
	GoodsName     string          `gorm:"column:goods_name"`
	OriginalPrice decimal.Decimal `gorm:"column:original_price;type:decimal(10,2)"`
	CreateTime    time.Time       `gorm:"column:create_time"`
	UpdateTime    time.Time       `gorm:"column:update_time"`
}

func (Sku) TableName() string { return "sku" }

// SCSkuActivity 渠道商品活动关联
type SCSkuActivity struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	Source     string    `gorm:"column:source"`
	Channel    string    `gorm:"column:channel"`
	ActivityID int64     `gorm:"column:activity_id"`
	GoodsID    string    `gorm:"column:goods_id"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

func (SCSkuActivity) TableName() string { return "sc_sku_activity" }

// GroupBuyOrder 拼团组队单
type GroupBuyOrder struct {
	ID             uint64          `gorm:"column:id;primaryKey"`
	TeamID         string          `gorm:"column:team_id"`
	ActivityID     int64           `gorm:"column:activity_id"`
	Source         string          `gorm:"column:source"`
	Channel        string          `gorm:"column:channel"`
	OriginalPrice  decimal.Decimal `gorm:"column:original_price;type:decimal(8,2)"`
	DeductionPrice decimal.Decimal `gorm:"column:deduction_price;type:decimal(8,2)"`
	PayPrice       decimal.Decimal `gorm:"column:pay_price;type:decimal(8,2)"`
	TargetCount    int             `gorm:"column:target_count"`
	CompleteCount  int             `gorm:"column:complete_count"`
	LockCount      int             `gorm:"column:lock_count"`
	Status         int             `gorm:"column:status"`
	ValidStartTime time.Time       `gorm:"column:valid_start_time"`
	ValidEndTime   time.Time       `gorm:"column:valid_end_time"`
	NotifyType     string          `gorm:"column:notify_type"`
	NotifyURL      *string         `gorm:"column:notify_url"`
	CreateTime     time.Time       `gorm:"column:create_time"`
	UpdateTime     time.Time       `gorm:"column:update_time"`
}

func (GroupBuyOrder) TableName() string { return "group_buy_order" }

// GroupBuyOrderList 拼团订单明细
type GroupBuyOrderList struct {
	ID             uint64          `gorm:"column:id;primaryKey"`
	UserID         string          `gorm:"column:user_id"`
	TeamID         string          `gorm:"column:team_id"`
	OrderID        string          `gorm:"column:order_id"`
	ActivityID     int64           `gorm:"column:activity_id"`
	StartTime      time.Time       `gorm:"column:start_time"`
	EndTime        time.Time       `gorm:"column:end_time"`
	GoodsID        string          `gorm:"column:goods_id"`
	Source         string          `gorm:"column:source"`
	Channel        string          `gorm:"column:channel"`
	OriginalPrice  decimal.Decimal `gorm:"column:original_price;type:decimal(8,2)"`
	DeductionPrice decimal.Decimal `gorm:"column:deduction_price;type:decimal(8,2)"`
	PayPrice       decimal.Decimal `gorm:"column:pay_price;type:decimal(8,2)"`
	Status         int             `gorm:"column:status"`
	OutTradeNo     string          `gorm:"column:out_trade_no"`
	OutTradeTime   *time.Time      `gorm:"column:out_trade_time"`
	BizID          string          `gorm:"column:biz_id"`
	CreateTime     time.Time       `gorm:"column:create_time"`
	UpdateTime     time.Time       `gorm:"column:update_time"`
}

func (GroupBuyOrderList) TableName() string { return "group_buy_order_list" }

// NotifyTask 回调任务
type NotifyTask struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	ActivityID     int64     `gorm:"column:activity_id"`
	TeamID         string    `gorm:"column:team_id"`
	NotifyCategory *string   `gorm:"column:notify_category"`
	NotifyType     string    `gorm:"column:notify_type"`
	NotifyMQ       *string   `gorm:"column:notify_mq"`
	NotifyURL      *string   `gorm:"column:notify_url"`
	NotifyCount    int       `gorm:"column:notify_count"`
	NotifyStatus   int       `gorm:"column:notify_status"`
	ParameterJSON  string    `gorm:"column:parameter_json"`
	UUID           string    `gorm:"column:uuid"`
	CreateTime     time.Time `gorm:"column:create_time"`
	UpdateTime     time.Time `gorm:"column:update_time"`
}

func (NotifyTask) TableName() string { return "notify_task" }

// CrowdTags 人群标签
type CrowdTags struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	TagID      string    `gorm:"column:tag_id"`
	TagName    string    `gorm:"column:tag_name"`
	TagDesc    string    `gorm:"column:tag_desc"`
	Statistics int       `gorm:"column:statistics"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

func (CrowdTags) TableName() string { return "crowd_tags" }

// CrowdTagsDetail 人群明细
type CrowdTagsDetail struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	TagID      string    `gorm:"column:tag_id"`
	UserID     string    `gorm:"column:user_id"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

func (CrowdTagsDetail) TableName() string { return "crowd_tags_detail" }

// CrowdTagsJob 人群标签任务
type CrowdTagsJob struct {
	ID            uint64    `gorm:"column:id;primaryKey"`
	TagID         string    `gorm:"column:tag_id"`
	BatchID       string    `gorm:"column:batch_id"`
	TagType       int       `gorm:"column:tag_type"`
	TagRule       string    `gorm:"column:tag_rule"`
	StatStartTime time.Time `gorm:"column:stat_start_time"`
	StatEndTime   time.Time `gorm:"column:stat_end_time"`
	Status        int       `gorm:"column:status"`
	CreateTime    time.Time `gorm:"column:create_time"`
	UpdateTime    time.Time `gorm:"column:update_time"`
}

func (CrowdTagsJob) TableName() string { return "crowd_tags_job" }
