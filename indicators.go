package fxtechnical

/*
***************************
BollingerBand structs and methods
***************************
*/

//BollingerBand contains unmarshaled prices data and methods to populate struct fields
type BollingerBand struct {
	UpperBand   float64
	Average     float64
	LowerBand   float64
	Instrument  string
	Count       string
	Granularity string
	Error       error
}

//Init populates and returns BollingerBand struct
//FIXME currently not doing any error checking here
func (b BollingerBand) Init(instrument string, count string, granularity string) *BollingerBand {
	candles, _ := Candles(instrument, count, granularity)
	average, pricesSlice := CloseAverage(candles, count)
	b.Average = average
	sd := StandardDeviation(average, pricesSlice)
	b.UpperBand = average + (sd * 2)
	b.LowerBand = average - (sd * 2)
	b.Instrument = instrument
	b.Count = count
	b.Granularity = granularity

	return &b
}

//DoubleBollingerBand contains unmarshaled prices data and methods to populate struct fields
type DoubleBollingerBand struct {
	UpperBand   float64
	Average     float64
	LowerBand   float64
	Instrument  string
	Count       string
	Granularity string
	Error       error
}

//Init populates and returns DoubleBollingerBand struct
//FIXME currently not doing any error checking here
func (d DoubleBollingerBand) Init(instrument string, count string, granularity string) DoubleBollingerBand {
	candles, _ := Candles(instrument, count, granularity)
	average, pricesSlice := CloseAverage(candles, count)
	d.Average = average
	sd := StandardDeviation(average, pricesSlice)
	d.UpperBand = average + (sd * 3)
	d.LowerBand = average - (sd * 3)
	d.Instrument = instrument
	d.Count = count
	d.Granularity = granularity

	return d
}