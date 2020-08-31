package main

import (
	"context"
	"encoding/json"
	"errors"
	"fireynis/velocity_checker/pkg/models"
	"fireynis/velocity_checker/pkg/models/postgres"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type application struct {
	loads interface {
		Get(id int64) (*models.Load, error)
		GetByTransactionId(customerId int64, transactionId int64) (*models.Load, error)
		GetByCustomerTransactionsByDateRange(customerId int64, startDate time.Time, endDate time.Time) ([]*models.Load, error)
		Insert(customerId int64, transactionId int64, amount int64, transactionTime time.Time, accepted bool) (int64, error)
		Update(model *models.Load) error
		WithinLimits(customerId int64, amount int64, transactionTime time.Time) bool
	}
}

func main() {

	var flagDsn = flag.String("dsn", "", "The connection string for the postgres database. Overrides the .env DATABASE_DSN")
	var flagPort = flag.String("port", "8080", "Sets the port to listen on for the server. Can be set in .env which overrides this option. Defaults to 8080")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var dsn string
	if len(*flagDsn) >= 1 {
		dsn = *flagDsn
	} else if len(os.Getenv("DATABASE_DSN")) >= 1 {
		dsn = os.Getenv("DATABASE_DSN")
	} else {
		log.Fatalf("A databse DSN is required")
	}

	var port string
	if len(os.Getenv("APP_PORT")) >= 1 {
		port = os.Getenv("APP_PORT")
	} else if len(*flagPort) >= 1 {
		port = *flagPort
	} else {
		log.Fatalf("A port is required.")
	}

	dbConn, err := pgx.Connect(context.Background(), dsn)

	if err != nil {
		log.Fatalf("Unable to connect to database. %s", err)
	}
	defer dbConn.Close(context.Background())

	app := &application{
		loads: &postgres.LoadModel{DB: dbConn},
	}

	router := http.NewServeMux()

	router.HandleFunc("/", app.parseLoad)

	err = http.ListenAndServe(":"+port, router)
	log.Fatal(err)
}

func (a *application) parseLoad(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	decoder := json.NewDecoder(r.Body)

	var inData loadJson

	err := decoder.Decode(&inData)

	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse json"), 403)
		return
	}

	transactionId, err := strconv.ParseInt(inData.TransactionId, 10, 64)
	if err != nil {
		log.Printf("Unable to parse json. %s.", err)
		http.Error(w, fmt.Sprintf("Unable to parse json. %s.", err), 403)
		return
	}

	customerId, err := strconv.ParseInt(inData.CustomerId, 10, 64)
	if err != nil {
		log.Printf("Unable to parse json. %s", err)
		http.Error(w, fmt.Sprintf("Unable to parse json. %s", err), 403)
		return
	}

	cleanedAmount, err := strconv.ParseFloat(strings.ReplaceAll(inData.Amount, "$", ""), 64)
	if err != nil {
		log.Printf("Unable to parse json. %s", err)
		http.Error(w, fmt.Sprintf("Unable to parse json. %s", err), 403)
		return
	}
	intAmount := int64(cleanedAmount * 100)

	_, err = a.loads.GetByTransactionId(customerId, transactionId)
	//Ignoring a second load with the same id on a customer
	if err == nil {
		_, _ = w.Write([]byte("Record already exists"))
		http.Error(w, fmt.Sprintf("Record already exists"), 403)
		return
	} else if !errors.Is(err, models.ErrNoRecord) {
		log.Printf("Error checking for duplicate record. %s", err)
		http.Error(w, fmt.Sprintf("Error checking for duplicate record. %s", err), 500)
		return
	}

	inData.Accepted = a.loads.WithinLimits(customerId, intAmount, inData.Time)

	_, err = a.loads.Insert(customerId, transactionId, intAmount, inData.Time, inData.Accepted)
	if err != nil {
		log.Printf("Unable to insert into loads table. %s", err)
		http.Error(w, fmt.Sprintf("Unable to insert into loads table. %s", err), 500)
		return
	}
	outJson, err := json.Marshal(jsonOutput{
		Id:         transactionId,
		CustomerId: customerId,
		Accepted:   inData.Accepted,
	})
	if err != nil {
		log.Printf("Unable to marshall output json. %s", err)
		http.Error(w, fmt.Sprintf("Unable to marshall output json. %s", err), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(outJson)
}

type loadJson struct {
	TransactionId string    `json:"id"`
	CustomerId    string    `json:"customer_id"`
	Amount        string    `json:"load_amount"`
	Time          time.Time `json:"time"`
	Accepted      bool      `json:"-"`
}

type jsonOutput struct {
	Id         int64 `json:"id"`
	CustomerId int64 `json:"customer_id"`
	Accepted   bool  `json:"accepted"`
}
