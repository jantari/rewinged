//go:generate goversioninfo

package main

import (
    "fmt"

    "github.com/gin-gonic/gin"
//    "github.com/gin-gonic/gin/binding"
)

func Must(i interface{}, err error) interface{} {
    if err != nil {
        panic(err)
    }
    return i
}

func main() {
    fmt.Println("Hello world")

    var manifests = make(map[string][]Versions)
    manifests = GetManifests("./packages")
    //manifests = GetManifests("./winget-pkgs/manifests")
    fmt.Println("Found", len(manifests), "package manifests.")

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
                results = GetPackagesByKeyword(manifests, post.Query.KeyWord)
            } else if (post.Inclusions != nil && len(post.Inclusions) > 0) || (post.Filters != nil && len(post.Filters) > 0) {
                fmt.Println("advanced search with inclusions[] and/or filters[]")
                results = GetPackagesByMatchFilter(manifests, post.Inclusions, post.Filters)
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

        //var pkg = GetPackageByIdentifier(manifests, c.Param("package_identifier"))
        var pkg = manifests[c.Param("package_identifier")]
        if pkg != nil && len(pkg) > 0 {
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
    router.RunTLS(":8443", "./cert.pem", "./server.key") // listen and serve on 0.0.0.0:8443 (for windows "localhost:8443")
}
