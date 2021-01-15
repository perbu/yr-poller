package yrsensor

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

func emit(location Location) {
	var obs Observation
	now := time.Now().UTC()
	forecastsCache.mu.RLock()
	defer forecastsCache.mu.RUnlock()
	ts := forecastsCache.observations[location.Id].ts
	firstAfter := 0

	// Find out where we are in the time series.
	for i := range ts {
		if ts[i].Time.After(now) {
			firstAfter = i
			break
		}
	}
	// First measurement is still in the future so we can't interpolate:
	if firstAfter == 0 {
		obs.Id = location.Id
		obs.Time = ts[0].Time
		obs.AirTemperature = ts[0].AirTemperature
		obs.AirPressureAtSeaLevel = ts[0].AirPressureAtSeaLevel
	} else {
		// Interpolate the two relevant measurements
		last := ts[firstAfter]
		first := ts[firstAfter-1]
		timeDelta := last.Time.Sub(first.Time).Seconds() // Typically 60mins
		howFarInto := now.Sub(first.Time).Seconds()
		factor := float64(howFarInto) / float64(timeDelta)
		obs.Id = location.Id
		obs.Time = now
		// Interpolating here:
		obs.AirTemperature = last.AirTemperature*factor + last.AirTemperature*(1.0-factor)
		obs.AirPressureAtSeaLevel = last.AirPressureAtSeaLevel*factor + last.AirPressureAtSeaLevel*(1.0-factor)
	}
	jsonData, err := json.MarshalIndent(obs, "", "  ")
	if err != nil {
		log.Fatal("Brain damage! Can't marshal internal structure to JSON.")
	}
	fmt.Println(string(jsonData))
}

// waits for observations to arrive. Blocks until they are present.
func waitForObservations() {
	for true {
		forecastsCache.mu.RLock()
		if len(forecastsCache.observations) == len(locations) {
			log.Debug("Observations are present. Starting emit loop.")
			forecastsCache.mu.RUnlock()
			break
		} else {
			log.Debug("Observations are not yet present. Sleeping.")
		}
		forecastsCache.mu.RUnlock()
		time.Sleep(time.Second)
	}
}

func emitter(control *bool, finished chan bool, emitterInterval time.Duration) {
	log.Info("Starting emitter")
	waitForObservations()
	for *control {
		forecastsCache.mu.RLock()
		emitNeeded := time.Now().UTC().Sub(forecastsCache.lastEmitted) > emitterInterval
		forecastsCache.mu.RUnlock()
		if emitNeeded {
			log.Debug("Emit triggered")
			for _, loc := range locations {
				emit(loc)
			}
			forecastsCache.mu.Lock()
			forecastsCache.lastEmitted = time.Now().UTC()
			forecastsCache.mu.Unlock()
		} else {
			log.Debug("No emit needed at this point")
			time.Sleep(10 * time.Second)
		}
	}
	log.Info("Emitter ending.")
	finished <- true

}
