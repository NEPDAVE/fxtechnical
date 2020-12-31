package fxtechnical

import (
	"fmt"
	"github.com/nepdave/oanda"
	"strconv"
)

type TradeTargets struct {
	TargetOne       string
	TargetOneStatus string
	TargetTwo       string
}

func ManageTrade(tradeID string) (string, error) {
	return "OPEN", nil
}

func GetHalfOfCurrentUnits(trade *oanda.Trade) (string, error) {
	currentUnits, err := strconv.ParseFloat(trade.CurrentUnits, 64)

	if err != nil {
		return "", err
	}

	unitsToReduceFloat := currentUnits / 2
	return fmt.Sprint(int(unitsToReduceFloat)), nil
}
