package main

import (
  "fmt"
  "log"
  "os"
  "errors"
  "strings"
  "reflect"

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

    // temporary map collecting all files belonging to a particular
    // packageidentifier + PackageVersion combination for later processing
    var nonSingletonsMap = make(map[string][]string)

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
              var manifest = parseManifestFile(path + "/" + file.Name())
              fmt.Println("  Found singleton manifest for package", manifest.PackageIdentifier)
              models.Manifests.Set(manifest.PackageIdentifier, basemanifest.PackageVersion, manifest.Versions[0])
            } else {
              nonSingletonsMap[basemanifest.PackageIdentifier + "/" + basemanifest.PackageVersion] =
                append(nonSingletonsMap[basemanifest.PackageIdentifier + "/" + basemanifest.PackageVersion], path + "/" + file.Name())
            }
          }
        }
      }
    }

    if len(nonSingletonsMap) > 0 {
      for key, value := range nonSingletonsMap {
        fmt.Println("  Found multi-file manifests for package", key)
        var merged_manifest, err = parseMultiFileManifest(value...)
        if err != nil {
          log.Println("  Could not parse the manifest files for this package", err)
        } else {
          for _, version := range merged_manifest.Versions {
            // Replace the existing PkgId + PkgVersion entry with this one
            models.Manifests.Set(merged_manifest.PackageIdentifier, version.PackageVersion, version)
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

func parseMultiFileManifest (filenames ...string) (*models.Manifest, error) {
  if len(filenames) <= 0 {
    return nil, errors.New("you must provide at least one filename for reading Values")
  }

  var packageidentifier string
  versions      := []models.VersionManifest{}
  installers    := []models.InstallerManifest{}
  locales       := []models.LocaleManifest{}
  defaultlocale := &models.DefaultLocaleManifest{}

  for _, file := range filenames {
    var basemanifest, err = parseFileAsBaseManifest(file)
    if err != nil {
      log.Printf("error unmarshaling YAML file '%v' as BaseManifest: %v, skipping", file, err)
      continue
    }
    packageidentifier = basemanifest.PackageIdentifier

    yamlFile, err := os.ReadFile(file)
    if err != nil {
      log.Printf("yamlFile.Get err   #%v ", err)
    }
    switch basemanifest.ManifestType {
      case "version":
        version := &models.VersionManifest{}
        err = yaml.Unmarshal(yamlFile, version)
        if err != nil {
          log.Printf("error unmarshalling version-manifest %v\n", err)
        }
        versions = append(versions, *version)
      case "installer":
        installer := &models.InstallerManifest{}
        err = yaml.Unmarshal(yamlFile, installer)
        if err != nil {
          log.Printf("error unmarshalling installer-manifest %v\n", err)
        }
        installers = append(installers, *installer)
      case "locale":
        locale := &models.LocaleManifest{}
        err = yaml.Unmarshal(yamlFile, locale)
        if err != nil {
          log.Printf("error unmarshalling locale-manifest %v\n", err)
        }
        locales = append(locales, *locale)
      case "defaultLocale":
        err = yaml.Unmarshal(yamlFile, defaultlocale)
        if err != nil {
          log.Printf("error unmarshalling defaultlocale-manifest %v\n", err)
        }
      default:
    }
  }

  // It's possible there were no installer or locale manifests or parsing them failed
  if len(installers) == 0 {
    return nil, errors.New("package manifests did not contain any (valid) installers")
  }
  if len(versions) == 0 {
    return nil, errors.New("package manifests did not contain any (valid) locales")
  }

  // This transforms the manifest data into the format the API will return.
  // This logic should probably be moved out of this function, so that it returns
  // the full unaltered data from the combined manifests - and restructuring to
  // API-format will happen somewhere else
  var apiLocales []models.Locale
  for _, locale := range locales {
    apiLocales = append(apiLocales, *localeManifestToAPILocale(locale))
  }

  versions_api := []models.Versions{
    {
      PackageVersion: versions[0].PackageVersion,
      DefaultLocale: *defaultLocaleManifestToAPIDefaultLocale(*defaultlocale),
      Channel: "",
      Locales: apiLocales,
      Installers: installerManifestToAPIInstallers(installers[0]),
    },
  }

  manifest := &models.Manifest {
    PackageIdentifier: packageidentifier,
    Versions: versions_api[:],
  }

  return manifest, nil //err
}

// This function takes two values and returns
// the one that's not set to its default value.
func nonDefault[T any] (optionA T, optionB T) T {
  if isDefault(reflect.ValueOf(optionA)) {
    return optionB
  }
  return optionA
}

func isDefault(v reflect.Value) bool {
  return v.IsZero()
}

// The installers in a manifest can contain 'global' properties
// that apply to all individual installers listed. In the API responses
// these have to be merged and set on all individual installers.
func installerManifestToAPIInstallers (instm models.InstallerManifest) []models.Installer {
  var apiInstallers []models.Installer

  for _, installer := range instm.Installers {
    apiInstallers = append(apiInstallers, models.Installer {
      Architecture: installer.Architecture, // Already mandatory per-Installer
      MinimumOSVersion: nonDefault(installer.MinimumOSVersion, instm.MinimumOSVersion), // Already mandatory per-Installer
      Platform: nonDefault(installer.Platform, instm.Platform),
      InstallerType: nonDefault(installer.InstallerType, instm.InstallerType),
      Scope: nonDefault(installer.Scope, instm.Scope),
      InstallerUrl: installer.InstallerUrl, // Already mandatory per-Installer
      InstallerSha256: installer.InstallerSha256, // Already mandatory per-Installer
      SignatureSha256: installer.SignatureSha256, // Can only be set per-Installer, impossible to copy from global manifest properties
      InstallModes: nonDefault(installer.InstallModes, instm.InstallModes),
      InstallerSuccessCodes: nonDefault(installer.InstallerSuccessCodes, instm.InstallerSuccessCodes),
      ExpectedReturnCodes: nonDefault(installer.ExpectedReturnCodes, instm.ExpectedReturnCodes),
      ProductCode: nonDefault(installer.ProductCode, instm.ProductCode),
      ReleaseDate: nonDefault(installer.ReleaseDate, instm.ReleaseDate),
    })
  }

  return apiInstallers
}

func defaultLocaleManifestToAPIDefaultLocale (locm models.DefaultLocaleManifest) *models.DefaultLocale {
  return &models.DefaultLocale{
    PackageLocale: locm.PackageLocale,
    Publisher: locm.Publisher,
    PublisherUrl: locm.PublisherUrl,
    PublisherSupportUrl: locm.PublisherSupportUrl,
    PrivacyUrl: locm.PrivacyUrl,
    Author: locm.Author,
    PackageName: locm.PackageName,
    PackageUrl: locm.PackageUrl,
    License: locm.License,
    LicenseUrl: locm.LicenseUrl,
    Copyright: locm.Copyright,
    CopyrightUrl: locm.CopyrightUrl,
    ShortDescription: locm.ShortDescription,
    Description: locm.Description,
    Moniker: locm.Moniker,
    Tags: locm.Tags,
    Agreements: locm.Agreements,
    ReleaseNotes: locm.ReleaseNotes,
    ReleaseNotesUrl: locm.ReleaseNotesUrl,
  }
}

func localeManifestToAPILocale (locm models.LocaleManifest) *models.Locale {
  return &models.Locale{
    PackageLocale: locm.PackageLocale,
    Publisher: locm.Publisher,
    PublisherUrl: locm.PublisherUrl,
    PublisherSupportUrl: locm.PublisherSupportUrl,
    PrivacyUrl: locm.PrivacyUrl,
    Author: locm.Author,
    PackageName: locm.PackageName,
    PackageUrl: locm.PackageUrl,
    License: locm.License,
    LicenseUrl: locm.LicenseUrl,
    Copyright: locm.Copyright,
    CopyrightUrl: locm.CopyrightUrl,
    ShortDescription: locm.ShortDescription,
    Description: locm.Description,
    Tags: locm.Tags,
    Agreements: locm.Agreements,
    ReleaseNotes: locm.ReleaseNotes,
    ReleaseNotesUrl: locm.ReleaseNotesUrl,
  }
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

func parseManifestFile (path string) *models.Manifest {
  yamlFile, err := os.ReadFile(path)
  if err != nil {
    log.Printf("error opening yaml file %v\n", err)
  }

  singleton := &models.SingletonManifest{}
  err = yaml.Unmarshal(yamlFile, singleton)
  if err != nil {
    log.Printf("error unmarshalling singleton %v\n", err)
  }

  manifest := singletonToStandardManifest(singleton)

  return manifest
}

func singletonToStandardManifest (singleton *models.SingletonManifest) *models.Manifest {
  manifest := &models.Manifest {
    PackageIdentifier: singleton.PackageIdentifier,
    Versions: []models.Versions {
      {
        PackageVersion: singleton.PackageVersion,
        DefaultLocale: models.DefaultLocale {
          PackageLocale: singleton.PackageLocale,
          PackageName: singleton.PackageName,
          Publisher: singleton.Publisher,
          ShortDescription: singleton.ShortDescription,
          License: singleton.License,
        },
        Channel: "",
        Locales: []models.Locale{},
        Installers: singleton.Installers[:],
      },
    },
  }

  return manifest
}

func caseInsensitiveHasSuffix(s, substr string) bool {
  s, substr = strings.ToUpper(s), strings.ToUpper(substr)
  return strings.HasSuffix(s, substr)
}

