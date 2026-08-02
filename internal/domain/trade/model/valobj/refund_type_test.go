package valobj

import (
	"testing"

	"group-buy-market/internal/types/enums"
)

func TestGetRefundStrategy(t *testing.T) {
	cases := []struct {
		team  enums.GroupBuyOrderStatus
		order TradeOrderStatus
		code  string
	}{
		{enums.GroupBuyProgress, TradeOrderCreate, "unpaid_unlock"},
		{enums.GroupBuyProgress, TradeOrderComplete, "paid_unformed"},
		{enums.GroupBuyComplete, TradeOrderComplete, "paid_formed"},
		{enums.GroupBuyCompleteFail, TradeOrderComplete, "paid_formed"},
	}
	for _, c := range cases {
		rt, err := GetRefundStrategy(c.team, c.order)
		if err != nil {
			t.Fatal(err)
		}
		if rt.Code != c.code {
			t.Fatalf("want %s got %s", c.code, rt.Code)
		}
	}
	_, err := GetRefundStrategy(enums.GroupBuyFail, TradeOrderCreate)
	if err == nil {
		t.Fatal("expect error")
	}
}

func TestGetRefundTypeByCode(t *testing.T) {
	rt, err := GetRefundTypeByCode("paid_unformed")
	if err != nil || rt.Strategy != "paid2RefundStrategy" {
		t.Fatalf("%v %v", rt, err)
	}
}
