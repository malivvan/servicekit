package mon

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/rs/zerolog"
)

type LogWriter struct{}

func (w LogWriter) Write(p []byte) (n int, err error) {

	// decode event
	var evt map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err = d.Decode(&evt)
	if err != nil {
		return 0, err
	}

	// decode and remove timestamp
	tsIf, ok := evt[zerolog.TimestampFieldName]
	if !ok {
		return 0, errors.New("cannot identify timestamp field")
	}
	tsStr, ok := tsIf.(string)
	if !ok {
		return 0, errors.New("timestamp is not a string")
	}
	ts, err := time.Parse(zerolog.TimeFieldFormat, tsStr)
	if err != nil {
		return 0, err
	}
	delete(evt, zerolog.TimestampFieldName)

	// decode and remove log level
	levelIf, ok := evt[zerolog.LevelFieldName]
	if !ok {
		return 0, errors.New("cannot identify timestamp field")
	}
	levelStr, ok := levelIf.(string)
	if !ok {
		return 0, errors.New("timestamp is not a string")
	}
	delete(evt, zerolog.LevelFieldName)

	// write to log measurement with all subfields
	Write("log", ts, map[string]string{zerolog.LevelFieldName: levelStr}, evt)
	return len(p), nil
}
