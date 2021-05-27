package fxtechnical

import (
	"fmt"
	"github.com/nepdave/oanda/trade"
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

func HalfOfCurrentUnits(trade *trade.Trade) (string, error) {
	currentUnits, err := strconv.ParseFloat(trade.CurrentUnits, 64)

	if err != nil {
		return "", err
	}

	unitsToReduceFloat := currentUnits / 2
	return fmt.Sprint(int(unitsToReduceFloat)), nil
}
