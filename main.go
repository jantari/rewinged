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
    Architecture Architecture `yaml:"Architecture"`
    InstallerType InstallerType `yaml:"InstallerType"`
    InstallerUrl string `yaml:"InstallerUrl"`
    InstallerSha256 string `yaml:"InstallerSha256"`
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
    Data *Manifest
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

    var package_manifests = FindManifestFiles()
    fmt.Println("Found", len(package_manifests), "package manifests.")

    var manifests = []SingletonManifest{}
    for _, file := range package_manifests {
        var manifest = ParseManifestFile("./packages/" + file.Name())
        manifests = append(manifests, *manifest)
    }

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
    router.RunTLS(":8080", "./cert.pem", "./server.key") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
