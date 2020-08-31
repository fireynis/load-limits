package main

import (
	"bufio"
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
	"os"
	"path/filepath"
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

	var flagPathToFile = flag.String("file", "", "The path to the file to be read in. Overrides the .env INPUT_FILE.")
	var flagPathToOutFile = flag.String("output_file", "", "The path to the file to be read in. Overrides the .env OUTPUT_FILE. Leave both blank to output to console")
	var flagDsn = flag.String("dsn", "", "The connection string for the postgres database. Overrides the .env DATABASE_DSN")
	flag.Parse()

	//I don't really need the env vars since the flags can override them.
	//Only if the flags are empty are they necessary
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var pathToFile string
	if len(*flagPathToFile) >= 1 {
		pathToFile = *flagPathToFile
	} else if len(os.Getenv("INPUT_FILE")) >= 1 {
		pathToFile = os.Getenv("INPUT_FILE")
	} else {
		log.Fatalf("A file path is requried")
	}

	var pathToOutFile string
	if len(*flagPathToOutFile) >= 1 {
		pathToOutFile = *flagPathToOutFile
	} else if len(os.Getenv("OUTPUT_FILE")) >= 1 {
		pathToOutFile = os.Getenv("OUTPUT_FILE")
	}

	var dsn string
	if len(*flagDsn) >= 1 {
		dsn = *flagDsn
	} else if len(os.Getenv("DATABASE_DSN")) >= 1 {
		dsn = os.Getenv("DATABASE_DSN")
	} else {
		log.Fatalf("A databse DSN is required")
	}

	dbConn, err := pgx.Connect(context.Background(), dsn)

	if err != nil {
		log.Fatalf("Unable to connect to database. %s", err)
	}
	defer dbConn.Close(context.Background())

	app := &application{
		loads: &postgres.LoadModel{DB: dbConn},
	}

	app.parseFile(pathToFile, pathToOutFile)
}

func (a *application) parseFile(filePath, pathToOutFile string) {
	cleanPath, err := filepath.Abs(filePath)

	if err != nil {
		log.Fatalf("Unable to clean file path. Error: %s", err)
	}

	file, err := os.Open(cleanPath)

	if err != nil {
		log.Fatalf("Unable to open file. Error: %s", err)
	}
	defer file.Close()

	var outToFile bool
	var output *os.File
	if len(pathToOutFile) > 1 {
		outToFile = true

		cleanPath, err := filepath.Abs(pathToOutFile)

		if err != nil {
			log.Fatalf("Unable to clean file path. Error: %s", err)
		}

		output, err = os.OpenFile(cleanPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			log.Fatalf("Unable to open output file. Error: %s", err)
		}
		defer output.Close()
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var tempLoad importLoad
		err = json.Unmarshal(scanner.Bytes(), &tempLoad)

		if err != nil {
			log.Printf("Unable to parse json. Line: %s. %s", scanner.Text(), err)
			continue
		}

		transactionId, err := strconv.ParseInt(tempLoad.TransactionId, 10, 64)
		if err != nil {
			log.Printf("Unable to parse json. Line: %s. %s", scanner.Text(), err)
			continue
		}

		customerId, err := strconv.ParseInt(tempLoad.CustomerId, 10, 64)
		if err != nil {
			log.Printf("Unable to parse json. Line: %s. %s", scanner.Text(), err)
			continue
		}

		cleanedAmount, err := strconv.ParseFloat(strings.ReplaceAll(tempLoad.Amount, "$", ""), 64)
		if err != nil {
			log.Printf("Unable to parse json. Line: %s. %s", scanner.Text(), err)
			continue
		}
		intAmount := int64(cleanedAmount * 100)

		_, err = a.loads.GetByTransactionId(customerId, transactionId)

		//Ignoring a second load with the same id on a customer
		if err == nil {
			log.Printf("Duplicate transaction. %s", scanner.Text())
			continue
		} else if !errors.Is(err, models.ErrNoRecord) {
			log.Printf("Error checking for duplicate record. %s", err)
		}

		tempLoad.Accepted = a.loads.WithinLimits(customerId, intAmount, tempLoad.Time)

		_, err = a.loads.Insert(customerId, transactionId, intAmount, tempLoad.Time, tempLoad.Accepted)
		if err != nil {
			log.Printf("Unable to insert into loads table. %s", err)
			continue
		}
		outJson, err := json.Marshal(jsonOutput{
			Id:         transactionId,
			CustomerId: customerId,
			Accepted:   tempLoad.Accepted,
		})
		if err != nil {
			log.Printf("Unable to marshall output json. %s", err)
			continue
		}

		if outToFile {
			_, _ = output.WriteString(string(outJson) + "\n")
		} else {
			fmt.Printf("%s\n", outJson)
		}
	}
}

type importLoad struct {
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
