package mLogger

import (
	"sync"

	"github.com/hashicorp/go-hclog"
)

var once sync.Once

// loggers added as Named
var loggers map[string]hclog.Logger

// top level logger
var logger hclog.Logger

func init() {
	once.Do(func() {
		logger = nil
		loggers = make(map[string]hclog.Logger)
	})
}

// New create a new top level logger
// Subsequent modules should call Get
func New(name string) hclog.Logger {
	m := sync.Mutex{}
	opts := hclog.LoggerOptions{
		Name:        "[" + name + "]",
		Level:       hclog.LevelFromString("info"),
		Mutex:       &m,
		DisableTime: true,
		Color:       hclog.AutoColor,
	}
	logger = hclog.New(&opts)
	loggers[name] = hclog.New(&opts)
	return loggers[name]
}

// Get returns a named logger by either creating a new named logger or
// returning existing one
func Get(name string) hclog.Logger {
	if loggers[name] == nil {
		loggers[name] = logger.Named(name)
	}
	return loggers[name]
}
