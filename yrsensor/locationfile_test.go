package yrsensor

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func Test_readLocations(t *testing.T) {

	const norm = `[
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
]`
	const invalidField = `[
  {
    "idx": "tryvannstua",
    "lat": 59.9981362,
    "long": 10.6660856
  },
  {
    "id": "skrindo",
    "lat": 60.6605926,
    "long": 8.5740604
  }
]`
	const invalidSyntax = `[
  {
    "id": "tryvannstua"
    "lat": 59.9981362,
    "long": 10.6660856
  },
  {
    "id": "skrindo",
    "lat": 60.6605926,
    "long": 8.5740604
  }
]`

	tests := []struct {
		name    string
		args    io.Reader
		want    []Location
		wantErr bool
	}{
		{
			name: "normal",
			args: strings.NewReader(norm),
			want: []Location{
				{
					Id:   "tryvannstua",
					Lat:  59.9981362,
					Long: 10.6660856,
				},
				{
					Id:   "skrindo",
					Lat:  60.6605926,
					Long: 8.5740604,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid field",
			args: strings.NewReader(invalidField),
			want: []Location{
				{
					Lat:  59.9981362,
					Long: 10.6660856,
				},
				{
					Id:   "skrindo",
					Lat:  60.6605926,
					Long: 8.5740604,
				},
			},
			wantErr: true,
		}, {
			name:    "invalid JSON syntax",
			args:    strings.NewReader(invalidSyntax),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readLocations(tt.args)
			if err != nil {
				fmt.Printf("Error (handled) from readLocations: %v\n", err.Error())
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("readLocations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readLocations() got = %v, want %v", got, tt.want)
			}
		})
	}
}
