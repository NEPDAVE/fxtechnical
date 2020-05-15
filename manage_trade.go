package fxtechnical

type TradeTargets struct {
	TargetOne       string
	TargetOneStatus string
	TargetTwo       string
}

func ManageTrade(tradeID string) (string, error) {
	return "OPEN", nil
}
