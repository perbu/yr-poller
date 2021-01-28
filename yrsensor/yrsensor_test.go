package yrsensor

import (
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	var result int
	setupLogging(log.DebugLevel, "")
	log.Debug("Log level set to DEBUG for test run")
	result = m.Run()
	// If we need any teardown code it goes in here.
	os.Exit(result)
}
