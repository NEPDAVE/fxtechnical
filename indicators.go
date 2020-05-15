package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"math"
	"strconv"
)

//Quote represents an Oanda Bid or Ask with the Price converted from
//a string to a float64
type Quote struct {
	Liquidity int     `json:"liquidity"`
	Price     float64 `json:"price"`
}

//MostLiquidAsk returns the most liquid Ask Quote out of all the Quotes
func MostLiquidAsk(asks []oanda.Asks) (*Quote, error) {
	liquidity := 0
	priceStr := ""

	for _, ask := range asks {
		if ask.Liquidity > liquidity {
			liquidity = ask.Liquidity
			priceStr = ask.Price
		}
	}

	price, err := strconv.ParseFloat(priceStr, 64)

	if err != nil {
		return nil, err
	}

	quote := &Quote{
		Liquidity: liquidity,
		Price:     price,
	}

	return quote, nil
}

//MostLiquidAsk returns the most liquid Ask Quote out of all the Quotes
func MostLiquidAsk(asks []oanda.Asks) (*Quote, error) {
	liquidity := 0
	priceStr := ""

	for _, ask := range asks {
		if ask.Liquidity > liquidity {
			liquidity = ask.Liquidity
			priceStr = ask.Price
		}
	}

	price, err := strconv.ParseFloat(priceStr, 64)

	if err != nil {
		return nil, err
	}

	quote := &Quote{
		Liquidity: liquidity,
		Price:     price,
	}

	return quote, nil
}

//CloseAverage returns the average and all of the values used to calculate the average
func CloseAverage(insHist *oanda.InstrumentHistory) (float64, error) {
	sum := 0.0

	for _, v := range insHist.Candles {
		f, err := strconv.ParseFloat(v.Mid.C, 64)
		if err != nil {
			return 0, err
		}
		sum = sum + f
		fmt.Println(f)
	}

	average := sum / float64(len(insHist.Candles))

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

	return trueRange, nil
}

//ATR=(1/n)*(i=1 N over SIGMA for TR) or ATR=(1/n) (n)(i=1)∑TR i
func AverageTrueRange(iH *oanda.InstrumentHistory) (float64, error) {

	n := len(iH.Candles)
	trueRangeSum := 0.0

	for i := len(iH.Candles) - 2; i >= 1; i-- {
		trueRange, err := TrueRange(iH.Candles[i].Mid, iH.Candles[i-1].Mid)
		if err != nil {
			return 0, err
		}
		trueRangeSum = trueRangeSum + trueRange
	}

	atrString := fmt.Sprintf("%.4f", (1.0/float64(n))*trueRangeSum)
	averageTrueRange, err := strconv.ParseFloat(atrString, 64)
	if err != nil {
		return 0, err
	}

	return averageTrueRange, nil
}
