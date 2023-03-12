package main

import (
  "fmt"
  "log"
  "os"
  "errors"
  "strings"

  "gopkg.in/yaml.v3"

  "rewinged/models"
)

func ingestManifestsWorker() error {
  for path := range jobs {
    files, err := os.ReadDir(path)
    if err != nil {
      log.Println("ingestManifestsWorker error", err)
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
            log.Printf("error unmarshaling YAML file '%v' as BaseManifest: %v, SKIPPING\n", path + "/" + file.Name(), err)
            continue
          }

          // There could be other, non winget-manifest YAML files in the manifestPath as well. Skip them.
          // All valid manifest files must have all basemanifest fields set as they are required by the schema
          if basemanifest.PackageIdentifier != "" && basemanifest.PackageVersion != "" &&
            basemanifest.ManifestType != "" && basemanifest.ManifestVersion != "" {
            if basemanifest.ManifestType == "singleton" {
              var manifest = parseFileAsSingletonManifest(path + "/" + file.Name())
              fmt.Println("  Found singleton manifest for package", basemanifest.PackageIdentifier)
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
        fmt.Println("  Found multi-file manifests for package", key.PackageIdentifier)
        var merged_manifest, err = parseMultiFileManifest(key, value...)
        if err != nil {
          log.Println("Could not parse the manifest files for this package", key.PackageIdentifier, err)
        } else {
          for _, version := range merged_manifest.GetVersions() {
            // Replace the existing PkgId + PkgVersion entry with this one
            models.Manifests.Set(merged_manifest.GetPackageIdentifier(), version.GetPackageVersion(), version)
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
    log.Println(err)
  }

  // wg.Add() before goroutine, see staticcheck check SA2000 and also
  // https://stackoverflow.com/questions/65213707/where-to-put-wg-add
  wg.Add(1)
  go func() {
    jobs <- path
  }()

  for _, file := range files {
    if file.IsDir() {
      fmt.Printf("Searching directory: %s\n", path + "/" + file.Name())

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
      log.Printf("yamlFile.Get err   #%v ", err)
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
          log.Println("Unsupported VersionManifest version", multifilemanifest.ManifestVersion, file)
          continue
        }
        err = yaml.Unmarshal(yamlFile, version)
        if err != nil {
          log.Printf("error unmarshalling version-manifest %v\n", err)
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
          log.Println("Unsupported InstallerManifest version", multifilemanifest.ManifestVersion, file)
          continue
        }
        err = yaml.Unmarshal(yamlFile, installer)
        if err != nil {
          log.Printf("error unmarshalling installer-manifest %v\n", err)
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
          log.Println("Unsupported LocaleManifest version", multifilemanifest.ManifestVersion, file)
          continue
        }
        err = yaml.Unmarshal(yamlFile, locale)
        if err != nil {
          log.Printf("error unmarshalling locale-manifest %v\n", err)
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
          log.Println("Unsupported DefaultLocaleManifest version", multifilemanifest.ManifestVersion, file)
          continue
        }
        err = yaml.Unmarshal(yamlFile, defaultlocale)
        if err != nil {
          log.Printf("error unmarshalling defaultlocale-manifest %v\n", err)
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

  manifest, err := newApiManifest(
    multifilemanifest.ManifestVersion,
    multifilemanifest.PackageIdentifier,
    versions[0].GetPackageVersion(),
    defaultlocale.ToApiDefaultLocale(),
    apiLocales,
    apiInstallers,
  )

  return manifest, err
}

func newApiManifest (
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
  var api_ret models.API_ManifestInterface
  var api_mvi models.API_ManifestVersionInterface

  if ManifestVersion == "1.1.0" {
    var apiLocales []models.API_Locale_1_1_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_1_0))
    }

    var apiInstallers []models.API_Installer_1_1_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, intf.(models.API_Installer_1_1_0))
    }

    api_mvi = models.API_ManifestVersion_1_1_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_1_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    api_ret = &models.API_Manifest_1_1_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ api_mvi },
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

    api_mvi = models.API_ManifestVersion_1_4_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_4_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    api_ret = &models.API_Manifest_1_4_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ api_mvi },
    }
  } else {
    return nil, errors.New("Converting manifest v" + ManifestVersion + " data for API responses is not yet supported.")
  }

  return api_ret, nil
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
    log.Printf("error opening yaml file %v\n", err)
  }

  singleton := &models.Manifest_SingletonManifest_1_1_0{}
  err = yaml.Unmarshal(yamlFile, singleton)
  if err != nil {
    log.Printf("error unmarshalling singleton %v\n", err)
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

