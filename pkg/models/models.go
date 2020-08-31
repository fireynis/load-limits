package models

import (
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: no matching record found")

//Storing money as in int (value * 100) means you don't lose precision. Effectively working in pennies.
type Load struct {
	Id            int64
	TransactionId int64
	CustomerId    int64
	Amount        int64
	Time          time.Time
	Accepted      bool
}
