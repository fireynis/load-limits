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

type ILoads interface {
	Get(id int64) (*Load, error)
	GetByTransactionId(customerId int64, transactionId int64) (*Load, error)
	GetByCustomerTransactionsByDateRange(customerId int64, startDate time.Time, endDate time.Time) ([]*Load, error)
	Insert(customerId int64, transactionId int64, amount int64, transactionTime time.Time, accepted bool) (int64, error)
	Update(model *Load) error
	WithinLimits(customerId int64, amount int64, transactionTime time.Time) bool
}
