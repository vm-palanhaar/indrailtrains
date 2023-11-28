package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Train struct {
	Name string
	No   string
}

func getRailTrainsApi() []Train {
	// calling rail stations API
	log.Printf("Requesting data from API")
	req, err := http.Get(os.Getenv("INDRAIL_TRAINS_API"))
	if err != nil {
		log.Fatal(err)
	}

	// response status code check 200
	log.Printf("Response status code: %s", req.Status)
	if req.StatusCode != http.StatusOK {
		log.Fatalf("API failed: %s", req.Status)
	}

	// read doc from response
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	docContentString := string(body)

	// format string to create a list of stations obj
	docContentString = strings.ReplaceAll(docContentString, "[", "")
	docContentString = strings.ReplaceAll(docContentString, "]", "")
	docContentString = strings.ReplaceAll(docContentString, "\"", "")
	railTrains := strings.Split(docContentString, ",")
	log.Printf("Total no. of rail trains (API): %d", len(railTrains))
	railTrainList := make([]Train, len(railTrains))
	for i, station := range railTrains {
		railTrainList[i].No = strings.Split(station, " - ")[0]
		railTrainList[i].Name = strings.Split(station, " - ")[1]
	}
	return railTrainList
}

func getRailTrainsDb(db *sql.DB) []Train {
	table := os.Getenv("TABLE_INDRAIL_TRAINS")
	// execute query
	log.Println("EXECUTING QUERY to train no, train name")
	rows, err := db.Query(fmt.Sprintf("SELECT train_no, train_name FROM %s", table))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	trainsDb := []Train{}
	for rows.Next() {
		rowData := Train{}
		err := rows.Scan(&rowData.No, &rowData.Name)
		if err != nil {
			log.Fatal(err)
		}
		trainsDb = append(trainsDb, rowData)
	}
	log.Printf("Total no. of rail trains (DB): %d", len(trainsDb))
	return trainsDb
}

func (trainDb Train) updateRailTrainDb(db *sql.DB, trainApi Train) {
	table := os.Getenv("TABLE_INDRAIL_TRAINS")
	if trainApi.No == trainDb.No && trainApi.Name == trainDb.Name {
		// DO NOTHING
	} else if trainApi.No == trainDb.No && trainApi.Name != trainDb.Name {
		stmt, err := db.Prepare(
			fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2",
				table, "train_name", "train_no"))
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(trainApi.Name, trainDb.No)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("---UPDATE SUCCESS [%s - %s]---", trainApi.Name, trainApi.No)
	}
}

func railTrains(db *sql.DB) {
	log.Println("START -> GET ALL rail trains API functionality <- START")
	trainsApi := getRailTrainsApi()
	log.Println("END -> GET ALL rail trains API functionality <- END")

	log.Println("START -> GET ALL rail trains DB functionality <- START")
	trainsDb := getRailTrainsDb(db)
	log.Println("END -> GET ALL rail trains DB functionality <- END")

	/*
		Iterate through each train to check for any changes in train name.
		1. If unchanged, do nothing
		2. If changed [train name], update train name w.r.t to train no
	*/
	if len(trainsApi) == len(trainsDb) {
		for _, sapi := range trainsApi {
			for _, sdb := range trainsDb {
				sdb.updateRailTrainDb(db, sapi)
			}
		}
	} else if len(trainsApi) < len(trainsDb) {
		/*
			Send mail to admin user(s)
		*/
		trains := []string{}
		for _, tdb := range trainsDb {
			checkTrain := true
			for _, tapi := range trainsApi {
				tdb.updateRailTrainDb(db, tapi)
				if tapi.No == tdb.No {
					checkTrain = false
					break
				}
			}
			if checkTrain {
				trains = append(trains, fmt.Sprintf("%s - %s", tdb.No, tdb.Name))
			}
		}
		log.Print("<---Rail Trains in DB--->")
		for i, train := range trains {
			log.Printf("%d. %s", i+1, train)
		}
	} else if len(trainsApi) > len(trainsDb) {
		table := os.Getenv("TABLE_INDRAIL_TRAINS")
		log.Print("<---Rail Trains in API--->")
		for i, tapi := range trainsApi {
			checkTrain := true
			for _, tdb := range trainsDb {
				tdb.updateRailTrainDb(db, tapi)
				if tapi.No == tdb.No {
					checkTrain = false
					break
				}
			}
			if checkTrain {
				stmt, err := db.Prepare(
					fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES ($1, $2)",
						table, "train_no", "train_name"))
				if err != nil {
					log.Fatal(err)
				}
				defer stmt.Close()
				_, err = stmt.Exec(tapi.No, tapi.Name)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("%d. INSERT SUCCESS [%s - %s]", i+1, tapi.Name, tapi.No)
			}
		}
	}

}
