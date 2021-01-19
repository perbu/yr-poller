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
	var err error
	setupLogging()
	log.Info("Yr poller 0.0.1")
	locations, err = readLocationsFromPath(locationFileLocation)
	if err != nil {
		log.Errorf("could not parse location file: %v", err.Error())
		log.Error("Example location file:")
		log.Error(locationFileExample())
		log.Fatal("Aborting")
	}
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
