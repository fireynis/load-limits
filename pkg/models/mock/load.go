package mock

import (
	"fireynis/velocity_checker/pkg/models"
	"time"
)

var loads = []*models.Load{
	{
		Id:            1,
		TransactionId: 1,
		CustomerId:    1,
		Amount:        250000,
		Time:          time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Accepted:      true,
	},
	{
		Id:            2,
		TransactionId: 2,
		CustomerId:    1,
		Amount:        250000,
		Time:          time.Date(2000, 1, 1, 6, 0, 0, 0, time.UTC),
		Accepted:      true,
	},
	{
		Id:            3,
		TransactionId: 3,
		CustomerId:    1,
		Amount:        250000,
		Time:          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Accepted:      true,
	},
	{
		Id:            4,
		TransactionId: 1,
		CustomerId:    2,
		Amount:        500000,
		Time:          time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Accepted:      true,
	},
	{
		Id:            5,
		TransactionId: 1,
		CustomerId:    3,
		Amount:        2000000,
		Time:          time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Accepted:      true,
	},
	{
		Id:            6,
		TransactionId: 1,
		CustomerId:    4,
		Amount:        250000,
		Time:          time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Accepted:      true,
	},
}

type Load struct{}

func (m *Load) Get(id int64) (*models.Load, error) {
	return loads[id-1], nil
}

func (m *Load) GetByTransactionId(customerId int64, transactionId int64) (*models.Load, error) {
	for _, load := range loads {
		if load.CustomerId == customerId && load.TransactionId == transactionId {
			return load, nil
		}
	}
	return nil, models.ErrNoRecord
}

func (m *Load) GetByCustomerTransactionsByDateRange(customerId int64, startDate time.Time, endDate time.Time) ([]*models.Load, error) {
	switch customerId {
	case 1:
		return loads[0:2], nil
	case 2:
		return []*models.Load{loads[3]}, nil
	case 3:
		return []*models.Load{loads[4]}, nil
	case 4:
		return []*models.Load{loads[5]}, nil
	}
	return []*models.Load{}, models.ErrNoRecord
}

func (m *Load) Insert(load *models.Load) (int64, error) {
	return 5, nil
}

func (m *Load) Update(model *models.Load) error {
	return nil
}
