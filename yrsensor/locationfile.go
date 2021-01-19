package yrsensor

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

func readLocations(locationsFile io.Reader) ([]Location, error) {
	var data []Location
	decoder := json.NewDecoder(locationsFile)
	decoder.DisallowUnknownFields() // force errors on unknown fields.
	err := decoder.Decode(&data)
	return data, err
}

func readLocationsFromPath(locationFilePath string) ([]Location, error) {
	var data []Location
	locationsFile, err := os.Open(locationFilePath)
	if err != nil {
		return data, err
	}
	data, parseErr := readLocations(locationsFile)
	if parseErr != nil {
		return nil, parseErr
	}
	err = locationsFile.Close()
	if err != nil {
		log.Panicf("closing file: %v", err.Error())
	}
	return data, nil
}

func locationFileExample() string {
	return `
[
  {
    "id": "tryvannstua",
    "lat": 59.9981362,
    "long": 10.6660856
  },
  {
    "id": "skrindo",
    "lat": 60.6605926,
    "long": 8.5740604
  }
]
`

}
