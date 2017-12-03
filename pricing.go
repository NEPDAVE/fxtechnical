package main

import (
	"fmt"
	"github.com/nepdave/oanda"
)

func bidAsk(instrument string) (string, string) {
	pricing := oanda.Pricing{}.UnmarshalPricing(oanda.GetPricing(instrument))

	return pricing.Prices[0].Bids[0].Price, pricing.Prices[0].Asks[0].Price
}

func main() {
	//blah, b := bidAsk("EUR_USD")
	//fmt.Println(blah)
	fmt.Println("start")
  oanda.GetCandles("EUR_USD", "10", "D")
}
