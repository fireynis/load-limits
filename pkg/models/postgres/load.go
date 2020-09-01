package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fireynis/velocity_checker/pkg/models"
	"github.com/jackc/pgx/v4"
	"time"
)

type LoadModel struct {
	DB *pgx.Conn
}

//Get retrieves a load from the database based on its ID
func (m *LoadModel) Get(id int64) (*models.Load, error) {
	stmt := "SELECT id, customer_id, transaction_id, load_amount, transaction_time, accepted FROM loads WHERE id = $1"
	row := m.DB.QueryRow(context.Background(), stmt, id)
	load, err := m.scanModel(row)
	return load, err
}

//GetByTransactionId finds the transaction based on the customer and id of the request.
func (m *LoadModel) GetByTransactionId(customerId int64, transactionId int64) (*models.Load, error) {
	stmt := "SELECT id, customer_id, transaction_id, load_amount, transaction_time, accepted FROM loads WHERE customer_id = $1 and transaction_id = $2"
	row := m.DB.QueryRow(context.Background(), stmt, customerId, transactionId)
	load, err := m.scanModel(row)
	return load, err
}

func (m *LoadModel) GetByCustomerTransactionsByDateRange(customerId int64, startDate time.Time, endDate time.Time) ([]*models.Load, error) {
	stmt := "SELECT id, customer_id, transaction_id, load_amount, transaction_time, accepted FROM loads WHERE customer_id = $1 and transaction_time >= $2 and transaction_time <= $3"

	rows, err := m.DB.Query(context.Background(), stmt, customerId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	loadModels := make([]*models.Load, 0)

	for rows.Next() {
		var tempModel models.Load
		err := rows.Scan(&tempModel.Id, &tempModel.CustomerId, &tempModel.TransactionId, &tempModel.Amount, &tempModel.Time, &tempModel.Accepted)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, models.ErrNoRecord
			} else {
				return nil, err
			}
		}
		loadModels = append(loadModels, &tempModel)
	}
	return loadModels, nil
}

//Insert saves the record to the database
func (m *LoadModel) Insert(load *models.Load) (int64, error) {
	stmt := "INSERT INTO loads (customer_id, transaction_id, load_amount, transaction_time, accepted) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	var lastInsertId int64
	err := m.DB.QueryRow(context.Background(), stmt, load.CustomerId, load.TransactionId, load.Amount, load.Time, load.Accepted).Scan(&lastInsertId)
	if err != nil {
		return 0, err
	}
	load.Id = lastInsertId
	return lastInsertId, nil
}

func (m *LoadModel) Update(model *models.Load) error {
	stmt := "UPDATE loads SET customer_id = $1, transaction_id = $2, load_amount = $3, transaction_time = $4, accepted = $5 WHERE id = $6"

	//Using Exec as I don't need to know anything other than if it works, which the Error will determine
	_, err := m.DB.Exec(context.Background(), stmt, model.CustomerId, model.TransactionId, model.Amount, model.Time, model.Accepted, model.Id)
	return err
}

func (m *LoadModel) hasExceededWeeklyLoadLimit(customerId int64, transactionTime time.Time, transactionAmount int64) bool {
	startTime := time.Date(transactionTime.Year(), transactionTime.Month(), transactionTime.Day(), 0, 0, 0, 0, time.UTC)
	for startTime.Weekday() != time.Monday {
		startTime = startTime.AddDate(0, 0, -1)
	}
	loadModels, err := m.GetByCustomerTransactionsByDateRange(customerId, startTime, transactionTime)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		} else {
			//This should pretty well never happen but safer to reject it than accept it
			return true
		}
	}

	amount := int64(0)
	for _, load := range loadModels {
		amount += load.Amount
	}
	amount += transactionAmount
	if amount > 2000000 {
		return true
	}
	return false
}

//scanModel is a helper function to scan a row into a load struct.
func (m LoadModel) scanModel(row pgx.Row) (*models.Load, error) {
	load := &models.Load{}
	err := row.Scan(&load.Id, &load.CustomerId, &load.TransactionId, &load.Amount, &load.Time, &load.Accepted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return load, nil
}
