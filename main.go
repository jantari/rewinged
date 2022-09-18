//go:generate goversioninfo

package main

import (
    "fmt"
    "log"
    "os"
    "flag"
    "sync"
    "path/filepath"

    "github.com/gin-gonic/gin"
    "github.com/rjeczalik/notify" // for live-reload of manifests
)

// These variables are overwritten at compile/link time using -ldflags
var version = "development-build"
var commit = "unknown"
var compileTime = "unknown"
var releaseMode = "false"

var wg sync.WaitGroup
var jobs chan string = make(chan string)

func getMapValues[M ~map[K]V, K comparable, V any](m M) []V {
    r := make([]V, 0, len(m))
    for _, v := range m {
        r = append(r, v)
    }
    return r
}

// Internal in-memory data store of all manifest data
type ManifestsStore struct {
    sync.RWMutex
    internal map[string]map[string]Versions
}

func (ms *ManifestsStore) Set(packageidentifier string, packageversion string, value Versions) {
    ms.Lock()
    vmap, ok := ms.internal[packageidentifier]
    if !ok {
        vmap = make(map[string]Versions)
        ms.internal[packageidentifier] = vmap
    }
    vmap[packageversion] = value
    ms.Unlock()
}

func (ms *ManifestsStore) GetAllVersions(packageidentifier string) (value []Versions) {
    ms.RLock()
    result := getMapValues(ms.internal[packageidentifier])
    ms.RUnlock()
    return result
}

func (ms *ManifestsStore) Get(packageidentifier string, packageversion string) (value Versions) {
    ms.RLock()
    result := ms.internal[packageidentifier][packageversion]
    ms.RUnlock()
    return result
}

func (ms *ManifestsStore) GetAll() (value map[string][]Versions) {
    ms.RLock()
    var m = make(map[string][]Versions)
    for k := range ms.internal {
        m[k] = getMapValues(ms.internal[k])
    }
    ms.RUnlock()
    return m
}

var manifests = ManifestsStore{
    internal: make(map[string]map[string]Versions),
}

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
    manifests.RLock()
    fmt.Println("Found", len(manifests.internal), "package manifests.")
    manifests.RUnlock()

    fmt.Println("Watching manifestPath for changes ...")
    // Make the channel buffered to ensure no event is dropped. Notify will drop
    // an event if the receiver is not able to keep up the sending pace.
    fileEventsChannel := make(chan notify.EventInfo, 1)

    // Recursively listen for Create, Write and Remove events in the manifestPath
    if err := notify.Watch(*packagePathPtr + "/...", fileEventsChannel, notify.Create, notify.Remove, notify.Write); err != nil {
        log.Fatal(err)
    }
    defer notify.Stop(fileEventsChannel)

    // If an event is received, push its directory-path to the jobs channel
    go func() {
      for ei := range fileEventsChannel {
        //ei := <-c
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
    router.GET("/information", func(c *gin.Context) {
        response := new(Information)
        response.Data.SourceIdentifier = "rewinged"
        response.Data.ServerSupportedVersions = []string{"1.1.0"}
        c.JSON(200, response)
    })
    router.GET("/packages", func(c *gin.Context) {
        response := new(PackageMultipleResponse)

        manifests.RLock()
        for k := range manifests.internal {
            response.Data = append(response.Data, Package{
                PackageIdentifier: k,
            })
        }
        manifests.RUnlock()

        fmt.Println(response)
        c.JSON(200, response)
    })
    router.POST("/manifestSearch", func(c *gin.Context) {
        var post ManifestSearch
        if err := c.BindJSON(&post); err == nil {
            fmt.Printf("%+v\n", post)
            response := &ManifestSearchResult{
                RequiredPackageMatchFields: []PackageMatchField{},
                Data: []ManifestSearchResponse {},
            }

            // results is a map where the PackageIdentifier is the key
            // and the values are arrays of manifests with that PackageIdentifier.
            // This means the values will be different versions of the package.
            var results map[string][]Versions

            if post.Query.KeyWord != "" {
                fmt.Println("someone searched the repo for:", post.Query.KeyWord)
                results = getPackagesByKeyword(manifests.GetAll(), post.Query.KeyWord)
            } else if (post.Inclusions != nil && len(post.Inclusions) > 0) || (post.Filters != nil && len(post.Filters) > 0) {
                fmt.Println("advanced search with inclusions[] and/or filters[]")
                results = getPackagesByMatchFilter(manifests.GetAll(), post.Inclusions, post.Filters)
            }

            fmt.Println("... with", len(results), "results:")

            if len(results) > 0 {
                for packageId, packageVersions := range results {
                    fmt.Println("  package", packageId, "with", len(packageVersions), "versions.")
                    var versions []ManifestSearchVersion

                    for _, version := range packageVersions {
                        versions = append(versions, ManifestSearchVersion{
                            PackageVersion: version.PackageVersion,
                            Channel: "",
                            PackageFamilyNames: []string{},
                            ProductCodes: []string{ version.Installers[0].ProductCode },
                        })
                    }

                    response.Data = append(response.Data, ManifestSearchResponse{
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

        response := ManifestSingleResponse {
            RequiredQueryParameters: []QueryParameter{},
            UnsupportedQueryParameters: []QueryParameter{},
            Data: nil,
        }

        var pkg []Versions = manifests.GetAllVersions(c.Param("package_identifier"))
        if len(pkg) > 0 {
            fmt.Println("the package was found!")

            response.Data = &Manifest{
                PackageIdentifier: c.Param("package_identifier"),
                Versions: pkg,
            }

            c.JSON(200, response)
        } else {
            fmt.Println("the package was NOT found!")
            c.JSON(404, WingetApiError{
                ErrorCode: 404,
                ErrorMessage: "The specified package was not found.",
            })
        }
    })

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
