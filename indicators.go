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

//MostLiquidBid returns the most liquid Bid Quote out of all the Quotes
func MostLiquidBid(bids []oanda.Bids) (*Quote, error) {
	liquidity := 0
	priceStr := ""

	for _, bid := range bids {
		if bid.Liquidity > liquidity {
			liquidity = bid.Liquidity
			priceStr = bid.Price
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

//CloseAverage returns the average and close of an array of candles
func CloseAverage(insHist *oanda.InstrumentHistory) (float64, error) {
	sum := 0.0

	for _, c := range insHist.Candles {
		f, err := strconv.ParseFloat(c.Mid.C, 64)
		if err != nil {
			return 0, err
		}
		sum = sum + f
	}

	average := sum / float64(len(insHist.Candles))

	return average, nil
}

//VolumeAverage returns the volume average of an array of candles
func VolumeAverage(insHist *oanda.InstrumentHistory) (float64, error) {
	sum := 0

	for _, c := range insHist.Candles {
		volume := c.Volume
		sum = sum + volume
	}

	average := float64(sum) / float64(len(insHist.Candles))

	return average, nil
}

//ATR=(1/n)*(i=1 N over SIGMA for TR) or ATR=(1/n) (n)(i=1)∑TR i
func AverageTrueRange(insHist *oanda.InstrumentHistory) (float64, error) {

	n := len(insHist.Candles)
	trueRangeSum := 0.0

	for i := len(insHist.Candles) - 2; i >= 1; i-- {
		trueRange, err := TrueRange(insHist.Candles[i].Mid, insHist.Candles[i-1].Mid)
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
