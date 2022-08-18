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

type Manifest struct {
    PackageIdentifier string
    Versions []Versions
}

type Versions struct {
    PackageVersion string
    DefaultLocale Locale
    Channel string
    Locales []Locale
    Installers []Installer
}

type Installer struct {
    Architecture Architecture
    InstallerType InstallerType
    InstallerUrl string
    InstallerSha256 string
}

type Locale struct {
    PackageLocale string
//    Moniker // Is this needed for DefaultLocale?
    Publisher string
//    PublisherUrl
//    PublisherSupportUrl
//    PrivacyUrl
//    Author
    PackageName string
//    PackageUrl
    License string
//    LicenseUrl
//    Copyright
//    CopyrightUrl
    ShortDescription string
//    Description
//    Tags
//    ReleaseNotes
//    ReleaseNotesUrl
//    Agreements
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

type Architecture string

const (
    neutral Architecture = "neutral"
    x86 = "x86"
    x64 = "x64"
    arm = "arm"
    arm64 = "arm64"
)

type InstallerType string

const (
    msix InstallerType = "msix"
    msi = "msi"
    appx = "appx"
    exe = "exe"
    zip = "zip"
    inno = "inno"
    nullsoft = "nullsoft"
    wix = "wix"
    burn = "burn"
    pwa = "pwa"
    msstore = "msstore"
)

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

type QueryParameter string

const (
    Version QueryParameter = "Version"
    Channel = "Channel"
//    Market = "Market" // Already declared in PackageMatchField enum
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
    Inclusions []SearchRequestPackageMatchFilter
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

type ManifestSingleResponse struct {
    Data Manifest
    RequiredQueryParameters []QueryParameter
    UnsupportedQueryParameters []QueryParameter
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
            Data: Manifest {
                PackageIdentifier: "bottom",
                Versions: []Versions {
                    {
                        PackageVersion: "0.6.8",
                        DefaultLocale: Locale {
                            PackageLocale: "en-US",
                            PackageName: "bottom",
                            Publisher: "Clement Tsang",
                            ShortDescription: "test package",
                            License: "GPLv2",
                        },
                        Channel: "",
                        Locales: []Locale{},
                        Installers: []Installer {
                            {
                                Architecture: "x64",
                                InstallerType: "msi",
                                InstallerUrl: "https://github.com/ClementTsang/bottom/releases/download/0.6.8/bottom_x86_64_installer.msi",
                                InstallerSha256: "43A860A1ECAC287CAF9745D774B0B2CE9C0A2A79D4048893E7649B0D8048EE58",
                            },
                        },
                    },
                },
            },
            RequiredQueryParameters: []QueryParameter{},
            UnsupportedQueryParameters: []QueryParameter{},
        }
        c.JSON(200, response)
    })
    router.RunTLS(":8080", "./cert.pem", "./server.key") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
