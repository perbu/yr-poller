package yrsensor

import (
	"github.com/perbu/yrpoller/statushttp"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func setupLogging(level log.Level) {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(level)
}

func addLocationsToStatus(ds *statushttp.DaemonStatus, locs []Location) {
	for _, loc := range locs {
		ds.AddLocation(loc.Id)
	}
}

func Run(userAgent string, apiUrl string, apiVersion string, emitterInterval time.Duration, locationFileLocation string) {
	var locations []Location
	var forecastsCache ObservationCache
	var err error

	setupLogging(log.DebugLevel)
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
	daemonStats := statushttp.Run(":8080")
	forecastsCache.observations = make(map[string]ObservationTimeSeries)
	addLocationsToStatus(daemonStats, locations)
	pollerControl := true
	pollerFinished := make(chan bool)
	emitterControl := true
	emitterFinished := make(chan bool)

	// There is likely a prettier way to do this:
	go poller(&pollerControl, pollerFinished, apiUrl, apiVersion, userAgent, locations, &forecastsCache, daemonStats)
	go emitter(&emitterControl, emitterFinished, emitterInterval, locations, &forecastsCache, daemonStats)
	// pollerControl = false
	// Listen for signals:
	mainControl := make(chan os.Signal)
	signal.Notify(mainControl, os.Interrupt, syscall.SIGINT)
	signal.Notify(mainControl, os.Interrupt, syscall.SIGTERM)
	log.Info("Daemon running")
	<-mainControl
	log.Info("signal caught, winding down gracefully.")
	pollerControl = false
	emitterControl = false
	<-pollerFinished
	<-emitterFinished
	log.Info("End of program")
	os.Exit(0)
}
