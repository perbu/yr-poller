package yrsensor

import (
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	var result int
	setupLogging(log.DebugLevel)
	log.Debug("Log level set to DEBUG for test run")
	result = m.Run()
	/*
		Draft code to have the test suite timeout.
		timeout := time.After(3 * time.Second)
		done := make(chan bool)
		go func() {
			result = m.Run()
			done <- true
		}()
		select {
		case <-timeout:
			panic("Test timeout.")
		case <-done:
		}
	*/

	// If we need any teardown code it goes in here.
	os.Exit(result)
}
