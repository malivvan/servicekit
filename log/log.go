package log

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kardianos/service"
	"github.com/malivvan/servicekit"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	fileLogger *lumberjack.Logger
	logger     zerolog.Logger
)

type Config struct {
	// Level of the logger. Valid options: debug, info, warn, error, disable.
	Level string `json:"level"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `json:"maxage"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"maxbackups"`

	// Compress determines if the rotated log files should be compressed
	// using gzip.
	Compress bool `json:"compress"`
}

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFieldName = "time"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
}

func Start(config Config, logWriters ...io.Writer) error {

	// create file logger path
	path := servicekit.Workdir("log", servicekit.Name()+".json")
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return err
	}

	// create lumberjack file logger
	fileLogger = &lumberjack.Logger{
		Filename:   path,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		Compress:   config.Compress,
	}
	logWriters = append(logWriters, fileLogger)

	// if interactive write formated log to stderr, disable color on windows
	if service.Interactive() {
		logWriters = append(logWriters, zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: runtime.GOOS == "windows",
		})
	}

	// set loglevel
	switch config.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "disable":
		zerolog.SetGlobalLevel(zerolog.Disabled)
	default:
		return errors.New("unrecognized logging level '" + config.Level + "'")
	}

	// initialize logger
	logger = zerolog.New(io.MultiWriter(logWriters...)).With().Timestamp().Logger()
	return nil
}

// Stop closes the underlying file logger. Should be called at the end
// of the service Stop method.
func Stop() error {
	return fileLogger.Close()
}

// Debug starts a new message with debug level.
//
// You must call Msg on the returned event in order to send the event.
func Debug() *zerolog.Event {
	return logger.Debug()
}

// Info starts a new message with info level.
//
// You must call Msg on the returned event in order to send the event.
func Info() *zerolog.Event {
	return logger.Info()
}

// Warn starts a new message with warn level.
//
// You must call Msg on the returned event in order to send the event.
func Warn() *zerolog.Event {
	return logger.Warn()
}

// Error starts a new message with error level.
//
// You must call Msg on the returned event in order to send the event.
func Error() *zerolog.Event {
	return logger.Error()
}

// With creates a new logger context.
//
// You must call Logger on the returned Context in order to get a logger.
func With() zerolog.Context {
	return logger.With()
}
