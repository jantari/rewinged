package controllers

import (
    "fmt"
    "strings"
    "net/http"
    "encoding/json"

    "rewinged/logging"
    "rewinged/settings"
    "rewinged/models"
)

func GetPackages(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    groups, ok := ctx.Value("groups").([]string)
    if !ok {
        logging.Logger.Error().Msg("no or invalid groups in request context")
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        return
    }

    allowedPackages := []models.API_Package{}
    globalDecision := EvaluateGlobalRule(groups)

    if globalDecision != Denied {
        for _, pkg := range models.Manifests.GetAllPackageIdentifiers() {
            if FilterAuthorizedPackage(globalDecision, pkg.PackageIdentifier, groups) {
                allowedPackages = append(allowedPackages, pkg)
            }
        }
    }

    response := &models.API_PackageMultipleResponse{
        Data: allowedPackages,
    }

    logging.Logger.Debug().Msgf("%v", response)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

type GetPackageHandler struct {
    TlsEnabled bool
    InternalizationEnabled bool
}

func (this *GetPackageHandler) GetPackage(w http.ResponseWriter, r *http.Request) {
  logging.Logger.Debug().Msgf("/packageManifests: Someone tried to GET package '%v' with query params: %v", r.PathValue("package_identifier"), r.URL.Query())
  logging.Logger.Debug().Msgf("client requested API version %v", r.Header.Get("Version"))

  response := models.API_ManifestSingleResponse_1_1_0 {
    RequiredQueryParameters: []string{},
    UnsupportedQueryParameters: []string{},
    Data: nil,
  }

  ctx := r.Context()
  groups, ok := ctx.Value("groups").([]string)
  if !ok {
      logging.Logger.Error().Msg("no or invalid groups in request context")
      http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
      return
  }

  var pkg []models.API_ManifestVersionInterface
  globalDecision := EvaluateGlobalRule(groups)

  if globalDecision != Denied {
      if FilterAuthorizedPackage(globalDecision, r.PathValue("package_identifier"), groups) {
          pkg = models.Manifests.GetAllVersions(r.PathValue("package_identifier"))
      }
  }

  if this.InternalizationEnabled {
      var rewrittenOrigin string
      // TODO: Only accept these headers if c.RemoteIP() is a trusted proxy configured by the user
      xfProto := r.Header.Values("X-Forwarded-Proto")
      xfHost  := r.Header.Values("X-Forwarded-Host")

      if len(xfProto) == 1 && xfProto[0] != "" && len(xfHost) == 1 && xfHost[0] != "" {
          rewrittenOrigin = fmt.Sprintf(
              "%s://%s", xfProto[0], xfHost[0],
          )
      } else {
          var proto string
          if this.TlsEnabled {
              proto = "https"
          } else {
              proto = "http"
          }
          rewrittenOrigin = fmt.Sprintf(
              "%s://%s", proto, r.Host,
          )
      }

      // We cannot use a range loop over the installers here because range loops
      // always put the current element in the loop into the same one memory address.
      // So if we collect/append the pointers to the installers in the loop, they
      // will all just point to the same memory location chosen by range.
      for i := 0; i < len(pkg); i++ {
          installers := pkg[i].GetInstallers()

          for j := 0; j < len(installers); j++ {
              // Only rewrite this installers InstallerUrl if it was marked for it on ingest
              if models.InternalizedInstallers[installers[j].GetInstallerSha()] {
                  installers[j].SetInstallerUrl(
                      fmt.Sprintf(
                          "%s/installers/%s",
                          rewrittenOrigin,
                          strings.ToLower(installers[j].GetInstallerSha()),
                      ),
                  )

                  // If source authentication is configured, any internalized installers
                  // are also only downloadable with valid authentication. The winget client
                  // does not pass authentication along with the Installer download request
                  // unless the Installer metadata explicitly asks for it, which is only
                  // supported starting with API schema 1.10.0.
                  // This means two things IF authentication is enabled in rewinged:
                  //   1. Source (REST API) authentication was added in schema 1.7.0, but
                  //      InstallerAuthentication is only allowed with schema 1.10.0+. This
                  //      means, for AutoInternalized packages, for which rewinged hosts the
                  //      installer no schemas older than 1.10.0 can be used.
                  //   2. We have to explicitly edit the metadata of internalized installers
                  //      to say they will require authentication to download to make winget CLI
                  //      pass credentials with the download request
                  if settings.SourceAuthenticationType == "microsoftEntraId" {
                      // Currently hardcoded to insert a v1.10.0 Authentication struct, but it doesn't matter as the
                      // struct itself hasn't changed since being introduced in API 1.7.0 and it's only allowed in
                      // installers as of API 1.10.0. Technically, when adding support for the next API version after
                      // 1.10.0 this should type-switch on installers[j] and add the version-matched struct.
                      v, ok := installers[j].(models.API_InstallerWithAuthInterface[models.API_Authentication_1_10_0])
                      if ok {
                        v.SetInstallerAuthentication(&models.API_Authentication_1_10_0{
                          AuthenticationType: settings.SourceAuthenticationType,
                          MicrosoftEntraIdAuthenticationInfo: struct {
                            Resource string `yaml:"Resource"`
                            Scope string `yaml:"Scope" json:",omitempty"`
                          }{
                            Resource: settings.SourceAuthenticationEntraIDResource, // Entra Application ID
                            Scope: "user_impersonation",
                          },
                        })
                      } else {
                        logging.Logger.Warn().Msgf("manifest version of this package %T is too old and does not support InstallerAuthentication, client download will likely fail", installers[j])
                      }
                  }
              }
          }
      }
  }

  if len(pkg) > 0 {
    logging.Logger.Debug().Msgf("the package was found")

    response.Data = &models.API_Manifest_1_1_0{
      PackageIdentifier: r.PathValue("package_identifier"),
      Versions: pkg,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
  } else {
    logging.Logger.Debug().Msgf("the package was not found")
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotFound)
    json.NewEncoder(w).Encode(models.API_WingetApiError{
      ErrorCode: 404,
      ErrorMessage: "The specified package was not found.",
    })
  }
}

func SearchForPackage(w http.ResponseWriter, r *http.Request) {
  var post models.API_ManifestSearchRequest_1_1_0

  d := json.NewDecoder(r.Body)
  d.DisallowUnknownFields() // catch unwanted fields

  err := d.Decode(&post)
  if err != nil {
    // bad JSON or unrecognized json field
    logging.Logger.Error().Err(err).Msg("error deserializing json post body")
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  logging.Logger.Debug().Msgf("%+v", post)
  response := &models.API_ManifestSearchResult[models.API_ManifestSearchVersion_1_1_0]{
    RequiredPackageMatchFields: []string{},
    Data: []models.API_ManifestSearchResponse[models.API_ManifestSearchVersion_1_1_0] {},
  }

  ctx := r.Context()
  groups, ok := ctx.Value("groups").([]string)
  if !ok {
    logging.Logger.Error().Msg("no or invalid groups in request context")
    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
    return
  }

  globalDecision := EvaluateGlobalRule(groups)

  // results is a map where the PackageIdentifier is the key
  // and the values are arrays of manifests with that PackageIdentifier.
  // This means the values will be different versions of the package.
  var results map[string][]models.API_ManifestVersionInterface

  if globalDecision != Denied {
    if post.Query.KeyWord != "" {
      logging.Logger.Debug().Msgf("someone searched the repo for: %v", post.Query.KeyWord)
      results = models.Manifests.GetByKeyword(post.Query.KeyWord)
    } else if (post.Inclusions != nil && len(post.Inclusions) > 0) || (post.Filters != nil && len(post.Filters) > 0) {
      logging.Logger.Debug().Msg("advanced search with inclusions[] and/or filters[]")
      results = models.Manifests.GetByMatchFilter(post.Inclusions, post.Filters)
    }

    logging.Logger.Debug().Msgf("with %v results", len(results))

    for packageId := range results {
      if ! FilterAuthorizedPackage(globalDecision, packageId, groups) {
        // It is safe to delete map elements during iteration
        delete(results, packageId)
      }
    }
  } else {
    logging.Logger.Debug().Msg("user is globally denied, not even looking for packages")
  }

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
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)

  } else {
    // winget REST-API specification calls for a 204 return code if no results were found
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNoContent)
    json.NewEncoder(w).Encode(response)
  }
}

