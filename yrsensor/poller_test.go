package yrsensor

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

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
	const URL_PARAMS = "?lat=10.000000&lon=20.000000"
	const USERAGENT = "myuseragent"

	generatedForecast := generateTestForecast()

	var forecastBody, err = json.Marshal(generatedForecast)

	Client = &ClientMock{
		response: map[string][]byte{
			URL + "/locationforecast/" + API_VERSION + "/compact" + URL_PARAMS: forecastBody,
		},
	}
	loc := generateOneTestLocation("nada")

	forecast, err := getNewForecast(loc, URL, API_VERSION, USERAGENT)
	assert.Nil(t, err)
	assert.NotNil(t, forecast)

	assert.Equal(t, len(generatedForecast.Properties.Timeseries), len(forecast.Properties.Timeseries))

}

func Test_transformForecast(t *testing.T) {
	forecast := generateTestForecast()
	obsTimeSeries := transformForecast(forecast)
	expected := generateTestObservationTimeSeries()
	assert.Equal(t, &expected, obsTimeSeries)
	// expected := ObservationTimeSeries{}
}

func Test_refreshData(t *testing.T) {
	const ID = "tryvannstua"
	const URL = "test://test.yr"
	const API_VERSION = "2"
	const URL_PARAMS = "?lat=10.000000&lon=20.000000"
	const USERAGENT = "myuseragent"

	// Generate an expired cache...
	obsCache := generateTestObservationCache(ID, -2*time.Hour)
	locs := generateTestLocations(ID)

	// Set up the mock.

	generatedForecast := generateTestForecast()
	forecastBody, err := json.Marshal(generatedForecast)
	assert.Nil(t, err, "can't marshall forecast.")
	Client = &ClientMock{
		response: map[string][]byte{
			URL + "/locationforecast/" + API_VERSION + "/compact" + URL_PARAMS: forecastBody,
		},
	}
	var pc = PollerConfig{
		Finished:            make(chan bool),
		ApiUrl:              URL,
		ApiVersion:          API_VERSION,
		UserAgent:           USERAGENT,
		Locations:           *locs,
		ObservationCachePtr: obsCache,
		DaemonStatusPtr:     nil,
	}

	refreshData(&pc)
	// Todo. We should inspect the data structures here and see there are new datapoints.

}
