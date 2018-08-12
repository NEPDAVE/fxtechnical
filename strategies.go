package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	// twilio "github.com/nepdave/twilio"
	"log"
	"sync"
	"time"
)

var wg sync.WaitGroup

/*
***************************
Raider is a trading algorithm that implements the Bolinger Band indicator
***************************
*/

//ExecuteRaider raider executes the Raider trading algorithm
func ExecuteRaider(instrument string, units string) {
	bb := BollingerBand{}.Init(instrument, "20", "D")

	//anonymous go func executing concurrently to update bb everyday at midnight
	//link to good time example becasue this has not been tested
	//http://www.golangprograms.com/get-year-month-day-hour-min-and-second-from-current-date-and-time.html
	wg.Add(1)
	go func() {
		for {
			now := time.Now()
			if now.Hour() == 00 && now.Minute() == 0 && now.Second() < 5 {
				bb = BollingerBand{}.Init(instrument, "20", "D")
			}
		}
	}()

	raiderChan := make(chan Raider)
	//FIXME where do we really want to set the number of units?
	wg.Add(1)
	go Raider{}.ContinuousRaid(bb, units, raiderChan)

	fmt.Println("entering range over raider channel")
	for raider := range raiderChan {
		if raider.Error != nil {
			log.Println(raider.Error)
			continue
		}

		//calls to marshaling the order data and submiting order to Oanda
		if raider.ExecuteOrder != 1 {
			raider.Orders.OrderData.Units = units

			//creating []byte order data for creating order
			ordersByte := oanda.MarshalOrders(raider.Orders)

			//creating/submiting the order to oanda
			createOrderByte, err := oanda.CreateOrder(ordersByte)

			//checking CreateOrder error
			if err != nil {
				log.Println(err)
			}

			//unmarshaling the returned createOrderByte into a native struct
			orderCreateTransaction := oanda.OrderCreateTransaction{}.
				UnmarshalOrderCreateTransaction(createOrderByte)

			//accessing the orderID field and saving it to a variable
			orderID := orderCreateTransaction.OrderFillTransaction.OrderID

			//using the orderID to check the order status
			checkOrderByte, err := oanda.CheckOrder(orderID)

			//checking the the CheckOrder error
			if err != nil {
				log.Println(err)
			}

			fmt.Println("Check Order Response:")
			fmt.Println(string(checkOrderByte))

		}
		wg.Wait()
	}

}

//Raider contains order data and the execute order decision
type Raider struct {
	Orders       oanda.Orders
	ExecuteOrder int //1 = execute order. 0 = don't execute order
	Error        error
}

//SingleRaid compares a single PricesData to a BollingerBand and returns a trading decision
func (r Raider) SingleRaid(bb BollingerBand, units string) Raider {
	//initializing pricesData struct
	pricesData := PricesData{}.Init(bb.Instrument)

	if pricesData.Error != nil {
		return Raider{Error: pricesData.Error}
	}

	//calling CalculateSpread also sets bid/ask fields in struct
	//pricesData.CalculateSpread()

	//FIXME need to have better error handling here
	if pricesData.Bid > bb.UpperBand {
		return Raider{
			Orders: oanda.Orders{}.MarketSellOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument,
				units),
			ExecuteOrder: 1,
		}
	} else if pricesData.Ask < bb.LowerBand {
		return Raider{
			Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument,
				units),
			ExecuteOrder: 1,
		}
	} else {
		return Raider{
			Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument,
				units),
			ExecuteOrder: 0,
		}
	}

}

//ContinuousRaid ranges over a channel of PricesData comparing each PricesData to a
//BollingerBand and sends a trading decision over a channel to the caller
//FIXME if running this function coninuosly be careful to generate a new
//bollinger band at the start of each day
func (r Raider) ContinuousRaid(bb *BollingerBand, units string, out chan Raider) {
	oandaChan := make(chan PricesData)
	go StreamBidAsk(bb.Instrument, oandaChan)

	for pricesData := range oandaChan {
		if pricesData.Error != nil {
			out <- Raider{Error: pricesData.Error}
		}

		if pricesData.Heartbeat != nil {
			fmt.Println(pricesData.Heartbeat)
			continue
		}

		//print prices data
		//fmt.Println(pricesData)

		//FIXME need to have better error handling here
		if pricesData.Bid > bb.UpperBand {
			out <- Raider{
				Orders: oanda.Orders{}.MarketSellOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument,
					units),
				ExecuteOrder: 1,
			}
		} else if pricesData.Ask < bb.LowerBand {
			out <- Raider{
				Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument,
					units),
				ExecuteOrder: 1,
			}
		} else {
			out <- Raider{
				Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument,
					units),
				ExecuteOrder: 0,
			}
		}
	}
}
