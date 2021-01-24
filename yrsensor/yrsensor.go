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

func addLocationsToStatus(ds *statushttp.DaemonStatus, locs Locations) {
	locs.mu.RLock()
	defer locs.mu.RUnlock()
	for _, loc := range locs.Locations {
		ds.AddLocation(loc.Id)
	}
}

func Shutdown(pc *PollerConfig, ec *EmitterConfig) {
	pc.mu.Lock()
	pc.Control = false
	pc.mu.Unlock()
	ec.mu.Lock()
	ec.Control = false
	ec.mu.Unlock()
	<-pc.Finished
	<-ec.Finished
}

func Run(userAgent string, apiUrl string, apiVersion string, emitterInterval time.Duration,
	locationFileLocation string, awsRegion string, awsTimeseriesDbname string) {
	var locations Locations
	var err error
	var forecastsCache ObservationCache

	forecastsCache.observations = make(map[string]ObservationTimeSeries)

	setupLogging(log.DebugLevel)
	locations.mu.Lock()
	locations.Locations, err = readLocationsFromPath(locationFileLocation)
	locations.mu.Unlock()
	if err != nil {
		log.Errorf("could not parse location file: %v", err.Error())
		log.Error("Example location file:")
		log.Error(locationFileExample())
		log.Fatal("Aborting")
	}
	locations.mu.RLock()
	for _, loc := range locations.Locations {
		log.Debugf("Polling location set: %s (%f, %f)", loc.Id, loc.Lat, loc.Long)
	}
	locations.mu.RUnlock()
	var ds = statushttp.Run(":8080")

	var pc = PollerConfig{
		Control:             true,
		Finished:            make(chan bool),
		ApiUrl:              apiUrl,
		ApiVersion:          apiVersion,
		UserAgent:           userAgent,
		Locations:           locations,
		ObservationCachePtr: &forecastsCache,
		DaemonStatusPtr:     ds,
	}

	var ec = EmitterConfig{
		Control:             true,
		Finished:            make(chan bool),
		EmitterInterval:     emitterInterval,
		Locations:           locations,
		ObservationCachePtr: &forecastsCache,
		AwsRegion:           awsRegion,
		AwsTimestreamDbname: awsTimeseriesDbname,
		DaemonStatusPtr:     ds,
	}

	addLocationsToStatus(ds, locations)

	// There is likely a prettier way to do this:
	go poller(&pc)
	go emitter(&ec)
	// pollerControl = false
	// Listen for signals:
	mainControl := make(chan os.Signal)
	signal.Notify(mainControl, os.Interrupt, syscall.SIGINT)
	signal.Notify(mainControl, os.Interrupt, syscall.SIGTERM)
	log.Info("Daemon running")
	<-mainControl
	Shutdown(&pc, &ec)
	log.Info("signal caught, winding down gracefully.")
	log.Info("end of program")
	os.Exit(0)
}
