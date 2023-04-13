package controllers

import (
    "github.com/gin-gonic/gin"

    "rewinged/logging"
    "rewinged/models"
)

func GetPackages(c *gin.Context) {
    response := &models.API_PackageMultipleResponse{
        Data: models.Manifests.GetAllPackageIdentifiers(),
    }

    logging.Logger.Debug().Msgf("%v", response)
    c.JSON(200, response)
}

func GetPackage(c *gin.Context) {
  logging.Logger.Debug().Msgf("/packageManifests: Someone tried to GET package '%v' with query params: %v", c.Param("package_identifier"), c.Request.URL.Query())

  response := models.API_ManifestSingleResponse_1_1_0 {
    RequiredQueryParameters: []string{},
    UnsupportedQueryParameters: []string{},
    Data: nil,
  }

  var pkg []models.API_ManifestVersionInterface = models.Manifests.GetAllVersions(c.Param("package_identifier"))
  if len(pkg) > 0 {
    logging.Logger.Debug().Msgf("the package was found")

    response.Data = &models.API_Manifest_1_1_0{
      PackageIdentifier: c.Param("package_identifier"),
      Versions: pkg,
    }

    c.JSON(200, response)
  } else {
    logging.Logger.Debug().Msgf("the package was not found")
    c.JSON(404, models.API_WingetApiError{
      ErrorCode: 404,
      ErrorMessage: "The specified package was not found.",
    })
  }
}

func SearchForPackage(c *gin.Context) {
  var post models.API_ManifestSearchRequest_1_1_0
  if err := c.BindJSON(&post); err == nil {
    logging.Logger.Debug().Msgf("%+v", post)
    response := &models.API_ManifestSearchResult[models.API_ManifestSearchVersion_1_1_0]{
      RequiredPackageMatchFields: []string{},
      Data: []models.API_ManifestSearchResponse[models.API_ManifestSearchVersion_1_1_0] {},
    }

    // results is a map where the PackageIdentifier is the key
    // and the values are arrays of manifests with that PackageIdentifier.
    // This means the values will be different versions of the package.
    var results map[string][]models.API_ManifestVersionInterface

    if post.Query.KeyWord != "" {
      logging.Logger.Debug().Msgf("someone searched the repo for: %v", post.Query.KeyWord)
      results = models.Manifests.GetByKeyword(post.Query.KeyWord)
    } else if (post.Inclusions != nil && len(post.Inclusions) > 0) || (post.Filters != nil && len(post.Filters) > 0) {
      logging.Logger.Debug().Msg("advanced search with inclusions[] and/or filters[]")
      results = models.Manifests.GetByMatchFilter(post.Inclusions, post.Filters)
    }

    logging.Logger.Debug().Msgf("with %v results", len(results))

    if len(results) > 0 {
      for packageId, packageVersions := range results {
        logging.Logger.Debug().Msgf("package %v with %v versions", packageId, len(packageVersions))
        var versions []models.API_ManifestSearchVersion_1_1_0

        for _, version := range packageVersions {
          versions = append(versions, models.API_ManifestSearchVersion_1_1_0{
            PackageVersion: version.GetPackageVersion(),
            Channel: "",
            PackageFamilyNames: []string{},
            ProductCodes: version.GetInstallerProductCodes(),
          })
        }

        response.Data = append(response.Data, models.API_ManifestSearchResponse[models.API_ManifestSearchVersion_1_1_0]{
          PackageIdentifier: packageId,
          PackageName: packageVersions[0].GetDefaultLocalePackageName(),
          Publisher: packageVersions[0].GetDefaultLocalePublisher(),
          Versions: versions,
        })
      }
      logging.Logger.Debug().Msgf("%+v", response)
      c.JSON(200, response)
    } else {
      // winget REST-API specification calls for a 204 return code if no results were found
      c.JSON(204, response)
    }
  } else {
    logging.Logger.Error().Err(err).Msg("error deserializing json post body")
  }
}

