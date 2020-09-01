package helpers

import (
	"errors"
	"fireynis/velocity_checker/pkg/models"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func InputToLoad(input ImportLoad) (load models.Load, err error) {
	load.TransactionId, err = strconv.ParseInt(input.TransactionId, 10, 64)
	if err != nil {
		return models.Load{}, errors.New(fmt.Sprintf("unable to parse json. %s", err))
	}

	load.CustomerId, err = strconv.ParseInt(input.CustomerId, 10, 64)
	if err != nil {
		return models.Load{}, errors.New(fmt.Sprintf("unable to parse json. %s", err))
	}

	cleanedAmount, err := strconv.ParseFloat(strings.ReplaceAll(input.Amount, "$", ""), 64)
	if err != nil {
		return models.Load{}, errors.New(fmt.Sprintf("Unable to parse json.  %s", err))
	}
	load.Amount = int64(cleanedAmount * 100)
	load.Time = input.Time
	return load, nil
}

type ImportLoad struct {
	TransactionId string    `json:"id"`
	CustomerId    string    `json:"customer_id"`
	Amount        string    `json:"load_amount"`
	Time          time.Time `json:"time"`
}
