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
func (p PricesData) Init(instrument string, spreadType string) PricesData {
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

	//control stucture to calculate the correct spread "size", IE are we looking
	//for liquidity or price
	if spreadType == "tightestSpread" {
		//calling CalculateWideSpread also sets bid/ask fields in struct
		p.CalculateTightestSpread()
	} else if spreadType == "mostLiquidSpread" {
		//calling CalculateWideSpread also sets bid/ask fields in struct
		p.CalculateMostLiquidSpread()
	}

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

//MostLiquidBid sets the value of Bid and BidLiquidity by finding the MostLiquidBid
func (p *PricesData) MostLiquidBid() float64 {
	//setting this to ensure p.BidLiquidity and p.Bid always contain a values
	p.BidLiquidity = p.Prices.Bids[0].Liquidity
	bid, err := strconv.ParseFloat(p.Prices.Bids[0].Price, 64)

	if err != nil {
		p.Error = err
	}

	p.Bid = bid

	for _, val := range p.Prices.Bids {
		check := val.Liquidity

		if check > p.BidLiquidity {
			p.BidLiquidity = check
			bid, err := strconv.ParseFloat(val.Price, 64)

			if err != nil {
				p.Error = err
			}

			p.Bid = bid
		}
	}
	return p.Bid
}

//MostLiquidAsk sets the value of Ask and AskLiquidity by finding the MostLiquidAsk
func (p *PricesData) MostLiquidAsk() float64 {
	p.AskLiquidity = p.Prices.Asks[0].Liquidity

	ask, err := strconv.ParseFloat(p.Prices.Asks[0].Price, 64)

	if err != nil {
		p.Error = err
	}

	p.Ask = ask

	for _, val := range p.Prices.Asks {
		check := val.Liquidity

		if check > p.AskLiquidity {
			p.AskLiquidity = check
			ask, err := strconv.ParseFloat(val.Price, 64)

			if err != nil {
				p.Error = err
			}

			p.Ask = ask
		}
	}
	return p.Ask
}

//CalculateTightSpread calcuates the spread between the LowestAsk and HighestBid
func (p *PricesData) CalculateTightestSpread() {
	p.Spread = p.LowestAsk() - p.HighestBid()
}

//CalculateMostLiquidSpread calcuates the spread between the LowestAsk and HighestBid
func (p *PricesData) CalculateMostLiquidSpread() {
	p.Spread = p.MostLiquidAsk() - p.MostLiquidBid()
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
			//FIXME this should not happen here ...
			pricesData.CalculateTightestSpread()
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
