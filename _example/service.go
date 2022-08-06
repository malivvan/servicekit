package main

import (
	"fmt"

	"github.com/malivvan/servicekit"
	"github.com/malivvan/servicekit/conf"
	"github.com/malivvan/servicekit/log"
)

func main() { servicekit.Wrap(info, new(service)) }

var info = servicekit.Info{
	Name:        "example_service",
	Version:     "0.0.1",
	Description: "some example on how to use this package",
}

var config = struct {
	Test    string     `encrypt:"true"`
	Logging log.Config `json:"logging"`
}{
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
	fmt.Println(config)

	// 3. Configure Logging.
	err = log.Start(config.Logging)
	if err != nil {
		return err
	}
	log.Info().Msg("starting service")

	log.Warn().Msg("warn test")
	log.Error().Msg("error test")

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
	return nil
}
