package statushttp

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"runtime"
	"time"
)

func (ds *DaemonStatus) IncPoll(location string) {
	ds.Pollers[location].LastPollTime = time.Now().UTC()
	ds.Pollers[location].NoOfPolls++
}
func (ds *DaemonStatus) IncPollError(location string, errMsg string) {
	ds.Pollers[location].LastPollErrorTime = time.Now().UTC()
	ds.Pollers[location].LastPollErrorMessage = errMsg
	ds.Pollers[location].NoOfPollErrors++
}

func (ds *DaemonStatus) IncEmitError(errMsg string) {
	ds.Emitter.LastEmitErrorTime = time.Now().UTC()
	ds.Emitter.LastEmitErrorMessage = errMsg
	ds.Emitter.NoOfEmitErrors++
}
func (ds *DaemonStatus) IncEmit() {
	ds.Emitter.LastEmitTime = time.Now().UTC()
	ds.Emitter.NoOfEmits++
}

func (ds *DaemonStatus) AddLocation(location string) {
	ds.Pollers[location] = new(PollerStatus)
}

func (ds *DaemonStatus) updateMemoryUsage() {
	var m runtime.MemStats
	var memstat MemStats

	runtime.ReadMemStats(&m)

	memstat.MemAlloc = m.Alloc
	memstat.MemTotalAlloc = m.TotalAlloc
	memstat.MemSys = m.Sys
	memstat.MemGC = m.NumGC

	ds.MemoryStats = memstat

}

func Run(addr string) (stats DaemonStatus) {
	stats.RunningSince = time.Now().UTC()
	stats.Status = "running"
	stats.Pollers = make(map[string]*PollerStatus)
	stats.Emitter = new(EmitterStatus)
	handler := stats.statsHandler
	// This is a very neat way of injecting state into a handler:
	http.HandleFunc("/", handler)
	log.Infof("starting stats server on %s", addr)
	go func() {
		log.Fatal(http.ListenAndServe(addr, nil))
	}()
	return stats
}
