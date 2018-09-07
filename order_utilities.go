package fxtechnical

import (
	"fmt"
	oanda "github.com/nepdave/oanda"
	"log"
	//"sync"
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
	State      string //closed/pending/open.
	OrderID    string //OrderID of order
	Error      error
}

//SideFilled holds the data for whether or not a long or short order was filled
type SideFilled struct {
	Long  bool
	Short bool
}

//CancelOppositeOrder cancels the opposite long/short that was not filled
func CancelOrder(OrderID string) []byte {
	cancelOrderByte, err := oanda.CancelOrder(OrderID)

	if err != nil {
		log.Println(err)
	}

	fmt.Println("cancel order returning cancelOrderByte")
	return cancelOrderByte
}

func CancelOrderAndGetConfirmation(OrderID string) string {
	cancelOrderByte := CancelOrder(OrderID)
	orderCancelTransaction := oanda.OrderCancelTransaction{}.UnmarshalOrderCancelTransaction(cancelOrderByte)
	_type := orderCancelTransaction.OrderCancelTransactionData.Type
	return _type
}

//CancelOppositeOrder cancels the opposite long/short that was not filled
func CancelOppositeOrder(longOrderID string,
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
func CreateOrderAndGetOrderID(instrument string,
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

	fmt.Println("STRING CREATE ORDERS BYTE:")
 	fmt.Println(string(createOrderByte))
  fmt.Println("")

	//checking CreateOrder error
	if err != nil {
		log.Println(err)
	}

	//unmarshaling the returned createOrderByte into a native struct
	orderCreateTransaction := oanda.OrderCreateTransaction{}.
		UnmarshalOrderCreateTransaction(createOrderByte)

	//accessing the orderID field and saving it to a variable
	//orderID := orderCreateTransaction.OrderFillTransaction.OrderID
	orderID := orderCreateTransaction.OrderCreateTransaction.ID

	return orderID
}

//GetOrderState uses an OrderID to get the latest order state
//IE closed/pending/open
func GetOrderState(OrderID string) string {
	//capturing panic raised by Unmarshaling returned getOrderStatusByte
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("UnmarshalOrderStatus() panicked")
			log.Println(err)
		}
	}()

	//using the orderID to check the order status
	getOrderDataByte, err := oanda.GetOrderData(OrderID)

	//checking the GetOrderState error
	if err != nil {
		log.Println(err)
	}

	// fmt.Println("string getOrderDatByte:")
	// fmt.Println(string(getOrderDataByte))

	orderData := oanda.OrderStatus{}.UnmarshalOrderState(getOrderDataByte)
	state := orderData.OrderStatusData.State
	return state
}

//ContinuousGetOrderState uses an infinite for loop  to continually call
//GetOrderState and send an OrderState struct over the channel
func ContinuousGetOrderState(OrderID string, OrderStateChan chan OrderState) {
	for {
		orderState := OrderState{}
		orderState.State = GetOrderState(OrderID)
		orderState.OrderID = OrderID
		OrderStateChan <- orderState
	}
}
