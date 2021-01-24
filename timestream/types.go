package timestream

import (
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"net/http"
	"time"
)

type TimestreamState struct {
	AwsRegion           string
	AwsTimestreamDbname string
	WriteSession        *timestreamwrite.TimestreamWrite
	WriteBuffer         map[string][]*timestreamwrite.Record // a hash with table name as key.
	Transport           *http.Transport
}

type TimestreamEntry struct {
	Time      time.Time
	SensorId  string
	TableName string
	Value     string
}
