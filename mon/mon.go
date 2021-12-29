package mon

import (
	"os"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/malivvan/servicekit"
)

var (
	influxClient          influxdb2.Client
	influxWriteAPI        api.WriteAPI
	influxNameSeperator   string
	influxNamePrefix      string
	measurementsRegister_ sync.Mutex
	measurementsRegister  = []*Measurement{}
)

type Config struct {
	URL       string `json:"url"`
	Token     string `json:"token" encrypt:"true"`
	Org       string `json:"org"`
	Bucket    string `json:"bucket"`
	Prefix    string `json:"prefix,omitempty"`
	Seperator string `json:"seperator,omitempty"`
}

func Start(config Config) error {

	// get runtime information
	host, err := os.Hostname()
	if err != nil {
		return err
	}
	osid, err := OSID()
	if err != nil {
		return err
	}

	// create client and set path building variables
	influxNameSeperator = config.Seperator
	influxNamePrefix = config.Prefix
	influxClient = influxdb2.NewClientWithOptions(config.URL, config.Token, influxdb2.DefaultOptions().
		SetPrecision(time.Nanosecond).
		AddDefaultTag("version", servicekit.Version()).
		AddDefaultTag("host", host).
		AddDefaultTag("osid", osid))
	influxWriteAPI = influxClient.WriteAPI(config.Org, config.Bucket)
	return nil
}

func Stop() {

	// cleanup measurement routines
	measurementsRegister_.Lock()
	for _, measurement := range measurementsRegister {
		measurement.Stop()
	}
	measurementsRegister_.Unlock()

	// flush influx connection and close
	influxWriteAPI.Flush()
	influxClient.Close()
}

func Write(name string, t time.Time, tags map[string]string, fields map[string]interface{}) {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	influxWriteAPI.WritePoint(write.NewPoint(influxNamePrefix+servicekit.Name()+influxNameSeperator+name, tags, fields, t))
}
