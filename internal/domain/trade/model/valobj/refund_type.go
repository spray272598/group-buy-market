package valobj

import (
	"fmt"

	"group-buy-market/internal/types/enums"
)

// RefundType 退单类型
type RefundType struct {
	Code     string
	Strategy string
	Info     string
}

var (
	RefundUnpaidUnlock = RefundType{"unpaid_unlock", "unpaid2RefundStrategy", "未支付，未成团"}
	RefundPaidUnformed = RefundType{"paid_unformed", "paid2RefundStrategy", "已支付，未成团"}
	RefundPaidFormed   = RefundType{"paid_formed", "paidTeam2RefundStrategy", "已支付，已成团"}
)

func GetRefundStrategy(teamStatus enums.GroupBuyOrderStatus, orderStatus TradeOrderStatus) (RefundType, error) {
	if teamStatus == enums.GroupBuyProgress && orderStatus == TradeOrderCreate {
		return RefundUnpaidUnlock, nil
	}
	if teamStatus == enums.GroupBuyProgress && orderStatus == TradeOrderComplete {
		return RefundPaidUnformed, nil
	}
	if (teamStatus == enums.GroupBuyComplete || teamStatus == enums.GroupBuyCompleteFail) && orderStatus == TradeOrderComplete {
		return RefundPaidFormed, nil
	}
	return RefundType{}, fmt.Errorf("不支持的退款状态组合: team=%d order=%d", teamStatus, orderStatus)
}

func GetRefundTypeByCode(code string) (RefundType, error) {
	switch code {
	case "unpaid_unlock":
		return RefundUnpaidUnlock, nil
	case "paid_unformed":
		return RefundPaidUnformed, nil
	case "paid_formed":
		return RefundPaidFormed, nil
	default:
		return RefundType{}, fmt.Errorf("退单类型枚举值不存在: %s", code)
	}
}
