package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"strconv"
)

//CloseAverage returns the average and all of the values used to calculate the average
func CloseAverage(ic *oanda.IC) (float64, error) {
	var sum = 0.0

	for _, v := range ic.Candles {
		f, err := strconv.ParseFloat(v.Mid.C, 64)
		if err != nil {
			return 0, err
		}
		sum = sum + f
		fmt.Println(f)
	}

	average := sum / float64(len(ic.Candles))

	return average, nil
}
