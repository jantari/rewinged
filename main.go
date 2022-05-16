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

type WingetApiError struct {
    ErrorCode    int
    ErrorMessage string
}

type Package struct {
    PackageIdentifier string
}

type PackageMultipleResponse struct {
    Data []Package
    ContinuationToken string
}

type Information struct {
    Data struct {
        SourceIdentifier        string
        ServerSupportedVersions []string
    }
}

type MatchType string

const (
    Exact MatchType = "Exact"
    CaseInsensitive = "CaseInsensitive"
    StartsWith      = "StartsWith"
    Substring       = "Substring"
    Wildcard        = "Wildcard"
    Fuzzy           = "Fuzzy"
    FuzzySubstring  = "FuzzySubstring"
)

type PackageMatchField string

const (
    PackageIdentifier PackageMatchField = "PackageIdentifier"
    PackageName = "PackageName"
    Moniker = "Moniker"
    Command = "Command"
    Tag = "Tag"
    PackageFamilyName = "PackageFriendlyName"
    ProductCode = "ProductCode"
    NormalizedPackageNameAndPublisher = "NormalizedPackageNameAndPublisher"
    Market = "Market"
)

type SearchRequestMatch struct {
    KeyWord string
    MatchType MatchType
}

type SearchRequestPackageMatchFilter struct {
    PackageMatchField PackageMatchField
    RequestMatch SearchRequestMatch
}

type ManifestSearch struct {
    MaximumResults int
    FetchAllManifests bool
    Query SearchRequestMatch
    Inclusions SearchRequestPackageMatchFilter
    Filters []SearchRequestPackageMatchFilter
}

type ManifestSearchVersion struct {
    PackageVersion string
//    Channel string
//    PackageFamilyNames []string // TODO: NOT THE ACTUAL DATATYPE!
//    ProductCodes []string // TODO: NOT THE ACTUAL DATATYPE!
}

type ManifestSearchResponse struct {
//    Package Package
    PackageIdentifier string
    PackageName string
    Publisher string
    Versions []ManifestSearchVersion
}

type ManifestSearchResult struct {
    Data []ManifestSearchResponse
    RequiredPackageMatchFields []PackageMatchField
    UnsupportedPackageMatchFields []PackageMatchField
}

//

func main() {
    fmt.Println("Hello world")
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
                Data: []ManifestSearchResponse {
                    {
                        PackageIdentifier: "kek",
                        PackageName: "kek",
                        Publisher: "test",
                        Versions: []ManifestSearchVersion {
                            {
                                PackageVersion: "1.0.0",
                            },
                        },
                    },
                },
            }

            fmt.Printf("%+v\n", response)

            c.JSON(200, response)
        } else {
            fmt.Println("error deserializing json post body")
            fmt.Println(err)
        }
    })
    router.RunTLS(":8080", "./cert.pem", "./server.key") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
