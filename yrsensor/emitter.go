package yrsensor

import (
	"fmt"
	"github.com/perbu/yrpoller/timestream"
	log "github.com/sirupsen/logrus"
	"time"
)

func interpolateObservations(first *Observation, last *Observation, when time.Time) Observation {
	var obs Observation
	timeDelta := last.Time.Sub(first.Time).Seconds() // Typically 60mins
	howFarInto := when.Sub(first.Time).Seconds()
	factor := float64(howFarInto) / float64(timeDelta)
	obs.Time = when
	// Interpolating here:
	obs.AirTemperature = last.AirTemperature*factor + first.AirTemperature*(1.0-factor)
	obs.AirPressureAtSeaLevel = last.AirPressureAtSeaLevel*factor + first.AirPressureAtSeaLevel*(1.0-factor)
	return obs
}

// Emit data
func emitLocation(tsconfig timestream.TimestreamState, location Location,
	obsCache *ObservationCache, when time.Time) {
	var obs Observation
	obsCache.mu.RLock()
	defer obsCache.mu.RUnlock()
	ts := obsCache.observations[location.Id].ts
	firstAfter := 0

	// Find out where we are in the time series.
	for i := range ts {
		if ts[i].Time.After(when) {
			firstAfter = i
			break
		}
	}
	// First measurement is still in the future so we can't interpolate:
	if firstAfter == 0 {
		obs.Time = ts[0].Time
		obs.AirTemperature = ts[0].AirTemperature
		obs.AirPressureAtSeaLevel = ts[0].AirPressureAtSeaLevel
	} else {
		// Interpolate the two relevant measurements
		last := ts[firstAfter]
		first := ts[firstAfter-1]
		obs = interpolateObservations(&first, &last, when)
	}
	// add the Id (place)
	obs.Id = location.Id

	// jsonData, err := json.MarshalIndent(obs, "TS: ", "  ")
	tsconfig.MakeObservation(timestream.TimestreamEntry{
		Time:      obs.Time,
		SensorId:  obs.Id,
		TableName: "air_temperature",
		Value:     fmt.Sprintf("%v", obs.AirTemperature),
	})
	tsconfig.MakeObservation(timestream.TimestreamEntry{
		Time:      obs.Time,
		SensorId:  obs.Id,
		TableName: "air_pressure_at_sealevel",
		Value:     fmt.Sprintf("%v", obs.AirPressureAtSeaLevel),
	})
	return
}

// waits for observations to arrive. Returns true or false
// false if not enough observations are present.
// true if the number of obs matches the fc cache.
func waitForObservations(fc *ObservationCache, locs *Locations) bool {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if len(fc.observations) == len(locs.Locations) {
		log.Debug("Observations are present.")
		return true
	} else {
		log.Debug("Observations are not yet present.")
	}
	return false
}

func emitter(config *EmitterConfig) {
	log.Info("Starting emitter")

	tsconfig := timestream.Factory(config.AwsRegion, config.AwsTimestreamDbname)

	err := tsconfig.CheckAndCreateTables()
	if err != nil {
		panic(err.Error())
	}
	for waitForObservations(config.ObservationCachePtr, &config.Locations) == false {
		time.Sleep(100 * time.Millisecond)
	}

	// run until until the channel closes.
	for {
		select {
		default:
			config.ObservationCachePtr.mu.RLock()
			emitNeeded := time.Now().UTC().Sub(config.ObservationCachePtr.lastEmitted) > config.EmitterInterval
			config.ObservationCachePtr.mu.RUnlock()
			if emitNeeded {
				log.Debug("Emit triggered")
				for _, loc := range config.Locations.Locations {
					emitLocation(tsconfig, loc, config.ObservationCachePtr, time.Now().UTC())
				}
				errs := tsconfig.FlushAwsTimestreamWrites()
				if len(errs) > 0 {
					for _, err := range errs {
						if err != nil {
							if config.DaemonStatusPtr != nil {
								config.DaemonStatusPtr.IncEmitError(err.Error())
							}
						}
					}
				} else {
					if config.DaemonStatusPtr != nil {
						config.DaemonStatusPtr.IncEmit()
					}
				}
				config.ObservationCachePtr.mu.Lock()
				config.ObservationCachePtr.lastEmitted = time.Now().UTC()
				config.ObservationCachePtr.mu.Unlock()
				log.Debugf("Emit done at %s", config.ObservationCachePtr.lastEmitted)
			} else {
				log.Debugf("No emit needed at this point (last emit: %s)",
					config.ObservationCachePtr.lastEmitted.Format(time.RFC3339))
				time.Sleep(time.Second)
			}
		case <-config.Finished:
			log.Info("Emitter ending.")
			config.Finished <- true
			return
		}

	}
}
