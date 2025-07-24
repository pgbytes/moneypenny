package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"runtime/debug"
	"strings"
	"time"
)

const (
	// Logger name for global logs
	loggerNameRoot = "root"
)

var (
	globalLogger  *zap.SugaredLogger
	defaultLogger *zap.SugaredLogger
)

// LoggingConfig has the basic configuration for log
type LoggingConfig struct {
	Level  string
	Format string
}

func (l LoggingConfig) LogLevel() string {
	return l.Level
}

func (l LoggingConfig) LogFormat() string {
	return l.Format
}

// NewLoggingConfig returns an instance of log config. Acceptable values:
// level: info, debug, warn, warning, error
// format: json, console
func NewLoggingConfig(level, format string) LoggingConfig {
	return LoggingConfig{
		Level:  level,
		Format: format,
	}
}

type Config interface {
	LogLevel() string
	LogFormat() string
}

// Logger log interface to be used for passing as method arguments where required.
type Logger interface {
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
}

// SetupLogging sets a global logger according to the configuration passed,
// that can then be used using from other packages using log.GetLogger or
// log.GetLoggerWithRequestId
func SetupLogging(config Config) error {
	if globalLogger != nil {
		return nil
	}
	format := parseLogFormat(config.LogFormat())
	var err error
	globalLogger, err = setupNewLogger(format, config.LogLevel())
	if err != nil {
		return err
	}
	globalLogger.Debug("GlobalLogger is now setup")
	return err
}

// GetLogger returns the configured global logger
// audit fields extracted from context if present:
// field `log-name` is populated as: `com.egym.<service-name>, `service-name` is extracted from go module name
//
// NOTE: You should run SetupLogging first before using this function in order
// to use the configuration passed to the Device Gateway, otherwise it will use
// a default configuration: using json encoding and info as minimum logging level
func GetLogger() *zap.SugaredLogger {
	return getGlobalLogger().Named(getLoggerNameForType(loggerNameRoot))
}

// SetupNewLogger creates a new logger using the encoding and minimum logging
// level defined. In addition, it uses RFC3339 with UTC Time, json encoding,
// and customizes the keys used for message, level and time to fit the
// conventions used in GCP
func setupNewLogger(encoding, minLogLevel string) (*zap.SugaredLogger, error) {
	loggerConfig := getConfig(encoding, parseMinLogLevel(minLogLevel))
	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("could not build a logger from the config: %v", err)
	}
	logger.Debug("Logger is now setup")
	logger.Debug("Counter metric (call GetLogCounter()) is ready to be registered")
	return logger.Sugar(), nil
}

func getConfig(encoding string, minLogLevel zapcore.Level) zap.Config {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(minLogLevel)
	loggerConfig.EncoderConfig.EncodeTime = utcRFC3339TimeEncoder
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	loggerConfig.EncoderConfig.MessageKey = "message"
	loggerConfig.EncoderConfig.LevelKey = "severity"
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.CallerKey = "method"
	loggerConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	loggerConfig.EncoderConfig.NameKey = "log-name"
	loggerConfig.EncoderConfig.EncodeName = zapcore.FullNameEncoder
	loggerConfig.OutputPaths = []string{"stdout"}
	loggerConfig.Encoding = "json"
	if encoding == "console" {
		loggerConfig.Encoding = "console"
	}
	return loggerConfig
}

// UTCRFC3339TimeEncoder encodes the time as UTC and formats it according to RFC3339, i.e: 2020-03-12T17:51:03Z
func utcRFC3339TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	t = t.UTC()
	enc.AppendString(t.Format(time.RFC3339))
}

// parseMinLogLevel reads the string from the config, parses it into fit into
// one of the zapcore Levels used to configure the loggers minimum logging
// level
func parseMinLogLevel(minLogLevel string) zapcore.Level {
	switch strings.ToLower(minLogLevel) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warning":
		fallthrough
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		log.Printf("unknown log level: %v. Will default to info", minLogLevel)
		return zapcore.InfoLevel
	}
}

func parseLogFormat(format string) string {
	encoding := format
	if format != "json" && format != "console" {
		log.Printf("unknown log format: %v. Will default to json", format)
		encoding = "json"
	}
	return encoding
}

func getGlobalLogger() *zap.SugaredLogger {
	if defaultLogger == nil && globalLogger == nil {
		var err error
		defaultLogger, err = setupNewLogger("json", "info")
		if err != nil {
			log.Fatal(err)
		}
		defaultLogger.Warn("GlobalLogger is not configured yet. Using default logger until `SetupLogging` is run")
		return defaultLogger
	} else if defaultLogger != nil && globalLogger == nil {
		return defaultLogger
	}
	return globalLogger
}

func getLoggerNameForType(loggerType string) string {
	serviceName := extractServiceName()
	switch loggerType {
	case loggerNameRoot:
		return strings.Join([]string{"com", "github", serviceName}, ".")
	default:
		return strings.Join([]string{"com", "github", loggerType, serviceName}, ".")
	}
}

// extractServiceName returns the module name from a running binary built with go module support
// Returns `UNSPECIFIED_SERVICE_NAME` if build info cannot be read
func extractServiceName() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "UNSPECIFIED_SERVICE_NAME"
	}
	moduleServiceName := buildInfo.Main.Path
	return moduleServiceName[strings.LastIndex(moduleServiceName, "/")+1:]
}
