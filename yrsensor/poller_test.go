package yrsensor

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
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
		ts: [100]Observation{
			{
				Time:                  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				AirTemperature:        -5.0,
				AirPressureAtSeaLevel: 1023.3,
			},
			{
				Time:                  time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
				AirTemperature:        -7.5,
				AirPressureAtSeaLevel: 1110.5,
			},
		},
	}
	return obs
}

type ClientMock struct {
	response map[string][]byte
	expires  map[string]time.Time
}

func generateTestLocation(id string) Location {
	loc := Location{
		Id:   id,
		Lat:  20.0,
		Long: 10.0,
	}
	return loc
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

// Not the greatest test, just tests if it runs and looks okish.
func Test_request(t *testing.T) {
	const URL = "test://randomurl.com/"
	const URL_PARAMS = "?a=alfa&b=beta&c=charlie"
	const USERAGENT = "myuseragent"
	response := []byte("this is a response")
	Client = &ClientMock{
		response: map[string][]byte{
			URL + URL_PARAMS: response,
		},
	}
	params := map[string]string{
		"a": "alfa",
		"b": "beta",
		"c": "charlie",
	}
	res, err := request(URL, params, USERAGENT)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.NotNil(t, res.Status)
	assert.Equal(t, 200, res.StatusCode, "Status 200 expected")
	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err, "could not read body")
	assert.Equal(t, response, body)

}

func Test_getNewForecast(t *testing.T) {
	const URL = "test://test.yr"
	const API_VERSION = "2"
	const URL_PARAMS = "?lat=20.000000&lon=10.000000"
	const USERAGENT = "myuseragent"

	generatedForecast := generateTestForecast()

	var forecastBody, err = json.Marshal(generatedForecast)

	Client = &ClientMock{
		response: map[string][]byte{
			URL + "/locationforecast/" + API_VERSION + "/compact" + URL_PARAMS: forecastBody,
		},
	}
	loc := generateTestLocation("nada")

	forecast, err := getNewForecast(loc, URL, API_VERSION, USERAGENT)
	assert.Nil(t, err)
	assert.NotNil(t, forecast)

	assert.Equal(t, len(generatedForecast.Properties.Timeseries), len(forecast.Properties.Timeseries))

}

func Test_transformForecast(t *testing.T) {
	forecast := generateTestForecast()
	obsTimeSeries := transformForecast(forecast)
	expected := generateTestObservationTimeSeries()
	assert.Equal(t, expected, obsTimeSeries)
	// expected := ObservationTimeSeries{}
}
