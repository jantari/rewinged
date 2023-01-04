package controllers

import (
    "fmt"
    "log"

    "github.com/gin-gonic/gin"

    "rewinged/models"
)

func GetPackages(c *gin.Context) {
    response := &models.PackageMultipleResponse{
        Data: models.Manifests.GetAllPackageIdentifiers(),
    }

    fmt.Println(response)
    c.JSON(200, response)
}

func GetPackage(c *gin.Context) {
  fmt.Println("/packageManifests: Someone tried to GET package '", c.Param("package_identifier"), "'")
  fmt.Println("with query params:", c.Request.URL.Query())

  response := models.ManifestSingleResponse {
    RequiredQueryParameters: []models.QueryParameter{},
    UnsupportedQueryParameters: []models.QueryParameter{},
    Data: nil,
  }

  var pkg []models.API_ManifestVersionInterface = models.Manifests.GetAllVersions(c.Param("package_identifier"))
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
}

func SearchForPackage(c *gin.Context) {
  var post models.ManifestSearch
  if err := c.BindJSON(&post); err == nil {
    fmt.Printf("%+v\n", post)
    response := &models.ManifestSearchResult[models.ManifestSearchVersion_1_1_0]{
      RequiredPackageMatchFields: []models.PackageMatchField{},
      Data: []models.ManifestSearchResponse[models.ManifestSearchVersion_1_1_0] {},
    }

    // results is a map where the PackageIdentifier is the key
    // and the values are arrays of manifests with that PackageIdentifier.
    // This means the values will be different versions of the package.
    var results map[string][]models.API_ManifestVersionInterface

    if post.Query.KeyWord != "" {
      fmt.Println("someone searched the repo for:", post.Query.KeyWord)
      results = models.Manifests.GetByKeyword(post.Query.KeyWord)
    } else if (post.Inclusions != nil && len(post.Inclusions) > 0) || (post.Filters != nil && len(post.Filters) > 0) {
      fmt.Println("advanced search with inclusions[] and/or filters[]")
      results = models.Manifests.GetByMatchFilter(post.Inclusions, post.Filters)
    }

    fmt.Println("... with", len(results), "results:")

    if len(results) > 0 {
      for packageId, packageVersions := range results {
        fmt.Println("  package", packageId, "with", len(packageVersions), "versions.")
        var versions []models.ManifestSearchVersion_1_1_0

        for _, version := range packageVersions {
          versions = append(versions, models.ManifestSearchVersion_1_1_0{
            PackageVersion: version.GetPackageVersion(),
            Channel: "",
            PackageFamilyNames: []string{},
            ProductCodes: version.GetInstallerProductCodes(),
          })
        }

        response.Data = append(response.Data, models.ManifestSearchResponse[models.ManifestSearchVersion_1_1_0]{
          PackageIdentifier: packageId,
          PackageName: packageVersions[0].GetDefaultLocalePackageName(),
          Publisher: packageVersions[0].GetDefaultLocalePublisher(),
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
}

