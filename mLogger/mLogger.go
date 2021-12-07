package mLogger

import (
	"io"
	"os"
	"sync"

	"github.com/hashicorp/go-hclog"
)

const defaultLevel = "info"

var once sync.Once
var output io.Writer
var mx sync.Mutex

// loggers added as Named
var loggers map[string]hclog.Logger

// top level logger
var logger hclog.Logger

func init() {
	once.Do(func() {
		mx = sync.Mutex{}
		logger = nil       // asserts New is called once
		output = os.Stderr // default out
		loggers = make(map[string]hclog.Logger)
	})
}

// New create a new top level logger with hclog.LevelFromString
// Subsequent modules should call Get.
// Default level is "info" & default out is stderr
func New(name, lvl string, out io.Writer) hclog.Logger {
	if lvl == "" {
		lvl = defaultLevel
	}
	if out != nil {
		output = out
	}
	opts := hclog.LoggerOptions{
		Name:        "[" + name + "]",
		Level:       hclog.LevelFromString(lvl),
		Mutex:       &sync.Mutex{},
		DisableTime: true,
		Color:       hclog.AutoColor,
		Output:      output,
	}
	logger = hclog.New(&opts)
	loggers[name] = hclog.New(&opts)
	return loggers[name]
}

// Get returns a named logger by either creating a sub logger or
// returning existing one. If no top level logger exists, the first call to Get
// creates a top level logger
func Get(name string) hclog.Logger {
	mx.Lock()
	defer mx.Unlock()
	if logger == nil {
		return New(name, defaultLevel, output)
	}
	if loggers[name] == nil {
		loggers[name] = logger.Named(name)
	}
	return loggers[name]
}
