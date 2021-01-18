package yrsensor

import (
	log "github.com/sirupsen/logrus"
	"time"
)

func setupLogging() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
}

func Run(userAgent string, apiUrl string, apiVersion string, emitterInterval time.Duration, locationFileLocation string) {
	var locations []Location
	var forecastsCache ObservationCache

	setupLogging()
	log.Info("Yr poller 0.0.1")
	locations = readLocations(locationFileLocation)
	for _, loc := range locations {
		log.Debugf("Polling location set: %s (%f, %f)", loc.Id, loc.Lat, loc.Long)
	}
	forecastsCache.observations = make(map[string]ObservationTimeSeries)
	pollerControl := true
	pollerFinished := make(chan bool)
	emitterControl := true
	emitterFinished := make(chan bool)
	go poller(&pollerControl, pollerFinished, apiUrl, apiVersion, userAgent, locations, &forecastsCache)
	go emitter(&emitterControl, emitterFinished, emitterInterval, locations, &forecastsCache)
	time.Sleep(5 * time.Second)
	// pollerControl = false
	log.Info("Daemon running")
	<-pollerFinished
	<-emitterFinished
}
