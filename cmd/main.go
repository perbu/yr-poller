package main

import (
	yrsensor "github.com/perbu/yrpoller/yrsensor"
	"time"
)

/*
  Poller deamon.

*/

const API_URL = "https://api.met.no/weatherapi"
const API_VERSION = "2.0"
const CLIENT_ID = "perbu-poller"
const EMITTERINTERVAL = time.Minute * 10
const LOCATIONFILEPATH = "locations.json"

func main() {
	// func run(userAgent string, apiUrl string, apiVersion string, emitterInterval time.Duration, locationFileLocation string) {
	yrsensor.Run(CLIENT_ID, API_URL, API_VERSION, EMITTERINTERVAL, LOCATIONFILEPATH)
}
