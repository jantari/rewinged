package main

import (
    "fmt"
//    "net"
//    "io/ioutil"

    "github.com/gin-gonic/gin"
//    "github.com/gin-gonic/gin/binding"
)

func Must(i interface{}, err error) interface{} {
    if err != nil {
        panic(err)
    }
    return i
}

//

func main() {
    fmt.Println("Hello world")

    var manifests = []SingletonManifest{}
    manifests = GetManifests("./packages")
    fmt.Println("Found", len(manifests), "package manifests.")

    router := gin.Default()
    router.SetTrustedProxies(nil)
    router.GET("/information", func(c *gin.Context) {
        response := new(Information)
        response.Data.SourceIdentifier = "testing"
        response.Data.ServerSupportedVersions = []string{"1.0.0", "1.1.0"}
        c.JSON(200, response)
    })
    router.GET("/packages", func(c *gin.Context) {
        response := new(PackageMultipleResponse)

        var test = Package{
            PackageIdentifier: "git.git",
        }

        response.Data = []Package{test}

        fmt.Println(response)
        c.JSON(200, response)
    })
    router.POST("/packages", func(c *gin.Context) {
        //
    })
    router.POST("/manifestSearch", func(c *gin.Context) {
//        body, _ := ioutil.ReadAll(c.Request.Body)
//        println(string(body))

        var post ManifestSearch
//        if err := c.ShouldBindBodyWith(&post, binding.JSON); err == nil {
        if err := c.BindJSON(&post); err == nil {
            fmt.Printf("%+v\n", post)
            response := &ManifestSearchResult{
                RequiredPackageMatchFields: []PackageMatchField{},
                Data: []ManifestSearchResponse {},
            }

            var results = []SingletonManifest{}

            if post.Query.KeyWord != "" {
                fmt.Println("someone searched the repo for:", post.Query.KeyWord)
                results = GetPackagesByKeyword(manifests, post.Query.KeyWord)
            } else if post.Inclusions != nil && len(post.Inclusions) > 0  {
                fmt.Println("advanced search with inclusions[]")
                var searchresults = GetPackagesByMatchFilter(manifests, post.Inclusions)
                results = append(results, searchresults...)
            }

            fmt.Println("... with", len(results), "results.")

            if len(results) > 0 {
                for _, result := range results {
                    response.Data = append(response.Data, ManifestSearchResponse{
                        PackageIdentifier: result.PackageIdentifier,
                        PackageName: result.PackageName,
                        Publisher: result.Publisher,
                        Versions: []ManifestSearchVersion {
                            {
                                PackageVersion: result.PackageVersion,
                            },
                        },
                    })
                }
            }

            /*
            response = &ManifestSearchResult{
                RequiredPackageMatchFields: []PackageMatchField{},
                Data: []ManifestSearchResponse {
                    {
                        PackageIdentifier: "bottom",
                        PackageName: "bottom",
                        Publisher: "test",
                        Versions: []ManifestSearchVersion {
                            {
                                PackageVersion: "0.6.8",
                            },
                        },
                    },
                },
            }
            */

            fmt.Printf("%+v\n", response)

            c.JSON(200, response)
        } else {
            fmt.Println("error deserializing json post body")
            fmt.Println(err)
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

        var pkg = GetPackageByIdentifier(manifests, c.Param("package_identifier"))
        if pkg != nil {
            fmt.Println("the package was found!", pkg.PackageIdentifier)

            response.Data = &Manifest {
                PackageIdentifier: pkg.PackageIdentifier,
                Versions: []Versions {
                    {
                        PackageVersion: pkg.PackageVersion,
                        DefaultLocale: Locale {
                            PackageLocale: pkg.PackageLocale,
                            PackageName: pkg.PackageName,
                            Publisher: pkg.Publisher,
                            ShortDescription: pkg.ShortDescription,
                            License: pkg.License,
                        },
                        Channel: "",
                        Locales: []Locale{},
                        Installers: pkg.Installers,
                    },
                },
            }
        } else {
            fmt.Println("the package was NOT found!")
        }

        c.JSON(200, response)
    })
    router.RunTLS(":8443", "./cert.pem", "./server.key") // listen and serve on 0.0.0.0:8443 (for windows "localhost:8443")
}
