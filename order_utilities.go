package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"log"
	"sync"
	"time"
)

/*
General flow
PrepareOrder()
CreateOrder()
GetOrderID()
CheckOrder()
*/

//ExecuteOrder sets the number of units to trade then creates the order
func ExecuteOrder(instrument string, units string, raider Raider) string {
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
func CheckOrder(OrderID string) string {
	//using the orderID to check the order status
	for {
		checkOrderByte, err := oanda.CheckOrder(r.OrderID)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("")
		fmt.Println("STRING CHECK ORDER BYTE:")
		fmt.Println(string(checkOrderByte))
		fmt.Println("")
		//FIXME need to have call to unmarshaling checkOrderByte
		//and way to see/check whether the order is close/pending/open
		status := "closed/pending/open"
		return status
	}
}

//ContinuousCheckOrder uses an OrderID to continuously get the latest order status
func ContinuousCheckOrder(CheckOrderChan chan string) {
	//using the orderID to check the order status
	for {
		checkOrderByte, err := oanda.CheckOrder(r.OrderID)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("")
		fmt.Println("STRING CHECK ORDER BYTE:")
		fmt.Println(string(checkOrderByte))
		fmt.Println("")
		//FIXME need to have call to unmarshaling checkOrderByte
		//and way to see/check whether the order is close/pending/open
		CheckOrderChan <- "closed/pending/open"
	}
}
