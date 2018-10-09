package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	twilio "github.com/nepdave/twilio"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

/*
***************************
Dragons is a trading algorithm that implements the London daybreak strategy
with a single market order
***************************
*/

var wg sync.WaitGroup

//Dragons holds the trading alogorithm state and neccesary data
type Dragons struct {
	Instrument             string
	LongUnits              string
	ShortUnits             string
	High                   float64 //high from the last three hours
	Low                    float64 //low from the last three hours
	Bid                    float64 //current highest Bid
	Ask                    float64 //current lowest Ask
	HighLowDifference      float64 //High - Low, provides volatility baseline
	AverageRange           float64 //average of (high - low)/number of candles
	FiftyPeriodSMA         float64
	MarketOrderCreated     bool
	TradeTimeOut           bool //program runs for four hours if no trade is placed
	LongOrders             OrderData
	ShortOrders            OrderData
	OrderCreateTransaction *oanda.OrderCreateTransaction
	TradeID                string
}

//Init kicks off the methods to check prices and create orders
func (d Dragons) Init(instrument string, units string) {
	d.SignalStart()
	d.Instrument = instrument
	d.LongUnits = units
	d.ShortUnits = "-" + units //adding -(negative sign) to denote short order
	fmt.Printf("Long Units: %s\n", d.LongUnits)
	fmt.Printf("Short Units: %s\n", d.ShortUnits)
	d.SetHighAndLow()
	fmt.Printf("High: %.5f\n", d.High)
	fmt.Printf("Low: %.5f\n", d.Low)
	d.SetHighLowDifference()
	d.SetAverageRange()
	d.SetFiftyPeriodSMA()
	d.PrepareLongOrders()
	d.PrepareShortOrders()
	wg.Add(2) //add before the go statement to prevent race conditions
	go d.TradeTimeOutTimer()
	go d.CloseOutPositionsTimer()
	d.MonitorPrices()
	//FIXME need to figure out how to properly set trailingStopLossOrder
	//d.SetTrailingStopLoss()
	d.SignalFinish()
	wg.Wait()
}

//SignalStart sends an SMS indicating that the algorithm has kicked off
func (d *Dragons) SignalStart() {
	message := fmt.Sprintf("Dragons Start: %s\n", time.Now().String())
	twilio.SendSms("15038411492", message)

}

//SetHighAndLow sets the previous three hour High and Low for the Dragons struct
func (d *Dragons) SetHighAndLow() {
	//getting and unmarshaling last three hourly candle data
	candles, err := Candles(d.Instrument, "3", "H1")

	if err != nil {
		log.Println(err)
	}

	//getting the previous three hour high and low
	d.High, d.Low = HighAndLow(candles)
}

//SetHighLowDifference is the total of the high - low
func (d *Dragons) SetHighLowDifference() {
	d.HighLowDifference = d.High - d.Low
}

//SetAverageRange is the (high - low)/number of candels
func (d *Dragons) SetAverageRange() {
	d.AverageRange = AverageRange(d.Instrument, "14", "H1")
}

func (d *Dragons) SetFiftyPeriodSMA() {
	count := "50"
	candles, _ := Candles(d.Instrument, count, "H1")
	average, _ := CloseAverage(candles, count)
	d.FiftyPeriodSMA = average
}

//SetBidAsk sets the current Bid and Ask for the Dragons struct
func (d *Dragons) SetBidAsk() {
	//FIXME should this should be bid and ask with the most liquidity?
	//currently using the highest bid and lowest ask...
	pricesData := PricesData{}.Init(d.Instrument, "mostLiquidSpread")
	d.Bid = pricesData.Bid
	d.Ask = pricesData.Ask
}

//PrepareLongOrders builds a order data populated []byte for API submission
func (d *Dragons) PrepareLongOrders() {
	//setting stop loss at 5 pips below the d.Low
	stopLossPrice := fmt.Sprintf("%.5f", d.Low-.0005)
	takeProfitSize := 3 * d.HighLowDifference
	//setting the take profit at 3x the HighLowDifference + the high + 10 pips
	takeProfitPrice := fmt.Sprintf("%.5f", (d.High + takeProfitSize + .0010))

	//building struct needed for marshaling data into a []byte
	d.LongOrders.Orders = MarketOrder(stopLossPrice, takeProfitPrice,
		d.Instrument, d.LongUnits)

	//marshaling the struct into a byte slice for order creation
	d.LongOrders.OrdersByte = oanda.ClientOrders{}.MarshalClientOrders(
		d.LongOrders.Orders)

	fmt.Println("Long Orders:")
	fmt.Println(string(d.LongOrders.OrdersByte))
	fmt.Println("")
}

//PrepareShortOrders builds a order data populated []byte for API submission
func (d *Dragons) PrepareShortOrders() {
	//setting stop loss 5 pips above the d.High
	stopLossPrice := fmt.Sprintf("%.5f", (d.High + .0005))
	takeProfitSize := 3 * d.HighLowDifference
	//setting the take profit at 3x the HighLowDifference - the low - 10 pips
	takeProfitPrice := fmt.Sprintf("%.5f", (d.Low - takeProfitSize - .0010))

	//building struct needed for marshaling data into a []byte
	d.ShortOrders.Orders = MarketOrder(stopLossPrice, takeProfitPrice,
		d.Instrument, d.ShortUnits)

	//marshaling the struct into a byte slice for order creation
	d.ShortOrders.OrdersByte = oanda.ClientOrders{}.MarshalClientOrders(
		d.ShortOrders.Orders)

	fmt.Println("Short Orders:")
	fmt.Println(string(d.ShortOrders.OrdersByte))
	fmt.Println("")
}

//FIXME may want to think about a way to stop both of the timers if needed

//TradeTimeOutTimer goes for 4 hours, if conditions to trade have not been met
//the timer signals the algorithm to begin the finish sequence
func (d *Dragons) TradeTimeOutTimer() {
	timer := time.NewTimer(4 * time.Hour) //4 hours

	//when the Timer expires, the current time will be sent on C indicating
	//the Timer is done
	<-timer.C
	d.TradeTimeOut = true
	wg.Done()
}

//CloseOutPositionsTimer goes for 8 hours and then closes out all Instrument
//positions to prevent positions from being carried past the London session
func (d *Dragons) CloseOutPositionsTimer() {
	timer := time.NewTimer(8 * time.Hour) //8 hours

	//when the Timer expires, the current time will be sent on C indicating
	//the Timer is done
	<-timer.C
	//FIXME may want to monitor the d.MarketOrderCreated and stop the timer if
	//no market order was created after the 4 hours timer alloted
	closePositionsByte, err := oanda.ClosePositions(d.Instrument)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Close Positions Response:")
	fmt.Println(string(closePositionsByte))
	wg.Done()
}

//MonitorPrices checks that the timer has not run out and that an order has not
//been created and continues to MonitorPrices for a breakout
func (d *Dragons) MonitorPrices() {
	//if a market order has not been created loop continues and the timer has
	//not run out the loop continues
	fmt.Println("Entering MonitorPrices loop...")
	fmt.Println("")
	for d.MarketOrderCreated == false && d.TradeTimeOut == false {
		//putting at least .5 seconds between requests to prevent blocked requests
		time.Sleep(500 * time.Millisecond)
		d.SetBidAsk()
		// fmt.Println("#######################")
		// fmt.Println(time.Now())
		// fmt.Printf("Highest Bid: %f\n", d.Bid)
		// fmt.Printf("BidDiff ABV: %.5f\n", d.BidDiff)
		// fmt.Println("")
		// fmt.Printf("Lowest Ask: %f\n", d.Ask)
		// fmt.Printf("AskDiff ABV: %.5f\n", d.AskDiff)
		// fmt.Println("")
		// fmt.Printf("Spread: %.5f\n", (d.Ask - d.Bid))

		if d.Ask > d.High && d.Ask > d.FiftyPeriodSMA {
			createOrdersByte, err := oanda.CreateOrder(d.LongOrders.OrdersByte)

			if err != nil {
				log.Println(err)
			}

			//unmarshaling the returned createOrdersByte into a native struct
			d.OrderCreateTransaction = oanda.OrderCreateTransaction{}.
				UnmarshalOrderCreateTransaction(createOrdersByte)

			fmt.Println("Long Order Create Transaction:")
			fmt.Println(string(createOrdersByte))
			fmt.Println("")

			d.MarketOrderCreated = true
			return

		} else if d.Bid < d.Low && d.Bid < d.FiftyPeriodSMA {
			createOrdersByte, err := oanda.CreateOrder(d.ShortOrders.OrdersByte)

			if err != nil {
				log.Println(err)
			}

			//unmarshaling the returned createOrdersByte into a native struct
			d.OrderCreateTransaction = oanda.OrderCreateTransaction{}.
				UnmarshalOrderCreateTransaction(createOrdersByte)

			fmt.Println("Short Order Create Transaction:")
			fmt.Println(string(createOrdersByte))
			fmt.Println("")

			d.MarketOrderCreated = true

			return
		}
	}
}

//SetTrailingStopLoss adds a trailingStopLossOrder to and existing trade
func (d *Dragons) SetTrailingStopLoss() {
	d.TradeID = d.OrderCreateTransaction.
		OrderFillTransaction.TradeOpened.TradeID

	if d.MarketOrderCreated == true && d.TradeID != "" {

		//FIXME this should not be hard coded. it should dynamically changed
		distance := "1"

		//building struct needed for marshaling data into a []byte
		trailingStopLossOrder := TrailingStopLossOrder(d.TradeID, distance)

		//marshaling the struct into a byte slice for order creation
		trailingStopLossOrderByte := oanda.ClientOrders{}.MarshalClientOrders(
			trailingStopLossOrder)

		createOrdersByte, err := oanda.CreateOrder(trailingStopLossOrderByte)

		if err != nil {
			log.Println(err)
		}

		fmt.Println("STOP LOSS ORDER RESPONSE:")
		fmt.Println(string(createOrdersByte))

	} else {
		fmt.Println("TradeID was empty string...")
	}
}

//SignalFinish sends an SMS to signal that the algo has finished
func (d *Dragons) SignalFinish() {
	fmt.Println("Writing to done.txt...")

	// use touch if log.txt does not exist, 0644 is standard permission
	file, err := os.OpenFile("done.txt", os.O_WRONLY, 0644)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	fmt.Fprintf(file, "Dragons finish: %s\n", time.Now().String())
	done := fmt.Sprintf("Dragons finish: %s\n", time.Now().String())

	marketOrderCreated := fmt.Sprintf("Market Order Created: %s\n",
		strconv.FormatBool(d.MarketOrderCreated))

	orderCreateTransaction := fmt.Sprintf("Order Create Transaction: %v\n",
		d.OrderCreateTransaction)

	message := done + marketOrderCreated + orderCreateTransaction

	twilio.SendSms("15038411492", message)
	fmt.Println("sent done sms")

	return
}
