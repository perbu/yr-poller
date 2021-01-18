package yrsensor

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

func emit(location Location, obsCache *ObservationCache) {
	var obs Observation
	now := time.Now().UTC()
	obsCache.mu.RLock()
	defer obsCache.mu.RUnlock()
	ts := obsCache.observations[location.Id].ts
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
func waitForObservations(fc *ObservationCache, locs []Location) {
	for true {
		fc.mu.RLock()
		if len(fc.observations) == len(locs) {
			log.Debug("Observations are present. Starting emit loop.")
			fc.mu.RUnlock()
			break
		} else {
			log.Debug("Observations are not yet present. Sleeping.")
		}
		fc.mu.RUnlock()
		time.Sleep(100 * time.Millisecond)
	}
}

func emitter(control *bool, finished chan bool, emitterInterval time.Duration, locs []Location, obs *ObservationCache) {
	log.Info("Starting emitter")
	waitForObservations(obs, locs)
	for *control {
		obs.mu.RLock()
		emitNeeded := time.Now().UTC().Sub(obs.lastEmitted) > emitterInterval
		obs.mu.RUnlock()
		if emitNeeded {
			log.Debug("Emit triggered")
			for _, loc := range locs {
				emit(loc, obs)
			}
			obs.mu.Lock()
			obs.lastEmitted = time.Now().UTC()
			obs.mu.Unlock()
		} else {
			log.Debug("No emit needed at this point")
			time.Sleep(100 * time.Millisecond)
		}
	}
	log.Info("Emitter ending.")
	finished <- true

}
