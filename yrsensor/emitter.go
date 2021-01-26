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
	obs.RelativeHumidity = last.RelativeHumidity*factor + first.RelativeHumidity*(1.0-factor)
	obs.WindSpeed = last.WindSpeed*factor + first.WindSpeed*(1.0-factor)
	obs.WindFromDirection = last.WindFromDirection*factor + first.WindFromDirection*(1.0-factor)
	return obs
}

// Emit data
func emitLocation(tsconfig timestream.TimestreamState, location Location,
	timeseries *ObservationTimeSeries, when time.Time) {
	var obs Observation
	firstAfter := 0

	// Find out where we are in the time series.
	for i := range timeseries.ts {
		if timeseries.ts[i].Time.After(when) {
			firstAfter = i
			break
		}
	}
	// First measurement is still in the future so we can't interpolate:
	if firstAfter == 0 {
		obs.Time = timeseries.ts[0].Time
		obs.AirTemperature = timeseries.ts[0].AirTemperature
		obs.AirPressureAtSeaLevel = timeseries.ts[0].AirPressureAtSeaLevel
		obs.RelativeHumidity = timeseries.ts[0].RelativeHumidity
		obs.WindSpeed = timeseries.ts[0].WindSpeed
		obs.WindFromDirection = timeseries.ts[0].WindFromDirection
	} else {
		// Interpolate the two relevant measurements
		last := timeseries.ts[firstAfter]
		first := timeseries.ts[firstAfter-1]
		obs = interpolateObservations(&first, &last, when)
	}
	// add the Id (place)
	obs.Id = location.Id

	// jsonData, err := json.MarshalIndent(obs, "TS: ", "  ")
	tsconfig.MakeEntry(timestream.TimestreamEntry{
		Time:      obs.Time,
		SensorId:  obs.Id,
		TableName: "air_temperature",
		Value:     fmt.Sprintf("%v", obs.AirTemperature),
	})
	tsconfig.MakeEntry(timestream.TimestreamEntry{
		Time:      obs.Time,
		SensorId:  obs.Id,
		TableName: "air_pressure_at_sealevel",
		Value:     fmt.Sprintf("%v", obs.AirPressureAtSeaLevel),
	})
	tsconfig.MakeEntry(timestream.TimestreamEntry{
		Time:      obs.Time,
		SensorId:  obs.Id,
		TableName: "relative_humidity",
		Value:     fmt.Sprintf("%v", obs.RelativeHumidity),
	})
	tsconfig.MakeEntry(timestream.TimestreamEntry{
		Time:      obs.Time,
		SensorId:  obs.Id,
		TableName: "wind_speed",
		Value:     fmt.Sprintf("%v", obs.WindSpeed),
	})
	tsconfig.MakeEntry(timestream.TimestreamEntry{
		Time:      obs.Time,
		SensorId:  obs.Id,
		TableName: "wind_from_direction",
		Value:     fmt.Sprintf("%v", obs.WindFromDirection),
	})

	return
}

// waits for observations to arrive. Returns true or false
// false if not enough observations are present.
// true if the number of obs matches the fc cache.

// We don't lock here, so we are subject to races. But
// the worst that could happen is that we delay startup a few
// milliseconds.
func waitForObservations(fc *ObservationCache, locs *Locations) bool {

	if len(fc.observations) == len(locs.Locations) {
		log.Debug("(emitter) Observations are present.")
		return true
	} else {
		log.Debug("(emitter) Observations are not yet present.")
	}
	return false
}

func emitter(config *EmitterConfig) {
	var previousEmit time.Time
	log.Info("Starting emitter")

	tsconfig := timestream.Factory(config.AwsRegion, config.AwsTimestreamDbname)

	err := tsconfig.CheckAndCreateTables([]string{
		"air_temperature", "air_pressure_at_sealevel", "relative_humidity",
		"wind_speed", "wind_from_direction"})
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
			emitNeeded := time.Now().UTC().Sub(previousEmit) > config.EmitterInterval
			if emitNeeded {
				log.Debug("(emitter) Emit triggered")
				for _, loc := range config.Locations.Locations {
					// Fire off the emit. This will put an Rlock on the obs cache.
					resCh := make(chan ObservationTimeSeries)
					config.TsRequestChannel <- TimeSeriesRequest{
						Location:        loc.Id,
						ResponseChannel: resCh,
					}
					resTimeSeries := <-resCh

					emitLocation(tsconfig, loc, &resTimeSeries, time.Now().UTC())
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
				previousEmit = time.Now().UTC()
				log.Debugf("(emitter) Emit done at %s", previousEmit)
			} else {
				log.Debugf("(emitter) No emit needed at this point (last emit: %s)",
					previousEmit.Format(time.RFC3339))
				time.Sleep(5 * time.Second)
			}
		case <-config.Finished:
			log.Info("Emitter ending.")
			config.Finished <- true
			return
		}
	}
}
