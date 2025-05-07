package logging

import (
    "fmt"

    "os"
    "time"
    "strings"

    // Structured logging
    "github.com/rs/zerolog"
)

var Logger zerolog.Logger

func InitLogger(level string, releaseMode bool) {
    zerolog.TimeFieldFormat = time.RFC3339
    // Use colorful non-JSON log output during development builds
    if releaseMode {
        Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
    } else {
        Logger = zerolog.New(
            zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05", FormatLevel: func(i interface{}) string {
                return strings.ToUpper(fmt.Sprintf("%-5s", i))
            }},
        ).With().Timestamp().Logger()
    }

    switch strings.ToLower(level) {
        case "error":
            zerolog.SetGlobalLevel(zerolog.ErrorLevel)
        case "warn":
            zerolog.SetGlobalLevel(zerolog.WarnLevel)
        case "info":
            zerolog.SetGlobalLevel(zerolog.InfoLevel)
        case "debug":
            zerolog.SetGlobalLevel(zerolog.DebugLevel)
        case "trace":
            zerolog.SetGlobalLevel(zerolog.TraceLevel)
        case "disable":
            zerolog.SetGlobalLevel(zerolog.Disabled)
        default:
            Logger.Fatal().Msgf("error parsing commandline arguments: invalid value \"%v\" for flag -logLevel: pass one of: disable, error, warn, info, debug, trace", level)
    }
}

