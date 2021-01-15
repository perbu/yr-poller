package yrsensor

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func readLocations(locationFilePath string) []Location {
	locationsFile, err := os.Open(locationFilePath)
	if err != nil {
		panic(err.Error())
	}
	defer locationsFile.Close()
	locationData, _ := ioutil.ReadAll(locationsFile)
	var data []Location
	json.Unmarshal(locationData, &data)
	return data
}
