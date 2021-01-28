package timestream

/*
   Integration with AWS Timestream for the yrpoller
*/

import (
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

func Factory(awsRegion string, awsTimestreamDbname string) TimestreamState {

	transport, err := createTimestreamTransport()
	if err != nil {
		panic(err.Error())
	}
	state := TimestreamState{
		AwsRegion:           awsRegion,
		AwsTimestreamDbname: awsTimestreamDbname,
		Transport:           transport,
		WriteSession:        createTimestreamWriteSession(awsRegion, transport),
		WriteBuffer:         make(map[string][]*timestreamwrite.Record, 100),
	}
	return state
}

func createTimestreamTransport() (*http.Transport, error) {
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
	err := http2.ConfigureTransport(&tr)
	return &tr, err
}

func createTimestreamWriteSession(awsRegion string, tr *http.Transport) *timestreamwrite.TimestreamWrite {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion),
		MaxRetries: aws.Int(10),
		HTTPClient: &http.Client{Transport: tr}})
	if err != nil {
		log.Fatalf("could not estalish write session to AWS: %s", err.Error())
	}
	writeSvc := timestreamwrite.New(sess)
	return writeSvc
}

// helper to check if a table exists in the supplied tableOutput
func tableExists(table string, tableOutput *timestreamwrite.ListTablesOutput) bool {
	for _, b := range tableOutput.Tables {
		if *b.TableName == table {
			return true
		}
	}
	return false
}

// get a list of the tables in the database and create the ones we need if they don't exist.
func (c *TimestreamState) CheckAndCreateTables(tables []string) error {
	var maxTables int64 = 20
	var err error

	listTablesInput := &timestreamwrite.ListTablesInput{
		DatabaseName: aws.String(c.AwsTimestreamDbname),
		MaxResults:   &maxTables,
	}
	listTablesOutput, err := c.WriteSession.ListTables(listTablesInput)
	if err != nil {
		return err
	}

	for _, table := range tables {
		log.Debugf("(timestream) checking table %s", table)
		if !tableExists(table, listTablesOutput) {
			createTableInput := &timestreamwrite.CreateTableInput{
				DatabaseName: aws.String(c.AwsTimestreamDbname),
				TableName:    aws.String(table),
			}
			_, err := c.WriteSession.CreateTable(createTableInput)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func (c *TimestreamState) MakeEntry(entry TimestreamEntry) {
	rec := timestreamwrite.Record{
		Dimensions: []*timestreamwrite.Dimension{
			{
				Name:  aws.String("sensor"),
				Value: aws.String(entry.SensorId),
			},
		},
		MeasureName:      aws.String(entry.SensorId),
		MeasureValue:     aws.String(entry.Value),
		MeasureValueType: aws.String("DOUBLE"),
		Time:             aws.String(strconv.FormatInt(entry.Time.Unix(), 10)),
		TimeUnit:         aws.String("SECONDS"),
	}
	c.WriteBuffer[entry.TableName] = append(c.WriteBuffer[entry.TableName], &rec)
}

func (c *TimestreamState) FlushAwsTimestreamWrites() []error {
	var errs = make([]error, 0)
	for table, buffer := range c.WriteBuffer {
		// construct a write
		write := &timestreamwrite.WriteRecordsInput{
			DatabaseName: aws.String(c.AwsTimestreamDbname),
			TableName:    aws.String(table),
			Records:      buffer,
		}
		_, err := c.WriteSession.WriteRecords(write)
		if err != nil {
			errs = append(errs, err)
		} else {
			log.Debugf("(timestream) pushed %d records to timestream table %s, flushing buffer",
				len(buffer), table)
			c.WriteBuffer[table] = nil
		}
	}
	return errs
}
