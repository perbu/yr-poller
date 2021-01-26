package main

import (
	"flag"
	"github.com/perbu/yrpoller/yrsensor"
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
const AWS_REGION = "eu-west-1"
const DBNAME = "yrpoller-fjas"
const BINDADDRESS = ":8080"

func main() {
	// func run(userAgentPtr string, apiUrlPtr string, apiVersionPtr string, emitterIntervalPtr time.Duration, locationFileLocation string) {
	locationPathPtr := flag.String("locationsfile", LOCATIONFILEPATH, "JSON file containing locations")
	userAgentPtr := flag.String("user-agent", CLIENT_ID, "User-agent to use")
	apiUrlPtr := flag.String("api-url", API_URL, "Baseurl for Yr API")
	apiVersionPtr := flag.String("api-version", API_VERSION, "API version to use. Appended to URL")
	emitterIntervalPtr := flag.Duration("interval", EMITTERINTERVAL, "How often to emit data")
	awsRegionPtr := flag.String("aws-region", AWS_REGION, "AWS region")
	awsTimeseriesDbnamePtr := flag.String("dbname", DBNAME, "DB name in AWS Timestream")
	bindAddressPtr := flag.String("bind", BINDADDRESS, "bind address")

	flag.Parse()
	// Note: these are all pointers.
	yrsensor.Run(*userAgentPtr, *apiUrlPtr, *apiVersionPtr,
		*emitterIntervalPtr, *locationPathPtr, *awsRegionPtr,
		*awsTimeseriesDbnamePtr, *bindAddressPtr)
}
