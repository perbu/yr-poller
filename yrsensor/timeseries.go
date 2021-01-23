package yrsensor

/*
 Timeseries stuff, useing AWS Timestream.

 With Go 1.6 we could drop the dependency on golang/x/net/http2.

*/

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"strconv"
	"time"
)

const TimestreamDatabase = "yrpoll-dev"

func createTimestreamTransport() *http.Transport {
	tr := http.Transport{
		ResponseHeaderTimeout: 20 * time.Second,
		// Using DefaultTransport values for other parameters: https://golang.org/pkg/net/http/#RoundTripper
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// So client makes HTTP/2 requests
	http2.ConfigureTransport(&tr)
	return &tr
}

func createTimestreamWriteSession() *timestreamwrite.TimestreamWrite {
	const REGION = "eu-west-1"
	tr := createTimestreamTransport()
	sess, err := session.NewSession(&aws.Config{Region: aws.String(REGION),
		MaxRetries: aws.Int(10),
		HTTPClient: &http.Client{Transport: tr}})
	if err != nil {
		panic(err.Error())
	}
	writeSvc := timestreamwrite.New(sess)
	return writeSvc
}

func tableExists(table string, tableOutput *timestreamwrite.ListTablesOutput) bool {
	for _, b := range tableOutput.Tables {
		if *b.TableName == table {
			return true
		}
	}
	return false
}

func checkAndCreateTables(sess *timestreamwrite.TimestreamWrite) (bool, error) {
	const DBNAME = "yrpoller-dev"
	var maxTables int64 = 20

	tables := []string{
		"air_temperature", "air_pressure_at_sealevel",
	}

	listTablesInput := &timestreamwrite.ListTablesInput{
		DatabaseName: aws.String(DBNAME),
		MaxResults:   &maxTables,
	}
	listTablesOutput, err := sess.ListTables(listTablesInput)

	for _, table := range tables {
		log.Debugf("checking table %s", table)
		if !tableExists(table, listTablesOutput) {
			createTableInput := &timestreamwrite.CreateTableInput{
				DatabaseName: aws.String(DBNAME),
				TableName:    aws.String(table),
			}
			_, err := sess.CreateTable(createTableInput)
			if err != nil {
				panic(err.Error())
			}
		}
	}

	return true, err
}

func timestreamWriteObservation(sess *timestreamwrite.TimestreamWrite, obs Observation) {
	// Note really pretty.

	const DBNAME = "yrpoller-dev"
	writeRecordsInput := &timestreamwrite.WriteRecordsInput{
		DatabaseName: aws.String(DBNAME),
		TableName:    aws.String("air_temperature"),
		Records: []*timestreamwrite.Record{
			&timestreamwrite.Record{
				Dimensions: []*timestreamwrite.Dimension{
					&timestreamwrite.Dimension{
						Name:  aws.String("sensor"),
						Value: aws.String(obs.Id),
					},
				},
				MeasureName:      aws.String("air_temperature"),
				MeasureValue:     aws.String(fmt.Sprintf("%v", obs.AirTemperature)),
				MeasureValueType: aws.String("DOUBLE"),
				Time:             aws.String(strconv.FormatInt(obs.Time.Unix(), 10)),
				TimeUnit:         aws.String("SECONDS"),
			},
		},
	}
	_, err := sess.WriteRecords(writeRecordsInput)
	if err != nil {
		panic(err.Error())
	}
	writeRecordsInput = &timestreamwrite.WriteRecordsInput{
		DatabaseName: aws.String(DBNAME),
		TableName:    aws.String("air_pressure_at_sealevel"),
		Records: []*timestreamwrite.Record{
			&timestreamwrite.Record{
				Dimensions: []*timestreamwrite.Dimension{
					&timestreamwrite.Dimension{
						Name:  aws.String("sensor"),
						Value: aws.String(obs.Id),
					},
				},
				MeasureName:      aws.String("air_pressure_at_sealevel"),
				MeasureValue:     aws.String(fmt.Sprintf("%v", obs.AirPressureAtSeaLevel)),
				MeasureValueType: aws.String("DOUBLE"),
				Time:             aws.String(strconv.FormatInt(obs.Time.Unix(), 10)),
				TimeUnit:         aws.String("SECONDS"),
			},
		},
	}
	_, err = sess.WriteRecords(writeRecordsInput)
	if err != nil {
		panic(err.Error())
	}
}
