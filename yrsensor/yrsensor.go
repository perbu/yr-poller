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
	// You could enable this for a bit more verbose logging.
	// log.SetReportCaller(true)
}

func addLocationsToStatus(ds *statushttp.DaemonStatus, locs Locations) {
	for _, loc := range locs.Locations {
		ds.AddLocation(loc.Id)
	}
}

func Run(userAgent string, apiUrl string, apiVersion string, emitterInterval time.Duration,
	locationFileLocation string, awsRegion string, awsTimeseriesDbname string, bindAddress string) {
	var locations Locations
	var err error
	var forecastsCache ObservationCache

	forecastsCache.observations = make(map[string]ObservationTimeSeries)

	setupLogging(log.DebugLevel)
	locations.Locations, err = readLocationsFromPath(locationFileLocation)

	if err != nil {
		log.Errorf("could not parse location file: %v", err.Error())
		log.Error("Example location file:")
		log.Error(locationFileExample())
		log.Fatal("Aborting")
	}
	for _, loc := range locations.Locations {
		log.Debugf("Polling location set: %s (%f, %f)", loc.Id, loc.Lat, loc.Long)
	}
	var ds = statushttp.Run(bindAddress)
	var tsReqChannel = make(chan TimeSeriesRequest)

	var pc = PollerConfig{
		Finished:            make(chan bool),
		ApiUrl:              apiUrl,
		ApiVersion:          apiVersion,
		UserAgent:           userAgent,
		Locations:           locations,
		ObservationCachePtr: &forecastsCache,
		DaemonStatusPtr:     &ds,
		TsRequestChannel:    tsReqChannel,
	}

	var ec = EmitterConfig{
		Finished:            make(chan bool),
		EmitterInterval:     emitterInterval,
		Locations:           locations,
		ObservationCachePtr: &forecastsCache,
		AwsRegion:           awsRegion,
		AwsTimestreamDbname: awsTimeseriesDbname,
		DaemonStatusPtr:     &ds,
		TsRequestChannel:    tsReqChannel,
	}

	addLocationsToStatus(&ds, locations)

	go poller(&pc)
	go emitter(&ec)
	// pollerControl = false
	// Listen for signals:
	mainControl := make(chan os.Signal)
	signal.Notify(mainControl, os.Interrupt, syscall.SIGINT)
	signal.Notify(mainControl, os.Interrupt, syscall.SIGTERM)
	log.Info("Daemon running")
	<-mainControl // block and wait for signals.
	log.Info("signal caught, winding down gracefully.")
	pc.Finished <- true
	<-pc.Finished
	ec.Finished <- true
	<-ec.Finished
	log.Info("end of program")
	os.Exit(0)
}
