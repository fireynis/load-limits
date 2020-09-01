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
	loads models.ILoads
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

		customerId, transactionId, intAmount, err := a.sanitizeInput(&tempLoad)

		if err != nil {
			log.Print(err)
			continue
		}

		accepted, err := a.withinLimits(customerId, transactionId, intAmount, tempLoad.Time)

		if err != nil {
			log.Print(err)
			continue
		}

		_, err = a.loads.Insert(customerId, transactionId, intAmount, tempLoad.Time, accepted)
		if err != nil {
			log.Printf("Unable to insert into loads table. %s", err)
			continue
		}

		outJson, err := json.Marshal(jsonOutput{
			Id:         transactionId,
			CustomerId: customerId,
			Accepted:   accepted,
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

func (a *application) sanitizeInput(tempLoad *importLoad) (customerId, transactionId, loadAmount int64, err error) {

	transactionId, err = strconv.ParseInt(tempLoad.TransactionId, 10, 64)
	if err != nil {
		return 0, 0, 0, errors.New(fmt.Sprintf("unable to parse json. %s", err))
	}

	customerId, err = strconv.ParseInt(tempLoad.CustomerId, 10, 64)
	if err != nil {
		return 0, 0, 0, errors.New(fmt.Sprintf("unable to parse json. %s", err))
	}

	cleanedAmount, err := strconv.ParseFloat(strings.ReplaceAll(tempLoad.Amount, "$", ""), 64)
	if err != nil {
		return 0, 0, 0, errors.New(fmt.Sprintf("Unable to parse json.  %s", err))
	}
	loadAmount = int64(cleanedAmount * 100)
	return customerId, transactionId, loadAmount, nil
}

func (a *application) withinLimits(customerId int64, transactionId int64, amount int64, transactionTime time.Time) (bool, error) {
	_, err := a.loads.GetByTransactionId(customerId, transactionId)

	//Ignoring a second load with the same id on a customer
	if err == nil {
		return false, errors.New("duplicate transaction")
	} else if !errors.Is(err, models.ErrNoRecord) {
		errors.New(fmt.Sprintf("error checking for duplicate record. %s", err))
	}

	return a.loads.WithinLimits(customerId, amount, transactionTime), nil
}

type importLoad struct {
	TransactionId string    `json:"id"`
	CustomerId    string    `json:"customer_id"`
	Amount        string    `json:"load_amount"`
	Time          time.Time `json:"time"`
}

type jsonOutput struct {
	Id         int64 `json:"id"`
	CustomerId int64 `json:"customer_id"`
	Accepted   bool  `json:"accepted"`
}
