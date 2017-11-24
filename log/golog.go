package log

import (
	"strings"
	"errors"
	"fmt"
	"github.com/mkideal/log/provider"
	"github.com/mkideal/log/logger"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
)

const (
	LvFATAL = logger.FATAL
	LvERROR = logger.ERROR
	LvWARN  = logger.WARN
	LvINFO  = logger.INFO
	LvDEBUG = logger.DEBUG
	LvTRACE = logger.TRACE
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

// Uninit uninits log package
func Uninit(glogger logger.Logger) {
	glogger.Quit()
}

// InitWithLogger inits global logger with a specified logger
func InitWithLogger(l logger.Logger) (logger.Logger, error) {
	glogger := l
	glogger.Run()
	return glogger, nil
}

// InitWithProvider inits global logger(sync) with a specified provider
func InitWithProvider(p logger.Provider) (logger.Logger, error) {
	l := logger.New(p)
	l.SetLevel(LvINFO)
	return InitWithLogger(l)
}

// InitSyncWithProvider inits global logger(async) with a specified provider
func InitSyncWithProvider(p logger.Provider) (logger.Logger, error) {
	l := logger.NewSync(p)
	l.SetLevel(LvINFO)
	return InitWithLogger(l)
}

// Init inits global logger with providerType and opts (opts is a json string or empty)
func Init(providerTypes string, opts interface{}) (logger.Logger, error) {
	// splits providerTypes by '/'
	types := strings.Split(providerTypes, "/")
	if len(types) == 0 || len(providerTypes) == 0 {
		err := errors.New("empty providers")
		//glogger.Error(1, "init log error: %v", err)
		return nil, err
	}
	// gets opts string
	optsString := ""
	switch c := opts.(type) {
	case string:
		optsString = c
	case jsonStringer:
		optsString = c.JSON()
	case fmt.Stringer:
		optsString = c.String()
	default:
		optsString = fmt.Sprintf("%v", opts)
	}

	// clean repeated provider type
	usedTypes := map[string]bool{}
	for _, typ := range types {
		typ = strings.TrimSpace(typ)
		usedTypes[typ] = true
	}

	// creates providers
	var providers []logger.Provider
	for typ := range usedTypes {
		creator := logger.Lookup(typ)
		if creator == nil {
			err := errors.New("unregistered provider type: " + typ)
			//glogger.Error(1, "init log error: %v", err)
			return nil, err
		}
		p := creator(optsString)
		if len(usedTypes) == 1 {
			return InitWithProvider(p)
		}
		providers = append(providers, p)
	}
	return InitWithProvider(provider.NewMixProvider(providers[0], providers[1:]...))
}

// InitConsole inits with console provider by toStderrLevel
func InitConsole(toStderrLevel logger.Level) (logger.Logger, error) {
	return Init("console", makeConsoleOpts(toStderrLevel))
}

func InitColoredConsole(toStderrLevel logger.Level) (logger.Logger, error) {
	return Init("colored_console", makeConsoleOpts(toStderrLevel))
}

// InitFile inits with file provider by log file fullpath
func InitFile(fullpath string) (logger.Logger, error) {
	return Init("file", makeFileOpts(fullpath))
}

func makeFileOpts(fullpath string) string {
	dir, filename := filepath.Split(fullpath)
	if dir == "" {
		dir = "."
	}
	return fmt.Sprintf(`{"dir":%s,"filename":%s}`, strconv.Quote(dir), strconv.Quote(filename))
}
func makeConsoleOpts(toStderrLevel logger.Level) string {
	return fmt.Sprintf(`{"tostderrlevel":%d}`, toStderrLevel)
}

// InitFileAndConsole inits with console and file providers
func InitFileAndConsole(fullpath string, toStderrLevel logger.Level) (logger.Logger, error) {
	fileOpts := makeFileOpts(fullpath)
	consoleOpts := makeConsoleOpts(toStderrLevel)
	p := provider.NewMixProvider(provider.NewFile(fileOpts), provider.NewConsole(consoleOpts))
	return InitWithProvider(p)
}

// InitMultiFile inits with multifile provider
func InitMultiFile(rootdir, filename string) (logger.Logger, error) {
	return Init("multifile", makeMultiFileOpts(rootdir, filename))
}

func makeMultiFileOpts(rootdir, filename string) string {
	return fmt.Sprintf(`{"rootdir":"%s","filename":"%s"}`, rootdir, filename)
}

// InitMultiFileAndConsole inits with console and multifile providers
func InitMultiFileAndConsole(rootdir, filename string, toStderrLevel logger.Level) (logger.Logger, error) {
	multifileOpts := makeMultiFileOpts(rootdir, filename)
	consoleOpts := makeConsoleOpts(toStderrLevel)
	p := provider.NewMixProvider(provider.NewMultiFile(multifileOpts), provider.NewConsole(consoleOpts))
	return InitWithProvider(p)
}

func NoHeader(l logger.Logger)                     { l.NoHeader() }
func GetLevel(l logger.Logger) logger.Level        { return l.GetLevel() }
func SetLevel(l logger.Logger, level logger.Level) { l.SetLevel(level) }

// HTTPHandlerGetLevel returns a http handler for getting log level
func HTTPHandlerGetLevel(l logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, GetLevel(l).String())
	})
}

// HTTPHandlerSetLevel sets new log level and returns old log level
// Returns status code `StatusBadRequest` if parse log level fail
func HTTPHandlerSetLevel(l logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		level := r.FormValue("level")
		lv, ok := ParseLevel(level)
		// invalid parameter
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "invalid log level: "+level)
			return
		}
		// not modified
		oldLevel := GetLevel(l)
		if lv == oldLevel {
			w.WriteHeader(http.StatusNotModified)
			io.WriteString(w, oldLevel.String())
			return
		}
		// updated
		SetLevel(l, lv)
		io.WriteString(w, oldLevel.String())
	})
}

// SetLevelFromString parses level from string and set parsed level
// (NOTE): set level to INFO if parse failed
func SetLevelFromString(l logger.Logger, s string) logger.Level {
	level, _ := ParseLevel(s)
	l.SetLevel(level)
	return level
}

// ParseLevel parses log level from string
func ParseLevel(s string) (logger.Level, bool) { return logger.ParseLevel(s) }

// MustParseLevel is similar to ParseLevel, but panics if parse failed
func MustParseLevel(s string) logger.Level { return logger.MustParseLevel(s) }

type GoLogger struct {
	logger.Logger
}

func (l *GoLogger) LSetLevel(level logger.Level) {
	l.SetLevel(level)
}

// Trace outputs trace-level logs
func (l *GoLogger) LTrace(format string, args ...interface{}) {
	l.Trace(20, format, args...)
}

// Debug outputs debug-level logs
func (l *GoLogger) LDebug(format string, args ...interface{}) {
	l.Debug(20, format, args...)

}

// Info outputs info-level logs
func (l *GoLogger) LInfo(format string, args ...interface{}) {
	l.Info(20, format, args...)
}

// Warn outputs warn-level logs
func (l *GoLogger) LWarn(format string, args ...interface{}) {
	l.Warn(20, format, args...)
}

// Error outputs error-level logs
func (l *GoLogger) LError(format string, args ...interface{}) {
	l.Error(20, format, args...)
}

// Fatal outputs fatal-level logs
func (l *GoLogger) LFatal(format string, args ...interface{}) {
	l.Fatal(20, format, args...)
}

func GetLogger(s string) (GoLogger) {
	logger, err := InitFile(s)
	_logger := GoLogger{logger}
	if err != nil {
		_logger.LError("get logger error")
		return _logger
	}
	return _logger
}

func DeferLogger(_logger logger.Logger) {
	_logger.Quit()
}
