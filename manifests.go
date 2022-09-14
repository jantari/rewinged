package main

import (
  "fmt"
  "errors"
  "io/ioutil"
  "strings"
  "os"
  "reflect"

  "gopkg.in/yaml.v3"
  //"github.com/mitchellh/mapstructure"
)

func GetManifests (path string) map[string][]Versions {
  // global map accumulating all manifests parsed
  var manifests = make(map[string][]Versions)
  // temporary map collecting all files belonging to a particular
  // packageidentifier + PackageVersion combination for later processing
  var nonSingletonsMap = make(map[string][]string)

  files, err := os.ReadDir(path)
  if err != nil {
    fmt.Println(err)
  }

  for _, file := range files {
    fmt.Printf("Path: %s IsDir: %v\n", path + "/" + file.Name(), file.IsDir())
    if file.IsDir() {
      var manifests_from_dir = GetManifests(path + "/" + file.Name())
      for k, v := range manifests_from_dir {
        manifests[k] = append(manifests[k], v...)
      }
    } else {
      if strings.HasSuffix(file.Name(), ".yml") || strings.HasSuffix(file.Name(), ".yaml") {
        var basemanifest, err = ParseFileAsBaseManifest(path + "/" + file.Name())
        if err != nil {
          fmt.Printf("Error unmarshaling YAML file '%v' as BaseManifest: %v, SKIPPING\n", path + "/" + file.Name(), err)
          continue
        }
        fmt.Printf("  BaseManifest: %+v\n", basemanifest)

        if basemanifest.ManifestType == "singleton" {
          var manifest = ParseManifestFile(path + "/" + file.Name())
          manifests[manifest.PackageIdentifier] = append(manifests[manifest.PackageIdentifier], manifest.Versions...)
        } else {
          nonSingletonsMap[basemanifest.PackageIdentifier + "/" + basemanifest.PackageVersion] =
            append(nonSingletonsMap[basemanifest.PackageIdentifier + "/" + basemanifest.PackageVersion], path + "/" + file.Name())
        }
      }
    }
  }

  if len(nonSingletonsMap) > 0 {
    fmt.Println("  There are multi-file manifests in this directory")
    fmt.Printf("%+v\n", nonSingletonsMap)
    for key, value := range nonSingletonsMap {
      fmt.Println("    Merging manifests for package", key)
      var merged_manifest, err = ParseMultiFileManifest(value...)
      if err != nil {
        fmt.Println("    Could not parse the manifest files for this package", err)
      } else {
        manifests[merged_manifest.PackageIdentifier] = append(manifests[merged_manifest.PackageIdentifier], merged_manifest.Versions...)
      }
    }
  }

  return manifests
}

//
func ParseMultiFileManifest (filenames ...string) (*Manifest, error) {
  if len(filenames) <= 0 {
    return nil, errors.New("You must provide at least one filename for reading Values")
  }

  var packageidentifier string
  versions      := []VersionManifest{}
  installers    := []InstallerManifest{}
  locales       := []LocaleManifest{}
  defaultlocale := &DefaultLocaleManifest{}

  for _, file := range filenames {
    var basemanifest, err = ParseFileAsBaseManifest(file)
    if err != nil {
      fmt.Printf("Error unmarshaling YAML file '%v' as BaseManifest: %v, skipping", file, err)
      continue
    }
    packageidentifier = basemanifest.PackageIdentifier

    yamlFile, err := ioutil.ReadFile(file)
    if err != nil {
      fmt.Printf("yamlFile.Get err   #%v ", err)
    }
    switch basemanifest.ManifestType {
      case "version":
        fmt.Println("Parsing version manifest ...")
        version := &VersionManifest{}
        err = yaml.Unmarshal(yamlFile, version)
        if err != nil {
          fmt.Printf("unmarshal version err   #%v ", err)
        } else {
          fmt.Println("  Successfully unmarshalled version")
        }
        versions = append(versions, *version)
      case "installer":
        fmt.Println("Parsing installer manifest ...")
        installer := &InstallerManifest{}
        err = yaml.Unmarshal(yamlFile, installer)
        if err != nil {
          fmt.Printf("unmarshal installer err   #%v ", err)
        } else {
          fmt.Println("  Successfully unmarshalled installer")
        }
        installers = append(installers, *installer)
      case "locale":
        fmt.Println("Parsing locale manifest ...")
        locale := &LocaleManifest{}
        err = yaml.Unmarshal(yamlFile, locale)
        if err != nil {
          fmt.Printf("unmarshal locale err   #%v ", err)
        } else {
          fmt.Println("  Successfully unmarshalled locale")
        }
        locales = append(locales, *locale)
      case "defaultLocale":
        fmt.Println("Parsing defaultlocale manifest ...")
        err = yaml.Unmarshal(yamlFile, defaultlocale)
        if err != nil {
          fmt.Printf("unmarshal defaultlocale err   #%v ", err)
        } else {
          fmt.Println("  Successfully unmarshalled defaultlocale")
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
    fmt.Println("installer didnt't have field set, using value from global manifest property", optionB)
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

func GetLocaleByName (locales []Locale, localename string) *Locale {
  fmt.Println("GetLocaleByName: looking for the locale", localename, "in", len(locales), "total locales")
  for _, locale := range locales {
    if locale.PackageLocale == localename {
      return &locale
    }
  }

  return nil
}

func ParseFileAsBaseManifest (path string) (*BaseManifest, error) {
  manifest := &BaseManifest{}
  yamlFile, err := ioutil.ReadFile(path)
  if err != nil {
    return manifest, err
  }

  err = yaml.Unmarshal(yamlFile, manifest)
  return manifest, err
}

func ParseManifestFile (path string) *Manifest {
  yamlFile, err := ioutil.ReadFile(path)
  if err != nil {
    fmt.Printf("yamlFile.Get err   #%v ", err)
  }

  singleton := &SingletonManifest{}
  err = yaml.Unmarshal(yamlFile, singleton)
  if err != nil {
    fmt.Printf("Unmarshal singleton error: %v", err)
  }

  manifest := SingletonToStandardManifest(singleton)

  return manifest
}

func SingletonToStandardManifest (singleton *SingletonManifest) *Manifest {
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

func CaseInsensitiveContains(s, substr string) bool {
  s, substr = strings.ToUpper(s), strings.ToUpper(substr)
  return strings.Contains(s, substr)
}

func CaseInsensitiveHasSuffix(s, substr string) bool {
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


func GetPackagesByMatchFilter (manifests map[string][]Versions, inclusions []SearchRequestPackageMatchFilter, filters []SearchRequestPackageMatchFilter) map[string][]Versions {
  var manifestResultsMap = make(map[string][]Versions)

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
            requestMatchValue = strings.ReplaceAll(
              strings.ReplaceAll(
                strings.ToLower(packageVersions[0].DefaultLocale.PackageName),
              " ", ""),
            "-", "")
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
            if !CaseInsensitiveContains(requestMatchValue, filter.RequestMatch.KeyWord) {
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
            requestMatchValue = strings.ReplaceAll(
              strings.ReplaceAll(
                strings.ToLower(packageVersions[0].DefaultLocale.PackageName),
              " ", ""),
            "-", "")
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
            if CaseInsensitiveContains(requestMatchValue, inclusion.RequestMatch.KeyWord) {
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

func GetPackagesByKeyword (manifests map[string][]Versions, keyword string) map[string][]Versions {
  var manifestResultsMap = make(map[string][]Versions)
  for packageIdentifier, packageVersions := range manifests {
    for _, version := range packageVersions {
      if CaseInsensitiveContains(version.DefaultLocale.PackageName, keyword) || CaseInsensitiveContains(version.DefaultLocale.ShortDescription, keyword) {
        manifestResultsMap[packageIdentifier] = append(manifestResultsMap[packageIdentifier], version)
      }
    }
  }

  return manifestResultsMap
}

func GetPackageByIdentifier (manifests []Manifest, packageidentifier string) *Manifest {
  for _, manifest := range manifests {
    if manifest.PackageIdentifier == packageidentifier {
      fmt.Println("searched ID", packageidentifier, "equals a package:", manifest.PackageIdentifier)
      return &manifest
    }
  }

  return nil
}

