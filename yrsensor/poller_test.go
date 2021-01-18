package yrsensor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type ClientMock struct {
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	fmt.Println(req.Header.Get("User-Agent"))
	fmt.Println(req.URL)
	return &http.Response{}, nil
}

func Test_request(t *testing.T) {
	Client = &ClientMock{}
	loc := Location{
		Id:   "nada",
		Lat:  20.0,
		Long: 10.0,
	}
	params := map[string]string{
		"lat": fmt.Sprintf("%f", loc.Lat),
		"lon": fmt.Sprintf("%f", loc.Long),
	}

	fc, err := request("https://www.yr.no/mock/getforecast", params, "myuseragent")
	assert.Nil(t, err)
	assert.NotNil(t, fc)
	fmt.Println("Successful request.")
}
