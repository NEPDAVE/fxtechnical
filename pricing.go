package main

import (
	"fmt"
	"github.com/nepdave/oanda"
	"log"
	"os"
	"strconv"
)

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

func main() {
	count := os.Args
	candles := Candles("EUR_USD", count[1], "D")
	bid, _ := BidAsk("EUR_USD")
	fmt.Println("BBBBBBIIIIIIDDDDD:")
	fmt.Println(bid)
	closeAverage := CloseAverage(candles, count[1])
	fmt.Println("**********")
	fmt.Println("AVERAGE:")
	fmt.Printf("%6.6f\n", closeAverage)
	//fmt.Println(closeAverage.Candles[0])
	fmt.Println(candles.Candles[0])
	fmt.Println("**********")
}
