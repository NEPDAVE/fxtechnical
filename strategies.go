package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"log"
	"sync"
	"time"
)

/*
***************************
Raider is a trading algorithm that implements the Bolinger Band indicator
***************************
*/

//Raider holds the algo state and neccesary algo data
type Raider struct {
	OrderStatus     string       //closed/pending/open.
	CreateOrderCode int          //0 = dont execute 1 = execute
	OrderID         string       //OrderID of current order
	Orders          oanda.Orders //Order SL/TP Limit/Market data
	Error           error
}

//Init kicks off the select pattern to create orders and check orders
func (r Raider) Init(instrument string, units string) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	RaiderChan := make(chan Raider)
	CheckOrderChan := make(chan string)

	wg.Add(2)
	//Checks whether or not conditions are right to trade
	go r.ContinuousRaid(instrument, RaiderChan)
	//Checks whether orders are close, pending, or open
	go r.ContinuousCheckOrder(CheckOrderChan)

	for {
		select {
		case raider := <-RaiderChan:

			if raider.Error != nil {
				log.Println(raider.Error)
				continue
			}

			if r.CreateOrderCode == 1 && r.OrderStatus == "closed" {
				fmt.Println("received create order signal...")
				mu.Lock()
				//doing exspensive IO calls but need to verify OrderStatus before assigning
				r.OrderID = r.ExecuteOrder(instrument, units, raider)
				r.OrderStatus = r.CheckOrder(r.OrderID)
				r.OrderStatus = "pending"
				mu.Unlock()
			} else {
				//fmt.Println(raider)
				fmt.Println("...")
			}
		//FIXME need to make sure you understand the checkOrder data structures
		case r.OrderStatus = <-CheckOrderChan:
			if r.OrderStatus == "closed" {
				mu.Lock()
				r.OrderStatus = "closed"
				mu.Unlock()
				fmt.Printf("order %s = closed", r.OrderID)
			} else if r.OrderStatus == "pending" {
				fmt.Printf("order %s = pending", r.OrderID)
			} else if r.OrderStatus == "open" {
				mu.Lock()
				r.OrderStatus = "open"
				mu.Unlock()
				fmt.Printf("order %s = open", r.OrderID)
			}
			// default:
			// 	fmt.Println("no data...")
			// 	fmt.Println(r.Orders)
		}
		fmt.Println("")
		fmt.Println("")
	}

	wg.Wait()
}

//ExecuteOrder sets the number of units to trade then submits/creates the order
func (r *Raider) ExecuteOrder(instrument string, units string, raider Raider) string {
	r.Orders.OrderData.Units = units

	//creating []byte order data for the order HTTP body
	ordersByte := oanda.MarshalOrders(r.Orders)

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
}

//CheckOrder used an OrderID to get the latest order status
func (r *Raider) CheckOrder(OrderID string) string {
	//using the orderID to check the order status
	for {
		checkOrderByte, err := oanda.CheckOrder(r.OrderID)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(checkOrderByte))
		//FIXME need to have call to unmarshaling checkOrderByte
		//and way to see/check whether the order is close/pending/open
		status := "closed/pending/open"
		return status
	}
}

//ContinuousCheckOrder uses an OrderID to continuously get the latest order status
func (r *Raider) ContinuousCheckOrder(CheckOrderChan chan string) {
	//using the orderID to check the order status
	for {
		checkOrderByte, err := oanda.CheckOrder(r.OrderID)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(checkOrderByte))
		//FIXME need to have call to unmarshaling checkOrderByte
		//and way to see/check whether the order is close/pending/open
		CheckOrderChan <- "closed/pending/open"
	}
}

//SingleRaid compares a single PricesData to a BollingerBand and returns Orders
//and the CreateOrderCode. 1 = execute order, 0 = don't execute order
func (r Raider) SingleRaid(instrument string) (oanda.Orders, int) {
	bb := BollingerBand{}.Init(instrument, "20", "D")
	pricesData := PricesData{}.Init(instrument)

	if pricesData.Error != nil {
		log.Println(pricesData.Error)
		return oanda.Orders{}, 0
	}

	//setting all units to 0 here so that proper amount of units can be set later
	if pricesData.Bid > bb.UpperBand {
		return oanda.Orders{}.MarketSellOrder(pricesData.Bid,
			pricesData.Ask,
			bb.Instrument,
			"0"), 1
	} else if pricesData.Ask < bb.LowerBand {
		return oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
			pricesData.Ask,
			bb.Instrument,
			"0"), 1
	} else {
		return oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
			pricesData.Ask,
			bb.Instrument,
			"0"), 0
	}
}

//ContinuousRaid ranges over a channel of PricesData comparing each PricesData to a
//BollingerBand and sends a trading decision over a channel to the caller
func (r Raider) ContinuousRaid(instrument string, out chan Raider) {
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
					bb.Instrument, "0"),
				CreateOrderCode: 1,
			}
		} else if pricesData.Ask < bb.LowerBand {
			out <- Raider{
				Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument, "0"),
				CreateOrderCode: 1,
			}
		} else {
			out <- Raider{
				Orders: oanda.Orders{}.MarketBuyOrder(pricesData.Bid,
					pricesData.Ask,
					bb.Instrument, "0"),
				CreateOrderCode: 0,
			}
		}
	}
}
