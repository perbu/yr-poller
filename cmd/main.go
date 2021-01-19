package main

import (
	"flag"
	yrsensor "github.com/perbu/yrpoller/yrsensor"
	"time"
)

/*
  Poller daemon.
*/

const API_URL = "https://api.met.no/weatherapi"
const API_VERSION = "2.0"
const CLIENT_ID = "yr-poller"
const EMITTERINTERVAL = time.Minute * 10
const LOCATIONFILEPATH = "locations.json"

func main() {
	// func run(userAgent string, apiUrl string, apiVersion string, emitterInterval time.Duration, locationFileLocation string) {
	locationPath := flag.String("locationsfile", LOCATIONFILEPATH, "JSON file containing locations")
	userAgent := flag.String("user-agent", CLIENT_ID, "User-agent to use")
	apiUrl := flag.String("api-url", API_URL, "Baseurl for Yr API")
	apiVersion := flag.String("api-version", API_VERSION, "API version to use. Appended to URL")
	emitterInterval := flag.Duration("interval", EMITTERINTERVAL, "How often to emit data")
	flag.Parse()
	// Note: these are all pointers.
	yrsensor.Run(*userAgent, *apiUrl, *apiVersion, *emitterInterval, *locationPath)
}
