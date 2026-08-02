package valobj

import "testing"

func TestVisibleEnable(t *testing.T) {
	v := &GroupBuyActivityDiscountVO{TagScope: ""}
	if !v.IsVisible() || !v.IsEnable() {
		t.Fatal("empty scope allow")
	}
	v.TagScope = "1"
	if v.IsVisible() {
		t.Fatal("1 refuse visible")
	}
	if !v.IsEnable() {
		t.Fatal("only visible restricted")
	}
	v.TagScope = "1,2"
	if v.IsVisible() || v.IsEnable() {
		t.Fatal("both refuse")
	}
	v.TagScope = "2"
	if !v.IsVisible() || v.IsEnable() {
		t.Fatal("only enable refuse")
	}
}
