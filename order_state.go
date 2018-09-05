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
OrderState holds the order state
***************************
*/

//OrderData holds the Instrument order state and an OrderID
type OrderState struct {
	Instrument string
	mu         sync.Mutex
	State      string //closed/pending/open.
	OrderID    string //OrderID of order
	Error      error
}
