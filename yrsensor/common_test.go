package yrsensor

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

func generateTestForecast() LocationForecast {
	data := []Timestep{
		{
			Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Data: TimestepData{
				InstantData{
					Details: ForecastTimeInstant{
						AirTemperature:        -5.0,
						AirPressureAtSeaLevel: 1023.3,
						RelativeHumidity:      64.4,
						WindSpeed:             2.32,
						WindFromDirection:     4.2,
					},
				},
			},
		},
		{
			Time: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Data: TimestepData{
				InstantData{
					Details: ForecastTimeInstant{
						AirTemperature:        -7.5,
						AirPressureAtSeaLevel: 1110.5,
						RelativeHumidity:      64.4,
						WindSpeed:             2.32,
						WindFromDirection:     4.2,
					},
				},
			},
		},
	}
	fc := LocationForecast{
		Properties: Properties{
			Timeseries: data,
		},
		Expires: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC),
	}
	return fc
}

// Match whats above
func generateTestObservationTimeSeries() ObservationTimeSeries {
	obs := ObservationTimeSeries{
		expires: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC),
		ts: []Observation{
			{
				Time:                  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				AirTemperature:        -5.0,
				AirPressureAtSeaLevel: 1023.3,
				RelativeHumidity:      64.4,
				WindSpeed:             2.32,
				WindFromDirection:     4.2,
			},
			{
				Time:                  time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
				AirTemperature:        -7.5,
				AirPressureAtSeaLevel: 1110.5,
				RelativeHumidity:      64.4,
				WindSpeed:             2.32,
				WindFromDirection:     4.2,
			},
		},
	}
	return obs
}

type ClientMock struct {
	response map[string][]byte
	expires  map[string]time.Time
}

func generateOneTestLocation(id string) Location {
	loc := Location{
		Id:   id,
		Lat:  10.0,
		Long: 20.0,
	}
	return loc
}

func generateTestLocations(id string) *Locations {
	locs := new(Locations)
	locs.Locations = make([]Location, 0)
	locs.Locations = append(locs.Locations, generateOneTestLocation(id))
	return locs
}

// Generate an premade obs cache. With a duration of 0 the cache will be valid for one hour past the
// last entry. Can be set negative (-2 hours) to make an invalid cache.
func generateTestObservationCache(id string, expireDelta time.Duration) *ObservationCache {

	fc := ObservationCache{
		observations: map[string]ObservationTimeSeries{
			id: {
				ts: []Observation{
					{Id: id,
						Time:                  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						AirTemperature:        -10.0,
						AirPressureAtSeaLevel: 1000.0,
						RelativeHumidity:      60.0,
						WindSpeed:             10.0,
						WindFromDirection:     1.0,
					},
					{Id: id,
						Time:                  time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
						AirTemperature:        -20.0,
						AirPressureAtSeaLevel: 1100.0,
						RelativeHumidity:      70.0,
						WindSpeed:             0,
						WindFromDirection:     11.0,
					},
				},
				expires: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC).Add(expireDelta),
			},
		},
	}
	return &fc
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	var expiresHeader string
	bodyBytes := c.response[req.URL.String()]
	expires, ok := c.expires[req.URL.String()]
	if ok == false {
		// default is one hour on expiry.
		expiresHeader = time.Now().Add(time.Hour).UTC().Format(http.TimeFormat)
	} else {
		expiresHeader = expires.Format(http.TimeFormat)
	}
	resp := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBuffer(bodyBytes)),
		StatusCode: 200,
		Request:    req,
		Header:     make(http.Header, 1),
	}
	resp.Header.Add("Expires", expiresHeader)
	return resp, nil
}
