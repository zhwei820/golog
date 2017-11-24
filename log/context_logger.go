package log

import (
	"encoding/json"
)

// Formatter formats the data of context
type Formatter interface {
	Format(v interface{}) []byte
}

type JSONFormatter struct{}

var jsonFormatter = JSONFormatter{}

func (f JSONFormatter) Format(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

type jsonStringer interface {
	JSON() string
}

// M aliases map
type M map[string]interface{}

func (m M) JSON() string {
	b, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// S aliases slice
type S []interface{}

// Context represents a context of logger
type Context interface {
	With(values ...interface{}) ContextLogger
	WithJSON(values ...interface{}) ContextLogger
	SetFormatter(f Formatter) ContextLogger
}

type ContextLogger interface {
	Context
	Trace(format string, args ...interface{}) ContextLogger
	Debug(format string, args ...interface{}) ContextLogger
	Info(format string, args ...interface{}) ContextLogger
	Warn(format string, args ...interface{}) ContextLogger
	Error(format string, args ...interface{}) ContextLogger
	Fatal(format string, args ...interface{}) ContextLogger
}
