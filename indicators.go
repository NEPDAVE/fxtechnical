package fxtechnical

/*
***************************
BollingerBand structs and methods
***************************
*/

//BollingerBand contains unmarshaled prices data and methods to populate struct fields
type BollingerBand struct {
	UpperBand float64
	Average float64
	lowerBand float64
	Error error
}

//Init populates and returns BollingerBand struct
func (b BolingerBand) Init(instrument string, count int, granularity string) BollingerBand {
	candles, _ := Candles("instrument", count, granularity)
	b.Average, pricesSlice = CloseAverage(candles, count)
	sd := fxtech.StandardDeviation(average, pricesSlice)
	b.UpperBand = average + (sd * 2)
	b.LowerBand = average - (sd * 2)
}

//DoubleBollingerBand contains unmarshaled prices data and methods to populate struct fields
type DoubleBollingerBand struct {
	UpperBand float64
	Average float64
	lowerBand float64
	Error error
}

//Init populates and returns DoubleBollingerBand struct
func (d DoubleBolingerBand) Init(instrument string, count int, granularity string) DoubleBollingerBand {
	candles, _ := Candles("instrument", count, granularity)
	d.Average, pricesSlice = CloseAverage(candles, count)
	sd := fxtech.StandardDeviation(average, pricesSlice)
	d.UpperBand = average + (sd * 3)
	d.LowerBand = average - (sd * 3)
}
