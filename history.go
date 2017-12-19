package fxtechnical

import (
	//"log"
	//"strconv"
	"errors"
	oanda "github.com/nepdave/oanda"
)

func Candles(instrument string, count string, granularity string) (*oanda.Candles, error) {
	candlesByte, err := oanda.GetCandles(instrument, count, granularity)

	if err != nil {
		//FIXME not really sure what to do here becuase of the type this func
		//is expecting to return
		return oanda.Candles{}.UnmarshalCandles([]byte{}), errors.New("Candles Error")
	}
	return oanda.Candles{}.UnmarshalCandles(candlesByte), errors.New("Candles Error")
}
