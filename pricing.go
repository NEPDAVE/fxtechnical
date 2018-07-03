package fxtechnical

import (
	"errors"
	"fmt"
	oanda "github.com/nepdave/oanda"
	"strconv"
)

//PricesData contains unmarshaled prices data and methods to find low/high bid/ask
type PricesData struct {
	Prices           *oanda.Prices
	Heartbeat        *oanda.Heartbeat
	HighBid          float64
	HighBidLiquidity int64
	LowAsk           float64
	LowAskLiquidity  int64
	Error            error
}

//FIXME HighestBid need to test this
func (p *PricesData) HighestBid() {
	for _, val := range p.Prices.Bids {
		check, err := strconv.ParseFloat(val.Price, 64)

		if err != nil {
			p.Error = err
		}

		if check > p.HighBid {
			p.HighBid = check
			p.HighBidLiquidity = val.Liquidity
		}
	}
}

//FIXME LowestAsk need to test this
func (p *PricesData) LowestAsk() {
	firstAsk, err := strconv.ParseFloat(p.Prices.Asks[0].Price, 64)

	if err != nil {
		p.Error = err
	}

	//setting these to ensure p.LowAsk always contains a valid price
	p.LowAsk = firstAsk

	for _, val := range p.Prices.Asks {
		check, err := strconv.ParseFloat(val.Price, 64)

		if err != nil {
			p.Error = err
		}

		if check < p.LowAsk {
			p.LowAsk = check
			p.LowAskLiquidity = val.Liquidity
		}
	}
}

//StreamBidAsk capturing panic raised by Unmarshaling
func StreamBidAsk(instrument string, out chan PricesData) {
	defer func() {
		if err := recover(); err != nil {
			out <- PricesData{Error: errors.New("error unmarshaling json")}
		}
	}()

	streamResultChan := make(chan oanda.StreamResult)
	go oanda.StreamPricing(instrument, streamResultChan)

	//ranging over values coming into channel
	for streamResult := range streamResultChan {
		if streamResult.Error != nil {
			out <- PricesData{Error: streamResult.Error}
		}

		//FIXME approximating length of returned byte slice. need more precision
		priceByte := streamResult.PriceByte
		if len(priceByte) > 100 {
			prices := oanda.Prices{}.UnmarshalPrices(priceByte)
			out <- PricesData{Prices: prices}
		} else {
			heartbeat := oanda.Heartbeat{}.UnmarshalHeartbeat(priceByte)
			out <- PricesData{Heartbeat: heartbeat}
		}
	}
}

//FIXME think about if you need this func to return the lowest ask and highest bid
//FIXME think about having this func return float64 instead of string so you
//can immediatly do math with the return values
// func BidAsk(instrument string) (string, string) {
// 	//capturing panic raised by Unmarshaling
// 	defer func() {
// 		if err := recover(); err != nil {
// 			fmt.Println("StreamBidAsk panicked")
// 		}
// 	}()
//
// 	pricingByte, err := oanda.GetPricing(instrument)
//
// 	if err != nil {
// 		//FIXME think this through... if the caller is trying to do a type
// 		//conversion then this will err out if you return a word
// 		//one possibility...
// 		return "0", "0"
// 	}
// 	pricing := oanda.Pricing{}.UnmarshalPricing(pricingByte)
// 	return pricing.Prices[0].Bids[0].Price, pricing.Prices[0].Asks[0].Price
// }

//FIXME need to look into what the "real" price is, currently just taking
//the first one at face value and using it
func BidAskMultiple(instruments ...string) map[string]string {
	//capturing panic raised by Unmarshaling
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("StreamBidAsk panicked")
		}
	}()

	instrumentsMap := make(map[string]string)
	pricingByte, err := oanda.GetPricing(instruments...)

	if err != nil {
		for _, v := range instruments {
			instrumentsMap[v] = "prices unavailable"
		}
		return instrumentsMap
	}

	pricing := oanda.Pricing{}.UnmarshalPricing(pricingByte)

	for i := range pricing.Prices {
		instrument := pricing.Prices[i].Instrument
		price := pricing.Prices[i].Asks[0].Price
		instrumentsMap[instrument] = price
	}
	return instrumentsMap
}

//Spread calcuates the spread between the LowestAsk and HighestBid
func Spread(bid string, ask string) float64 {
	bidF, _ := strconv.ParseFloat(bid, 64)
	askF, _ := strconv.ParseFloat(ask, 64)
	return askF - bidF
}
