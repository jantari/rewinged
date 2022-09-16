//go:generate goversioninfo

package main

import (
    "fmt"
    "log"
    "os"
    "flag"

    "github.com/gin-gonic/gin"
)

// These variables are overwritten at compile/link time using -ldflags
var version = "development-build"
var commit = "unknown"
var compileTime = "unknown"
var releaseMode = "false"

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

    var manifests = make(map[string][]Versions)
    fmt.Println("Searching for manifests...")
    manifests = getManifests(*packagePathPtr)
    fmt.Println("Found", len(manifests), "package manifests.")

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

        for k := range manifests {
            response.Data = append(response.Data, Package{
                PackageIdentifier: k,
            })
        }

        fmt.Println(response)
        c.JSON(200, response)
    })
    router.POST("/packages", func(c *gin.Context) {
        //
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
                results = getPackagesByKeyword(manifests, post.Query.KeyWord)
            } else if (post.Inclusions != nil && len(post.Inclusions) > 0) || (post.Filters != nil && len(post.Filters) > 0) {
                fmt.Println("advanced search with inclusions[] and/or filters[]")
                results = getPackagesByMatchFilter(manifests, post.Inclusions, post.Filters)
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
            }

            fmt.Printf("%+v\n", response)

            c.JSON(200, response)
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

        var pkg = manifests[c.Param("package_identifier")]
        if len(pkg) > 0 {
            fmt.Println("the package was found!")

            response.Data = &Manifest{
                PackageIdentifier: c.Param("package_identifier"),
                Versions: pkg,
            }
        } else {
            fmt.Println("the package was NOT found!")
        }

        c.JSON(200, response)
    })

    if *tlsEnablePtr {
        router.RunTLS(*listenAddrPtr, *tlsCertificatePtr, *tlsPrivateKeyPtr)
    } else {
        router.Run(*listenAddrPtr)
    }
}
