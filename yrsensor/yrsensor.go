package yrsensor

import (
	"log"
	"time"
)

var locations []Location
var forecastsCache ObservationCache
var ready bool

func Run(userAgent string, apiUrl string, apiVersion string, emitterInterval time.Duration, locationFileLocation string) {
	log.Print("Yr poller 0.0.1")
	locations = readLocations(locationFileLocation)
	for _, loc := range locations {
		log.Printf("Polling location set: %s (%f, %f)", loc.Id, loc.Lat, loc.Long)
	}
	forecastsCache.observations = make(map[string]ObservationTimeSeries)
	pollerControl := true
	pollerFinished := make(chan bool)
	emitterControl := true
	emitterFinished := make(chan bool)
	go poller(&pollerControl, pollerFinished, apiUrl, apiVersion, userAgent)
	go emitter(&emitterControl, emitterFinished, emitterInterval)
	time.Sleep(5 * time.Second)
	// pollerControl = false
	log.Print("Waiting for poller to finish")
	<-pollerFinished
}
