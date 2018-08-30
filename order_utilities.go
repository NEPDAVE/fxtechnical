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

//OrderUtilities is a collection of method signatures
type OrderUtilities interface {
	ExecuteOrder()
	GetOrderStatus()
	ContinuousGetOrderStatus()
}

//ExecuteOrder sets the number of units to trade then creates the order
func ExecuteOrder(instrument string, units string, raider Raider) string {
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
func GetOrderStatus(OrderID string) string {
	//using the orderID to check the order status
	getOrderStatusByte, err := oanda.GetOrderStatus(OrderID)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println("")
	// fmt.Println("STRING CHECK ORDER BYTE:")
	// fmt.Println(string(GetOrderStatusByte))
	// fmt.Println("")
	//FIXME this is assuming the order we want is the 0 element in the list
	state := oanda.OrderStatus{}.UnmarshalOrderStatus(getOrderStatusByte).OrderStatusData[0].State
	fmt.Println("")
	fmt.Println("")
	fmt.Println("STATE:")
	fmt.Println(state)
	fmt.Println("")
	fmt.Println("")
	return state
}

//ContinuousGetOrderStatus uses an OrderID to continuously get the latest order status
func ContinuousGetOrderStatus(orderID string, GetOrderStatusChan chan string) {
	//using the orderID to check the order status
	for {
		getOrderStatusByte, err := oanda.GetOrderStatus(orderID)
		if err != nil {
			fmt.Println(err)
		}
		// fmt.Println("")
		// fmt.Println("STRING CHECK ORDER BYTE:")
		// fmt.Println(string(GetOrderStatusByte))
		// fmt.Println("")
		//FIXME this is assuming the order we want is the 0 element in the list
		state := oanda.OrderStatus{}.UnmarshalOrderStatus(getOrderStatusByte).OrderStatusData[0].State
		fmt.Println("")
		fmt.Println("")
		fmt.Println("STATE:")
		fmt.Println(state)
		fmt.Println("")
		fmt.Println("")
		GetOrderStatusChan <- state
	}
}
