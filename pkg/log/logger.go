package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	black = iota + 30
	red
	green
	yellow
	violet
	magenta
	cyan
	white

	bold     = 1
	darkGray = 90

	readableTimeFormat = "2006-01-02 15:04:05.000000"

	levelTraceValue = "trace"
	levelDebugValue = "debug"
	levelInfoValue  = "info"
	levelWarnValue  = "warn"
	levelErrorValue = "error"
	levelFatalValue = "fatal"
	levelPanicValue = "panic"
)

var (
	config Config
	Logger *zerolog.Logger
)

type Config struct {
	// Enable console logging
	ConsoleLoggingEnabled bool
	// Is console logging will be colorized
	Colorized bool
}

func init() {
	var writers []io.Writer

	zerolog.TimeFieldFormat = readableTimeFormat

	// buildInfo, _ := debug.ReadBuildInfo()

	infoSampler := &zerolog.BurstSampler{
		Burst:  5,
		Period: 1 * time.Second,
	}
	warnSampler := &zerolog.BurstSampler{
		Burst:       5,
		Period:      1 * time.Second,
		NextSampler: &zerolog.BasicSampler{N: 5},
	}

	config = Config{
		ConsoleLoggingEnabled: true,
		Colorized:             true,
	}
	if config.ConsoleLoggingEnabled {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		writers = append(writers, consoleLogger())
	}

	mw := io.MultiWriter(writers...)

	logger := zerolog.New(mw).
		With().
		Timestamp().
		Caller().
		// Str("go_version", buildInfo.GoVersion).
		Logger().
		Sample(zerolog.LevelSampler{
			InfoSampler: infoSampler,
			WarnSampler: warnSampler,
		})

	logger.Info().
		Msg("logging configured")

	Logger = &logger
}

func consoleLogger() zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out: os.Stdout, TimeFormat: readableTimeFormat,
		FormatTimestamp: func(i interface{}) string {
			return colorize(darkGray, fmt.Sprintf("|%s|", i))
		},
		FormatLevel: func(i interface{}) string {

			s := strings.ToUpper(fmt.Sprintf("[%s]", i))
			switch i {
			case levelTraceValue:
				return s
			case levelDebugValue:
				return colorize(red, s)
			case levelInfoValue:
				return colorize(white, s)
			case levelWarnValue:
				return colorize(yellow, s)
			case levelErrorValue:
				return colorize(red, s)
			case levelFatalValue:
				return colorize(bold, colorize(red, s))
			case levelPanicValue:
				return colorize(bold, colorize(red, s))
			default:
				return s
			}
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("| %s |", i)
		},
		FormatCaller: func(i interface{}) string {
			for skip := 5; skip < 15; skip++ { // Перебираем стек вызовов
				_, file, line, ok := runtime.Caller(skip)
				if ok && !strings.Contains(file, "/pkg/mod/") {
					return fmt.Sprintf("%s:%d", file, line) // Возвращаем первый корректный путь
				}
			}
			return filepath.Base(fmt.Sprintf("%s", i))
		},
	}
}

func colorize(c int, s string) string {
	if !config.Colorized {
		return s
	}
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", c, s)
}
