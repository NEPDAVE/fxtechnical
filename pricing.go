package fxtechnical

import (
	"fmt"
	"strconv"
	oanda "github.com/nepdave/oanda"
)

func StreamBidAsk(instruments string, out chan oanda.StreamResult) {
	//haha not sure if im doing this right...
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("StreamBidAsk panicked")
		}
	}()

	oandaChan := make(chan oanda.StreamResult)
	go oanda.StreamPricing("EUR_USD", oandaChan)

	//ranging over values coming into channel
	for streamResult := range oandaChan {
		if streamResult.Error != nil {
			fmt.Println(streamResult.Error)
		}
	//approximating length of returned byte slice. need more precision
	priceByte := streamResult.PriceByte
		if len(priceByte) > 100 {
			prices := oanda.Prices{}.UnmarshalPrices(priceByte)
			fmt.Println(prices)
		} else if len(priceByte) < 100 {
			heartbeat := oanda.Heartbeat{}.UnmarshalHeartbeat(priceByte)
			fmt.Println(heartbeat)
		} else {
			fmt.Println("Neither Price Nor Heartbeat...")
		}
	}

}

//FIXME think about having this func return float64 instead of string so you
//can immediatly do math with the return values
func BidAsk(instrument string) (string, string) {
	pricingByte, err := oanda.GetPricing(instrument)

	if err != nil {
		//FIXME think this through... if the caller is trying to do a type
		//conversion then this will err out if you return a word
		//one possibility...
		return "0", "0"
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
