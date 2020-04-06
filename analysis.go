package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
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

//AverageTrueRange return the AverageTrueRange until 
//func AverageTrueRange(i *oanda.InstrumentHistory) (float64, error) {

//sigma notation in golang format
func AverageTrueRange() {
	fmt.Println("Hello, playground")

	sum := 0
	for i := 0; i <= 3; i++ {
		sum = sum + i
		fmt.Printf("i: %d sum: %d\n", i, sum)
	}
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
	


}
