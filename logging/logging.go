package logging

import (
    "fmt"

    "os"
    "time"
    "strings"
    "net/http"

    // Structured logging
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/hlog"
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

func RequestLogger(next http.Handler) http.Handler {
    h := hlog.NewHandler(Logger)

    // TODO: Add client_id / client_ip, check trusted proxies!
    accessHandler := hlog.AccessHandler(
        func(r *http.Request, status, size int, duration time.Duration) {
            hlog.FromRequest(r).Info().
                Str("method", r.Method).
                Stringer("url", r.URL).
                Int("status_code", status).
                Int("response_size_bytes", size).
                Dur("elapsed_ms", duration).
                Msg("incoming request")
        },
    )

    userAgentHandler := hlog.UserAgentHandler("http_user_agent")

    return h(accessHandler(userAgentHandler(next)))
}
