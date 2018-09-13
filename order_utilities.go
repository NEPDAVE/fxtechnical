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

//CancelOrder cancels an order and retursns a []byte slice to unmarshal
func CancelOrder(OrderID string) []byte {
	cancelOrderByte, err := oanda.CancelOrder(OrderID)

	if err != nil {
		log.Println(err)
	}

	fmt.Println("cancel order returning cancelOrderByte")
	return cancelOrderByte
}

//CancelOrderAndGetConfirmation cancels an orders and returns a confirmation string
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

//CreateClientOrdersAndGetOrderID sets the number of units to trade then creates the order using
//the oanda package CreateOrder primitive function and returns an OrderID
func CreateClientOrdersAndGetOrderID(instrument string,
	units string, orders oanda.ClientOrders) string {
	//capturing panic raised by Unmarshaling returned createOrderByte
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("UnmarshalOrderCreateTransaction() panicked")
			log.Println(err)
		}
	}()

	orders.Orders.Units = units

	//creating []byte order data for the order HTTP body
	ordersByte := oanda.ClientOrders{}.MarshalClientOrders(orders)

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

//GetOrderState uses an OrderID to to call oanda.GetOrder() and then unmarshals
//the struct and returns the order state IE open/pending/closed
func GetOrderState(orderID string) string {
	//capturing panic raised by Unmarshaling returned getOrderStatusByte
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("UnmarshalOrderStatus() panicked")
			log.Println(err)
		}
	}()

	//using the orderID to check the order status
	getOrderByte, err := oanda.GetOrder(orderID)

	//checking the GetOrderState error
	if err != nil {
		log.Println(err)
	}

	fmt.Println("string getOrderByte:")
	fmt.Println(string(getOrderByte))

	order := oanda.Order{}.UnmarshalOrder(getOrderByte)
	state := order.OrderData.State
	return state
}

//ContinuousGetOrder uses an infinite for loop  to continually call
//GetOrder and send an OrderState struct over the channel
func ContinuousGetOrder(OrderID string, OrderStateChan chan OrderState) {
	for {
		orderState := OrderState{}
		orderState.State = GetOrderState(OrderID)
		orderState.OrderID = OrderID
		OrderStateChan <- orderState
	}
}


/*
{
	"orderCreateTransaction": {
		"type": "LIMIT_ORDER",
		"instrument": "GBP_USD",
		"units": "2",
		"price": "1.31242",
		"timeInForce": "GTC",
		"triggerCondition": "DEFAULT",
		"partialFill": "DEFAULT",
		"positionFill": "DEFAULT",
		"takeProfitOnFill": {
			"price": "1.31842",
			"timeInForce": "GTC"
		},
		"stopLossOnFill": {
			"price": "1.31042",
			"timeInForce": "GTC"
		},
		"reason": "CLIENT_ORDER",
		"id": "10782",
		"accountID": "101-001-6395930-001",
		"userID": 6395930,
		"batchID": "10782",
		"requestID": "24460451218915493",
		"time": "2018-09-13T17:12:07.249639551Z"
	},
	"orderFillTransaction": {
		"type": "ORDER_FILL",
		"orderID": "10782",
		"instrument": "GBP_USD",
		"units": "2",
		"requestedUnits": "2",
		"price": "1.31113",
		"pl": "0.0000",
		"financing": "0.0000",
		"commission": "0.0000",
		"accountBalance": "100346.9816",
		"gainQuoteHomeConversionFactor": "1",
		"lossQuoteHomeConversionFactor": "1",
		"guaranteedExecutionFee": "0.0000",
		"halfSpreadCost": "0.0002",
		"fullVWAP": "1.31113",
		"reason": "LIMIT_ORDER",
		"tradeOpened": {
			"price": "1.31113",
			"tradeID": "10783",
			"units": "2",
			"guaranteedExecutionFee": "0.0000",
			"halfSpreadCost": "0.0002",
			"initialMarginRequired": "0.1311"
		},
		"fullPrice": {
			"closeoutBid": "1.31073",
			"closeoutAsk": "1.31138",
			"timestamp": "2018-09-13T17:12:06.390456028Z",
			"bids": [{
				"price": "1.31098",
				"liquidity": "10000000"
			}],
			"asks": [{
				"price": "1.31113",
				"liquidity": "10000000"
			}]
		},
		"id": "10783",
		"accountID": "101-001-6395930-001",
		"userID": 6395930,
		"batchID": "10782",
		"requestID": "24460451218915493",
		"time": "2018-09-13T17:12:07.249639551Z"
	},
	"relatedTransactionIDs": ["10782", "10783", "10784", "10785"],
	"lastTransactionID": "10785"
}

*/
