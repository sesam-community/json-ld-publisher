package main

import (
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	sesamAPI string
	sesamJWT string
	port     int
	datasets []string
)

func loadConfig() {
	// get config
	log.Println("Loading Config ---------------------- ")
	sesamAPI = os.Getenv("SESAM_API")
	sesamJWT = os.Getenv("SESAM_JWT")
	port, _ = strconv.Atoi(os.Getenv("SERVICE_PORT"))
	var datasetNames = os.Getenv("SESAM_DATASETS")
	var splitNames = strings.Split(datasetNames, ";")
	for _, n := range splitNames {
		datasets = append(datasets, n)
	}

	log.Println("API: " + sesamAPI)
	log.Println("JWT: " + sesamJWT)
	log.Println("PORT: " + strconv.Itoa(port))
	log.Println("Datasets: " + datasetNames)
	log.Println("Loaded Config  ---------------------- ")

}

func main() {
	log.Println("Starting JSON-LD Publisher")
	loadConfig()
}
