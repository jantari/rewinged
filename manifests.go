package main

import (
  "fmt"
  "os"
  "errors"
  "strings"
  "io"
  "io/fs"
  "net/http"

  "gopkg.in/yaml.v3"

  "rewinged/logging"
  "rewinged/models"
)

func ingestManifestsWorker(autoInternalize bool, globalInstallerUrl string) error {
  for path := range jobs {
    files, err := os.ReadDir(path)
    if err != nil {
      logging.Logger.Error().Err(err).Msg("ingestManifestsWorker error")
      wg.Done()
      continue
    }

    // temporary map collecting all files belonging to a particular package
    var nonSingletonsMap = make(map[models.MultiFileManifest][]models.ManifestTypeAndPath)

    for _, file := range files {
      if !file.IsDir() {
        if caseInsensitiveHasSuffix(file.Name(), ".yml") || caseInsensitiveHasSuffix(file.Name(), ".yaml") {
          var basemanifest, err = parseFileAsBaseManifest(path + "/" + file.Name())
          if err != nil {
            logging.Logger.Error().Err(err).Str("file", path + "/" + file.Name()).Msgf("cannot unmarshal YAML file as BaseManifest")
            continue
          }

          // There could be other, non winget-manifest YAML files in the manifestPath as well. Skip them.
          // All valid manifest files must have all basemanifest fields set as they are required by the schema
          if basemanifest.PackageIdentifier != "" && basemanifest.PackageVersion != "" &&
            basemanifest.ManifestType != "" && basemanifest.ManifestVersion != "" {
            if basemanifest.ManifestType == "singleton" {
              var manifest = parseFileAsSingletonManifest(path + "/" + file.Name())
              logging.Logger.Debug().Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("found singleton manifest")

              // Singleton manifests can only contain version of a package each
              var version = manifest.GetVersions()[0]

              // Internalization logic
              if (autoInternalize) {
                var installers []models.API_InstallerInterface = version.GetInstallers()

                for _, v := range installers {
                  var destFile string = fmt.Sprintf("./installers/%s", strings.ToLower(v.GetInstallerSha()))
                  // Why os.OpenFile instead of os.Create:
                  // https://stackoverflow.com/a/22483001
                  out, err := os.OpenFile(destFile, os.O_RDWR | os.O_CREATE | os.O_EXCL, 0666)
                  defer out.Close()
                  if err != nil {
                    if errors.Is(err, fs.ErrExist) {
                      logging.Logger.Debug().Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("file already exists, not redownloading %s", destFile)
                    } else {
                      logging.Logger.Error().Err(err).Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("cannot create file %s", destFile)
                      continue
                    }
                  } else {
                    // No error, we could open the file for writing and it does not exist yet - so download it
                    logging.Logger.Debug().Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("downloading installer")

                    resp, err := http.Get(v.GetInstallerUrl())
                    if err != nil {
                      logging.Logger.Error().Err(err).Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("cannot download file %s", v.GetInstallerUrl())
                    }
                    defer resp.Body.Close()

                    n, err := io.Copy(out, resp.Body)
                    logging.Logger.Debug().Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("downloaded installer, %d bytes written", n)
                  }

                  // Rewrite the installers' InstallerUrl
                  logging.Logger.Debug().Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("internalizing InstallerUrl")
                  v.SetInstallerUrl(fmt.Sprintf("%s/%s", globalInstallerUrl, strings.ToLower(v.GetInstallerSha())))
                }

                // Recreate manifest object, but with overwritten values (InstallerUrl(s))
                manifest, err = newAPIManifest(
                  basemanifest.ManifestVersion,
                  basemanifest.PackageIdentifier,
                  basemanifest.PackageVersion,
                  version.GetDefaultLocale(),
                  version.GetLocales(),
                  installers,
                )

                if err != nil {
                  logging.Logger.Error().Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("error reconstructing package after overwrite")
                }

                version = manifest.GetVersions()[0]
              }
              // End internalization logic

              models.Manifests.Set(manifest.GetPackageIdentifier(), basemanifest.PackageVersion, version)
            } else {
              typeAndPath := models.ManifestTypeAndPath{
                ManifestType: basemanifest.ManifestType,
                FilePath: path + "/" + file.Name(),
              }
              nonSingletonsMap[basemanifest.ToMultiFileManifest()] = append(nonSingletonsMap[basemanifest.ToMultiFileManifest()], typeAndPath)
            }
          }
        }
      }
    }

    if len(nonSingletonsMap) > 0 {
      for key, value := range nonSingletonsMap {
        logging.Logger.Debug().Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("found multi-file manifest")
        var mergedManifest, err = parseMultiFileManifest(key, value...)
        if err != nil {
          logging.Logger.Error().Err(err).Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("could not parse all manifest files for this package")
        } else {
          for _, version := range mergedManifest.GetVersions() {
            // Internalization logic
            if (autoInternalize) {
              var installers []models.API_InstallerInterface = version.GetInstallers()

              for _, v := range installers {
                var destFile string = fmt.Sprintf("./installers/%s", strings.ToLower(v.GetInstallerSha()))
                // Why os.OpenFile instead of os.Create:
                // https://stackoverflow.com/a/22483001
                out, err := os.OpenFile(destFile, os.O_RDWR | os.O_CREATE | os.O_EXCL, 0666)
                defer out.Close()
                if err != nil {
                  if errors.Is(err, fs.ErrExist) {
                    logging.Logger.Debug().Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("file already exists, not redownloading %s", destFile)
                  } else {
                    logging.Logger.Error().Err(err).Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("cannot create file %s", destFile)
                    continue
                  }
                } else {
                  // No error, we could open the file for writing and it does not exist yet - so download it
                  logging.Logger.Debug().Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("downloading installer")

                  resp, err := http.Get(v.GetInstallerUrl())
                  if err != nil {
                    logging.Logger.Error().Err(err).Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("cannot download file %s", v.GetInstallerUrl())
                  }
                  defer resp.Body.Close()

                  n, err := io.Copy(out, resp.Body)
                  logging.Logger.Debug().Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("downloaded installer, %d bytes written", n)
                }

                // Rewrite the installers' InstallerUrl
                logging.Logger.Debug().Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("internalizing InstallerUrl")
                v.SetInstallerUrl(fmt.Sprintf("%s/%s", globalInstallerUrl, strings.ToLower(v.GetInstallerSha())))
              }

              // Recreate manifest object, but with overwritten values (InstallerUrl(s))
              overwrittenMergedManifest, err := newAPIManifest(
                key.ManifestVersion,
                key.PackageIdentifier,
                key.PackageVersion,
                version.GetDefaultLocale(),
                version.GetLocales(),
                installers,
              )

              if err != nil {
                logging.Logger.Error().Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("error reconstructing package after overwrite")
              }

              version = overwrittenMergedManifest.GetVersions()[0]
            }
            // End internalization logic

            // Replace the existing PkgId + PkgVersion entry with this one
            models.Manifests.Set(mergedManifest.GetPackageIdentifier(), version.GetPackageVersion(), version)
          }
        }
      }
    }

    wg.Done()
  }

  return nil
}

// Finds and parses all package manifest files in a directory
// recursively and returns them as a map of PackageIdentifier
// and PackageVersions
func getManifests (path string) {
  files, err := os.ReadDir(path)
  if err != nil {
    logging.Logger.Error().Err(err)
  }

  // wg.Add() before goroutine, see staticcheck check SA2000 and also
  // https://stackoverflow.com/questions/65213707/where-to-put-wg-add
  wg.Add(1)
  go func() {
    jobs <- path
  }()

  for _, file := range files {
    if file.IsDir() {
      logging.Logger.Trace().Msgf("searching directory %s", path + "/" + file.Name())

      getManifests(path + "/" + file.Name())
    }
  }
}

func unmarshalVersionManifest (manifestVersion string, yamlData []byte) (models.Manifest_VersionManifestInterface, error) {
    var version models.Manifest_VersionManifestInterface

    switch manifestVersion {
        case "1.1.0":
          version = &models.Manifest_VersionManifest_1_1_0{}
        case "1.2.0":
          version = &models.Manifest_VersionManifest_1_2_0{}
        case "1.4.0":
          version = &models.Manifest_VersionManifest_1_4_0{}
        case "1.5.0":
          version = &models.Manifest_VersionManifest_1_5_0{}
        default:
          return nil, errors.New("unsupported VersionManifest version " + manifestVersion)
    }

    err := yaml.Unmarshal(yamlData, version)
    if err != nil {
      return nil, err
    }

    return version, nil
}

func unmarshalInstallerManifest (manifestVersion string, yamlData []byte) (models.Manifest_InstallerManifestInterface, error) {
    var installer models.Manifest_InstallerManifestInterface

    switch manifestVersion {
        case "1.1.0":
            installer = &models.Manifest_InstallerManifest_1_1_0{}
        case "1.2.0":
            installer = &models.Manifest_InstallerManifest_1_2_0{}
        case "1.4.0":
            installer = &models.Manifest_InstallerManifest_1_4_0{}
        case "1.5.0":
            installer = &models.Manifest_InstallerManifest_1_5_0{}
        default:
            return nil, errors.New("unsupported InstallerManifest version " + manifestVersion)
    }

    err := yaml.Unmarshal(yamlData, installer)
    if err != nil {
      return nil, err
    }

    return installer, nil
}

func unmarshalLocaleManifest (manifestVersion string, yamlData []byte) (models.Manifest_LocaleManifestInterface, error) {
    var locale models.Manifest_LocaleManifestInterface

    switch manifestVersion {
        case "1.1.0":
            locale = &models.Manifest_LocaleManifest_1_1_0{}
        case "1.2.0":
            locale = &models.Manifest_LocaleManifest_1_2_0{}
        case "1.4.0":
            locale = &models.Manifest_LocaleManifest_1_4_0{}
        case "1.5.0":
            locale = &models.Manifest_LocaleManifest_1_5_0{}
        default:
            return nil, errors.New("unsupported LocaleManifest version " + manifestVersion)
    }

    err := yaml.Unmarshal(yamlData, locale)
    if err != nil {
        return nil, err
    }

    return locale, nil
}

func unmarshalDefaultLocaleManifest (manifestVersion string, yamlData []byte) (models.Manifest_DefaultLocaleManifestInterface, error) {
    var defaultlocale models.Manifest_DefaultLocaleManifestInterface

    switch manifestVersion {
        case "1.1.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_1_0{}
        case "1.2.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_2_0{}
        case "1.4.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_4_0{}
        case "1.5.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_5_0{}
        default:
            return nil, errors.New("unsupported DefaultLocaleManifest version " + manifestVersion)
    }

    err := yaml.Unmarshal(yamlData, defaultlocale)
    if err != nil {
        return nil, err
    }

    return defaultlocale, nil
}

func parseMultiFileManifest (multifilemanifest models.MultiFileManifest, files ...models.ManifestTypeAndPath) (models.API_ManifestInterface, error) {
  if len(files) <= 0 {
    return nil, errors.New("you must provide at least one filename for reading values")
  }

  versions   := []models.Manifest_VersionManifestInterface{}
  installers := []models.Manifest_InstallerManifestInterface{}
  locales    := []models.Manifest_LocaleManifestInterface{}
  var defaultlocale models.Manifest_DefaultLocaleManifestInterface

  for _, file := range files {
    yamlFile, err := os.ReadFile(file.FilePath)
    if err != nil {
      logging.Logger.Error().Str("file", file.FilePath).Err(err).Msg("cannot read file")
      continue
    }
    switch file.ManifestType {
      case "version":
        var version models.Manifest_VersionManifestInterface
        version, err = unmarshalVersionManifest(multifilemanifest.ManifestVersion, yamlFile)
        if err != nil {
          logging.Logger.Error().Str("file", file.FilePath).Err(err).Msg("cannot unmarshal manifest file")
          continue
        }
        versions = append(versions, version)
      case "installer":
        var installer models.Manifest_InstallerManifestInterface
        installer, err = unmarshalInstallerManifest(multifilemanifest.ManifestVersion, yamlFile)
        if err != nil {
          logging.Logger.Error().Str("file", file.FilePath).Err(err).Msg("cannot unmarshal manifest file")
          continue
        }
        installers = append(installers, installer)
      case "locale":
        var locale models.Manifest_LocaleManifestInterface
        locale, err = unmarshalLocaleManifest(multifilemanifest.ManifestVersion, yamlFile)
        if err != nil {
          logging.Logger.Error().Str("file", file.FilePath).Err(err).Msg("cannot unmarshal manifest file")
          continue
        }
        locales = append(locales, locale)
      case "defaultLocale":
        defaultlocale, err = unmarshalDefaultLocaleManifest(multifilemanifest.ManifestVersion, yamlFile)
        if err != nil {
          logging.Logger.Error().Str("file", file.FilePath).Err(err).Msg("cannot unmarshal manifest file")
          continue
        }
      default:
    }
  }

  // It's possible there were no installer or locale manifests or parsing them failed
  if len(installers) == 0 {
    return nil, errors.New("no (valid) installer manifest")
  }
  if len(versions) == 0 {
    return nil, errors.New("no (valid) locale manifest")
  }

  // This transforms the manifest data into the format the API will return.
  // This logic should probably be moved out of this function, so that it returns
  // the full unaltered data from the combined manifests - and restructuring to
  // API-format will happen somewhere else
  var apiLocales []models.API_LocaleInterface
  for _, locale := range locales {
    apiLocales = append(apiLocales, locale.ToApiLocale())
  }

  var apiInstallers []models.API_InstallerInterface

  // In case there are more than 1 InstallerManifests per package (not sure if officially allowed)
  // we get all the installers defined in all the InstallerManifest files.
  for _, v := range installers {
    apiInstallers = append(apiInstallers, v.ToApiInstallers()...)
  }

  manifest, err := newAPIManifest(
    multifilemanifest.ManifestVersion,
    multifilemanifest.PackageIdentifier,
    versions[0].GetPackageVersion(),
    defaultlocale.ToApiDefaultLocale(),
    apiLocales,
    apiInstallers,
  )

  return manifest, err
}

func newAPIManifest (
  ManifestVersion string,
  PackageIdentifier string,
  pv string,
  dl models.API_DefaultLocaleInterface,
  l []models.API_LocaleInterface,
  inst []models.API_InstallerInterface,
) (
  models.API_ManifestInterface,
  error,
) {
  var apiReturnManifest models.API_ManifestInterface
  var apiMvi models.API_ManifestVersionInterface

  if ManifestVersion == "1.1.0" {
    var apiLocales []models.API_Locale_1_1_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_1_0))
    }

    var apiInstallers []models.API_Installer_1_1_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, *intf.(*models.API_Installer_1_1_0))
    }

    apiMvi = models.API_ManifestVersion_1_1_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_1_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    apiReturnManifest = &models.API_Manifest_1_1_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ apiMvi },
    }
  } else if ManifestVersion == "1.2.0" || ManifestVersion == "1.4.0" {
    // There is no API schema 1.2.0, so both v1.2.0 and v1.4.0
    // packages are returned to clients as v1.4.0 API responses
    var apiLocales []models.API_Locale_1_4_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_4_0))
    }

    var apiInstallers []models.API_Installer_1_4_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, *intf.(*models.API_Installer_1_4_0))
    }

    apiMvi = models.API_ManifestVersion_1_4_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_4_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    apiReturnManifest = &models.API_Manifest_1_4_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ apiMvi },
    }
  } else if ManifestVersion == "1.5.0" {
    var apiLocales []models.API_Locale_1_5_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_5_0))
    }

    var apiInstallers []models.API_Installer_1_5_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, *intf.(*models.API_Installer_1_5_0))
    }

    apiMvi = models.API_ManifestVersion_1_5_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_5_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    apiReturnManifest = &models.API_Manifest_1_5_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ apiMvi },
    }
  } else {
    return nil, errors.New("Converting manifest v" + ManifestVersion + " data for API responses is not yet supported.")
  }

  return apiReturnManifest, nil
}

func parseFileAsBaseManifest (path string) (*models.BaseManifest, error) {
  manifest := &models.BaseManifest{}
  yamlFile, err := os.ReadFile(path)
  if err != nil {
    return manifest, err
  }

  err = yaml.Unmarshal(yamlFile, manifest)
  return manifest, err
}

func parseFileAsSingletonManifest (path string) models.API_ManifestInterface {
  yamlFile, err := os.ReadFile(path)
  if err != nil {
    logging.Logger.Error().Err(err).Msg("error opening yaml file")
  }

  singleton := &models.Manifest_SingletonManifest_1_1_0{}
  err = yaml.Unmarshal(yamlFile, singleton)
  if err != nil {
    logging.Logger.Error().Err(err).Msg("error unmarshalling singleton")
  }

  manifest := singletonToStandardManifest(singleton)

  return manifest
}

func singletonToStandardManifest (singleton *models.Manifest_SingletonManifest_1_1_0) *models.API_Manifest_1_1_0 {
  manifest := &models.API_Manifest_1_1_0 {
    PackageIdentifier: singleton.PackageIdentifier,
    Versions: []models.API_ManifestVersionInterface{ models.API_ManifestVersion_1_1_0 {
      PackageVersion: singleton.PackageVersion,
      DefaultLocale: models.API_DefaultLocale_1_1_0 {
        PackageLocale: singleton.PackageLocale,
        PackageName: singleton.PackageName,
        Publisher: singleton.Publisher,
        ShortDescription: singleton.ShortDescription,
        License: singleton.License,
      },
      Channel: "",
      Locales: []models.API_Locale_1_1_0{},
      Installers: []models.API_Installer_1_1_0{singleton.Installers[0].ToApiInstaller()},
    },
  },
  }

  return manifest
}

func caseInsensitiveHasSuffix(s, substr string) bool {
  s, substr = strings.ToUpper(s), strings.ToUpper(substr)
  return strings.HasSuffix(s, substr)
}

