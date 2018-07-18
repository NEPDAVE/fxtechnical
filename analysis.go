package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"log"
	"math"
	"strconv"
)

//CloseAverage returns the average and all of the values used to calculate the average
//so we can determine the StandardDeviation
//FIXME this should have a unit test!
func CloseAverage(candles *oanda.Candles, count string) (float64, []float64) {
	sum := 0.0
	pricesSlice := []float64{}
	i, err := strconv.Atoi(count)
	if err != nil {
		log.Fatal(err)
	}

	for _, val := range candles.Candles {

		f, err := strconv.ParseFloat(val.Mid.Close, 64)
		if err != nil {
			log.Fatal(err)
		}
		pricesSlice = append(pricesSlice, f)
		sum = sum + f

	}
	return (sum / float64(i)), pricesSlice
}

//StandardDeviation returns the standard deviate for the parameters passed
func StandardDeviation(average float64, pricesSlice []float64) float64 {
	sd := 0.0
	counter := 0.0

	for _, val := range pricesSlice {
		// The use of Pow math function func Pow(x, y float64) float64
		sd += math.Pow(val-average, 2)
		counter++
	}
	// The use of Sqrt math function func Sqrt(x float64) float64
	//gotta know if bolliner bands use n or n-1 for this part...
	sd = math.Sqrt(sd / (counter - 1))

	fmt.Println("The Standard Deviation is : ", sd)

	return sd
}
