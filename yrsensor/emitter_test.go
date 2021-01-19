package yrsensor

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func Test_waitForObservations(t *testing.T) {
	const ID = "nada"
	fc := ObservationCache{
		observations: nil,
		lastEmitted:  time.Time{},
		mu:           sync.RWMutex{},
	}
	var obs = ObservationTimeSeries{
		ts: [100]Observation{
			{ID,
				time.Now().UTC(),
				25.5,
				1023.2,
			},
		},
		expires: time.Now().UTC(),
	}
	locs := generateLocations(ID)

	assert.False(t, waitForObservations(&fc, locs))
	fc.observations = make(map[string]ObservationTimeSeries)
	fc.observations[ID] = obs
	assert.True(t, waitForObservations(&fc, locs))

}

func generateObservationCache() ObservationCache {
	const ID = "tryvannstua"

	fc := ObservationCache{
		lastEmitted: time.Time{},
		mu:          sync.RWMutex{},
		observations: map[string]ObservationTimeSeries{
			ID: ObservationTimeSeries{
				ts: [100]Observation{
					{Id: ID,
						Time:                  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						AirTemperature:        -10.0,
						AirPressureAtSeaLevel: 1000.0,
					},
					{Id: ID,
						Time:                  time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
						AirTemperature:        -20.0,
						AirPressureAtSeaLevel: 1100.0,
					},
				},
				expires: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC),
			},
		},
	}
	return fc
}

func generateLocations(id string) []Location {

	locs := []Location{
		{
			Id:   id,
			Lat:  10.0,
			Long: 20.0,
		},
	}
	return locs
}

func Test_emit(t *testing.T) {
	const ID = "tryvannstua"
	const expectedResult = `{
  "Id": "tryvannstua",
  "time": "2020-01-01T00:30:00Z",
  "air_temperature": -15.0,
  "air_pressure_at_sealevel": 1050.0
}`
	when := time.Date(2020, 1, 1, 0, 30, 0, 0, time.UTC)
	fc := generateObservationCache()
	loc := generateLocation(ID)
	actual := emit(loc, &fc, when)
	assert.JSONEq(t, expectedResult, actual, "JSON comparison from emit failed")
}
