package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	traderepo "group-buy-market/internal/domain/trade/adapter/repository"
	"group-buy-market/internal/domain/trade/model/aggregate"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/infrastructure/dao/po"
	"group-buy-market/internal/infrastructure/dcc"
	"group-buy-market/internal/infrastructure/metrics"
	redisx "group-buy-market/internal/infrastructure/redis"
	"group-buy-market/internal/types/common"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
)

var _ traderepo.ITradeRepository = (*TradeRepository)(nil)

// TradeRepository 交易仓储实现
type TradeRepository struct {
	db               *gorm.DB
	redis            *redisx.Service
	dcc              *dcc.Service
	topicTeamSuccess string
	topicTeamRefund  string
}

func NewTradeRepository(db *gorm.DB, rdb *redisx.Service, dccSvc *dcc.Service, topicSuccess, topicRefund string) *TradeRepository {
	return &TradeRepository{
		db:               db,
		redis:            rdb,
		dcc:              dccSvc,
		topicTeamSuccess: topicSuccess,
		topicTeamRefund:  topicRefund,
	}
}

func (r *TradeRepository) QueryMarketPayOrderEntityByOutTradeNo(ctx context.Context, userID, outTradeNo string) (*entity.MarketPayOrderEntity, error) {
	var row po.GroupBuyOrderList
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND out_trade_no = ?", userID, outTradeNo).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity.MarketPayOrderEntity{
		TeamID:           row.TeamID,
		OrderID:          row.OrderID,
		OriginalPrice:    row.OriginalPrice,
		DeductionPrice:   row.DeductionPrice,
		PayPrice:         row.PayPrice,
		TradeOrderStatus: valobj.TradeOrderStatus(row.Status),
	}, nil
}

func (r *TradeRepository) LockMarketPayOrder(ctx context.Context, agg *aggregate.GroupBuyOrderAggregate) (*entity.MarketPayOrderEntity, error) {
	user := agg.UserEntity
	act := agg.PayActivityEntity
	disc := agg.PayDiscountEntity
	notify := disc.NotifyConfig
	userTake := agg.UserTakeOrderCount

	var result *entity.MarketPayOrderEntity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		teamID := act.TeamID
		if teamID == "" {
			teamID = randomNumeric(8)
			now := time.Now()
			end := now.Add(time.Duration(act.ValidTime) * time.Minute)
			var notifyURL *string
			if notify != nil && notify.NotifyUrl != "" {
				u := notify.NotifyUrl
				notifyURL = &u
			}
			notifyType := string(valobj.NotifyHTTP)
			if notify != nil {
				notifyType = string(notify.NotifyType)
			}
			order := po.GroupBuyOrder{
				TeamID:         teamID,
				ActivityID:     act.ActivityID,
				Source:         disc.Source,
				Channel:        disc.Channel,
				OriginalPrice:  disc.OriginalPrice,
				DeductionPrice: disc.DeductionPrice,
				PayPrice:       disc.PayPrice,
				TargetCount:    act.TargetCount,
				CompleteCount:  0,
				LockCount:      1,
				Status:         0,
				ValidStartTime: now,
				ValidEndTime:   end,
				NotifyType:     notifyType,
				NotifyURL:      notifyURL,
				CreateTime:     now,
				UpdateTime:     now,
			}
			if err := tx.Create(&order).Error; err != nil {
				return err
			}
		} else {
			res := tx.Model(&po.GroupBuyOrder{}).
				Where("team_id = ? AND lock_count < target_count", teamID).
				Updates(map[string]any{
					"lock_count":  gorm.Expr("lock_count + 1"),
					"update_time": time.Now(),
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return exception.NewAppException(enums.E0005)
			}
		}

		orderID := randomNumeric(12)
		bizID := fmt.Sprintf("%d_%s_%d", act.ActivityID, user.UserID, userTake+1)
		now := time.Now()
		list := po.GroupBuyOrderList{
			UserID:         user.UserID,
			TeamID:         teamID,
			OrderID:        orderID,
			ActivityID:     act.ActivityID,
			StartTime:      act.StartTime,
			EndTime:        act.EndTime,
			GoodsID:        disc.GoodsID,
			Source:         disc.Source,
			Channel:        disc.Channel,
			OriginalPrice:  disc.OriginalPrice,
			DeductionPrice: disc.DeductionPrice,
			PayPrice:       disc.PayPrice,
			Status:         int(valobj.TradeOrderCreate),
			OutTradeNo:     disc.OutTradeNo,
			BizID:          bizID,
			CreateTime:     now,
			UpdateTime:     now,
		}
		if err := tx.Create(&list).Error; err != nil {
			if isDuplicate(err) {
				return exception.NewAppException(enums.INDEX_EXCEPTION)
			}
			return err
		}
		result = &entity.MarketPayOrderEntity{
			OrderID:          orderID,
			OriginalPrice:    disc.OriginalPrice,
			DeductionPrice:   disc.DeductionPrice,
			PayPrice:         disc.PayPrice,
			TradeOrderStatus: valobj.TradeOrderCreate,
			TeamID:           teamID,
		}
		return nil
	})
	return result, err
}

func (r *TradeRepository) QueryGroupBuyProgress(ctx context.Context, teamID string) (*valobj.GroupBuyProgressVO, error) {
	var row po.GroupBuyOrder
	err := r.db.WithContext(ctx).
		Select("target_count, complete_count, lock_count").
		Where("team_id = ?", teamID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &valobj.GroupBuyProgressVO{
		TargetCount:   row.TargetCount,
		CompleteCount: row.CompleteCount,
		LockCount:     row.LockCount,
	}, nil
}

func (r *TradeRepository) QueryGroupBuyActivityEntityByActivityID(ctx context.Context, activityID int64) (*entity.GroupBuyActivityEntity, error) {
	var act po.GroupBuyActivity
	err := r.db.WithContext(ctx).Where("activity_id = ?", activityID).First(&act).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity.GroupBuyActivityEntity{
		ActivityID:     act.ActivityID,
		ActivityName:   act.ActivityName,
		DiscountID:     act.DiscountID,
		GroupType:      act.GroupType,
		TakeLimitCount: act.TakeLimitCount,
		Target:         act.Target,
		ValidTime:      act.ValidTime,
		Status:         enums.ActivityStatus(act.Status),
		StartTime:      act.StartTime,
		EndTime:        act.EndTime,
		TagID:          act.TagID,
		TagScope:       act.TagScope,
	}, nil
}

func (r *TradeRepository) QueryOrderCountByActivityID(ctx context.Context, activityID int64, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&po.GroupBuyOrderList{}).
		Where("user_id = ? AND activity_id = ?", userID, activityID).
		Count(&count).Error
	return int(count), err
}

func (r *TradeRepository) QueryGroupBuyTeamByTeamID(ctx context.Context, teamID string) (*entity.GroupBuyTeamEntity, error) {
	var row po.GroupBuyOrder
	err := r.db.WithContext(ctx).Where("team_id = ?", teamID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	url := ""
	if row.NotifyURL != nil {
		url = *row.NotifyURL
	}
	return &entity.GroupBuyTeamEntity{
		TeamID:         row.TeamID,
		ActivityID:     row.ActivityID,
		TargetCount:    row.TargetCount,
		CompleteCount:  row.CompleteCount,
		LockCount:      row.LockCount,
		Status:         enums.GroupBuyOrderStatus(row.Status),
		ValidStartTime: row.ValidStartTime,
		ValidEndTime:   row.ValidEndTime,
		NotifyConfig: &valobj.NotifyConfigVO{
			NotifyType: valobj.NotifyType(row.NotifyType),
			NotifyUrl:  url,
			NotifyMQ:   r.topicTeamSuccess,
		},
	}, nil
}

func (r *TradeRepository) SettlementMarketPayOrder(ctx context.Context, agg *aggregate.GroupBuyTeamSettlementAggregate) (*entity.NotifyTaskEntity, error) {
	user := agg.UserEntity
	team := agg.GroupBuyTeamEntity
	pay := agg.TradePaySuccessEntity
	notifyCfg := team.NotifyConfig

	var notifyEntity *entity.NotifyTaskEntity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&po.GroupBuyOrderList{}).
			Where("out_trade_no = ? AND user_id = ? AND status = 0", pay.OutTradeNo, user.UserID).
			Updates(map[string]any{
				"status":         1,
				"out_trade_time": pay.OutTradeTime,
				"update_time":    time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return exception.NewAppException(enums.UPDATE_ZERO)
		}

		res = tx.Model(&po.GroupBuyOrder{}).
			Where("team_id = ? AND complete_count < target_count", team.TeamID).
			Updates(map[string]any{
				"complete_count": gorm.Expr("complete_count + 1"),
				"update_time":    time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return exception.NewAppException(enums.UPDATE_ZERO)
		}

		// 最后一人成团
		if team.TargetCount-team.CompleteCount == 1 {
			res = tx.Model(&po.GroupBuyOrder{}).
				Where("team_id = ? AND status = 0", team.TeamID).
				Updates(map[string]any{"status": 1, "update_time": time.Now()})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return exception.NewAppException(enums.UPDATE_ZERO)
			}

			var outTradeNos []string
			if err := tx.Model(&po.GroupBuyOrderList{}).
				Where("team_id = ? AND status = 1", team.TeamID).
				Pluck("out_trade_no", &outTradeNos).Error; err != nil {
				return err
			}

			param, _ := json.Marshal(map[string]any{
				"teamId":         team.TeamID,
				"outTradeNoList": outTradeNos,
			})
			category := string(valobj.TaskTradeSettlement)
			uuid := team.TeamID + common.Underline + category + common.Underline + pay.OutTradeNo
			notifyType := string(valobj.NotifyHTTP)
			var notifyMQ, notifyURL *string
			if notifyCfg != nil {
				notifyType = string(notifyCfg.NotifyType)
				if notifyCfg.NotifyType == valobj.NotifyMQ {
					mq := notifyCfg.NotifyMQ
					if mq == "" {
						mq = r.topicTeamSuccess
					}
					notifyMQ = &mq
				}
				if notifyCfg.NotifyType == valobj.NotifyHTTP && notifyCfg.NotifyUrl != "" {
					u := notifyCfg.NotifyUrl
					notifyURL = &u
				}
			}
			task := po.NotifyTask{
				ActivityID:     team.ActivityID,
				TeamID:         team.TeamID,
				NotifyCategory: &category,
				NotifyType:     notifyType,
				NotifyMQ:       notifyMQ,
				NotifyURL:      notifyURL,
				NotifyCount:    0,
				NotifyStatus:   0,
				ParameterJSON:  string(param),
				UUID:           uuid,
				CreateTime:     time.Now(),
				UpdateTime:     time.Now(),
			}
			if err := tx.Create(&task).Error; err != nil {
				return err
			}
			notifyEntity = toNotifyEntity(&task)
		}
		return nil
	})
	return notifyEntity, err
}

func (r *TradeRepository) IsSCBlackIntercept(source, channel string) bool {
	return r.dcc.IsSCBlackIntercept(source, channel)
}

func (r *TradeRepository) QueryUnExecutedNotifyTaskList(ctx context.Context) ([]*entity.NotifyTaskEntity, error) {
	var rows []po.NotifyTask
	err := r.db.WithContext(ctx).
		Where("notify_status IN (0, 2)").
		Order("id ASC").
		Limit(50).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	list := make([]*entity.NotifyTaskEntity, 0, len(rows))
	for i := range rows {
		list = append(list, toNotifyEntity(&rows[i]))
	}
	return list, nil
}

func (r *TradeRepository) QueryUnExecutedNotifyTaskListByTeamID(ctx context.Context, teamID string) ([]*entity.NotifyTaskEntity, error) {
	var row po.NotifyTask
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND notify_status IN (0, 2)", teamID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*entity.NotifyTaskEntity{}, nil
	}
	if err != nil {
		return nil, err
	}
	return []*entity.NotifyTaskEntity{toNotifyEntity(&row)}, nil
}

func (r *TradeRepository) UpdateNotifyTaskStatusSuccess(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	res := r.db.WithContext(ctx).Model(&po.NotifyTask{}).
		Where("team_id = ? AND uuid = ?", task.TeamID, task.UUID).
		Updates(map[string]any{
			"notify_status": 1,
			"notify_count":  gorm.Expr("notify_count + 1"),
			"update_time":   time.Now(),
		})
	return int(res.RowsAffected), res.Error
}

func (r *TradeRepository) UpdateNotifyTaskStatusError(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	res := r.db.WithContext(ctx).Model(&po.NotifyTask{}).
		Where("team_id = ? AND uuid = ?", task.TeamID, task.UUID).
		Updates(map[string]any{
			"notify_status": 3,
			"notify_count":  gorm.Expr("notify_count + 1"),
			"update_time":   time.Now(),
		})
	return int(res.RowsAffected), res.Error
}

func (r *TradeRepository) UpdateNotifyTaskStatusRetry(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	res := r.db.WithContext(ctx).Model(&po.NotifyTask{}).
		Where("team_id = ? AND uuid = ?", task.TeamID, task.UUID).
		Updates(map[string]any{
			"notify_status": 2,
			"notify_count":  gorm.Expr("notify_count + 1"),
			"update_time":   time.Now(),
		})
	return int(res.RowsAffected), res.Error
}

func (r *TradeRepository) OccupyTeamStock(ctx context.Context, teamStockKey, recoveryTeamStockKey string, target, validTime int) (bool, error) {
	recoveryCount, err := r.redis.GetInt64(ctx, recoveryTeamStockKey)
	if err != nil {
		metrics.ObserveStock("occupy", "error")
		return false, err
	}
	occupy, err := r.redis.Incr(ctx, teamStockKey)
	if err != nil {
		metrics.ObserveStock("occupy", "error")
		return false, err
	}
	// rollback 撤销本次 Incr：超卖/加锁失败时若不回滚，会产生「幽灵占位」，
	// 导致实际可售库存越卖越少（TOCTOU 竞态）
	rollback := func() {
		if _, e := r.redis.Decr(ctx, teamStockKey); e != nil {
			metrics.ObserveStock("occupy", "error")
			slog.Error("库存占用回滚失败", "key", teamStockKey, "err", e)
		}
	}
	occupy = occupy + 1 // 对齐 Java：已有占用量 +1
	if occupy > int64(target)+recoveryCount {
		rollback()
		metrics.ObserveStock("occupy", "fail")
		return false, nil
	}
	lockKey := teamStockKey + common.Underline + fmt.Sprintf("%d", occupy)
	ok, err := r.redis.SetNX(ctx, lockKey, time.Duration(validTime+60)*time.Minute)
	if err != nil {
		rollback()
		metrics.ObserveStock("occupy", "error")
		return false, err
	}
	if !ok {
		rollback()
		metrics.ObserveStock("occupy", "fail")
		slog.Info("组队库存加锁失败", "lockKey", lockKey)
		return false, nil
	}
	metrics.ObserveStock("occupy", "success")
	return true, nil
}

func (r *TradeRepository) RecoveryTeamStock(ctx context.Context, recoveryTeamStockKey string, validTime int) error {
	if recoveryTeamStockKey == "" {
		return nil
	}
	_, err := r.redis.Incr(ctx, recoveryTeamStockKey)
	if err != nil {
		metrics.ObserveStock("recovery", "error")
		return err
	}
	metrics.ObserveStock("recovery", "success")
	return nil
}

// Refund2AddRecovery 对齐 Java refund2AddRecovery：orderId 维度防重锁 + recovery 自增
func (r *TradeRepository) Refund2AddRecovery(ctx context.Context, recoveryTeamStockKey, orderID string) error {
	if recoveryTeamStockKey == "" || orderID == "" {
		return nil
	}
	lockKey := "refund_lock_" + orderID
	// 30 天防重复恢复
	ok, err := r.redis.SetNX(ctx, lockKey, 30*24*time.Hour)
	if err != nil {
		metrics.ObserveStock("refund_recovery", "error")
		return err
	}
	if !ok {
		metrics.ObserveStock("refund_recovery", "skip")
		slog.Warn("订单恢复库存操作已在进行中，跳过", "orderId", orderID)
		return nil
	}
	if _, err := r.redis.Incr(ctx, recoveryTeamStockKey); err != nil {
		_ = r.redis.Del(ctx, lockKey) // 失败释放锁允许 MQ 重试
		metrics.ObserveStock("refund_recovery", "error")
		return err
	}
	metrics.ObserveStock("refund_recovery", "success")
	slog.Info("订单恢复库存成功", "orderId", orderID, "key", recoveryTeamStockKey)
	return nil
}

func (r *TradeRepository) Unpaid2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return r.doRefund(ctx, agg, 0, string(valobj.TaskTradeUnpaid2Refund), valobj.RefundUnpaidUnlock.Code, true, false)
}

func (r *TradeRepository) Paid2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return r.doRefund(ctx, agg, 1, string(valobj.TaskTradePaid2Refund), valobj.RefundPaidUnformed.Code, true, true)
}

func (r *TradeRepository) PaidTeam2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	order := agg.TradeRefundOrderEntity
	progress := agg.GroupBuyProgress
	var notifyEntity *entity.NotifyTaskEntity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&po.GroupBuyOrderList{}).
			Where("user_id = ? AND order_id = ? AND status = 1", order.UserID, order.OrderID).
			Updates(map[string]any{"status": 2, "update_time": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return exception.NewAppException(enums.UPDATE_ZERO)
		}

		// 查询当前组队完成情况
		var team po.GroupBuyOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("team_id = ?", order.TeamID).First(&team).Error; err != nil {
			return err
		}
		// complete_count + progress.CompleteCount（-1）
		// 若只剩 0 人完成则组队失败 status=2，否则完成含退单 status=3
		newComplete := team.CompleteCount + progress.CompleteCount
		newLock := team.LockCount + progress.LockCount
		newStatus := int(enums.GroupBuyCompleteFail)
		if newComplete <= 0 {
			newStatus = int(enums.GroupBuyFail)
		}
		res = tx.Model(&po.GroupBuyOrder{}).
			Where("team_id = ?", order.TeamID).
			Updates(map[string]any{
				"lock_count":     newLock,
				"complete_count": newComplete,
				"status":         newStatus,
				"update_time":    time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return exception.NewAppException(enums.UPDATE_ZERO)
		}

		task, err := r.insertRefundNotify(tx, order, string(valobj.TaskTradePaidTeam2Refund), valobj.RefundPaidFormed.Code)
		if err != nil {
			return err
		}
		notifyEntity = task
		return nil
	})
	return notifyEntity, err
}

// doRefund 通用退单：fromStatus 0=未支付 1=已支付
func (r *TradeRepository) doRefund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate, fromStatus int, category, refundType string, decLock, decComplete bool) (*entity.NotifyTaskEntity, error) {
	order := agg.TradeRefundOrderEntity
	progress := agg.GroupBuyProgress
	var notifyEntity *entity.NotifyTaskEntity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&po.GroupBuyOrderList{}).
			Where("user_id = ? AND order_id = ? AND status = ?", order.UserID, order.OrderID, fromStatus).
			Updates(map[string]any{"status": 2, "update_time": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			slog.Error("逆向流程退单，更新订单状态失败", "userId", order.UserID, "orderId", order.OrderID)
			return exception.NewAppException(enums.UPDATE_ZERO)
		}

		updates := map[string]any{"update_time": time.Now()}
		if decLock {
			updates["lock_count"] = gorm.Expr("lock_count + ?", progress.LockCount)
		}
		if decComplete {
			updates["complete_count"] = gorm.Expr("complete_count + ?", progress.CompleteCount)
		}
		res = tx.Model(&po.GroupBuyOrder{}).Where("team_id = ?", order.TeamID).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return exception.NewAppException(enums.UPDATE_ZERO)
		}

		task, err := r.insertRefundNotify(tx, order, category, refundType)
		if err != nil {
			return err
		}
		notifyEntity = task
		return nil
	})
	return notifyEntity, err
}

func (r *TradeRepository) insertRefundNotify(tx *gorm.DB, order *entity.TradeRefundOrderEntity, category, refundType string) (*entity.NotifyTaskEntity, error) {
	param, _ := json.Marshal(map[string]any{
		"type":       refundType,
		"userId":     order.UserID,
		"teamId":     order.TeamID,
		"orderId":    order.OrderID,
		"outTradeNo": order.OutTradeNo,
		"activityId": order.ActivityID,
	})
	uuid := order.TeamID + common.Underline + category + common.Underline + order.OrderID
	mq := r.topicTeamRefund
	cat := category
	task := po.NotifyTask{
		ActivityID:     order.ActivityID,
		TeamID:         order.TeamID,
		NotifyCategory: &cat,
		NotifyType:     string(valobj.NotifyMQ),
		NotifyMQ:       &mq,
		NotifyCount:    0,
		NotifyStatus:   0,
		ParameterJSON:  string(param),
		UUID:           uuid,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	if err := tx.Create(&task).Error; err != nil {
		return nil, err
	}
	return toNotifyEntity(&task), nil
}

func (r *TradeRepository) QueryTimeoutUnpaidOrderList(ctx context.Context) ([]*entity.TimeoutUnpaidOrderEntity, error) {
	var lists []po.GroupBuyOrderList
	err := r.db.WithContext(ctx).
		Where("status = 0 AND out_trade_time IS NULL AND ? > end_time", time.Now()).
		Limit(10).
		Find(&lists).Error
	if err != nil {
		return nil, err
	}
	if len(lists) == 0 {
		return []*entity.TimeoutUnpaidOrderEntity{}, nil
	}
	teamIDs := make([]string, 0, len(lists))
	for _, l := range lists {
		teamIDs = append(teamIDs, l.TeamID)
	}
	var teams []po.GroupBuyOrder
	if err := r.db.WithContext(ctx).Where("team_id IN ?", teamIDs).Find(&teams).Error; err != nil {
		return nil, err
	}
	teamMap := make(map[string]po.GroupBuyOrder, len(teams))
	for _, t := range teams {
		teamMap[t.TeamID] = t
	}
	result := make([]*entity.TimeoutUnpaidOrderEntity, 0, len(lists))
	for _, l := range lists {
		t, ok := teamMap[l.TeamID]
		if !ok {
			continue
		}
		result = append(result, &entity.TimeoutUnpaidOrderEntity{
			UserID:         l.UserID,
			TeamID:         t.TeamID,
			ActivityID:     t.ActivityID,
			TargetCount:    t.TargetCount,
			CompleteCount:  t.CompleteCount,
			LockCount:      t.LockCount,
			ValidStartTime: t.ValidStartTime,
			ValidEndTime:   t.ValidEndTime,
			OutTradeNo:     l.OutTradeNo,
			Source:         l.Source,
			Channel:        l.Channel,
		})
	}
	return result, nil
}

func (r *TradeRepository) RefundOrderExist(ctx context.Context, teamID, category, orderID string) (bool, error) {
	uuid := teamID + common.Underline + category + common.Underline + orderID
	var count int64
	err := r.db.WithContext(ctx).Model(&po.NotifyTask{}).Where("uuid = ?", uuid).Count(&count).Error
	return count > 0, err
}

func toNotifyEntity(t *po.NotifyTask) *entity.NotifyTaskEntity {
	mq, url := "", ""
	if t.NotifyMQ != nil {
		mq = *t.NotifyMQ
	}
	if t.NotifyURL != nil {
		url = *t.NotifyURL
	}
	return &entity.NotifyTaskEntity{
		TeamID:        t.TeamID,
		NotifyType:    t.NotifyType,
		NotifyMQ:      mq,
		NotifyUrl:     url,
		NotifyCount:   t.NotifyCount,
		Status:        valobj.NotifyTaskStatus(t.NotifyStatus),
		ParameterJSON: t.ParameterJSON,
		UUID:          t.UUID,
		ActivityID:    t.ActivityID,
	}
}

func randomNumeric(n int) string {
	const digits = "0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = digits[rand.Intn(10)]
	}
	return string(b)
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "Duplicate") || contains(msg, "duplicate") || contains(msg, "UNIQUE")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
