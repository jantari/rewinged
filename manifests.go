package main

import (
  "fmt"
  "log"
  "os"
  "errors"
  "strings"
  "reflect"

  "gopkg.in/yaml.v3"
)

func ingestManifestsWorker() error {
  for path := range jobs {
    files, err := os.ReadDir(path)
    if err != nil {
      wg.Done()
      return err
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

          if basemanifest.ManifestType == "singleton" {
            var manifest = parseManifestFile(path + "/" + file.Name())
            fmt.Println("  Found singleton manifest for package", manifest.PackageIdentifier)
            manifests2.AppendValues(manifest.PackageIdentifier, manifest.Versions)
          } else {
            nonSingletonsMap[basemanifest.PackageIdentifier + "/" + basemanifest.PackageVersion] =
              append(nonSingletonsMap[basemanifest.PackageIdentifier + "/" + basemanifest.PackageVersion], path + "/" + file.Name())
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
          manifests2.AppendValues(merged_manifest.PackageIdentifier, merged_manifest.Versions)
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

  go func() {
    wg.Add(1)
    jobs <- path
  }()

  for _, file := range files {
    if file.IsDir() {
      fmt.Printf("Searching directory: %s\n", path + "/" + file.Name())

      getManifests(path + "/" + file.Name())
    }
  }
}

func parseMultiFileManifest (filenames ...string) (*Manifest, error) {
  if len(filenames) <= 0 {
    return nil, errors.New("you must provide at least one filename for reading Values")
  }

  var packageidentifier string
  versions      := []VersionManifest{}
  installers    := []InstallerManifest{}
  locales       := []LocaleManifest{}
  defaultlocale := &DefaultLocaleManifest{}

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
        version := &VersionManifest{}
        err = yaml.Unmarshal(yamlFile, version)
        if err != nil {
          log.Printf("error unmarshalling version-manifest %v\n", err)
        }
        versions = append(versions, *version)
      case "installer":
        installer := &InstallerManifest{}
        err = yaml.Unmarshal(yamlFile, installer)
        if err != nil {
          log.Printf("error unmarshalling installer-manifest %v\n", err)
        }
        installers = append(installers, *installer)
      case "locale":
        locale := &LocaleManifest{}
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

  // This transforms the manifest data into the format the API will return.
  // This logic should probably be moved out of this function, so that it returns
  // the full unaltered data from the combined manifests - and restructuring to
  // API-format will happen somewhere else
  var apiLocales []Locale
  for _, locale := range locales {
    apiLocales = append(apiLocales, *localeManifestToAPILocale(locale))
  }

  versions_api := []Versions{
    {
      PackageVersion: versions[0].PackageVersion,
      DefaultLocale: *defaultLocaleManifestToAPIDefaultLocale(*defaultlocale),
      Channel: "",
      Locales: apiLocales,
      Installers: installerManifestToAPIInstallers(installers[0]),
    },
  }

  manifest := &Manifest {
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
func installerManifestToAPIInstallers (instm InstallerManifest) []Installer {
  var apiInstallers []Installer

  for _, installer := range instm.Installers {
    apiInstallers = append(apiInstallers, Installer {
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

func defaultLocaleManifestToAPIDefaultLocale (locm DefaultLocaleManifest) *DefaultLocale {
  return &DefaultLocale{
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

func localeManifestToAPILocale (locm LocaleManifest) *Locale {
  return &Locale{
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

func parseFileAsBaseManifest (path string) (*BaseManifest, error) {
  manifest := &BaseManifest{}
  yamlFile, err := os.ReadFile(path)
  if err != nil {
    return manifest, err
  }

  err = yaml.Unmarshal(yamlFile, manifest)
  return manifest, err
}

func parseManifestFile (path string) *Manifest {
  yamlFile, err := os.ReadFile(path)
  if err != nil {
    log.Printf("error opening yaml file %v\n", err)
  }

  singleton := &SingletonManifest{}
  err = yaml.Unmarshal(yamlFile, singleton)
  if err != nil {
    log.Printf("error unmarshalling singleton %v\n", err)
  }

  manifest := singletonToStandardManifest(singleton)

  return manifest
}

func singletonToStandardManifest (singleton *SingletonManifest) *Manifest {
  manifest := &Manifest {
    PackageIdentifier: singleton.PackageIdentifier,
    Versions: []Versions {
      {
        PackageVersion: singleton.PackageVersion,
        DefaultLocale: DefaultLocale {
          PackageLocale: singleton.PackageLocale,
          PackageName: singleton.PackageName,
          Publisher: singleton.Publisher,
          ShortDescription: singleton.ShortDescription,
          License: singleton.License,
        },
        Channel: "",
        Locales: []Locale{},
        Installers: singleton.Installers[:],
      },
    },
  }

  return manifest
}

func caseInsensitiveContains(s, substr string) bool {
  s, substr = strings.ToUpper(s), strings.ToUpper(substr)
  return strings.Contains(s, substr)
}

func caseInsensitiveHasSuffix(s, substr string) bool {
  s, substr = strings.ToUpper(s), strings.ToUpper(substr)
  return strings.HasSuffix(s, substr)
}

// Modified from: https://stackoverflow.com/a/38407429
func findField(v interface{}, name string) reflect.Value {
  // create queue of values to search. Start with the function arg.
  queue := []reflect.Value{reflect.ValueOf(v)}
  for len(queue) > 0 {
    v := queue[0]
    queue = queue[1:]
    // dereference pointers
    for v.Kind() == reflect.Ptr {
        v = v.Elem()
    }
    // check all elements in slices
    if v.Kind() == reflect.Slice {
      if v.Len() > 0 {
        for i := 0; i < v.Len(); i++ {
          queue = append(queue, v.Index(i))
        }
      } else {
        //fmt.Println("CONTINUE (EMPTY SLICE)")
        continue
      }
    }
    // ignore if this is not a struct
    if v.Kind() != reflect.Struct {
        //fmt.Println("CONTINUE (NOT STRUCT)")
        continue
    }
    // iterate through fields looking for match on name
    t := v.Type()
    for i := 0; i < v.NumField(); i++ {
        //fmt.Println("TESTING FIELD", t.Field(i).Name)
        if t.Field(i).Name == name {
            // found it!
            //fmt.Println("FOUND THE FIELD", name, "!")
            return v.Field(i)
        }
        // push field to queue
        queue = append(queue, v.Field(i))
    }
  }
  return reflect.Value{}
}


func getPackagesByMatchFilter (manifests map[string][]Versions, inclusions []SearchRequestPackageMatchFilter, filters []SearchRequestPackageMatchFilter) map[string][]Versions {
  var manifestResultsMap = make(map[string][]Versions)
  normalizeReplacer := strings.NewReplacer(" ", "", "-", "", "+", "")

  for packageIdentifier, packageVersions := range manifests {
    // Loop through every version of the package as well because MatchFields like
    // ProductCode, PackageName etc. can change between versions.
    NEXT_VERSION:
    for _, packageVersion := range packageVersions {
      // From what I can gather from https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/CosmosDataStore.cs#L452
      // the difference between inclusions and filters are that inclusions are evaluated with a logical OR (only one of them has to match) and filters are evaluated with a logical AND
      // (all filter specified have to match) - so this is what I implemented here. But I am not 100% sure this is the correct/intended use for inclusions vs. filters.

      // process filters (if any)
      for _, filter := range filters {
        var requestMatchValue string

        switch filter.PackageMatchField {
          case NormalizedPackageNameAndPublisher:
            // winget only ever sends the package / software name, the publisher isn't included so to
            // enable proper matching we also only compare against the normalized packagename.
            requestMatchValue = normalizeReplacer.Replace(strings.ToLower(packageVersion.DefaultLocale.PackageName))
          case PackageIdentifier:
            // We don't need to recursively search for this field, it's easy to get to
            requestMatchValue = packageIdentifier
          case PackageName:
            fallthrough
          case Moniker:
            fallthrough
          case Command:
            fallthrough
          case Tag:
            fallthrough
          case PackageFamilyName:
            fallthrough
          case ProductCode:
            fallthrough
          case Market:
            fallthrough
          default:
            // Just search the whole struct for a field with the right name
            // Get the value of a nested struct field passing in the field name to search for as a string
            // Source: https://stackoverflow.com/a/38407429
            f := findField(packageVersion, string(filter.PackageMatchField))
            requestMatchValue = string(f.String())
        }

        // Because all filters (if any) must match (logical AND)
        // we just skip to the next packageversion if any did not match
        switch filter.RequestMatch.MatchType {
          // TODO: `winget list -s rewinged-local -q lapce` searches for the ProductCode with MatchType Exact
          // Why does it use MatchType Exact?? Does the reference / official source normalize all ProductCodes on ingest??
          case Exact:
            if requestMatchValue != filter.RequestMatch.KeyWord {
              continue NEXT_VERSION
            }
          case CaseInsensitive:
            if !strings.EqualFold(requestMatchValue, filter.RequestMatch.KeyWord) {
              continue NEXT_VERSION
            }
          case StartsWith:
            // StartsWith is implemented as case-sensitive, because it is that way in the reference implementation as well:
            // https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/PredicateGenerator.cs#L92-L102
            if !strings.HasPrefix(requestMatchValue, filter.RequestMatch.KeyWord) {
              continue NEXT_VERSION
            }
          case Substring:
            // Substring comparison is case-insensitive, because it is that way in the reference implementation as well:
            // https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/PredicateGenerator.cs#L92-L102
            if !caseInsensitiveContains(requestMatchValue, filter.RequestMatch.KeyWord) {
              continue NEXT_VERSION
            }
          default:
            // Unimplemented: Wildcard, Fuzzy, FuzzySubstring
        }
      }

      if len(inclusions) == 0 {
        manifestResultsMap[packageIdentifier] = append(manifestResultsMap[packageIdentifier], packageVersion)
        continue NEXT_VERSION
      }

      var anyInclusionMatched bool
      anyInclusionMatched = true
      // process inclusions (if any)
      NEXT_INCLUSION:
      for _, inclusion := range inclusions {
        var requestMatchValue string
        anyInclusionMatched = false

        switch inclusion.PackageMatchField {
          case NormalizedPackageNameAndPublisher:
            // winget only ever sends the package / software name, the publisher isn't included so to
            // enable proper matching we also only compare against the normalized packagename.
            requestMatchValue = normalizeReplacer.Replace(strings.ToLower(packageVersion.DefaultLocale.PackageName))
          case PackageIdentifier:
            // We don't need to recursively search for this field, it's easy to get to
            requestMatchValue = packageIdentifier
          case PackageName:
            fallthrough
          case Moniker:
            fallthrough
          case Command:
            fallthrough
          case Tag:
            fallthrough
          case PackageFamilyName:
            fallthrough
          case ProductCode:
            fallthrough
          case Market:
            fallthrough
          default:
            // Just search the whole struct for a field with the right name
            // Get the value of a nested struct field passing in the field name to search for as a string
            // Source: https://stackoverflow.com/a/38407429
            f := findField(packageVersion, string(inclusion.PackageMatchField))
            requestMatchValue = string(f.String())
        }

        switch inclusion.RequestMatch.MatchType {
          // TODO: `winget list -s rewinged-local -q lapce` searches for the ProductCode with MatchType Exact
          // Why does it use MatchType Exact?? Does the reference / official source normalize all ProductCodes on ingest??
          case Exact:
            if requestMatchValue == inclusion.RequestMatch.KeyWord {
              // Break out of the inclusions loop after one successful match
              anyInclusionMatched = true
              break NEXT_INCLUSION
            }
          case CaseInsensitive:
            if strings.EqualFold(requestMatchValue, inclusion.RequestMatch.KeyWord) {
              // Break out of the inclusions loop after one successful match
              anyInclusionMatched = true
              break NEXT_INCLUSION
            }
          case StartsWith:
            // StartsWith is implemented as case-sensitive, because it is that way in the reference implementation as well:
            // https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/PredicateGenerator.cs#L92-L102
            if strings.HasPrefix(requestMatchValue, inclusion.RequestMatch.KeyWord) {
              // Break out of the inclusions loop after one successful match
              anyInclusionMatched = true
              break NEXT_INCLUSION
            }
          case Substring:
            // Substring comparison is case-insensitive, because it is that way in the reference implementation as well:
            // https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/PredicateGenerator.cs#L92-L102
            if caseInsensitiveContains(requestMatchValue, inclusion.RequestMatch.KeyWord) {
              // Break out of the inclusions loop after one successful match
              anyInclusionMatched = true
              break NEXT_INCLUSION
            }
          default:
            // Unimplemented: Wildcard, Fuzzy, FuzzySubstring
        }
      }

      // All filters and inclusions have passed for this manifest, add it to the returned map
      if anyInclusionMatched {
        fmt.Println("Adding to the results map:", packageIdentifier, "version", packageVersion.PackageVersion)
        manifestResultsMap[packageIdentifier] = append(manifestResultsMap[packageIdentifier], packageVersion)
      }
    }
  }

  return manifestResultsMap
}

func getPackagesByKeyword (manifests map[string][]Versions, keyword string) map[string][]Versions {
  var manifestResultsMap = make(map[string][]Versions)
  for packageIdentifier, packageVersions := range manifests {
    for _, version := range packageVersions {
      if caseInsensitiveContains(version.DefaultLocale.PackageName, keyword) || caseInsensitiveContains(version.DefaultLocale.ShortDescription, keyword) {
        manifestResultsMap[packageIdentifier] = append(manifestResultsMap[packageIdentifier], version)
      }
    }
  }

  return manifestResultsMap
}

