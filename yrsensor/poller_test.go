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

func generateForecast() LocationForecast {

	data := []Timestep{
		{
			Time: time.Now().UTC().String(),
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
			Time: time.Now().UTC().String(),
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
		Type: "",
		Properties: Properties{
			Timeseries: data,
		},
		Expires: time.Time{},
	}
	return fc
}

type ClientMock struct {
	response map[string]string
}

func generateLocation() Location {
	loc := Location{
		Id:   "nada",
		Lat:  20.0,
		Long: 10.0,
	}
	return loc
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	fc := generateForecast()
	bodyBytes, err := json.Marshal(fc)
	resp := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBuffer(bodyBytes)),
		StatusCode: 200,
		Request:    req,
	}
	return resp, err
}

// Not the greatest test, just tests if it runs and looks okish.
func Test_request(t *testing.T) {
	const URL = "https://crazyurls.com"
	const USERAGENT = "myuseragent"
	Client = &ClientMock{
		response: make(map[string]string),
	}

	params := map[string]string{
		"foo": "quux",
		"bar": "zoo",
	}

	res, err := request(URL, params, USERAGENT)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.NotNil(t, res.Status)
	assert.Equal(t, 200, res.StatusCode, "Status 200 expected")
}

func Test_getNewForecast(t *testing.T) {
	var forecast LocationForecast

	Client = &ClientMock{
		response: make(map[string]string),
	}
	loc := generateLocation()
	params := map[string]string{
		"lat": fmt.Sprintf("%f", loc.Lat),
		"lon": fmt.Sprintf("%f", loc.Long),
	}

	res, err := request("https://www.yr.no/mock/getforecast", params, "myuseragent")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)

	assert.Equal(t, 200, res.StatusCode, "Status 200 expected")

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err, "Could not ready body of response.")
	err = json.Unmarshal(body, &forecast)
	assert.Nil(t, err, "Unmarshalling failure.")
}
