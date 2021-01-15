package yrsensor

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

func readLocations(locationFilePath string) []Location {
	locationsFile, err := os.Open(locationFilePath)
	if err != nil {
		log.Fatalf("While opening file: %s", err.Error())
	} else {
		log.Debug("Location file successfully opened.")
	}
	defer locationsFile.Close()
	locationData, _ := ioutil.ReadAll(locationsFile)
	var data []Location
	json.Unmarshal(locationData, &data)

	return data
}
