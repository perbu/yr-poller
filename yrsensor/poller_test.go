package yrsensor

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	response map[string]string
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

	bodyBytes := []byte(c.response[req.URL.String()])
	resp := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBuffer(bodyBytes)),
		StatusCode: 200,
		Request:    req,
	}
	return resp, nil
}

// Not the greatest test, just tests if it runs and looks okish.
func Test_request(t *testing.T) {
	const URL = "test://randomurl.com/"
	const URL_PARAMS = "?foo=bar"
	const USERAGENT = "myuseragent"
	const RESPONSE = "This is a response"
	Client = &ClientMock{
		response: map[string]string{
			URL + URL_PARAMS: RESPONSE,
		},
	}
	params := map[string]string{
		"foo": "bar",
	}
	res, err := request(URL, params, USERAGENT)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.NotNil(t, res.Status)
	assert.Equal(t, 200, res.StatusCode, "Status 200 expected")
	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err, "could not read body")
	assert.Equal(t, []byte(RESPONSE), body)

}

func Test_getNewForecast(t *testing.T) {
	var forecast LocationForecast

	Client = &ClientMock{
		response: make(map[string]string),
	}
	loc := generateTestLocation("nada")
	params := map[string]string{
		"lat": fmt.Sprintf("%f", loc.Lat),
		"lon": fmt.Sprintf("%f", loc.Long),
	}

	res, err := request("test://www.yr.no/mock/getforecast", params, "myuseragent")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)

	assert.Equal(t, 200, res.StatusCode, "Status 200 expected")

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err, "Could not ready body of response.")
	err = json.Unmarshal(body, &forecast)
	assert.Nil(t, err, "Unmarshalling failure.")
}

func Test_transformForecast(t *testing.T) {
	forecast := generateTestForecast()
	obsTimeSeries := transformForecast(forecast)
	expected := generateTestObservationTimeSeries()
	assert.Equal(t, expected, obsTimeSeries)
	// expected := ObservationTimeSeries{}
}
