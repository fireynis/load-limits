package main

import (
	"context"
	"encoding/json"
	"errors"
	"fireynis/velocity_checker/pkg/helpers"
	"fireynis/velocity_checker/pkg/models"
	"fireynis/velocity_checker/pkg/models/postgres"
	"fireynis/velocity_checker/pkg/validators"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"
)

type application struct {
	loads         models.ILoads
	loadValidator validators.ILoadValidator
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
		loads:         &postgres.LoadModel{DB: dbConn},
		loadValidator: &validators.LoadValidator{},
	}

	err = http.ListenAndServe(":"+port, app.routes())
	log.Fatal(err)
}

func (a *application) parseLoad(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	decoder := json.NewDecoder(r.Body)

	var inData helpers.ImportLoad

	err := decoder.Decode(&inData)

	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse json"), 400)
		return
	}

	load, err := helpers.InputToLoad(inData)

	if err != nil {
		http.Error(w, fmt.Sprintf("Data in is incorrect. %s", err), 400)
		return
	}

	_, err = a.loads.GetByTransactionId(load.CustomerId, load.TransactionId)
	//Ignoring a second load with the same id on a customer
	if err == nil {
		http.Error(w, "Record already exists", 400)
		return
	} else if !errors.Is(err, models.ErrNoRecord) {
		log.Printf("Error checking for duplicate record. %s", err)
		http.Error(w, fmt.Sprintf("Error checking for duplicate record. %s", err), 500)
		return
	}

	startDate := time.Date(load.Time.Year(), load.Time.Month(), load.Time.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(load.Time.Year(), load.Time.Month(), load.Time.Day(), 23, 59, 59, 999, time.UTC)
	loadModels, err := a.loads.GetByCustomerTransactionsByDateRange(load.CustomerId, startDate, endDate)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			load.Accepted = true
		} else {
			http.Error(w, fmt.Sprintf("Error retrieving data. %s", err), 500)
			return
		}
	}
	load.Accepted = a.loadValidator.LessThanThreeLoadsDaily(loadModels) &&
		a.loadValidator.LessThanFiveThousandLoadedDaily(loadModels, &load) &&
		a.loadValidator.LessThanTwentyThousandLoadedWeekly(loadModels, &load)

	_, err = a.loads.Insert(&load)
	if err != nil {
		log.Printf("Unable to insert into loads table. %s", err)
		http.Error(w, fmt.Sprintf("Unable to insert into loads table. %s", err), 500)
		return
	}
	outJson, err := json.Marshal(jsonOutput{
		Id:         load.TransactionId,
		CustomerId: load.CustomerId,
		Accepted:   load.Accepted,
	})
	if err != nil {
		log.Printf("Unable to marshall output json. %s", err)
		http.Error(w, fmt.Sprintf("Unable to marshall output json. %s", err), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(outJson)
}

type jsonOutput struct {
	Id         int64 `json:"id"`
	CustomerId int64 `json:"customer_id"`
	Accepted   bool  `json:"accepted"`
}
