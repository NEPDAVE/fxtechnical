package fxtechnical

import (
	//"log"
	"strconv"
	oanda "github.com/nepdave/oanda"
)

//FIXME think about having this funct return float64 instead of string so you
//can immediatly do math with the return values
func BidAsk(instrument string) (string, string) {
	pricingByte, err := oanda.GetPricing(instrument)

	if err != nil {
		//FIXME think this through... if the caller is trying to do a type
		//conversion then this will err out
		return "prices unavailable", "prices unavailable"
	}
	pricing := oanda.Pricing{}.UnmarshalPricing(pricingByte)
	return pricing.Prices[0].Bids[0].Price, pricing.Prices[0].Asks[0].Price
}

//FIXME need to look into what the "real" price is, currently just taking
//the first one at face value and using it
func BidAskMultiple(instruments ...string) map[string]string {
	instrumentsMap := make(map[string]string)
	pricingByte, err := oanda.GetPricing(instruments...)

	if err != nil {
		for _, v := range instruments {
			instrumentsMap[v] = "prices unavailable"
		}
		return instrumentsMap
	}

	pricing := oanda.Pricing{}.UnmarshalPricing(pricingByte)

	for i, _ := range pricing.Prices {
		instrument := pricing.Prices[i].Instrument
		price := pricing.Prices[i].Asks[0].Price
		instrumentsMap[instrument] = price
	}
	return instrumentsMap
}

func Spread(bid string, ask string) float64 {
	bidF, _ := strconv.ParseFloat(bid, 64)
	askF, _ := strconv.ParseFloat(ask, 64)
	return askF - bidF
}
