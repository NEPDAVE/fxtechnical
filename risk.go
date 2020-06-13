package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"strconv"
)

//CalculateRiskCapitol takes a margin available percent you want to risk
//on a trade and converts the percentage to a number of units to risk
func CalculateRiskCapitol(marginPercent float64) (string, error) {
	//get account summary
	accountSummary, err := oanda.GetAccountSummary()

	if err != nil {
		fmt.Println(err)
	}

	//get the margin available from the account summary
	marginAvailable, err := strconv.ParseFloat(accountSummary.Account.MarginAvailable, 64)

	if err != nil {
		return "", err
	}

	//setting the riskCapitol to the marginPercent passed in
	//typically the marginPercent should be about 2% or 3%
	//of the margin available
	riskCapitolFloat := marginAvailable * marginPercent

	return fmt.Sprintf("%d", int(riskCapitolFloat)), nil
}
