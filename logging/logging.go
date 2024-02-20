package logging

import (
    "fmt"

    "os"
    "time"
    "strings"

    // Structured logging
    "github.com/rs/zerolog"
    "github.com/gin-gonic/gin"
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

// https://learninggolang.com/it5-gin-structured-logging.html
func GinLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        path := c.Request.URL.Path
        raw := c.Request.URL.RawQuery

        // Process request
        c.Next()

        // Fill the params
        param := gin.LogFormatterParams{}

        param.TimeStamp = time.Now() // Stop timer
        param.ClientIP = c.ClientIP()
        param.Method = c.Request.Method
        param.StatusCode = c.Writer.Status()
        param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
        param.BodySize = c.Writer.Size()
        if raw != "" {
            path = path + "?" + raw
        }
        param.Path = path

        xff, ok := c.Request.Header["X-Forwarded-For"]
        if ok {
            var xffAllValues []string
            for _, h := range(xff) {
                for _, v := range(strings.Split(h, ",")) {
                    xffAllValues = append(xffAllValues, strings.TrimSpace(v))
                }
            }
            xff = xffAllValues
        } else {
            xff = []string{}
        }

        // Log using the params
        var logEvent *zerolog.Event
        if c.Writer.Status() >= 500 {
            logEvent = Logger.Error()
        } else {
            logEvent = Logger.Info()
        }

        logEvent.Str("remote_addr", c.Request.RemoteAddr).
            Strs("client_ips", xff).
            Str("method", param.Method).
            Int("status_code", param.StatusCode).
            Int("body_size", param.BodySize).
            Str("path", param.Path).
            Msg(param.ErrorMessage)
    }
}
