package logging

import (
    "fmt"

    "os"
    "time"
    "strings"
    "regexp"

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

func GetIPFromHeader(req *gin.Context) string{
    checkHeader := regexp.MustCompile(`(?i)x-real-ip`)
    for key, value := range req.Request.Header {
        if checkHeader.MatchString(key) {
            return value[0]
        }
    }
    return req.ClientIP()
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

        param.ClientIP =GetIPFromHeader(c)
        param.Method = c.Request.Method
        param.StatusCode = c.Writer.Status()
        param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
        //param.Latency = duration
        param.BodySize = c.Writer.Size()
        if raw != "" {
            path = path + "?" + raw
        }
        param.Path = path

        // Log using the params
        var logEvent *zerolog.Event
        if c.Writer.Status() >= 500 {
            logEvent = Logger.Error()
        } else {
            logEvent = Logger.Info()
        }

        logEvent.Str("client_id", param.ClientIP).
            Str("method", param.Method).
            Int("status_code", param.StatusCode).
            Int("body_size", param.BodySize).
            Str("path", param.Path).
            //Str("latency", param.Latency.String()).
            Msg(param.ErrorMessage)
    }
}
