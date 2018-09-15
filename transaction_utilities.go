package fxtechnical

import (
//"fmt"
oanda "github.com/nepdave/oanda"
//"log"
)

/*
***************************
TransactionState holds the order/trade state
***************************

could be possible to just either return an orderID or a tradeID and
depending on that continue to monitor the situation..

*/

//ClientTransaction creates transactions and queries oanda to see if a
//transaction is an order or trade and the "state" or the order/trade
//FIXME need to verify the above statement is how to do it... it appears
//"transactions" are the data type common across orders/trades/positions
//with oanda using them as a high level "driver" is worth looking into
type ClientTransaction struct {
	Instrument string
	Order      OrdersTransaction
	Trade      TradeTransaction
	Error      error
}

type OrdersTransaction struct {
	Data    oanda.OrderCreateTransactionData //oanda package primitive
	State   string                     //cancelled/closed/pending/open
	OrderID string
}

type TradeTransaction struct {
	Data    oanda.OrderFillTransactionData //oanda package primitive
	State   string                   //FIXME is this even a thing?
	TradeID string
}

func (t ClientTransaction) Init() {

}

func CreateOrdersTransaction() {
	ordersTransaction := CreateClientOrders(instrument, units, orders)
}

//MonitorTransaction finds out whether a transaction is an order or trade
//and the "state" of the order/trade
func MonitorTransaction() {

}
