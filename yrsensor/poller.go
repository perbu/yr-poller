package yrsensor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func request(url string, queryParams map[string]string, userAgent string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// This shouldn't happen.
		panic(err.Error())
	}
	// Add query params
	reqQuery := req.URL.Query()
	for k, v := range queryParams {
		reqQuery.Add(k, v)
	}
	req.URL.RawQuery = reqQuery.Encode()

	// Set User-Agent
	req.Header.Set("User-Agent", userAgent)

	// Send request
	res, err := http.DefaultClient.Do(req)

	return res, err
}

func getNewForecast(loc Location, apiUrl string, apiVersion string, userAgent string) (LocationForecast, error) {
	var forecast LocationForecast

	url := fmt.Sprintf("%s/locationforecast/%s/compact", apiUrl, apiVersion)
	params := map[string]string{
		"lat": fmt.Sprintf("%f", loc.Lat),
		"lon": fmt.Sprintf("%f", loc.Long),
	}
	res, err := request(url, params, userAgent)

	if err != nil {
		log.Fatal(err.Error())
		return forecast, err
	}
	if res.StatusCode != 200 && res.StatusCode != 203 {
		log.Fatal(err.Error())
		return forecast, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return forecast, err
	}
	json.Unmarshal(body, &forecast)
	forecast.Expires, err = http.ParseTime(res.Header.Get("Expires"))
	if err != nil {
		panic("Could not parse expires header")
	}
	return forecast, nil
}

/*
func getFullForecast(loc Location) (LocationForecast, error) {
	var forecast LocationForecast

	forecast, err != getFullForecast(loc)
}
*/
func transformForecast(forecast LocationForecast) ObservationTimeSeries {
	var m ObservationTimeSeries
	m.expires = forecast.Expires
	ts := forecast.Properties.Timeseries
	for i := 0; i < len(ts); i++ {
		var err error
		m.ts[i].AirTemperature = ts[i].Data.Instant.Details.AirTemperature
		m.ts[i].AirPressureAtSeaLevel = ts[i].Data.Instant.Details.AirPressureAtSeaLevel
		m.ts[i].Time, err = time.Parse(time.RFC3339, ts[i].Time)
		if err != nil {
			panic("Could not parse time on timeseries")
		}
	}
	log.Print("Forecast transformed")
	return m
}

func refreshData(apiUrl string, apiVersion string, userAgent string) {
	log.Printf("Polling the virtual nodes...(%s)", locations)
	for i, loc := range locations {
		log.Printf("%v - Polling %v", i, loc.Id)
		forecastsCache.mu.RLock()
		updateNeeded := forecastsCache.observations[loc.Id].expires.Before(time.Now().UTC())
		forecastsCache.mu.RUnlock()

		if updateNeeded {
			log.Printf("  Invalid or no data found for %v - refreshing.\n", loc.Id)
			// No data or invalid data. Refresh the dataset we have.
			forecast, err := getNewForecast(loc, apiUrl, apiVersion, userAgent)
			if err != nil {
				panic(err.Error())
			}
			m := transformForecast(forecast)
			if len(m.ts) > 0 {
				fmt.Print("Got it!")
			}
			forecastsCache.mu.Lock()
			forecastsCache.observations[loc.Id] = m
			forecastsCache.mu.Unlock()
		} else {
			log.Print("Current data is up to date.")
		}
	}
	ready = true // Let the emitter know we have data...
}

func poller(control *bool, finished chan bool, apiUrl string, apiVersion string, userAgent string) {
	log.Print("Starting poller...")
	for *control {
		refreshData(apiUrl, apiVersion, userAgent) // Todo: Make this async. Lock it so it doesn't run out of control.
		log.Print("refreshData() returned")
		time.Sleep(5 * time.Second)
	}
	log.Printf("Poller ending")
	finished <- true
}
