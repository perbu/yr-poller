package yrsensor

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func emit(location Location) {
	var obs Observation
	now := time.Now().UTC()
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
		fmt.Println(factor)
		obs.Time = now
		// Interpolating here:
		obs.AirTemperature = last.AirTemperature*factor + last.AirTemperature*(1.0-factor)
		obs.AirPressureAtSeaLevel = last.AirPressureAtSeaLevel*factor + last.AirPressureAtSeaLevel*(1.0-factor)
	}
	jsonData, err := json.MarshalIndent(obs, fmt.Sprintf("EMIT(%s):", location.Id), " ")
	if err != nil {
		panic("Brain damage!")
	}
	fmt.Print("=========== EMIT ============")
	fmt.Print(string(jsonData))
}

func emitter(control *bool, finished chan bool, emitterInterval time.Duration) {
	// Todo: locking
	log.Print("Starting emitter")
	// wait for data to arrive...
	for ready == false {
		time.Sleep(time.Second)
		log.Print("Waiting for data...")
	}
	for *control {
		// Do I need locking to do this?
		if time.Now().UTC().Sub(forecastsCache.lastEmitted) > emitterInterval {
			log.Print("Emit triggered")
			for _, loc := range locations {
				emit(loc)
			}
			forecastsCache.lastEmitted = time.Now().UTC()
		} else {
			fmt.Print("No emit")
			time.Sleep(10 * time.Second)
		}
	}
}
