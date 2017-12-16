package fxtechnical

import (
	"log"
	"strconv"
	oanda "github.com/nepdave/oanda"
)

//FIXME need to look into what the "real" price is, currently just taking
//the first one a face value and using it
func BidAskMultiple(instruments ...string) map[string]string {
	pricing := oanda.Pricing{}.UnmarshalPricing(oanda.GetPricing(instruments...))
	instrumentsMap := make(map[string]string)

	for i, _ := range pricing.Prices {
		instrument := pricing.Prices[i].Instrument
		price := pricing.Prices[i].Asks[0].Price
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

func Trend(instrument string, count string, granularity string) string {
	candles := Candles(instrument, count, "D")
	closeAverage := CloseAverage(candles, count)
	bid, _ := BidAsk(instrument)

	fBid, err := strconv.ParseFloat(bid, 64)
	if err != nil {
		log.Fatal(err)
	}

	if fBid > closeAverage {
		return "up"
	} else if fBid < closeAverage {
		return "down"
	} else if fBid == closeAverage {
		return "equal"
	} else {
		return "wtf"
	}

}

//FIXME some of the formatting code in here and shit should be looked at again
/*
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
*/
