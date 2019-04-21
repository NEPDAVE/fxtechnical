package fxtechnical

import (
	oanda "github.com/nepdave/oanda"
	"fmt"
)

//Candles returns a *oanda.Candles, used by CloseAverage
func Candles(instrument string, count string, granularity string) (*oanda.Candles, error) {
	candlesByte, err := oanda.GetCandles(instrument, count, granularity)

	if err != nil {
		fmt.Println(err)
	}

	return oanda.Candles{}.UnmarshalCandles(candlesByte), err
}
