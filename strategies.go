package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	// twilio "github.com/nepdave/twilio"
	"log"
	"sync"
	"time"
)

//trying to figure out select stuff
//https://play.golang.org/p/HrPk4uEO2tS

/*
***************************
Raider is a trading algorithm that implements the Bolinger Band indicator
***************************
*/

//Raider holds the algo state and neccesary algo data
type Raider struct {
	mu              sync.Mutex
	Status          int          //0 = orders closed. 1 = order submitted. 2 = orders open.
	OrderID         string       //OrderID of current submitted/filled order
	Orders          oanda.Orders //Order data
	CreateOrderCode int          //1 = execute order. 0 = don't execute order
	Error           error
}

//Init kicks off the select pattern to check on raider status
func (r Raider) Init(instrument string, units string) {

	RaiderChan := make(chan Raider)
	var wg sync.WaitGroup

	wg.Add(2)
	go r.ContinuousRaid(instrument, RaiderChan)
	go r.CheckOrder()

	/*
		  the two goroutines should still send a value of the channel
			this will make each of them more modular. you can then init
			use the value sent over the channel to and pass it to CheckConditions
			which is a func that checks the current Status and decides
			whether or not to submit the order and get the new order ID.
	*/

	for {
		select {
		case raider := <-RaiderChan:

			if Raider.Error != nil {
				log.Println(Raider.Error)
				continue
			}

			if raider.CreateOrderCode == 1 && raider.Status == 0 {
				fmt.Println("received create order signal...")
				mu.Lock()
				r.Status = 1
				r.OrderID = ExecuteOrder(instrument, units, raider)
				mu.Unlock()
			}

		default:
			fmt.Println("no data...")
		}
	}

	wg.Wait()
}

//ExecuteOrder submits and creates the order
func (r *Raider) ExecuteOrder(instrument string, units string, raider Raider) string {
	Raider.Orders.OrderData.Units = units

	//creating []byte order data for the order HTTP body
	ordersByte := oanda.MarshalOrders(Raider.Orders)

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

	return orderID
	wg.Wait()

}

//CheckOrder used an OrderID to get the latest order status
func (r *Raider) CheckOrder() {
	//using the orderID to check the order status
	for {
		if r.Status == 1 {
			checkOrderByte, err := oanda.CheckOrder(r.OrderID)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("check order byte:")
			fmt.Println(checkOrderByte)
		}
	}
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
			CreateOrderCode: 1,
		}
	} else if pricesData.Ask < bb.LowerBand {
		return Raider{
			Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument,
				units),
			CreateOrderCode: 1,
		}
	} else {
		return Raider{
			Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
				pricesData.Ask,
				bb.Instrument,
				units),
			CreateOrderCode: 0,
		}
	}

}

//ContinuousRaid ranges over a channel of PricesData comparing each PricesData to a
//BollingerBand and sends a trading decision over a channel to the caller
//FIXME if running this function coninuosly be careful to generate a new
//bollinger band at the start of each day
func (r Raider) ContinuousRaid(instrument, out chan Raider) {
	bb := BollingerBand{}.Init(instrument, "20", "D")
	var wg sync.WaitGroup

	//anonymous go func executing concurrently to update bb everyday at midnight
	wg.Add(1)
	go func() {
		for {
			now := time.Now()
			if now.Hour() == 00 && now.Minute() == 0 && now.Second() < 5 {
				bb = BollingerBand{}.Init(instrument, "20", "D")
			}
		}
	}()

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
				CreateOrderCode: 1,
			}
		} else if pricesData.Ask < bb.LowerBand {
			out <- Raider{
				Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument,
					units),
				CreateOrderCode: 1,
			}
		} else {
			out <- Raider{
				Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument,
					units),
				CreateOrderCode: 0,
			}
		}
	}
}
