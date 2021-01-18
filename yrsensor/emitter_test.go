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
	locs := []Location{
		{
			Id:   ID,
			Lat:  10.0,
			Long: 20.0,
		},
	}

	assert.False(t, waitForObservations(&fc, locs))
	time.Sleep(50 * time.Millisecond)
	// ok until here:
	fc.observations = make(map[string]ObservationTimeSeries)
	fc.observations[ID] = obs
	assert.True(t, waitForObservations(&fc, locs))

}
