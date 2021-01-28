package yrsensor

import (
	"github.com/perbu/yrpoller/statushttp"
	"time"
)

type TimeSeriesRequest struct {
	Location        string
	ResponseChannel chan ObservationTimeSeries
}

type EmitterConfig struct {
	Finished            chan bool
	EmitterInterval     time.Duration
	Locations           Locations
	ObservationCachePtr *ObservationCache
	AwsRegion           string
	AwsTimestreamDbname string
	DaemonStatusPtr     *statushttp.DaemonStatus
	TsRequestChannel    chan TimeSeriesRequest
}

type PollerConfig struct {
	Finished            chan bool
	ApiUrl              string
	UserAgent           string
	Locations           Locations
	ObservationCachePtr *ObservationCache
	DaemonStatusPtr     *statushttp.DaemonStatus
	TsRequestChannel    chan TimeSeriesRequest
}

type Location struct {
	Id   string  `json:"id"`
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

type Locations struct {
	Locations []Location
}

type ObservationCache struct {
	observations map[string]ObservationTimeSeries
}

type ObservationTimeSeries struct {
	ts      []Observation
	expires time.Time
}

type Observation struct {
	Id                    string    // Only used by the emitter
	Time                  time.Time `json:"time"`
	AirTemperature        float64   `json:"air_temperature"`
	AirPressureAtSeaLevel float64   `json:"air_pressure_at_sealevel"`
	RelativeHumidity      float64   `json:"relative_humidity"`
	WindSpeed             float64   `json:"wind_speed"`
	WindFromDirection     float64   `json:"wind_from_direction"`
}

/* Most code below is (c) 2020 Andreas Palm and used under a MIT licence
   Source at https://github.com/zapling/yr.no-golang-client
*/

// LocationForecast => METJSONForecast
type LocationForecast struct {
	Type       string        `json:"type"`
	Geometry   PointGeometry `json:"geometry"`
	Properties Properties    `json:"properties"`
	Expires    time.Time
}

type PointGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"` // [longitude, latitude, altitude]
}

// Properties => Forecast
type Properties struct {
	Meta       Meta       `json:"meta"`
	Timeseries []Timestep `json:"timeseries"`
}

// Meta => inline_model_0
type Meta struct {
	UpdatedAt string        `json:"updated_at"`
	Units     ForecastUnits `json:"units"`
}

type ForecastUnits struct {
	PrecipitationAmount    string `json:"precipitation_amount,omitempty"`
	PrecipitationAmountMin string `json:"precipitation_amount_min"`
	PrecipitationAmountMax string `json:"precipitation_amount_max"`

	CloudAreaFractionLow    string `json:"cloud_area_fraction_low"`
	CloudAreaFractionMedium string `json:"cloud_area_fraction_medium"`
	CloudAreaFractionHigh   string `json:"cloud_area_fraction_high"`

	FogAreaFraction   string `json:"fog_area_fraction"`
	CloudAreaFraction string `json:"cloud_area_fraction"`

	WindSpeed         string `json:"wind_speed"`
	WindSpeedOfGust   string `json:"wind_speed_of_gust"`
	WindFromDirection string `json:"wind_from_direction"`

	AirTemperature        string `json:"air_temperature"`
	AirTemperatureMin     string `json:"air_temperature_min"`
	AirTemperatureMax     string `json:"air_temperature_max"`
	AirPressureAtSeaLevel string `json:"air_pressure_at_sea_level"`

	ProbabillityOfPrecipitation string `json:"probability_of_precipitation"`
	ProbabillityOfThunder       string `json:"probability_of_thunder"`

	RelativeHumidity            string `json:"relative_humidity"`
	DewPointTemperature         string `json:"dew_point_temperature"`
	UltravioletIndexClearSkyMax string `json:"ultraviolet_index_clear_sky_max"`
}

// Timestep => ForecastTimestep
type Timestep struct {
	Time string       `json:"time"`
	Data TimestepData `json:"data"`
}

// TimestepData => inline_model

// Not sure about wanting to keep next* data. De really only care about the
// next hour or two.
type TimestepData struct {
	Instant    InstantData    `json:"instant"`
	Next1Hours Next1HoursData `json:"next_1_hours"`
	/*
		Next6Hours  Next1HoursData `json:"next_6_hours"`
		Next12Hours Next1HoursData `json:"next_12_hours"`

	*/
}

// InstantData => Inline Model 2
type InstantData struct {
	Details ForecastTimeInstant `json:"details"`
}

type ForecastTimeInstant struct {
	AirTemperature          float64 `json:"air_temperature"`
	AirPressureAtSeaLevel   float64 `json:"air_pressure_at_sea_level"`
	CloudAreaFraction       float64 `json:"cloud_area_fraction"`
	WindSpeed               float64 `json:"wind_speed"`
	RelativeHumidity        float64 `json:"relative_humidity"`
	DewPointTemperature     float64 `json:"dew_point_temperature"`
	WindFromDirection       float64 `json:"wind_from_direction"`
	FogAreaFraction         float64 `json:"fog_area_fraction"`
	CloudAreaFractionHigh   float64 `json:"cloud_area_fraction_hight"`
	WindSpeedOfGust         float64 `json:"wind_speed_of_gust"`
	CloudAreaFractionMedium float64 `json:"cloud_area_fraction_medium"`
	CloudAreaFractionLow    float64 `json:"cloud_area_fraction_low"`
}

// Next1HoursData => Inline Model 3
type Next1HoursData struct {
	Details ForecastTimePeriod `json:"details"`
	Summary ForecastSummary    `json:"summary"`
}

type ForecastTimePeriod struct {
	ProbabillityOfPrecipitation float64 `json:"probability_of_precipitation"`
	AirTemperatureMin           float64 `json:"air_temperature_min"`
	UltravioletIndexClearSkyMax float64 `json:"ultraviolet_index_clear_sky_max"`
	AirTemperatureMax           float64 `json:"air_temperature_max"`
	PrecipitationAmountMin      float64 `json:"precipitation_amount_min"`
	ProbabillityOfThunder       float64 `json:"probability_of_thunder"`
	PrecipitationAmountMax      float64 `json:"precipitation_amount_max"`
	PrecipitationAmount         float64 `json:"precipitation_amount"`
}

type ForecastSummary struct {
	SymbolCode string `json:"symbol_code"`
}
