package main

import (
  "os"
  "errors"
  "strings"

  "gopkg.in/yaml.v3"

  "rewinged/logging"
  "rewinged/models"
)

func ingestManifestsWorker() error {
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
            logging.Logger.Error().Msgf("error unmarshaling YAML file '%v' as BaseManifest: %v, SKIPPING", path + "/" + file.Name(), err)
            continue
          }

          // There could be other, non winget-manifest YAML files in the manifestPath as well. Skip them.
          // All valid manifest files must have all basemanifest fields set as they are required by the schema
          if basemanifest.PackageIdentifier != "" && basemanifest.PackageVersion != "" &&
            basemanifest.ManifestType != "" && basemanifest.ManifestVersion != "" {
            if basemanifest.ManifestType == "singleton" {
              var manifest = parseFileAsSingletonManifest(path + "/" + file.Name())
              logging.Logger.Debug().Msgf("Found singleton manifest for package %v", basemanifest.PackageIdentifier)
              models.Manifests.Set(manifest.GetPackageIdentifier(), basemanifest.PackageVersion, manifest.GetVersions()[0])
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
        logging.Logger.Debug().Msgf("found multi-file manifests for package %v", key.PackageIdentifier)
        var mergedManifest, err = parseMultiFileManifest(key, value...)
        if err != nil {
          logging.Logger.Error().Err(err).Msgf("Could not parse the manifest files for this package %v", key.PackageIdentifier)
        } else {
          for _, version := range mergedManifest.GetVersions() {
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
      logging.Logger.Error().Err(err).Msg("yamlFile.Get err")
      continue
    }
    switch file.ManifestType {
      case "version":
        var version models.Manifest_VersionManifestInterface
        if multifilemanifest.ManifestVersion == "1.1.0" {
          version = &models.Manifest_VersionManifest_1_1_0{}
        } else if multifilemanifest.ManifestVersion == "1.2.0" {
          version = &models.Manifest_VersionManifest_1_2_0{}
        } else if multifilemanifest.ManifestVersion == "1.4.0" {
          version = &models.Manifest_VersionManifest_1_4_0{}
        } else {
          logging.Logger.Error().Msgf("Unsupported VersionManifest version %v %v", multifilemanifest.ManifestVersion, file)
          continue
        }
        err = yaml.Unmarshal(yamlFile, version)
        if err != nil {
          logging.Logger.Error().Err(err).Msg("error unmarshalling version-manifest")
        }
        versions = append(versions, version)
      case "installer":
        var installer models.Manifest_InstallerManifestInterface
        if multifilemanifest.ManifestVersion == "1.1.0" {
          installer = &models.Manifest_InstallerManifest_1_1_0{}
        } else if multifilemanifest.ManifestVersion == "1.2.0" {
          installer = &models.Manifest_InstallerManifest_1_2_0{}
        } else if multifilemanifest.ManifestVersion == "1.4.0" {
          installer = &models.Manifest_InstallerManifest_1_4_0{}
        } else {
          logging.Logger.Error().Msgf("Unsupported InstallerManifest version %v %v", multifilemanifest.ManifestVersion, file)
          continue
        }
        err = yaml.Unmarshal(yamlFile, installer)
        if err != nil {
          logging.Logger.Error().Err(err).Msg("error unmarshalling installer-manifest")
        }
        installers = append(installers, installer)
      case "locale":
        var locale models.Manifest_LocaleManifestInterface
        if multifilemanifest.ManifestVersion == "1.1.0" {
          locale = &models.Manifest_LocaleManifest_1_1_0{}
        } else if multifilemanifest.ManifestVersion == "1.2.0" {
          locale = &models.Manifest_LocaleManifest_1_2_0{}
        } else if multifilemanifest.ManifestVersion == "1.4.0" {
          locale = &models.Manifest_LocaleManifest_1_4_0{}
        } else {
          logging.Logger.Error().Msgf("Unsupported LocaleManifest version", multifilemanifest.ManifestVersion, file)
          continue
        }
        err = yaml.Unmarshal(yamlFile, locale)
        if err != nil {
          logging.Logger.Error().Err(err).Msg("error unmarshalling locale-manifest")
        }
        locales = append(locales, locale)
      case "defaultLocale":
        if multifilemanifest.ManifestVersion == "1.1.0" {
          defaultlocale = &models.Manifest_DefaultLocaleManifest_1_1_0{}
        } else if multifilemanifest.ManifestVersion == "1.2.0" {
          defaultlocale = &models.Manifest_DefaultLocaleManifest_1_2_0{}
        } else if multifilemanifest.ManifestVersion == "1.4.0" {
          defaultlocale = &models.Manifest_DefaultLocaleManifest_1_4_0{}
        } else {
          logging.Logger.Error().Msgf("Unsupported DefaultLocaleManifest version", multifilemanifest.ManifestVersion, file)
          continue
        }
        err = yaml.Unmarshal(yamlFile, defaultlocale)
        if err != nil {
          logging.Logger.Error().Err(err).Msg("error unmarshalling defaultlocale-manifest")
        }
      default:
    }
  }

  // It's possible there were no installer or locale manifests or parsing them failed
  if len(installers) == 0 {
    return nil, errors.New(multifilemanifest.PackageVersion + " package manifests did not contain any (valid) installers")
  }
  if len(versions) == 0 {
    return nil, errors.New("package manifests did not contain any (valid) locales")
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
  apiInstallers = append(apiInstallers, installers[0].ToApiInstallers()...)

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
      apiInstallers = append(apiInstallers, intf.(models.API_Installer_1_1_0))
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
    var apiLocales []models.API_Locale_1_4_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_4_0))
    }

    var apiInstallers []models.API_Installer_1_4_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, intf.(models.API_Installer_1_4_0))
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

