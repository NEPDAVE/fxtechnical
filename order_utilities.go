package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"log"
	"sync"
)

/*
General flow
PrepareOrder()
CreateOrder()
GetOrderID()
GetOrderStatus()
*/

/*
***************************
OrderState holds the order state
***************************
*/

//OrderState holds the Instrument order state and an OrderID
type OrderState struct {
	Instrument string
	mu         sync.Mutex
	State      string //closed/pending/open.
	OrderID    string //OrderID of order
	Error      error
}

//SideFilled holds the data for whether or not a long or short order was filled
type SideFilled struct {
	Long  bool
	Short bool
}

//OrderUtilities is a collection of methods for checking order status/data and
//creating orders
type OrderUtilities struct {
}

//CancelOppositeOrder cancels the opposite long/short that was not filled
func (o OrderUtilities) CancelOppositeOrder(longOrderID string,
	shortOrderID string, sideFilledChan chan SideFilled) {

	for sideFilled := range sideFilledChan {

		if sideFilled.Long == true {
			cancelOrderByte, err := oanda.CancelOrder(shortOrderID)
			if err != nil {
				log.Println(err)
			}
			fmt.Println(string(cancelOrderByte))
		} else if sideFilled.Short == true {
			cancelOrderByte, err := oanda.CancelOrder(longOrderID)
			if err != nil {
				log.Println(err)
			}
			fmt.Println(string(cancelOrderByte))
		}
	}
}

//CreateOrderAndGetOrderID sets the number of units to trade then creates the order using
//the oanda package CreateOrder primitive function and returns an OrderID
func (o OrderUtilities) CreateOrderAndGetOrderID(instrument string,
	units string, orders oanda.Orders) string {
	//capturing panic raised by Unmarshaling returned createOrderByte
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("UnmarshalOrderCreateTransaction() panicked")
			log.Println(err)
		}
	}()

	orders.Order.Units = units

	//creating []byte order data for the order HTTP body
	ordersByte := oanda.Orders{}.MarshalOrders(orders)

	//creating the orders to oanda
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

//GetOrderState uses an OrderID to get the latest order state
//IE closed/pending/open
func (o OrderUtilities) GetOrderState(OrderID string) string {
	//capturing panic raised by Unmarshaling returned getOrderStatusByte
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("UnmarshalOrderStatus() panicked")
			log.Println(err)
		}
	}()

	//using the orderID to check the order status
	getOrderStatusByte, err := oanda.GetOrderStatus(OrderID)

	//checking the GetOrderState error
	if err != nil {
		log.Println(err)
	}

	orderStatus := oanda.OrderStatus{}.UnmarshalOrderStatus(getOrderStatusByte)
	//FIXME this is assuming the order we want is the 0 element in the list
	state := orderStatus.OrderStatusData[0].State
	return state
}
