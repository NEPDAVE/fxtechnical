package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"log"
)

/*
General flow
PrepareOrder()
CreateOrder()
GetOrderID()
GetOrderStatus()
*/

//OrderUtilities is a collection of methods for checking order status/data and
//executing orders
type OrderUtilities struct {
}

//ExecuteOrder sets the number of units to trade then creates the order using
//the oanda package CreateOrder primitive function and returns an OrderID
func (o OrderUtilities) ExecuteOrder(instrument string, units string, raider Raider) string {
	//capturing panic raised by Unmarshaling
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("UnmarshalOrderCreateTransaction() panicked")
			log.Println(err)
		}
	}()

	raider.Orders.OrderData.Units = units

	//creating []byte order data for the order HTTP body
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

	return orderID
}

//GetOrderStatus uses an OrderID to get the latest order status
func (o OrderUtilities) GetOrderStatus(OrderID string) string {
	//capturing panic raised by Unmarshaling
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("UnmarshalOrderStatus() panicked")
			log.Println(err)
		}
	}()

	//using the orderID to check the order status
	getOrderStatusByte, err := oanda.GetOrderStatus(OrderID)
	if err != nil {
		log.Println(err)
	}

	orderStatus := oanda.OrderStatus{}.UnmarshalOrderStatus(getOrderStatusByte)
	//FIXME this is assuming the order we want is the 0 element in the list
	state := orderStatus.OrderStatusData[0].State
	return state
}
