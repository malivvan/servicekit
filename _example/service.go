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
	Secret    string   `encrypt:"true"`
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

	// Configure Service.
	err := conf.LoadFile("", "secret", &config)
	if err != nil {
		return err
	}
	fmt.Println(config)

	// Configure Logging.
	err = log.Start(config.Logging)
	if err != nil {
		return err
	}
	log.Info().Msg("starting service")

	// ...

	log.Info().Msg("service started")
	return nil
}

func (s *service) Stop() error {
	log.Info().Msg("stopping service")

	// ...

	log.Info().Msg("stopped service")
	log.Stop()
	return nil
}
