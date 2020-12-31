package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"github.com/nepdave/xe"
	"strconv"
	"strings"
)

//CalculateRiskCapitol takes a margin available percent you want to risk
//on a trade and converts the percentage to a number of units to risk
func CalculateRiskCapitol(marginToRisk float64) (string, error) {
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

	//setting the riskCapitol to the marginToRisk passed in
	//typically the marginToRisk should be about 2% or 3%
	//of the margin available
	riskCapitolFloat := marginAvailable * marginToRisk

	//multiplying the risk capitol by 20 to account for leverage
	//currently using 20:1 leverage
	riskCapitolFloat = riskCapitolFloat * 20

	return fmt.Sprintf("%d", int(riskCapitolFloat)), nil
}

//UnitsToRisk takes a margin available percent you want to risk
//on a trade and converts the percentage to a number of units to risk
func UnitsToRisk(marginToRisk float64, marginRatio float64, instrument string) (string, error) {
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

	fmt.Printf("margin available: %v\n", marginAvailable)

	//convert home currency to base currency
	currencies := strings.Split(instrument, "_")

	exchangeRate, err := xe.GetExchangeRate(currencies[0], currencies[1])

	fmt.Printf("exchange rate: %v\n", exchangeRate.To[0].Mid)

	if err != nil {
		return "", err
	}

	unitsToRisk := ((marginAvailable * marginToRisk) * marginRatio) / exchangeRate.To[0].Mid

	return fmt.Sprintf("%d", int(unitsToRisk)), nil
}

/*
How This Tool Works
This calculation uses the following formula:

Margin Available * (margin ratio) / ({BASE}/{HOME Currency} Exchange Rate)
For example, suppose:

Home Currency: USD
Currency Pair: GBP/CHF
Margin Available: 100
Margin Ratio : 20:1
Base / Home Currency: GBP/USD = 1.584
Then,

Units = (100 * 20) / 1.584
Units = 1262
*/
