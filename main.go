//go:generate goversioninfo -platform-specific=true

package main

import (
    "fmt"
    "os"
    "flag"
    "sync"
    "time"
    "strings"
    "unicode"
    "net/http"
    "net/netip"
    "path/filepath"

    // Configuration
    "github.com/peterbourgon/ff/v3"
    "gopkg.in/yaml.v3"

    // for live-reload of manifests
    "github.com/rjeczalik/notify"

    "rewinged/settings"
    "rewinged/logging"
    "rewinged/models"
    "rewinged/controllers"
)

// These variables are overwritten at compile/link time using -ldflags
var version = "development-build"
var commit = "unknown"
var compileTime = "unknown"
var releaseMode = "false"

var wg sync.WaitGroup
var jobs chan string = make(chan string)

func main() {
    fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
    var (
        versionFlagPtr = fs.Bool("version", false, "Print the version information and exit")
        packagePathPtr = fs.String("manifestPath", "./packages", "The directory to search for package manifest files")

        tlsEnablePtr           = fs.Bool("https", false, "Serve encrypted HTTPS traffic directly from rewinged without the need for a proxy")
        tlsCertificatePtr      = fs.String("httpsCertificateFile", "./cert.pem", "The webserver certificate to use if HTTPS is enabled")
        tlsPrivateKeyPtr       = fs.String("httpsPrivateKeyFile", "./private.key", "The private key file to use if HTTPS is enabled")
        listenAddrPtr          = fs.String("listen", "localhost:8080", "The address and port for the REST API to listen on")
        autoInternalizePtr     = fs.Bool("autoInternalize", false, "Turn on the auto-internalization feature")
        autoInternalizePathPtr = fs.String("autoInternalizePath", "./installers", "The directory where auto-internalized installers will be stored")
        autoInternalizeSkipPtr = fs.String("autoInternalizeSkip", "", "List of hostnames excluded from auto-internalization (comma or space to separate)")
        sourceAuthTypePtr             = fs.String("sourceAuthType", "none", "Require authentication to interact with the REST API: none, microsoftEntraId")
        sourceAuthEntraIDResourcePtr  = fs.String("sourceAuthEntraIDResource", "", "ApplicationID of the EntraID App used for authenticating clients")
        sourceAuthEntraIDAuthorityURL = fs.String("sourceAuthEntraIDAuthorityURL", "", "Authority/Issuer URL of the EntraID App used for authenticating clients")
        packageAuthorizationRulesPtr = fs.String("packageAuthorizationRulesFile", "", "Path to a YAML file defining granular allow/deny rules (optional)")
        logLevelPtr            = fs.String("logLevel", "info", "Set log verbosity: disable, error, warn, info, debug or trace")
        trustedProxiesPtr      = fs.String("trustedProxies", "", "List of IPs from which to trust Client-IP headers (comma or space to separate)")
        _                      = fs.String("configFile", "", "Path to a json configuration file (optional)")
    )

    // Ingest configuration flags.
    // Commandline arguments > Environment variables > config file
    err := ff.Parse(fs, os.Args[1:],
        ff.WithEnvVarPrefix("REWINGED"),
        ff.WithConfigFileFlag("configFile"),
        ff.WithConfigFileParser(ff.JSONParser),
    )

    if err != nil {
        // Replicate default ExitOnError behavior of exiting with 0 when -h/-help/--help is used
        if strings.HasSuffix(err.Error(), "help requested") {
            os.Exit(0)
        }
        fmt.Println(err)
        os.Exit(2)
    }

    if *versionFlagPtr {
        fmt.Printf("rewinged %v\n\ncommit:\t\t%v\ncompiled:\t%v\n", version, commit, compileTime)
        os.Exit(0)
    }

    logging.InitLogger(*logLevelPtr, releaseMode == "true")

    if *sourceAuthTypePtr != "microsoftEntraId" && *sourceAuthTypePtr != "none" {
        logging.Logger.Fatal().Msg("sourceAuthType must be either none or microsoftEntraId")
    }

    // sourceAuthEntraIDResource is required if sourceAuthType is "microsoftEntraId"
    if *sourceAuthTypePtr == "microsoftEntraId" && *sourceAuthEntraIDResourcePtr == "" {
        logging.Logger.Fatal().Msg("sourceAuthEntraIDResource is required when sourceAuthType is set to microsoftEntraId")
    }

    // sourceAuthEntraIDAuthorityURL is required if sourceAuthType is "microsoftEntraId"
    if *sourceAuthTypePtr == "microsoftEntraId" && *sourceAuthEntraIDAuthorityURL == "" {
        logging.Logger.Fatal().Msg("sourceAuthEntraIDAuthorityURL is required when sourceAuthType is set to microsoftEntraId")
    }

    settings.SourceAuthenticationType = *sourceAuthTypePtr
    settings.SourceAuthenticationEntraIDResource = *sourceAuthEntraIDResourcePtr
    settings.SourceAuthenticationEntraIDAuthorityURL = *sourceAuthEntraIDAuthorityURL

    // Users can set 0.0.0.0/0 or ::/0 to trust all proxies if need be
    if (*trustedProxiesPtr != "") {
        trustedProxies := strings.FieldsFunc(*trustedProxiesPtr, func(c rune) bool {
            return unicode.IsSpace(c) || c == ','
        })

        for _, proxy := range(trustedProxies) {
            var prefix netip.Prefix
            var err error
            if !strings.Contains(proxy, "/") {
                var addr netip.Addr
                addr, err = netip.ParseAddr(proxy)
                // addr.Prefix() cannot error if called with addr.Bitlen()
                prefix, _ = addr.Prefix(addr.BitLen())
            } else {
                prefix, err = netip.ParsePrefix(proxy)
            }
            if err != nil {
                logging.Logger.Fatal().Err(err).Msg("invalid trustedProxies")
            }
            logging.TrustedProxies = append(logging.TrustedProxies, prefix)
        }
    }

    //authConfig := models.GetInitialAuthorizationConfig_1()
    if (*packageAuthorizationRulesPtr != "") {
        authConfigFile, err := os.Open(*packageAuthorizationRulesPtr)
        if err == nil {
            fileDecoder := yaml.NewDecoder(authConfigFile)
            err = fileDecoder.Decode(&settings.PackageAuthorizationConfig)
            if err != nil {
                logging.Logger.Fatal().Err(err).Msgf("error unmarshaling packageAuthorizationRules file")
            }
        } else {
            logging.Logger.Fatal().Err(err).Msgf("error opening packageAuthorizationRules file")
        }
    }
    logging.Logger.Debug().Msgf("packageAuthorizationRules: %+v", settings.PackageAuthorizationConfig)
    // TODO: actually enforce this ruleset

    logging.Logger.Debug().Msg("searching for manifests")

    autoInternalizeSkipHosts := strings.FieldsFunc(*autoInternalizeSkipPtr, func(c rune) bool {
        return unicode.IsSpace(c) || c == ','
    })

    // Start up 6 worker goroutines that can parse in manifest-files from one directory each
    for w := 1; w <= 6; w++ {
        go ingestManifestsWorker(*autoInternalizePtr, *autoInternalizePathPtr, autoInternalizeSkipHosts)
    }

    getManifests(*packagePathPtr)
    wg.Wait()

    // I don't know whether this is safe.
    // if manifests is just a reference-copy of manifests2 then it wouldn't be I think?
    // But *currently* since live-reload isn't implemented yet, manifests2 won't be written
    // to after this point so it's safe for now - TODO: only access manifests2 in a thread-safe way
    logging.Logger.Info().Msgf("found %v package manifests", models.Manifests.GetManifestCount())

    logging.Logger.Info().Msg("watching manifestPath for changes")
    // Make the channel buffered to try and not miss events. Notify will drop
    // an event if the receiver is not able to keep up the sending pace.
    fileEventsBuffer := 100
    fileEventsChannel := make(chan notify.EventInfo, fileEventsBuffer)

    // Recursively listen for Create and Write events in the manifestPath.
    // Currently not watching for remove / delete events because we couldn't
    // correlate filenames to packages anyway so there's no way to know which
    // package is affected by the event.
    if err := notify.Watch(*packagePathPtr + "/...", fileEventsChannel, notify.Create, notify.Write); err != nil {
        logging.Logger.Fatal().Err(err)
    }
    defer notify.Stop(fileEventsChannel)

    // If an event is received, push its directory-path to the jobs channel
    go func() {
        for {
            // Detect and handle channel overflow
            // This is a loop because it is possible for the channel to fill up
            // multiple times in a row if events are flooding in for a prolonged
            // period of time, thus necessitating further full rescans
            for len(fileEventsChannel) == fileEventsBuffer {
                // If the channel is ever full we are missing events as the notify package drops them at this point
                //log.Println("\x1b[31mfileEventsChannel full - we're missing events - will perform full manifest rescan\x1b[0m")
                logging.Logger.Info().Msg("fileEventsChannel full - we're missing events - will perform full manifest rescan")
                // Wait out the thundering herd - events have been lost anyway
                time.Sleep(5 * time.Second)
                // Drop all events to clear the channel, this also enables new events to stream in again
                CLEAR_CHANNEL: for { select { case <- fileEventsChannel:; default: break CLEAR_CHANNEL } }
                getManifests(*packagePathPtr)
                // wait for the synchronous full rescan to finish.
                // any events accumulated in the meantime will be processed after.
                wg.Wait()
            }

            ei := <- fileEventsChannel
            logging.Logger.Debug().Msgf("received event (type %T):\n\t%+v\n", ei, ei)
            wg.Add(1)
            jobs <- filepath.Dir(ei.Path())
        }
    }()

    var getPackagesConfig = &controllers.GetPackageHandler{
        TlsEnabled: *tlsEnablePtr,
        InternalizationEnabled: *autoInternalizePtr,
    }

    router := http.NewServeMux()

    // TODO: Recovery maybe?

    fileServer := http.FileServer(http.Dir(*autoInternalizePathPtr))
    router.HandleFunc("GET /api/information", controllers.GetInformation)

    switch settings.SourceAuthenticationType {
    case "none":
        router.Handle("/installers/", http.StripPrefix("/installers", hideDirectoryListings(fileServer)))
        router.Handle("GET /api/packages", http.HandlerFunc(controllers.GetPackages))
        router.Handle("POST /api/manifestSearch", http.HandlerFunc(controllers.SearchForPackage))
        router.Handle("GET /api/packageManifests/{package_identifier}", http.HandlerFunc(getPackagesConfig.GetPackage))
    case "microsoftEntraId":
        router.Handle("/installers/", http.StripPrefix("/installers", controllers.JWTAuthMiddleware(hideDirectoryListings(fileServer))))
        router.Handle("GET /api/packages", controllers.JWTAuthMiddleware(http.HandlerFunc(controllers.GetPackages)))
        router.Handle("POST /api/manifestSearch", controllers.JWTAuthMiddleware(http.HandlerFunc(controllers.SearchForPackage)))
        router.Handle("GET /api/packageManifests/{package_identifier}", controllers.JWTAuthMiddleware(http.HandlerFunc(getPackagesConfig.GetPackage)))
    default:
        logging.Logger.Fatal().Msg("sourceAuthType must be either none or microsoftEntraId")
    }

    logging_router := logging.RequestLogger(router)

    if *tlsEnablePtr {
        logging.Logger.Info().Msgf("starting server on https://%v", *listenAddrPtr)
        if err := http.ListenAndServeTLS(*listenAddrPtr, *tlsCertificatePtr, *tlsPrivateKeyPtr, logging_router); err != nil {
            logging.Logger.Fatal().Err(err).Msg("could not start webserver")
        }
    } else {
        logging.Logger.Info().Msgf("starting server on http://%v", *listenAddrPtr)
        if err := http.ListenAndServe(*listenAddrPtr, logging_router); err != nil {
            logging.Logger.Fatal().Err(err).Msg("could not start webserver")
        }
    }
}

func hideDirectoryListings(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasSuffix(r.URL.Path, "/") {
            http.NotFound(w, r)
            return
        }

        next.ServeHTTP(w, r)
    })
}

