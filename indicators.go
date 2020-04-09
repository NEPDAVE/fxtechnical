package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"math"
	"strconv"
)

//CloseAverage returns the average and all of the values used to calculate the average
func CloseAverage(i *oanda.InstrumentHistory) (float64, error) {
	var sum float64 = 0.0

	for _, v := range i.Candles {
		f, err := strconv.ParseFloat(v.Mid.C, 64)
		if err != nil {
			return 0, err
		}
		sum = sum + f
		fmt.Println(f)
	}

	average := sum / float64(len(i.Candles))

	return average, nil
}

//TrueRange returns the True Range, defined as TR=Max[(H−L),Abs(H−CP),Abs(L−CP)]
func TrueRange(midCurrent oanda.Mid, midPrevious oanda.Mid) (float64, error) {
	ranges := []float64{}

	//getting the H - L
	currentHigh, err := strconv.ParseFloat(midCurrent.H, 64)
	if err != nil {
		return 0, err
	}
	currentLow, err := strconv.ParseFloat(midPrevious.L, 64)
	if err != nil {
		return 0, err
	}
	ranges = append(ranges, currentHigh-currentLow)

	//getting the H - CP
	closePrevious, err := strconv.ParseFloat(midPrevious.C, 64)
	if err != nil {
		return 0, err
	}
	ranges = append(ranges, math.Abs(currentHigh-closePrevious))

	//getting the L - CP
	ranges = append(ranges, math.Abs(currentLow-closePrevious))

	//getting the max range ie, the true range
	trueRange := 0.0
	for _, v := range ranges {
		if v > trueRange {
			trueRange = v
		}
	}
	fmt.Printf("TRUE RANGE: %f\n", trueRange)
	return trueRange, nil
}

//sigma notation in golang format
//ATR=(1/n)*(i=1 N over SIGMA for TR)
func AverageTrueRange(iH *oanda.InstrumentHistory) (float64, error) {

	n := len(iH.Candles)
	trueRangeSum := 0.0

	for i := n - 1; i >= 1; i-- {
		trueRange, err := TrueRange(iH.Candles[i].Mid, iH.Candles[i-1].Mid)
		if err != nil {
			return 0, err
		}
		trueRangeSum = trueRangeSum + trueRange
	}

	averageTrueRange := (1.0 / float64(n)) * trueRangeSum

	return averageTrueRange, nil
}

/*
Where
TR= A particular true range
n = the time period - IE 14 days

TR=Max[(H − L),Abs(H − CP),Abs(L − CP)]

ATR=(1/n)*(i=1 N over SIGMA for TR)

ATR=(1/n) (n)(i=1)∑TR i
​
*/
