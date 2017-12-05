package main

import (
	"fmt"
  "strconv"
  "log"
	"github.com/nepdave/oanda"
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
func CloseAverage(candles *oanda.Candles) float64 {
  sum := 0.0
  counter := 0
  fmt.Println("LEN CANDLES:")
  fmt.Println(len(candles.Candles))

  for _, v := range candles.Candles {
    //FIXME doing a type assertion here. this is new to me...
    if str, ok := v.Mid["c"].(string); ok {
        /* act on str */
        f, err := strconv.ParseFloat(str, 64)
        //FIXME need to work on error handling
        if err != nil {
            log.Fatal(err)
        }
        sum = sum + f
    } else {
        /* not string */
        log.Fatal("CloseAverage type assertion error")
    }
    counter++


  }
  fmt.Println("COUNTER:")
  fmt.Println(counter)
  return sum/float64(len(candles.Candles))
}



func main() {
	//blah, b := bidAsk("EUR_USD")
	//fmt.Println(blah)
	fmt.Println("start")
	//oanda.GetCandles("EUR_USD", "1", "D")
	//close := close("EUR_USD", "2", "D")
  //fmt.Println(close)
	//fmt.Println(close.Candles[0].Mid["c"])

  candles := Candles("EUR_USD", "10", "D")
  closeAverage := CloseAverage(candles)
  fmt.Println("**********")
  fmt.Println("AVERAGE:")
  fmt.Println(closeAverage)
  fmt.Println("**********")
	fmt.Println("end")
}
