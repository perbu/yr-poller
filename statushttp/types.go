package statushttp

import (
	"time"
)

type LocationStatus struct {
	LastPollTime         time.Time `json:"last_poll"`
	LastEmitTime         time.Time `json:"last_emit"`
	NoOfPolls            uint64    `json:"no_of_polls"`
	NoOfPollErrors       uint64    `json:"no_of_poll_errors"`
	NoOfEmits            uint64    `json:"no_of_emits"`
	NoOfEmitErrors       uint64    `json:"no_of_emit_errors"`
	LastEmitErrorMessage string    `json:"last_emit_error_message"`
	LastEmitErrorTime    time.Time `json:"last_emit_error_time"`
	LastPollErrorMessage string    `json:"last_poll_error_message"`
	LastPollErrorTime    time.Time `json:"last_poll_error_time"`
}

type MemStats struct {
	MemAlloc      uint64 `json:"mem_alloc"`
	MemTotalAlloc uint64 `json:"mem_total_alloc"`
	MemSys        uint64 `json:"mem_sys"`
	MemGC         uint32 `json:"no_of_GCs"`
}

type DaemonStatus struct {
	Status       string                     `json:"status"`
	Locations    map[string]*LocationStatus `json:"locations"`
	RunningSince time.Time                  `json:"running_since"`
	MemoryStats  MemStats                   `json:"memory_stats"`
}
