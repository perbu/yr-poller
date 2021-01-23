package statushttp

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (ds *DaemonStatus) statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Errorf("invalid method for %s from %v", r.URL.String(), r.RemoteAddr)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ds.updateMemoryUsage()
	// bla bla
	jsonBytes, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		log.Fatal("Brain damage! Can't marshal internal structure to JSON.")
	}
	w.Write(jsonBytes)
}
