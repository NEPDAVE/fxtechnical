package main

import (
	"fmt"
	"github.com/nepdave/oanda"
	"log"
	"os"
	"strconv"
)

//FIXME you can add a second function that can mostly using the existing
//oanda endpoint function and a second bidask(multiple) or maybe you could one
//becasue your getting back a list of prices and we do not want to assume the
//prices are going to be in order, we will want to loop through the struct to pull
//out the prices, then we can use them for things like the index page or doing
//statistical arbitrage, so that's cool...

//TODO write a BidAsk multiple function and a GetPricing function that can
//handle the variadic input

//TODO figure out if you need to do anykind of analysis for the multiple prices
//that are returned.... that is annoying


//We need a list of all the instruments to reference them later...
func BidAskMultiple(intruments ...string) map[string]string {
	pricing := oanda.Pricing{}.UnmarshalPricing(oanda.GetMultiplePricing(instruments))

  instrumentsMap := map[string]string

	for i, v range := pricing {
		instrumentsMap[]
		instrument := pricing.Prices[i].Instrument
		price := pricing.Prices[i].Asks[0].Price
		//FIXME double check this is right way to add k/v for go
		instrumentsMap[instrument] = price
	}

	return instrumentsMap

}

func BidAsk(instrument string) (string, string) {
	pricing := oanda.Pricing{}.UnmarshalPricing(oanda.GetPricing(instrument))

	return pricing.Prices[0].Bids[0].Price, pricing.Prices[0].Asks[0].Price
}

func Candles(instrument string, count string, granularity string) *oanda.Candles {
	return oanda.Candles{}.UnmarshalCandles(oanda.GetCandles(instrument, count,
		granularity))
}


//FIXME this should have a unit test!
func CloseAverage(candles *oanda.Candles, count string) float64 {
	sum := 0.0
	i, err := strconv.Atoi(count)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range candles.Candles {

		f, err := strconv.ParseFloat(v.Mid.Close, 64)
		if err != nil {
			log.Fatal(err)
		}
		sum = sum + f

	}
	return sum / float64(i)
}

func trend(instrument string, count string, granularity string) string {
	candles := Candles(instrument, count, "D")
	closeAverage := CloseAverage(candles, count)
	bid, _ := BidAsk(instrument)

	fBid, err := strconv.ParseFloat(bid, 64)
	if err != nil {
		log.Fatal(err)
	}

	if fBid > closeAverage {
		return "up"
	}else if fBid < closeAverage {
		return "down"
	}else if fBid == closeAverage {
		return "equal"
	}else {
		return "wtf"
	}

}

func main() {
	count := os.Args
	candles := Candles("EUR_USD", count[1], "D")
	bid, _ := BidAsk("EUR_USD")
	fmt.Println("**********")
	fmt.Println("BREAK:BREAK:BREAK:BREAK:BREAK:BREAK:BREAK:BREAK:BREAK:")
	fmt.Println("**********")
	closeAverage := CloseAverage(candles, count[1])
	way := trend("EUR_USD", count[1], "D")
	fmt.Println("**********")
	fmt.Println("AVERAGE:")
	fmt.Printf("%6.6f\n", closeAverage)
	fmt.Println("BID:")
	fmt.Println(bid)
	fmt.Println("Trend:")
	//FIXME calling this prints out a whole bunch of shit...
  fmt.Println(way)
	fmt.Println("**********")
}
