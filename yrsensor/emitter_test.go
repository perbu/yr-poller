package yrsensor

import (
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/perbu/yrpoller/timestream"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_waitForObservations(t *testing.T) {
	const ID = "nada"
	fc := ObservationCache{
		observations: nil,
	}
	var obs = ObservationTimeSeries{
		ts: []Observation{
			{Id: ID,
				Time:                  time.Now().UTC(),
				AirTemperature:        25.5,
				AirPressureAtSeaLevel: 1023.2,
				RelativeHumidity:      64.4,
				WindSpeed:             2.32,
				WindFromDirection:     4.2,
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

func Test_emit(t *testing.T) {
	const ID = "tryvannstua"

	when := time.Date(2020, 1, 1, 0, 30, 0, 0, time.UTC)
	fc := generateTestObservationCache(ID, 0)
	loc := generateOneTestLocation(ID)

	// mock the timestream state:
	tsState := timestream.TimestreamState{
		WriteBuffer: make(map[string][]*timestreamwrite.Record),
	}
	locTimeseries := fc.observations[loc.Id]
	emitLocation(tsState, loc, &locTimeseries, when)
	assert.Equal(t, "-15", *tsState.WriteBuffer["air_temperature"][0].MeasureValue)
	assert.Equal(t, "1050", *tsState.WriteBuffer["air_pressure_at_sealevel"][0].MeasureValue)
}
