package main

import (
	"bufio"
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
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type application struct {
	loads         models.ILoads
	loadValidator validators.ILoadValidator
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
		loads:         &postgres.LoadModel{DB: dbConn},
		loadValidator: &validators.LoadValidator{},
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
		var tempLoad helpers.ImportLoad
		err = json.Unmarshal(scanner.Bytes(), &tempLoad)

		if err != nil {
			log.Printf("Unable to parse json. Line: %s. %s", scanner.Text(), err)
			continue
		}

		load, err := helpers.InputToLoad(tempLoad)

		if err != nil {
			log.Print(err)
			continue
		}

		err = a.withinLimits(&load)

		if err != nil {
			log.Print(err)
			continue
		}

		_, err = a.loads.Insert(&load)
		if err != nil {
			log.Printf("Unable to insert into loads table. %s", err)
			continue
		}

		outJson, err := json.Marshal(jsonOutput{
			Id:         strconv.FormatInt(load.TransactionId, 10),
			CustomerId: strconv.FormatInt(load.CustomerId, 10),
			Accepted:   load.Accepted,
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

func (a *application) withinLimits(load *models.Load) error {
	_, err := a.loads.GetByTransactionId(load.CustomerId, load.TransactionId)

	//Ignoring a second load with the same id on a customer
	if err == nil {
		return errors.New(fmt.Sprintf("duplicate transaction, %+v", load))
	} else if !errors.Is(err, models.ErrNoRecord) {
		return errors.New(fmt.Sprintf("error checking for duplicate record. %s", err))
	}

	startDate := time.Date(load.Time.Year(), load.Time.Month(), load.Time.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(load.Time.Year(), load.Time.Month(), load.Time.Day(), 23, 59, 59, 999, time.UTC)
	loadModels, err := a.loads.GetByCustomerTransactionsByDateRange(load.CustomerId, startDate, endDate)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			return nil
		} else {
			return err
		}
	}
	load.Accepted = a.loadValidator.LessThanThreeLoadsDaily(loadModels) &&
		a.loadValidator.LessThanFiveThousandLoadedDaily(loadModels, load) &&
		a.loadValidator.LessThanTwentyThousandLoadedWeekly(loadModels, load)
	return nil
}

type jsonOutput struct {
	Id         string `json:"id"`
	CustomerId string `json:"customer_id"`
	Accepted   bool   `json:"accepted"`
}
