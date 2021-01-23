package yrsensor

import (
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/perbu/yrpoller/statushttp"
	log "github.com/sirupsen/logrus"
	"time"
)

// Emit some JSON constituting a virtual sensor readout.
func emit(session *timestreamwrite.TimestreamWrite, location Location, obsCache *ObservationCache, when time.Time) {
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
		obs.Id = location.Id
		obs.Time = ts[0].Time
		obs.AirTemperature = ts[0].AirTemperature
		obs.AirPressureAtSeaLevel = ts[0].AirPressureAtSeaLevel
	} else {
		// Interpolate the two relevant measurements
		last := ts[firstAfter]
		first := ts[firstAfter-1]
		timeDelta := last.Time.Sub(first.Time).Seconds() // Typically 60mins
		howFarInto := when.Sub(first.Time).Seconds()
		factor := float64(howFarInto) / float64(timeDelta)
		obs.Id = location.Id
		obs.Time = when
		// Interpolating here:
		obs.AirTemperature = last.AirTemperature*factor + first.AirTemperature*(1.0-factor)
		obs.AirPressureAtSeaLevel = last.AirPressureAtSeaLevel*factor + first.AirPressureAtSeaLevel*(1.0-factor)
	}
	// jsonData, err := json.MarshalIndent(obs, "TS: ", "  ")
	if false {
		// Todo: Enable again when stuff works.
		timestreamWriteObservation(session, obs)
	}
	return
}

// waits for observations to arrive. Returns true or false
// false if not enough observations are present.
// true if the number of obs matches the fc cache.
func waitForObservations(fc *ObservationCache, locs []Location) bool {
	fc.mu.RLock()
	if len(fc.observations) == len(locs) {
		log.Debug("Observations are present.")
		fc.mu.RUnlock()
		return true
	} else {
		log.Debug("Observations are not yet present.")
	}
	fc.mu.RUnlock()
	return false
}

func emitter(control *bool, finished chan bool, emitterInterval time.Duration, locs []Location, obs *ObservationCache, ds *statushttp.DaemonStatus) {
	log.Info("Starting emitter")
	session := createTimestreamWriteSession()
	checkAndCreateTables(session)
	for waitForObservations(obs, locs) == false {
		time.Sleep(100 * time.Millisecond)
	}
	for *control {
		obs.mu.RLock()
		emitNeeded := time.Now().UTC().Sub(obs.lastEmitted) > emitterInterval
		obs.mu.RUnlock()
		if emitNeeded {
			log.Debug("Emit triggered")
			for _, loc := range locs {
				emit(session, loc, obs, time.Now().UTC())
				log.Infof("emitting data for %s", loc.Id)
				ds.IncEmit(loc.Id)
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
