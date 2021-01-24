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
	locs := generateTestLocations(ID)

	assert.False(t, waitForObservations(&fc, locs))
	fc.observations = make(map[string]ObservationTimeSeries)
	fc.observations[ID] = obs
	assert.True(t, waitForObservations(&fc, locs))

}

func generateTestEmitJson() string {
	return `{
  "Id": "tryvannstua",
  "time": "2020-01-01T00:30:00Z",
  "air_temperature": -15.0,
  "air_pressure_at_sealevel": 1050.0  }`
}

// Todo: Enable this again.

//func Test_emit(t *testing.T) {
//	const ID = "tryvannstua"
//
//	when := time.Date(2020, 1, 1, 0, 30, 0, 0, time.UTC)
//	fc := generateTestObservationCache(ID, 0)
//	loc := generateTestLocation(ID)
//	actual := emitLocation(nil, loc, &fc, when)
//	assert.JSONEq(t, generateTestEmitJson(), actual, "JSON comparison from emitLocation failed")
//}
