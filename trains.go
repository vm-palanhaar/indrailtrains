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
	fmt.Print(railTrainList)
	return railTrainList
}

func railTrains(db *sql.DB) {
	log.Println("START -> GET ALL rail trains API functionality <- START")
	getRailTrainsApi()
	log.Println("END -> GET ALL rail trains API functionality <- END")
}
