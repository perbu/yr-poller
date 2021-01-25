package statushttp

import (
	"time"
)

type PollerStatus struct {
	LastPollTime         time.Time `json:"last_poll"`
	NoOfPolls            uint64    `json:"no_of_polls"`
	NoOfPollErrors       uint64    `json:"no_of_poll_errors"`
	LastPollErrorMessage string    `json:"last_poll_error_message"`
	LastPollErrorTime    time.Time `json:"last_poll_error_time"`
}

type EmitterStatus struct {
	NoOfEmits            uint64    `json:"no_of_emits"`
	NoOfEmitErrors       uint64    `json:"no_of_emit_errors"`
	LastEmitTime         time.Time `json:"last_emit"`
	LastEmitErrorMessage string    `json:"last_emit_error_message"`
	LastEmitErrorTime    time.Time `json:"last_emit_error_time"`
}

type MemStats struct {
	MemAlloc      uint64 `json:"mem_alloc"`
	MemTotalAlloc uint64 `json:"mem_total_alloc"`
	MemSys        uint64 `json:"mem_sys"`
	MemGC         uint32 `json:"no_of_GCs"`
}

type DaemonStatus struct {
	Status       string                   `json:"status"`
	Pollers      map[string]*PollerStatus `json:"poller"`
	Emitter      *EmitterStatus           `json:"emitter"`
	RunningSince time.Time                `json:"running_since"`
	MemoryStats  MemStats                 `json:"memory_stats"`
}
