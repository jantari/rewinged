//go:generate goversioninfo

package main

import (
    "fmt"
    "log"
    "os"
    "flag"
    "sync"
    "time"
    "path/filepath"

    "github.com/gin-gonic/gin"
    "github.com/rjeczalik/notify" // for live-reload of manifests

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
    versionFlagPtr := flag.Bool("version", false, "Print the version information and exit")
    packagePathPtr := flag.String("manifestPath", "./packages", "The directory to search for package manifest files")

    tlsEnablePtr := flag.Bool("https", false, "Serve encrypted HTTPS traffic directly from rewinged without the need for a proxy")
    tlsCertificatePtr := flag.String("httpsCertificateFile", "./cert.pem", "The webserver certificate to use if HTTPS is enabled")
    tlsPrivateKeyPtr := flag.String("httpsPrivateKeyFile", "./private.key", "The private key file to use if HTTPS is enabled")
    listenAddrPtr := flag.String("listen", "localhost:8080", "The address and port for the REST API to listen on")

    flag.Parse()

    if *versionFlagPtr {
        fmt.Printf("rewinged %v\n\ncommit:\t\t%v\ncompiled:\t%v\n", version, commit, compileTime)
        os.Exit(0)
    }

    fmt.Println("Searching for manifests...")
    // Start up 10 worker goroutines that can parse in manifest-files from one directory each
    for w := 1; w <= 6; w++ {
        go ingestManifestsWorker()
    }

    getManifests(*packagePathPtr)
    wg.Wait()

    // I don't know whether this is safe.
    // if manifests is just a reference-copy of manifests2 then it wouldn't be I think?
    // But *currently* since live-reload isn't implemented yet, manifests2 won't be written
    // to after this point so it's safe for now - TODO: only access manifests2 in a thread-safe way
    fmt.Println("Found", models.Manifests.GetManifestCount(), "package manifests.")

    fmt.Println("Watching manifestPath for changes ...")
    // Make the channel buffered to try and not miss events. Notify will drop
    // an event if the receiver is not able to keep up the sending pace.
    fileEventsBuffer := 100
    fileEventsChannel := make(chan notify.EventInfo, fileEventsBuffer)

    // Recursively listen for Create and Write events in the manifestPath.
    // Currently not watching for remove / delete events because we couldn't
    // correlate filenames to packages anyway so there's no way to know which
    // package is affected by the event.
    if err := notify.Watch(*packagePathPtr + "/...", fileEventsChannel, notify.Create, notify.Write); err != nil {
        log.Fatal(err)
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
                log.Println("\x1b[31mfileEventsChannel full - we're missing events - will perform full manifest rescan\x1b[0m")
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
            log.Printf("Received event (type %T):\n\t%+v\n", ei, ei)
            wg.Add(1)
            jobs <- filepath.Dir(ei.Path())
        }
    }()

    if releaseMode == "true" {
        gin.SetMode(gin.ReleaseMode)
    }
    router := gin.Default()
    router.SetTrustedProxies(nil)
    router.GET("/information", controllers.GetInformation)
    router.GET("/packages", controllers.GetPackages)
    router.POST("/manifestSearch", func(c *gin.Context) {
        var post models.ManifestSearch
        if err := c.BindJSON(&post); err == nil {
            fmt.Printf("%+v\n", post)
            response := &models.ManifestSearchResult{
                RequiredPackageMatchFields: []models.PackageMatchField{},
                Data: []models.ManifestSearchResponse {},
            }

            // results is a map where the PackageIdentifier is the key
            // and the values are arrays of manifests with that PackageIdentifier.
            // This means the values will be different versions of the package.
            var results map[string][]models.Versions

            if post.Query.KeyWord != "" {
                fmt.Println("someone searched the repo for:", post.Query.KeyWord)
                results = getPackagesByKeyword(models.Manifests.GetAll(), post.Query.KeyWord)
            } else if (post.Inclusions != nil && len(post.Inclusions) > 0) || (post.Filters != nil && len(post.Filters) > 0) {
                fmt.Println("advanced search with inclusions[] and/or filters[]")
                results = getPackagesByMatchFilter(models.Manifests.GetAll(), post.Inclusions, post.Filters)
            }

            fmt.Println("... with", len(results), "results:")

            if len(results) > 0 {
                for packageId, packageVersions := range results {
                    fmt.Println("  package", packageId, "with", len(packageVersions), "versions.")
                    var versions []models.ManifestSearchVersion

                    for _, version := range packageVersions {
                        versions = append(versions, models.ManifestSearchVersion{
                            PackageVersion: version.PackageVersion,
                            Channel: "",
                            PackageFamilyNames: []string{},
                            ProductCodes: []string{ version.Installers[0].ProductCode },
                        })
                    }

                    response.Data = append(response.Data, models.ManifestSearchResponse{
                        PackageIdentifier: packageId,
                        PackageName: packageVersions[0].DefaultLocale.PackageName,
                        Publisher: packageVersions[0].DefaultLocale.Publisher,
                        Versions: versions,
                    })
                }
                fmt.Printf("%+v\n", response)
                c.JSON(200, response)
            } else {
                // winget REST-API specification calls for a 204 return code if no results were found
                c.JSON(204, response)
            }
        } else {
            log.Println("error deserializing json post body %v\n", err)
        }
    })
    router.GET("/packageManifests/:package_identifier", func(c *gin.Context) {
        fmt.Println("/packageManifests: Someone tried to GET package '", c.Param("package_identifier"), "'")
        fmt.Println("with query params:", c.Request.URL.Query())

        response := models.ManifestSingleResponse {
            RequiredQueryParameters: []models.QueryParameter{},
            UnsupportedQueryParameters: []models.QueryParameter{},
            Data: nil,
        }

        var pkg []models.Versions = models.Manifests.GetAllVersions(c.Param("package_identifier"))
        if len(pkg) > 0 {
            fmt.Println("the package was found!")

            response.Data = &models.Manifest{
                PackageIdentifier: c.Param("package_identifier"),
                Versions: pkg,
            }

            c.JSON(200, response)
        } else {
            fmt.Println("the package was NOT found!")
            c.JSON(404, models.WingetApiError{
                ErrorCode: 404,
                ErrorMessage: "The specified package was not found.",
            })
        }
    })

    fmt.Println("Starting server on", *listenAddrPtr)
    if *tlsEnablePtr {
        if err := router.RunTLS(*listenAddrPtr, *tlsCertificatePtr, *tlsPrivateKeyPtr); err != nil {
            log.Fatal("error could not start webserver:", err)
        }
    } else {
        if err := router.Run(*listenAddrPtr); err != nil {
            log.Fatal("error could not start webserver:", err)
        }
    }
}
