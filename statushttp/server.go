package statushttp

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"runtime"
	"time"
)

func StatusFactory() {

}

func (s *DaemonStatus) IncEmit(location string) {
	s.Locations[location].LastEmitTime = time.Now().UTC()
	s.Locations[location].NoOfEmits++
}

func (s *DaemonStatus) IncPoll(location string) {
	s.Locations[location].LastPollTime = time.Now().UTC()
	s.Locations[location].NoOfPolls++
}

func (s *DaemonStatus) IncEmitError(location string, errMsg string) {
	s.Locations[location].LastEmitErrorTime = time.Now().UTC()
	s.Locations[location].LastEmitErrorMessage = errMsg
	s.Locations[location].NoOfEmitErrors++
}

func (s *DaemonStatus) IncPollError(location string, errMsg string) {
	s.Locations[location].LastPollErrorTime = time.Now().UTC()
	s.Locations[location].LastPollErrorMessage = errMsg
	s.Locations[location].NoOfPollErrors++
}

func (s *DaemonStatus) AddLocation(location string) {
	s.Locations[location] = new(LocationStatus)
}

func (s *DaemonStatus) updateMemoryUsage() {
	var m runtime.MemStats
	var memstat MemStats

	runtime.ReadMemStats(&m)

	memstat.MemAlloc = m.Alloc
	memstat.MemTotalAlloc = m.TotalAlloc
	memstat.MemSys = m.Sys
	memstat.MemGC = m.NumGC

	s.MemoryStats = memstat

}

func Run(addr string) *DaemonStatus {
	stats := new(DaemonStatus)
	stats.RunningSince = time.Now().UTC()
	stats.Status = "running"
	stats.Locations = make(map[string]*LocationStatus)
	handler := stats.statsHandler
	http.HandleFunc("/", handler)
	log.Infof("starting stats server on %s", addr)
	go http.ListenAndServe(addr, nil)
	return stats
}
