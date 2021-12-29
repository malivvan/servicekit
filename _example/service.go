package main

import (
	"runtime"

	"github.com/malivvan/servicekit"
	"github.com/malivvan/servicekit/conf"
	"github.com/malivvan/servicekit/log"
	"github.com/malivvan/servicekit/mon"
)

func main() { servicekit.Wrap(info, new(service)) }

var info = servicekit.Info{
	Name:        "example_service",
	Version:     "0.0.1",
	Description: "some example on how to use this package",
}

var config = struct {
	Monitoring mon.Config `json:"monitoring"`
	Logging    log.Config `json:"logging"`
}{
	Monitoring: mon.Config{
		URL:       "https://influx.example.org",
		Token:     "my-token",
		Org:       "malivvan",
		Bucket:    "malivvan",
		Prefix:    "service/",
		Seperator: "/",
	},
	Logging: log.Config{
		MaxSize:    10,
		MaxAge:     60,
		MaxBackups: 100,
		Level:      "info",
		Compress:   true,
	},
}

type service struct{}

func (s *service) Start() error {

	// 1. Configure Service.
	err := conf.Load("secret", &config)
	if err != nil {
		return err
	}

	// 2. Start Monitoring.
	err = mon.Start(config.Monitoring)
	if err != nil {
		return err
	}

	// 3. Configure Logging.
	err = log.Start(config.Logging, mon.LogWriter{})
	if err != nil {
		return err
	}

	log.Info().Msg("starting service")

	log.Warn().Msg("warn test")
	log.Error().Int("threshold", 1337).Msg("error test")

	mon.New("stats", nil).
		Int("mem", mon.AVG, mon.MAX, mon.MIN, mon.COUNT).
		Start(10000, 1000, func(values map[string]interface{}) {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			values["mem"] = m.Alloc
		})

	// ...

	log.Info().Msg("service started")
	return nil
}

func (s *service) Stop() error {
	log.Info().Msg("stopping service")

	// ...

	log.Info().Msg("stopped service")

	// stop service packages

	log.Stop()
	mon.Stop()
	return nil
}
