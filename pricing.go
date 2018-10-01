package fxtechnical

import (
	"errors"
	"fmt"
	oanda "github.com/nepdave/oanda"
	"strconv"
)

/*
***************************
PricesData struct and methods
***************************
*/

//PricesData contains unmarshaled prices data and methods to populate struct fields
type PricesData struct {
	Prices       *oanda.Prices
	Heartbeat    *oanda.Heartbeat
	Bid          float64
	BidLiquidity int64
	Ask          float64
	AskLiquidity int64
	Spread       float64
	Error        error
}

//Init populates PricesData with data and returns itself
func (p PricesData) Init(instrument string) PricesData {
	//capturing panic raised by Unmarshaling
	defer func() {
		if err := recover(); err != nil {
			p.Error = errors.New("error unmarshaling json")
		}
	}()

	pricingByte, err := oanda.GetPricing(instrument)

	if err != nil {
		p.Error = err
		return p
	}

	pricing := oanda.Pricing{}.UnmarshalPricing(pricingByte)
	p.Prices = &pricing.Prices[0]

	//calling CalculateWideSpread also sets bid/ask fields in struct
	p.CalculateWideSpread()

	return p
}

//HighestBid sets the value of Bid and BidLiquidity by finding the Highest Bid
func (p *PricesData) HighestBid() float64 {
	for _, val := range p.Prices.Bids {
		check, err := strconv.ParseFloat(val.Price, 64)

		if err != nil {
			p.Error = err
		}

		if check > p.Bid {
			p.Bid = check
			p.BidLiquidity = val.Liquidity
		}
	}
	return p.Bid
}

//LowestAsk sets the value of Ask and AskLiquidity by finding the Lowest Ask
func (p *PricesData) LowestAsk() float64 {
	firstAsk, err := strconv.ParseFloat(p.Prices.Asks[0].Price, 64)

	if err != nil {
		p.Error = err
	}

	//setting these to ensure p.LowAsk always contains a valid price
	p.Ask = firstAsk
	p.AskLiquidity = p.Prices.Asks[0].Liquidity

	for _, val := range p.Prices.Asks {
		check, err := strconv.ParseFloat(val.Price, 64)

		if err != nil {
			p.Error = err
		}

		if check < p.Ask {
			p.Ask = check
			p.AskLiquidity = val.Liquidity
		}
	}
	return p.Ask
}

//LowestBid sets the value of Bid and BidLiquidity by finding the Lowest Bid
func (p *PricesData) LowestBid() float64 {
	//setting p.Bid to first value in slice to set a baseline
	firstBid, err := strconv.ParseFloat(p.Prices.Bids[0].Price, 64)

	if err != nil {
		p.Error = err
	}

	//setting these to ensure p.LowAsk always contains a valid price
	p.Bid = firstBid
	p.BidLiquidity = p.Prices.Bids[0].Liquidity

	for _, val := range p.Prices.Bids {
		check, err := strconv.ParseFloat(val.Price, 64)

		if err != nil {
			p.Error = err
		}

		if check < p.Bid {
			p.Bid = check
			p.BidLiquidity = val.Liquidity
		}
	}
	return p.Bid
}

//HighestAsk sets the value of Ask and AskLiquidity by finding the Highest Ask
func (p *PricesData) HighestAsk() float64 {
	firstAsk, err := strconv.ParseFloat(p.Prices.Asks[0].Price, 64)

	if err != nil {
		p.Error = err
	}

	//setting these to ensure p.LowAsk always contains a valid price
	p.Ask = firstAsk
	p.AskLiquidity = p.Prices.Asks[0].Liquidity

	for _, val := range p.Prices.Asks {
		check, err := strconv.ParseFloat(val.Price, 64)

		if err != nil {
			p.Error = err
		}

		if check > p.Ask {
			p.Ask = check
			p.AskLiquidity = val.Liquidity
		}
	}
	return p.Ask
}

//CalculateTightSpread calcuates the spread between the LowestAsk and HighestBid
func (p *PricesData) CalculateTightSpread() {
	p.Spread = p.LowestAsk() - p.HighestBid()
}

//CalculateWideSpread calcuates the spread between the LowestAsk and HighestBid
func (p *PricesData) CalculateWideSpread() {
	p.Spread = p.HighestAsk() - p.LowestBid()
}

/*
***************************
stand alone pricing functions
***************************
*/

//StreamBidAsk kicks off pricing stream
func StreamBidAsk(instrument string, out chan PricesData) {
	//capturing panic raised by Unmarshaling
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
			//most likely an actual price
			prices := oanda.Prices{}.UnmarshalPrices(priceByte)
			pricesData := PricesData{Prices: prices}
			pricesData.CalculateTightSpread()
			out <- pricesData
		} else {
			//most likely a heartbeat
			heartbeat := oanda.Heartbeat{}.UnmarshalHeartbeat(priceByte)
			out <- PricesData{Heartbeat: heartbeat}
		}
	}
}

//BidAsk returns the first Bid and Ask in the Prices struct
func BidAsk(instrument string) (string, string) {
	//capturing panic raised by Unmarshaling
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("StreamBidAsk panicked")
		}
	}()

	pricingByte, err := oanda.GetPricing(instrument)

	if err != nil {
		fmt.Println(err)
		return "0", "0"
	}

	pricing := oanda.Pricing{}.UnmarshalPricing(pricingByte)
	return pricing.Prices[0].Bids[0].Price, pricing.Prices[0].Asks[0].Price
}

//BidAskMultiple returns the first bid and the first ask for each intrument you
//pass it from Oanda
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
